package game

import "errors"

//Здесь будут расположены всяк разные интересные скиллы

//УБИВАЕМ ВРАЖЕСКУЮ КАРТУ НАХУЙ
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
