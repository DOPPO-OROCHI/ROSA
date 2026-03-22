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

var CardSkillHandlers = map[string]CardSkillHandler{
	"damage_splash":             castDamageSplash,
	"damage_single":             castDamageSingle,
	"apply_buff":                castBuffSelf,
	"apply_debuff":              castApplyDebuff,
	"summon_self_copy":          castSummonSelfCopy,
	"banish_unit":               castBanishUnit,
	"reveal_enemy_hand":         castRevealEnemyHand,
	"inc_enemy_cd_all_on_death": castIncEnemyCdAfterDeath,
	"inc_enemy_cd_single":       castIncEnemyCdSingle,
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
	if a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	slot, target := enemy.FindSlot(a.TargetInstanceID)
	if slot < 0 || target == nil {
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
	if hasTank && !target.IsTank {
		return ErrCardSkillTargetTankBlocked
	}
	dmg := caster.SkillValue
	target.HP -= dmg
	died := target.HP <= 0
	newHP := target.HP
	if died {
		enemy.RemoveAt(slot)
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
		u.HP -= caster.SkillValue
		died := u.HP <= 0
		newHP := u.HP
		if died {
			enemy.RemoveAt(s)
			newHP = 0
		}
		targets = append(targets, EventTarget{
			InstanceID: inst,
			TemplateID: tpl,
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
	AddEffect(caster, UnitEffect{EffectType: cards.DamageUpdate,
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
		mode = "atk_down"
	}
	switch mode {
	case "atk_down":
		AddEffect(target, UnitEffect{
			EffectType: cards.DamageUpdate,
			TurnsLeft:  caster.SkillDuration,
			Value:      caster.SkillValue,
		})
	case "dot_hp":
		AddEffect(target, UnitEffect{
			EffectType: cards.HealthPointsUpdate,
			TurnsLeft:  caster.SkillDuration,
			Value:      caster.SkillValue,
		})
		if target.HP <= 0 {
			enemy.RemoveAt(slot)
		}
	case "cd_up":
		AddEffect(target, UnitEffect{
			EffectType: cards.CoolDownUpdate,
			TurnsLeft:  caster.SkillDuration,
			Value:      -caster.SkillValue,
		})
	default:
		return ErrCardSkillUnsupported
	}
	newHP := target.HP
	died := true
	if _, u := enemy.FindSlot(a.TargetInstanceID); u != nil {
		newHP = u.HP
		died = u.HP <= 0
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
				InstanceID: a.TargetInstanceID,
				TemplateID: target.TemplateID,
				Amount:     caster.SkillValue,
				Died:       died,
				NewHP:      newHP,
			},
		},
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
		SourceKind:       string(SourceCard),
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
	enemy.RemoveAt(slot)
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
		m.Events = append(m.Events, Event{
			Type:                  string(EventCardSkill),
			PlayerIndex:           a.PlayerIndex,
			SourceKind:            string(SourceUnit),
			SourceInstanceID:      caster.InstanceID,
			SourceTemplateID:      caster.TemplateID,
			Targets:               targets,
			VisibleForPlayerIndex: &viewer,
		})
	}
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

// УВЕЛИЧЕНИЕ КД СОЛО КАРТЕ
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
				Amount:     target.SkillValue,
				Died:       false,
				NewHP:      0,
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
		
	})
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
