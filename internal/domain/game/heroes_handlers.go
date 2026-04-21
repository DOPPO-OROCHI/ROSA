package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"errors"
	"fmt"
)

func CastHeroAttackAndHPBuffAbility(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.BuffEffect != cards.BuffEffectAttackAndHP {
		return errors.New("bad hero spell")
	}
	if owner.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	if owner.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	var targets []*UnitState
	switch spec.Target {
	case cards.SkillTargetAllySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		_, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
	case cards.SkillTargetAllySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		targetSlot, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
		left := targetSlot - 1
		right := targetSlot + 1
		if left >= 0 && owner.Table[left] != nil {
			targets = append(targets, owner.Table[left])
		}
		if right < TableSize && owner.Table[right] != nil {
			targets = append(targets, owner.Table[right])
		}
	case cards.SkillTargetAllyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrHeroAbilityBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrHeroAbilityBadTarget
		}
	default:
		return errors.New("unsupported hero ability target")
	}
	heroInstanseID := fmt.Sprintf("hero:p%d", a.PlayerIndex)
	eventTargets := make([]EventTarget, 0, len(targets))
	for _, target := range targets {
		if target == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.BuffEffect,
			TurnsLeft:        spec.Duration,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceHero),
			Polarity:         "buff",
			SourceInstanceID: heroInstanseID,
			Dispellable:      true,
			Targeting:        spec.Target,
		}
		if err := AddEffect(target, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     spec.Power,
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if len(eventTargets) == 0 {
		return ErrHeroAbilityBadTarget
	}
	owner.Mana -= spec.ManaCost
	owner.HeroAbilityCooldown = spec.CoolDown
	heroBase := HeroBaseKey(owner.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: owner.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		ImpactVFXKey:   BuildVFXKey(heroBase, "impact"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "impact"),
		Targets:        eventTargets,
	})
	return nil
}

func CastHeroDebuffAbility(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.Kind != cards.SkillKindDebuff {
		return errors.New("bad hero debuff spell")
	}
	if spec.DebuffEffect == "" || spec.DebuffEffect == cards.DebuffEffectNone {
		return ErrHeroAbilityUnknown
	}
	if owner.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	if owner.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	var targets []*UnitState
	switch spec.Target {
	case cards.SkillTargetEnemySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		_, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrHeroAbilityBadTarget
		}
		if !spec.IgnoreTank && enemyHasTank(enemy) {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
	case cards.SkillTargetEnemySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrHeroAbilityBadTarget
		}
		if !spec.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
		left := targetSlot - 1
		right := targetSlot + 1
		if left >= 0 && enemy.Table[left] != nil {
			targets = append(targets, enemy.Table[left])
		}
		if right < TableSize && enemy.Table[right] != nil {
			targets = append(targets, enemy.Table[right])
		}
	case cards.SkillTargetEnemyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrHeroAbilityBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrHeroAbilityBadTarget
		}
	default:
		return ErrHeroAbilityBadTarget
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	heroInstanseID := fmt.Sprintf("hero:p%d", a.PlayerIndex)
	for _, target := range targets {
		if target == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.DebuffEffect,
			TurnsLeft:        spec.Duration,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceHero),
			Polarity:         "debuff",
			SourceInstanceID: heroInstanseID,
			Dispellable:      true,
			Targeting:        spec.Target,
		}
		if err := AddEffect(target, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     spec.Power,
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if len(eventTargets) == 0 {
		return ErrHeroAbilityBadTarget
	}
	owner.Mana -= spec.ManaCost
	owner.HeroAbilityCooldown = spec.CoolDown
	heroBase := HeroBaseKey(owner.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: owner.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		ImpactVFXKey:   BuildVFXKey(heroBase, "impact"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "impact"),
		Targets:        eventTargets,
	})
	return nil
}

func CastHeroBuffAbility(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.Kind != cards.SkillKindBuff {
		return errors.New("bad hero buff spell")
	}
	if spec.BuffEffect == "" || spec.BuffEffect == cards.BuffEffectNone {
		return ErrHeroAbilityUnknown
	}
	if owner.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	if owner.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	var targets []*UnitState
	switch spec.Target {
	case cards.SkillTargetAllySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		_, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
	case cards.SkillTargetAllySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		targetSlot, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrHeroAbilityBadTarget
		}
		targets = append(targets, target)
		left := targetSlot - 1
		right := targetSlot + 1
		if left >= 0 && owner.Table[left] != nil {
			targets = append(targets, owner.Table[left])
		}
		if right < TableSize && owner.Table[right] != nil {
			targets = append(targets, owner.Table[right])
		}
	case cards.SkillTargetAllyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrHeroAbilityBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrHeroAbilityBadTarget
		}
	default:
		return ErrHeroAbilityBadTarget
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	heroInstanseID := fmt.Sprintf("hero:p%d", a.PlayerIndex)
	for _, target := range targets {
		if target == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.BuffEffect,
			TurnsLeft:        spec.Duration,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceHero),
			Polarity:         "buff",
			SourceInstanceID: heroInstanseID,
			Dispellable:      true,
			Targeting:        spec.Target,
		}
		if err := AddEffect(target, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     spec.Power,
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if len(eventTargets) == 0 {
		return ErrHeroAbilityBadTarget
	}
	owner.Mana -= spec.ManaCost
	owner.HeroAbilityCooldown = spec.CoolDown
	heroBase := HeroBaseKey(owner.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: owner.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		ImpactVFXKey:   BuildVFXKey(heroBase, "impact"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "impact"),
		Targets:        eventTargets,
	})
	return nil
}

func CastHeroDamageAbility(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.Kind != cards.SkillKindDamage {
		return errors.New("bad hero damage spell")
	}
	if owner.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	if owner.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	targets := make([]EventTarget, 0, TableSize)
	switch spec.Target {
	case cards.SkillTargetEnemySingle:
		if a.AttackHero {
			if !spec.IgnoreTank && enemyHasTank(enemy) {
				return ErrHeroAbilityBadTarget
			}
			enemy.HeroHP -= spec.Power
			if enemy.HeroHP <= 0 {
				enemy.HeroHP = 0
			}
			heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
			targets = append(targets, EventTarget{
				InstanceID: heroID,
				Amount:     spec.Power,
				Died:       enemy.HeroHP <= 0,
				NewHP:      enemy.HeroHP,
			})
			if enemy.HeroHP <= 0 {
				m.Finished = true
				if a.PlayerIndex == 0 {
					m.Result = MatchWinP1
				} else {
					m.Result = MatchWinP2
				}
				return nil
			}
		} else {
			if a.TargetInstanceID == "" {
				return ErrHeroAbilityBadTarget
			}
			targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
			if target == nil || targetSlot < 0 {
				return ErrHeroAbilityBadTarget
			}
			if !spec.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
				return ErrHeroAbilityBadTarget
			}
			inst := target.InstanceID
			tplID := target.TemplateID
			res, err := applyDamageToUnit(m, 1-a.PlayerIndex, targetSlot, target,
				spec.Power, fmt.Sprintf("hero:p%d", a.PlayerIndex), a.PlayerIndex, true)
			if err != nil {
				return err
			}
			targets = append(targets, EventTarget{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     res.DamageToHP,
				Died:       res.Died,
				NewHP:      res.NewHP,
			})
			if res.ReflectedDamage > 0 {
				owner.HeroHP -= res.ReflectedDamage
				if owner.HeroHP < 0 {
					owner.HeroHP = 0
				}
				selfHeroID := fmt.Sprintf("hero:p%d", a.PlayerIndex)
				targets = append(targets, EventTarget{
					InstanceID: selfHeroID,
					Amount:     res.ReflectedDamage,
					Died:       owner.HeroHP <= 0,
					NewHP:      owner.HeroHP,
				})
			}
		}
	case cards.SkillTargetEnemySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrHeroAbilityBadTarget
		}
		if !spec.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return ErrHeroAbilityBadTarget
		}
		targetSlots := []int{targetSlot}
		left := targetSlot - 1
		right := targetSlot + 1
		if left >= 0 && enemy.Table[left] != nil {
			targetSlots = append(targetSlots, left)
		}
		if right < TableSize && enemy.Table[right] != nil {
			targetSlots = append(targetSlots, right)
		}
		for _, slot := range targetSlots {
			u := enemy.Table[slot]
			if u == nil {
				continue
			}
			damage := spec.Power
			if slot != targetSlot {
				damage = damage / 2
			}
			inst := u.InstanceID
			tplID := u.TemplateID
			res, err := applyDamageToUnit(m, 1-a.PlayerIndex, slot, u, damage,
				fmt.Sprintf("hero:p%d", a.PlayerIndex), a.PlayerIndex, true)
			if err != nil {
				return err
			}
			targets = append(targets, EventTarget{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     res.DamageToHP,
				Died:       res.Died,
				NewHP:      res.NewHP,
			})
			if res.ReflectedDamage > 0 {
				owner.HeroHP -= res.ReflectedDamage
				if owner.HeroHP < 0 {
					owner.HeroHP = 0
				}
				selfHeroID := fmt.Sprintf("hero:p%d", a.PlayerIndex)
				targets = append(targets, EventTarget{
					InstanceID: selfHeroID,
					Amount:     res.ReflectedDamage,
					Died:       owner.HeroHP <= 0,
					NewHP:      owner.HeroHP,
				})
			}
		}
	case cards.SkillTargetEnemyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrHeroAbilityBadTarget
		}
		hasTarget := false
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u == nil {
				continue
			}
			hasTarget = true
			inst := u.InstanceID
			tplID := u.TemplateID
			res, err := applyDamageToUnit(
				m,
				1-a.PlayerIndex,
				slot,
				u,
				spec.Power,
				fmt.Sprintf("hero:p%d", a.PlayerIndex),
				a.PlayerIndex,
				false,
			)
			if err != nil {
				return err
			}
			targets = append(targets, EventTarget{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     res.DamageToHP,
				Died:       res.Died,
				NewHP:      res.NewHP,
			})
		}
		if !hasTarget {
			return ErrHeroAbilityBadTarget
		}
	default:
		return errors.New("unsupported hero ability target")
	}
	if len(targets) == 0 {
		return ErrHeroAbilityBadTarget
	}
	owner.Mana -= spec.ManaCost
	owner.HeroAbilityCooldown = spec.CoolDown
	heroBase := HeroBaseKey(owner.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: owner.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		ImpactVFXKey:   BuildVFXKey(heroBase, "impact"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "impact"),
		Targets:        targets,
	})
	if owner.HeroHP <= 0 {
		m.Finished = true
		if a.PlayerIndex == 0 {
			m.Result = MatchWinP2
		} else {
			m.Result = MatchWinP1
		}
		return nil
	}
	return nil
}

// ТОЛЬКО ПЕРМАНЕНТНОЕ ИЗМЕНЕНИЕ
func CastHeroHybridAbility(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.Kind != cards.SkillKindHybrid {
		return errors.New("bad hero hybrid spell")
	}
	if spec.BuffEffect == "" || spec.BuffEffect == cards.BuffEffectNone {
		return ErrHeroAbilityUnknown
	}
	if owner.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	if owner.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	if spec.Target != cards.SkillTargetAllySingle {
		return ErrHeroAbilityBadTarget
	}
	if a.AttackHero || a.TargetInstanceID == "" {
		return ErrHeroAbilityBadTarget
	}
	_, target := owner.FindSlot(a.TargetInstanceID)
	if target == nil {
		return ErrHeroAbilityBadTarget
	}
	e := UnitEffect{
		EffectType:       spec.BuffEffect,
		TurnsLeft:        spec.Duration,
		Value:            spec.Power,
		ExtraValue:       0,
		SourceType:       string(SourceHero),
		Polarity:         "buff",
		SourceInstanceID: fmt.Sprintf("hero:p%d", a.PlayerIndex),
		Dispellable:      true,
		Targeting:        spec.Target,
	}
	if err := AddEffect(target, e); err != nil {
		return err
	}
	if spec.ExtraValue <= 0 {
		return ErrHeroAbilityUnknown
	}
	target.HP = spec.ExtraValue
	target.MaxHP = spec.ExtraValue
	owner.Mana -= spec.ManaCost
	owner.HeroAbilityCooldown = spec.CoolDown
	heroBase := HeroBaseKey(owner.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: owner.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		ImpactVFXKey:   BuildVFXKey(heroBase, "impact"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "impact"),
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     spec.Power,
				Died:       false,
				NewHP:      target.HP,
			},
		},
	})
	return nil
}
