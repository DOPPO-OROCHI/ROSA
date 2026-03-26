package game

import (
	"TheWar/internal/domain/cards"
	"errors"
	"fmt"
	"time"
)

//Файл целиком и полностью посвящен скилам карт. Логика реализации -общая под каждый тип скила

type CardSkillHandler func(m *MatchState,
	a Action, caster *UnitState, owner *PlayerState,
	enemy *PlayerState) error

var CardSkillHandlers map[string]CardSkillHandler

func init() {
	CardSkillHandlers = map[string]CardSkillHandler{
		cards.SkillDamageSplash:             castDamageSplash,
		cards.SkillDamageSingle:             castDamageSingle,
		cards.SkillHealSingle:               castHealSingle,
		cards.SkillApplyBuff:                castBuffSelf,
		cards.SkillApplyDebuff:              castApplyDebuff,
		cards.SkillSummonSelfCopy:           castSummonSelfCopy,
		cards.SkillBanishUnit:               castBanishUnit,
		cards.SkillRevealEnemyHand:          castRevealEnemyHand,
		cards.SkillIncEnemyCdAllOnDeath:     castIncEnemyCdAfterDeath,
		cards.SkillIncEnemyCdSingle:         castIncEnemyCdSingle,
		cards.SkillDecAllyCdSingle:          castDecAllyCdSingle,
		cards.SkillDeathAoe:                 castDeathAoe,
		cards.SkillResurrectTargetFromGrave: castResurrectTargetFromGrave,
	}
}

// ФУНКЦИЯ ПОД СОЛО ДАМАГ
func castDamageSingle(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || owner == nil || enemy == nil {
		return errors.New("nil state")
	}
	if !a.AttackHero && a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	hasTank := false
	for i := 0; i < TableSize; i++ {
		u := enemy.Table[i]
		if u != nil && u.IsTank {
			hasTank = true
			break
		}
	}
	ignoreTank := caster.SkillParamsJSON == cards.IgnoreTankTrue
	dmg := caster.SkillValue
	if a.AttackHero {
		if !ignoreTank && hasTank {
			return ErrCardSkillTargetTankBlocked
		}
		enemy.HeroHP -= dmg
		heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
		m.Events = append(m.Events, Event{
			Type:             string(EventCardSkill),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
			Targets: []EventTarget{{
				InstanceID: heroID,
				Amount:     dmg,
				Died:       enemy.HeroHP <= 0,
				NewHP:      enemy.HeroHP,
			}},
		})
		if enemy.HeroHP <= 0 {
			m.Finished = true
			if a.PlayerIndex == 0 {
				m.Result = MatchWinP1
			} else {
				m.Result = MatchWinP2
			}
		}
		return nil
	}
	slot, target := enemy.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
		return ErrCardSkillBadTarget
	}
	if !ignoreTank && hasTank && !target.IsTank {
		return ErrCardSkillTargetTankBlocked
	}
	target.HP -= dmg
	died := target.HP <= 0
	newHP := target.HP
	if died {
		if err := killUnitAt(m, 1-a.PlayerIndex, slot, caster.InstanceID, a.PlayerIndex); err != nil {
			return err
		}
		newHP = 0
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets: []EventTarget{{
			InstanceID: a.TargetInstanceID,
			TemplateID: target.TemplateID,
			Amount:     dmg,
			Died:       died,
			NewHP:      newHP,
		},
		},
	})
	return nil
}

// ФУНКЦИЯ ЛЕЧЕНИЯ СОЛО ЦЕЛИ
func castHealSingle(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	_ *PlayerState) error {
	if m == nil || caster == nil || owner == nil {
		return errors.New("nil state")
	}
	heal := caster.SkillValue
	slot, target := owner.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
		return ErrCardSkillBadTarget
	}
	target.HP += heal
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     heal,
				Died:       false,
				NewHP:      target.HP,
			},
		},
	})
	return nil
}

// ФУНКЦИЯ ПОД СПЛЕШ ДАМАГ +1 ОТ ТАРГЕТА
func castDamageSplash(
	m *MatchState, a Action,
	caster *UnitState,
	_ *PlayerState,
	enemy *PlayerState) error {
	if enemy == nil {
		return errors.New("nil enemy player")
	}
	centerSlot, center := enemy.FindSlot(a.TargetInstanceID)
	if center == nil || centerSlot < 0 {
		return ErrCardSkillBadTarget
	}
	targetSlots := []int{centerSlot}
	if left := centerSlot - 1; left >= 0 && enemy.Table[left] != nil {
		targetSlots = append(targetSlots, left)
	}
	if right := centerSlot + 1; right < TableSize && enemy.Table[right] != nil {
		targetSlots = append(targetSlots, right)
	}
	targets := make([]EventTarget, 0, len(targetSlots))
	for _, s := range targetSlots {
		u := enemy.Table[s]
		if u == nil {
			continue
		}
		inst, tpl := u.InstanceID, u.TemplateID
		baseDamage := caster.SkillValue
		damage := baseDamage
		if s != centerSlot {
			damage = baseDamage / 2
		}
		u.HP -= damage
		died := u.HP <= 0
		newHP := u.HP
		if died {
			if err := killUnitAt(m, 1-a.PlayerIndex, s, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
		}
		targets = append(targets, EventTarget{
			InstanceID: inst,
			TemplateID: tpl,
			Amount:     damage,
			Died:       died,
			NewHP:      newHP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       centerSlot,
		Targets:          targets,
	})
	return nil
}

// ФУНКЦИЯ ПОД БАФ СЕБЯ
func castBuffSelf(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || owner == nil {
		return errors.New("nil state in castBuffSelf")
	}
	effectType := ""
	switch caster.SkillParamsJSON {
	case "", cards.DamageUpdate:
		effectType = cards.DamageUpdate
	case cards.HealthPointsUpdate:
		effectType = cards.HealthPointsUpdate
	case cards.MaxHealthPointsUpdate:
		effectType = cards.MaxHealthPointsUpdate
	case cards.CoolDownUpdate:
		effectType = cards.CoolDownUpdate
	case cards.MakeTankUpdate:
		effectType = cards.MakeTankUpdate
	case cards.SkillDamageUpdate:
		effectType = cards.SkillDamageUpdate
	case cards.SkillCooldownUpdate:
		effectType = cards.SkillCooldownUpdate
	default:
		return ErrCardSkillUnsupported
	}
	AddEffect(caster, UnitEffect{EffectType: effectType,
		TurnsLeft: caster.SkillDuration, Value: caster.SkillValue})
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		Targets: []EventTarget{
			{
				InstanceID: a.CardInstanceID,
				TemplateID: caster.TemplateID,
				Amount:     caster.SkillValue,
				Died:       false,
				NewHP:      caster.HP,
			},
		},
	})
	return nil
}

// ДЕБАФ СОЛО ЦЕЛИ (ПРОТИВНИКА)
func castApplyDebuff(
	m *MatchState, a Action,
	caster *UnitState,
	_ *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || enemy == nil {
		return errors.New("nil state in castApplyDebuff")
	}
	if a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	slot, target := enemy.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
		return ErrCardSkillBadTarget
	}
	mode := caster.SkillParamsJSON
	if mode == "" {
		mode = cards.DotHPUpdate
	}
	targetSlots := []int{slot}
	if caster.SkillTarget == cards.TargetEnemySplash {
		left := slot - 1
		right := slot + 1
		if left >= 0 && enemy.Table[left] != nil {
			targetSlots = append(targetSlots, left)
		}
		if right < TableSize && enemy.Table[right] != nil {
			targetSlots = append(targetSlots, right)
		}
	}
	targets := make([]EventTarget, 0, len(targetSlots))
	for _, s := range targetSlots {
		u := enemy.Table[s]
		if u == nil {
			continue
		}
		switch mode {
		case cards.DotAttackUpdate:
			AddEffect(target, UnitEffect{
				EffectType: cards.DotAttackUpdate,
				TurnsLeft:  caster.SkillDuration,
				Value:      caster.SkillValue,
			})
		case cards.DotHPUpdate:
			AddEffect(u, UnitEffect{
				EffectType: cards.DotHPUpdate,
				TurnsLeft:  caster.SkillDuration,
				Value:      caster.SkillValue,
			})
		case cards.DotCooldownUpdate:
			AddEffect(u, UnitEffect{
				EffectType: cards.DotCooldownUpdate,
				TurnsLeft:  caster.SkillDuration,
				Value:      caster.SkillValue,
			})
		default:
			return ErrCardSkillUnsupported
		}
		targets = append(targets, EventTarget{
			InstanceID: u.InstanceID,
			TemplateID: u.TemplateID,
			Amount:     caster.SkillValue,
			Died:       false,
			NewHP:      u.HP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets:          targets,
	})
	return nil
}

// ФУНКЦИЯ ПОД ПРИЗЫВ КОПИЙ ПРИЗЫВАТЕЛЯ НА СТОЛ
func castSummonSelfCopy(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	_ *PlayerState) error {
	if m == nil || caster == nil || owner == nil {
		return errors.New("nil state in castSummonSelfCopy")
	}
	need := caster.SkillValue
	if need <= 0 {
		need = 1
	}
	summoned := 0
	targets := make([]EventTarget, 0, need)
	for i := 0; i < need; i++ {
		slot := firstFreeSlot(owner)
		if slot < 0 {
			break
		}
		newIDforSummonedCard := fmt.Sprintf("%s_copy_%d_%d", caster.InstanceID, time.Now().UnixNano(), i)
		u := &UnitState{
			InstanceID:        newIDforSummonedCard,
			TemplateID:        caster.TemplateID,
			GamerCardID:       0,
			CardLevel:         caster.CardLevel,
			HP:                caster.HP,
			Attack:            caster.Attack,
			SplashRadius:      caster.SplashRadius,
			CanBeUpgraded:     caster.CanBeUpgraded,
			Cooldown:          0,
			IsTank:            caster.IsTank,
			SummonedInTurn:    owner.Turns,
			Effects:           nil,
			CardType:          caster.CardType,
			SkillName:         caster.SkillName,
			SkillCode:         caster.SkillCode,
			SkillTrigger:      caster.SkillTrigger,
			SkillTarget:       caster.SkillTarget,
			SkillValue:        caster.SkillValue,
			SkillDuration:     caster.SkillDuration,
			SkillCooldown:     caster.SkillCooldown,
			SkillCooldownLeft: caster.SkillCooldown,
			SkillParamsJSON:   caster.SkillParamsJSON,
		}
		owner.Table[slot] = u
		summoned++
		targets = append(targets, EventTarget{
			InstanceID: u.InstanceID,
			TemplateID: u.TemplateID,
		})
	}
	if summoned == 0 {
		return ErrSlotOccupied
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		Targets:          targets,
	})
	return nil
}

// ФУНКЦИЯ УДАЛЕНИЯ КАРТЫ СО СТОЛА (убить карту противника мгновенно)
func castBanishUnit(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || enemy == nil {
		return errors.New("nil state in castBanishUnit")
	}
	if a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	slot, target := enemy.FindSlot(a.TargetInstanceID)
	if target == nil || slot < 0 {
		return ErrCardSkillBadTarget
	}
	if err := killUnitAt(m, 1-a.PlayerIndex, slot, caster.InstanceID, a.PlayerIndex); err != nil {
		return err
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Died:       true,
				NewHP:      0,
			},
		},
	})
	return nil
}

// РЕВИЛ РУКИ ИГРОКА
func castRevealEnemyHand(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || enemy == nil {
		return errors.New("nil state from castRevileEnemyHand")
	}
	viewer := a.PlayerIndex
	targets := make([]EventTarget, 0, len(enemy.Hand))
	for i := 0; i < len(enemy.Hand); i++ {
		hand := enemy.Hand[i]
		targets = append(targets, EventTarget{
			InstanceID: hand.InstanceID,
			TemplateID: hand.TemplateID,
		})
	}
	m.Events = append(m.Events, Event{
		Type:                  string(EventCardSkill),
		PlayerIndex:           a.PlayerIndex,
		SourceKind:            string(SourceUnit),
		SourceInstanceID:      caster.InstanceID,
		SourceTemplateID:      caster.TemplateID,
		Targets:               targets,
		VisibleForPlayerIndex: &viewer,
	})
	return nil
}

// УВЕЛИЧИТЬ КД ВСЕМ КАРТАМ ПРОТИВНИКА ПОСЛЕ СМЕРТИ
func castIncEnemyCdAfterDeath(
	m *MatchState, a Action,
	caster *UnitState,
	_ *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || enemy == nil {
		return errors.New("nil state from castIncEnemyCdAfterDeath")
	}
	targets := make([]EventTarget, 0, TableSize)
	for i := 0; i < TableSize; i++ {
		unit := enemy.Table[i]
		if unit == nil {
			continue
		}
		AddEffect(unit, UnitEffect{
			EffectType: cards.CoolDownUpdate,
			TurnsLeft:  caster.SkillDuration,
			Value:      -caster.SkillValue,
		})
		targets = append(targets, EventTarget{
			InstanceID: unit.InstanceID,
			TemplateID: unit.TemplateID,
			Amount:     caster.SkillValue,
			Died:       false,
			NewHP:      unit.HP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		Targets:          targets,
	})
	return nil
}

// УВЕЛИЧЕНИЕ КД СОЛО КАРТЕ ПРОТИВНИКА
func castIncEnemyCdSingle(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	enemy *PlayerState) error {
	if m == nil || caster == nil || enemy == nil {
		return errors.New("nil state from castIncEnemyCdSingle")
	}
	slot, target := enemy.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
		return ErrCardSkillBadTarget
	}
	AddEffect(target, UnitEffect{
		EffectType: cards.CoolDownUpdate,
		TurnsLeft:  caster.SkillDuration,
		Value:      -caster.SkillValue,
	})
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     caster.SkillValue,
				Died:       false,
				NewHP:      target.HP,
			},
		},
	})
	return nil
}

// ФУНКЦИЯ УМЕНЬШЕНИЯ КД АТАКИ СВОЮЗНОЙ КАРТЫ
func castDecAllyCdSingle(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	_ *PlayerState) error {
	if m == nil || caster == nil {
		return errors.New("nil state from castDecAllyCdSingle")
	}
	slot, target := owner.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
		return ErrCardSkillBadTarget
	}
	AddEffect(target, UnitEffect{
		EffectType: cards.CoolDownUpdate,
		TurnsLeft:  caster.SkillDuration,
		Value:      caster.SkillValue,
	})
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		TargetSlot:       slot,
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     caster.SkillValue,
				Died:       false,
				NewHP:      target.HP,
			},
		},
	})
	return nil
}

// СКИЛЛ ПОД РАЗЪЕБ ВСЕГО СТОЛА ПРИ СМЕРТИ (либо вражеского, либо вообще всего, зависит от таргета)
func castDeathAoe(
	m *MatchState, a Action,
	caster *UnitState,
	_ *PlayerState,
	_ *PlayerState) error {
	if m == nil || caster == nil {
		return errors.New("nil state from castDeathAoe")
	}
	targets, err := collectAoeUnits(m, a.PlayerIndex, caster.SkillTarget)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	outEvent := make([]EventTarget, 0, len(targets))
	for _, v := range targets {
		u := v.unit
		if u.InstanceID == caster.InstanceID {
			continue
		}
		u.HP -= caster.SkillValue
		died := u.HP <= 0
		newHP := u.HP
		if died {
			if err := killUnitAt(m, v.ownerIndex, v.slot, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
		}
		outEvent = append(outEvent, EventTarget{
			InstanceID: u.InstanceID,
			TemplateID: u.TemplateID,
			Amount:     caster.SkillValue,
			Died:       died,
			NewHP:      newHP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		Targets:          outEvent,
	})
	return nil
}

// СТРУКТУРА ВЫБОРА ЦЕЛИ ДЛЯ АОЕ СКИЛЛА
type aoeTarget struct {
	ownerIndex int
	slot       int
	unit       *UnitState
}

// ФУНКЦИЯ ВОЗВРАЩАЮЩАЯ ЦЕЛИ ДЛЯ АОЕ УРОНА
func collectAoeUnits(m *MatchState,
	playerIndex int, skillTarget string) ([]aoeTarget, error) {
	if m == nil || skillTarget == "" {
		return nil, errors.New("bad input in collectAoeUnits")
	}
	if playerIndex < 0 || playerIndex > 1 {
		return nil, errors.New("bad player index")
	}
	out := make([]aoeTarget, 0, TableSize*2)
	collect := func(ownerIdx int) {
		p := m.Players[ownerIdx]
		if p == nil {
			return
		}
		for s := 0; s < TableSize; s++ {
			u := p.Table[s]
			if u == nil {
				continue
			}
			out = append(out, aoeTarget{
				ownerIndex: ownerIdx,
				slot:       s,
				unit:       u,
			})
		}
	}
	switch skillTarget {
	case cards.TargetEnemyAll:
		collect(1 - playerIndex)
	case cards.TargetAllyAll:
		collect(playerIndex)
	case cards.TargetBothAll:
		collect(playerIndex)
		collect(1 - playerIndex)
	default:
		return nil, ErrCardSkillBadTarget
	}
	return out, nil
}

func getCardSkillHandler(code string) (CardSkillHandler, error) {
	h, ok := CardSkillHandlers[code]
	if !ok {
		return nil, ErrCardSkillUnsupported
	}
	return h, nil
}

func firstFreeSlot(p *PlayerState) int {
	for i := 0; i < TableSize; i++ {
		if p.Table[i] == nil {
			return i
		}
	}
	return -1
}

// ФУНКЦИЯ ПОДНЯТИЯ КАРТЫ ИЗ МОГИЛЫ КОНКРЕТНОЙ КАРТОЙ
func castResurrectTargetFromGrave(
	m *MatchState, a Action,
	caster *UnitState,
	owner *PlayerState,
	_ *PlayerState) error {
	if m == nil || caster == nil || owner == nil {
		return errors.New("bad input in castResurrectTargetFromGrave")
	}
	for i := range owner.GraveYard {
		if owner.GraveYard[i].Unit.InstanceID != a.TargetInstanceID {
			continue
		}
		slot := firstFreeSlot(owner)
		if slot < 0 {
			return ErrTablesFull
		}
		revived := owner.GraveYard[i].Unit
		if revived.HP <= 0 {
			revived.HP = caster.HP / 2
			if revived.HP < 1 {
				revived.HP = 1
			}
		}
		revived.SummonedInTurn = owner.Turns
		revived.ResurrectedUsed = true
		owner.Table[slot] = &revived
		last := len(owner.GraveYard) - 1
		owner.GraveYard[i] = owner.GraveYard[last]
		owner.GraveYard = owner.GraveYard[:last]
		m.Events = append(m.Events, Event{
			Type:             string(EventResurrect),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
			TargetSlot:       slot,
			Targets: []EventTarget{
				{
					InstanceID: revived.InstanceID,
					TemplateID: revived.TemplateID,
					Died:       false,
					NewHP:      revived.HP,
				},
			},
		})
		return nil
	}
	return ErrCardSkillBadTarget
}
