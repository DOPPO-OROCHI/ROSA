package game

import (
	"TheWar/internal/domain/cards"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type PassiveSkillHandler func(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error

type PassiveTriggerContext struct {
	ActorInstanceID  string
	TargetInstanceID string
	DeadInstanceID   string
	SourceOwnerIdx   int
	TargetOwnerIdx   int
	TargetSlot       int
}

var PassiveSkillsHandler map[string]PassiveSkillHandler

func init() {
	PassiveSkillsHandler = map[string]PassiveSkillHandler{
		"passive_damage_up": passiveDamageUpHandler,
	}
}

// ПАССИВКА ПОД АП УРОНА
func passiveDamageUpHandler(m *MatchState,
	ownerIdx int, source *UnitState,
	event string, ctx PassiveTriggerContext) error {
	if m == nil || source == nil || ownerIdx < 0 || ownerIdx > 1 {
		return nil
	}
	if !shouldTriggerPassive(source, event) {
		return nil
	}
	if source.PassiveTrigger == cards.PassiveTriggerHitMe && ctx.TargetInstanceID != source.InstanceID {
		return nil
	}
	matched := passiveMatchedCount(m, ownerIdx, source)
	if !passiveConditionOK(m, ownerIdx, source, matched) {
		return nil
	}
	value := calcPassiveFinalValue(source, matched)
	if value == 0 {
		return nil
	}
	targets := passiveTargets(m, ownerIdx, source, ctx)
	for _, t := range targets {
		if t == nil {
			continue
		}
		t.Attack += value
	}
	return nil
}

/*ДАЛЕЕ ХЕЛПЕРЫ*/
// СЧИТАЕМ СКОЛЬКО ЮНИТОВ ПОДХОДИТ ПОД ТРИГГЕР ПАССИВОК. ПО СТОРОНЕ И ТИПУ\КОДУ КАРТЫ
func passiveMatchedCount(m *MatchState, ownerIdx int, source *UnitState) int {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return 0
	}
	ally := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	countInTable := func(t [TableSize]*UnitState) int {
		c := 0
		for i := 0; i < TableSize; i++ {
			u := t[i]
			if u == nil {
				continue
			}
			if source.PassiveCountType != "" {
				if u.CardType == source.PassiveCountType {
					c++
				}
				continue
			}
			if source.PassiveCountCode != "" && u.TemplateID == source.PassiveCountCode {
				c++
			}
		}
		return c
	}
	switch source.PassiveCountOwner {
	case cards.PassiveCountOwnerEnemy:
		if enemy == nil {
			return 0
		}
		return countInTable(enemy.Table)
	case cards.PassiveCountOwnerBoth:
		total := 0
		if ally != nil {
			total += countInTable(ally.Table)
		}
		if enemy != nil {
			total += countInTable(enemy.Table)
		}
		return total
	default:
		if ally == nil {
			return 0
		}
		return countInTable(ally.Table)
	}
}

// ОПРЕДЕЛЯЕМ АКТИВНА ЛИ ПАССИВНАЯ СПОСОБНОСТЬ В МОМЕНТ ВРЕМЕНИ
func passiveConditionOK(m *MatchState, ownerIdx int, source *UnitState, matched int) bool {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return false
	}
	switch source.PassiveCondition {
	case "", cards.PassiveConditionAlways:
		return true
	case cards.PassiveConditionCountAtLeast:
		return matched >= source.PassiveConditionCount
	case cards.PassiveConditionCountAtMost:
		return matched <= source.PassiveConditionCount
	case cards.PassiveConditionExact:
		return matched == source.PassiveConditionCount
	case cards.PassiveConditionDemonicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.DemonicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionMechanicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.MechanicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionOrganicalOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.OrganicalCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	case cards.PassiveConditionHealerOnTable:
		tmp := *source
		tmp.PassiveCountType = cards.HealerCard
		tmp.PassiveCountCode = ""
		return passiveMatchedCount(m, ownerIdx, &tmp) > 0
	default:
		return false
	}
}

// ОПРЕДЕЛЯЕМ К КАКИМ ЮНИТАМ ПРИМЕНИТЬ ЭФФЕКТ (ally_all, random_enemy...)
func passiveTargets(m *MatchState, ownerIdx int, source *UnitState, ctx PassiveTriggerContext) []*UnitState {
	if m == nil || source == nil || ownerIdx > 1 || ownerIdx < 0 {
		return nil
	}
	ally := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	if ally == nil || enemy == nil {
		return nil
	}
	out := make([]*UnitState, 0, 6)
	add := func(u *UnitState) {
		if u != nil {
			out = append(out, u)
		}
	}
	addAll := func(t [TableSize]*UnitState) {
		for i := 0; i < TableSize; i++ {
			add(t[i])
		}
	}
	addByType := func(t [TableSize]*UnitState, cardType string) {
		for i := 0; i < TableSize; i++ {
			u := t[i]
			if u != nil && u.CardType == cardType {
				out = append(out, u)
			}
		}
	}
	switch source.PassiveTarget {
	case cards.PassiveTargetSelf:
		add(source)
	case cards.PassiveTargetAllyAll:
		addAll(ally.Table)
	case cards.PassiveTargetEnemyAll:
		addAll(enemy.Table)
	case cards.PassiveTargetBothAll:
		addAll(ally.Table)
		addAll(enemy.Table)
	case cards.PassiveTargetAllyLeftRight:
		slot, _ := ally.FindSlot(source.InstanceID)
		if slot-1 >= 0 {
			add(ally.Table[slot-1])
		}
		if slot+1 < TableSize {
			add(ally.Table[slot+1])
		}
	case cards.PassiveTargetRandomEnemy:
		var picked *UnitState
		seen := 0
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			seen++
			if rng.Intn(seen) == 0 {
				picked = u
			}
		}
		add(picked)
	case cards.PassiveTargetRandomAlly:
		var picked *UnitState
		seen := 0
		for i := 0; i < TableSize; i++ {
			u := ally.Table[i]
			if u == nil {
				continue
			}
			seen++
			if rng.Intn(seen) == 0 {
				picked = u
			}
		}
		add(picked)
	case cards.PassiveTargetAllyTypeDemonical:
		addByType(ally.Table, cards.DemonicalCard)
	case cards.PassiveTargetAllyTypeMechanical:
		addByType(ally.Table, cards.MechanicalCard)
	case cards.PassiveTargetAllyTypeOrganical:
		addByType(ally.Table, cards.OrganicalCard)
	case cards.PassiveTargetAllyTypeHealer:
		addByType(ally.Table, cards.HealerCard)
	case cards.PassiveTargetEnemyTypeDemonical:
		addByType(enemy.Table, cards.DemonicalCard)
	case cards.PassiveTargetEnemyTypeMechanical:
		addByType(enemy.Table, cards.MechanicalCard)
	case cards.PassiveTargetEnemyTypeOrganical:
		addByType(enemy.Table, cards.OrganicalCard)
	case cards.PassiveTargetEnemyTypeHealer:
		addByType(enemy.Table, cards.HealerCard)
	default:
		if ctx.TargetOwnerIdx == ownerIdx && ctx.TargetSlot >= 0 && ctx.TargetSlot < TableSize {
			add(ally.Table[ctx.TargetSlot])
		} else if ctx.TargetOwnerIdx == 1-ownerIdx && ctx.TargetSlot >= 0 && ctx.TargetSlot < TableSize {
			add(enemy.Table[ctx.TargetSlot])
		}
	}
	return out
}

// СЧИТАЕМ ФИНАЛЬНУЮ СИЛУ ТОГО ИЛИ ИНОГО ПАССИВНОГО ЭФФЕКТА
func calcPassiveFinalValue(source *UnitState, matchedCount int) int {
	if source == nil {
		return 0
	}
	switch source.PassiveScale {
	case cards.PassiveScalePerCount:
		return source.PassiveValue * matchedCount
	case cards.PassiveScaleFlat:
		fallthrough
	default:
		return source.PassiveValue
	}
}

// Проверяет должна ли пассивка сработать в момент времени или нет
func shouldTriggerPassive(u *UnitState, event string) bool {
	return u != nil && u.PassiveCode != "" && u.PassiveTrigger == event
}

func getPassiveSkillHandler(code string) (PassiveSkillHandler, error) {
	h, ok := PassiveSkillsHandler[code]
	if !ok {
		return nil, ErrCardPassiveSkillUnsupported
	}
	return h, nil
}
