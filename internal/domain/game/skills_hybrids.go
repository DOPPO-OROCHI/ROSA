package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

//Здесь будут расположены всяк разные интересные скиллы

// УБИВАЕМ ВРАЖЕСКУЮ КАРТУ НАХУЙ
func CastKillTargetSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if a.AttackHero || a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
	if target == nil || targetSlot < 0 {
		return ErrCardSkillBadTarget
	}
	if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
		return ErrCardSkillTargetTankBlocked
	}
	inst := target.InstanceID
	tplID := target.TemplateID
	targets := []EventTarget{
		{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     0,
			Died:       true,
			NewHP:      0,
		},
	}
	if err := killUnitAt(m, 1-a.PlayerIndex, targetSlot, caster.InstanceID, a.PlayerIndex); err != nil {
		return err
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// ПАЛИМ ЧУЖУЮ РУКУ (реально, буквально смотрим)
func CastVisionSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if a.AttackHero || a.TargetInstanceID != "" {
		return ErrCardSkillBadTarget
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil owner or enemy state")
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
		VFXKey:                BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:                BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:               targets,
		VisibleForPlayerIndex: &viewer,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

func CastSetFixedHPToAllySkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	var targets []*UnitState
	switch caster.Skill.Target {
	case cards.SkillTargetSelf:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, caster)
	case cards.SkillTargetAllySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrTargetNotFound
		}
		_, t := owner.FindSlot(a.TargetInstanceID)
		if t == nil {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, t)
	case cards.SkillTargetAllyAdjacent:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		casterSLot := -1
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u != nil && u.InstanceID == caster.InstanceID {
				casterSLot = slot
				break
			}
		}
		if casterSLot == -1 {
			return ErrCardSkillBadTarget
		}
		left := casterSLot - 1
		right := casterSLot + 1
		if left >= 0 && owner.Table[left] != nil {
			targets = append(targets, owner.Table[left])
		}
		if right < TableSize && owner.Table[right] != nil {
			targets = append(targets, owner.Table[right])
		}
	case cards.SkillTargetAllyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		for i := 0; i < TableSize; i++ {
			if owner.Table[i] == nil {
				continue
			}
			targets = append(targets, owner.Table[i])
		}
	}
	if len(targets) == 0 {
		return ErrCardSkillBadTarget
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	fixedHP := caster.Skill.ExtraValue
	if fixedHP <= 0 {
		return ErrCardSkillUnsupported
	}
	for i := 0; i < len(targets); i++ {
		t := targets[i]
		if t == nil {
			continue
		}
		beforeHP := t.HP
		t.HP = fixedHP
		t.MaxHP = fixedHP
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: t.InstanceID,
			TemplateID: t.TemplateID,
			Amount:     t.HP - beforeHP,
			Died:       false,
			NewHP:      t.HP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// СТАВИМ ФИКСИРОВАННОЕ КОЛ_ВО ХП ВРАЖЕСКОЙ КАРТЕ
func CastSetFixedHPToEnemySkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	var targets []*UnitState
	switch caster.Skill.Target {
	case cards.SkillTargetEnemySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		_, t := enemy.FindSlot(a.TargetInstanceID)
		if t == nil {
			return ErrCardSkillBadTarget
		}
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !t.IsTank {
			return ErrCardSkillTargetTankBlocked
		}
		targets = append(targets, t)
	case cards.SkillTargetEnemySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrTargetNotFound
		}
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return ErrCardSkillTargetTankBlocked
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
			return ErrCardSkillBadTarget
		}
		for i := 0; i < TableSize; i++ {
			if enemy.Table[i] == nil {
				continue
			}
			targets = append(targets, enemy.Table[i])
		}
	default:
		return ErrCardSkillUnsupported
	}
	if len(targets) == 0 {
		return ErrCardSkillBadTarget
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	fixedHP := caster.Skill.ExtraValue
	if fixedHP <= 0 {
		return ErrCardSkillUnsupported
	}
	for i := 0; i < len(targets); i++ {
		t := targets[i]
		if t == nil {
			continue
		}
		beforeHP := t.HP
		t.HP = fixedHP
		t.MaxHP = fixedHP
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: t.InstanceID,
			TemplateID: t.TemplateID,
			Amount:     t.HP - beforeHP,
			Died:       false,
			NewHP:      t.HP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}
