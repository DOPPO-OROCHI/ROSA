package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

func triggerEnemyOnDeathExplosion(m *MatchState, ownerIdx int, dead *UnitState, deadSlot int) error {
	if m == nil || ownerIdx > 1 || ownerIdx < 0 || dead == nil {
		return errors.New("bad state")
	}
	enemyIdx := 1 - ownerIdx
	enemy := m.Players[enemyIdx]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	var explosion *UnitEffect
	for i := range dead.Effects {
		e := &dead.Effects[i]
		if e.EffectType == cards.BuffEffectDeathExplosion {
			explosion = e
			break
		}
	}
	if explosion == nil {
		return nil
	}
	targetSlots := make([]int, 0, TableSize)
	switch explosion.Targeting {
	case cards.SkillTargetEnemySplash:
		if deadSlot >= 0 && deadSlot < TableSize && enemy.Table[deadSlot] != nil {
			targetSlots = append(targetSlots, deadSlot)
		}
		left := deadSlot - 1
		right := deadSlot + 1
		if left >= 0 && enemy.Table[left] != nil {
			targetSlots = append(targetSlots, left)
		}
		if right < TableSize && enemy.Table[right] != nil {
			targetSlots = append(targetSlots, right)
		}
	case cards.SkillTargetEnemyAll:
		for slot := 0; slot < TableSize; slot++ {
			if enemy.Table[slot] != nil {
				targetSlots = append(targetSlots, slot)
			}
		}
	default:
		return nil
	}
	damage := explosion.Value
	eventTargets := make([]EventTarget, 0, len(targetSlots))
	for _, slot := range targetSlots {
		u := enemy.Table[slot]
		if u == nil {
			continue
		}
		inst := u.InstanceID
		tplID := u.TemplateID
		res, err := applyDamageToUnit(m, enemyIdx, slot, u, damage, dead.InstanceID, ownerIdx, false)
		if err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     res.DamageToHP,
			Died:       res.Died,
			NewHP:      res.NewHP,
		})
	}
	if len(eventTargets) > 0 {
		m.Events = append(m.Events, Event{
			Type:             string(EventCardSkill),
			PlayerIndex:      ownerIdx,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: dead.InstanceID,
			SourceTemplateID: dead.TemplateID,
			VFXKey:           BuildVFXKey(dead.AssetBaseKey, "spell"),
			SFXKey:           BuildSFXKey(dead.AssetBaseKey, "spell"),
			ImpactVFXKey:     BuildVFXKey(dead.AssetBaseKey, "impact"),
			ImpactSFXKey:     BuildSFXKey(dead.AssetBaseKey, "impact"),
			Targets:          eventTargets,
		})
	}
	return nil
}

func triggerAllyOnDeathMassHeal(m *MatchState, ownerIdx int, dead *UnitState, deadSlot int) error {
	if m == nil || ownerIdx > 1 || ownerIdx < 0 || dead == nil {
		return errors.New("bad state")
	}
	owner := m.Players[ownerIdx]
	if owner == nil {
		return errors.New("nil owner state")
	}
	var healEffect *UnitEffect
	for i := range dead.Effects {
		e := &dead.Effects[i]
		if e.EffectType == cards.BuffEffectDeathMassHeal {
			healEffect = e
			break
		}
	}
	if healEffect == nil {
		return nil
	}
	targetSlots := make([]int, 0, TableSize)
	switch healEffect.Targeting {
	case cards.SkillTargetAllyAdjacent:
		left := deadSlot - 1
		right := deadSlot + 1
		if left >= 0 && left < TableSize && owner.Table[left] != nil {
			targetSlots = append(targetSlots, left)
		}
		if right >= 0 && right < TableSize && owner.Table[right] != nil {
			targetSlots = append(targetSlots, right)
		}
	case cards.SkillTargetAllyAll:
		for slot := 0; slot < TableSize; slot++ {
			if slot == deadSlot {
				continue
			}
			if owner.Table[slot] != nil {
				targetSlots = append(targetSlots, slot)
			}
		}
	default:
		return nil
	}
	heal := healEffect.Value
	eventTargets := make([]EventTarget, 0, len(targetSlots))
	for _, slot := range targetSlots {
		u := owner.Table[slot]
		if u == nil {
			continue
		}
		if HasEffect(u, cards.DebuffEffectNoHeal) {
			continue
		}
		beforeHP := u.HP
		u.HP += heal
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
		actualHeal := u.HP - beforeHP
		if actualHeal <= 0 {
			continue
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: u.InstanceID,
			TemplateID: u.TemplateID,
			Amount:     actualHeal,
			Died:       false,
			NewHP:      u.HP,
		})
	}
	if len(eventTargets) > 0 {
		m.Events = append(m.Events, Event{
			Type:             string(EventCardSkill),
			PlayerIndex:      ownerIdx,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: dead.InstanceID,
			SourceTemplateID: dead.TemplateID,
			VFXKey:           BuildVFXKey(dead.AssetBaseKey, "spell"),
			SFXKey:           BuildSFXKey(dead.AssetBaseKey, "spell"),
			ImpactVFXKey:     BuildVFXKey(dead.AssetBaseKey, "impact"),
			ImpactSFXKey:     BuildSFXKey(dead.AssetBaseKey, "impact"),
			Targets:          eventTargets,
		})
	}
	return nil
}
