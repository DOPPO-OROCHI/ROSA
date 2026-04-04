package game

import "errors"

/*Данный файл целиком и полностью посвящен хендлерам хил скилов. Ура!*/

//КАСТУЕМ ХИЛ НА СЕБЯ
func CastHealSelfSkill(m *MatchState, a Action, caster *UnitState) error {
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
	heal := caster.Skill.Power
	beforeHP := caster.HP
	if caster.HP >= caster.MaxHP {
		return errors.New("cant cast this skill while HP full")
	}
	caster.HP += heal
	if caster.HP > caster.MaxHP {
		caster.HP = caster.MaxHP
	}
	actualHeal := caster.HP - beforeHP
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets: []EventTarget{
			{
				InstanceID: caster.InstanceID,
				TemplateID: caster.TemplateID,
				Amount:     actualHeal,
				Died:       false,
				NewHP:      caster.HP,
			},
		},
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

//КАСТУЕМ ХИЛ НА КОГО НИБУДЬ СОЛО
func CastHealSingleSkill(m *MatchState, a Action, caster *UnitState) error {
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
	owner := m.Players[a.PlayerIndex]
	enemy := m.Players[1-a.PlayerIndex]
	if owner == nil || enemy == nil {
		return errors.New("nil enemy or owner state")
	}
	if _, enemyTarget := enemy.FindSlot(a.TargetInstanceID); enemyTarget != nil {
		return ErrHealerCannotHealEnemy
	}
	_, target := owner.FindSlot(a.TargetInstanceID)
	if target == nil {
		return ErrTargetNotFound
	}
	heal := caster.Skill.Power
	beforeHP := target.HP
	if target.HP >= target.MaxHP {
		return errors.New("cant cast this skill while HP full")
	}
	target.HP += heal
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	actualHeal := target.HP - beforeHP
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     actualHeal,
				Died:       false,
				NewHP:      target.HP,
			},
		},
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

//КАСТУЕМ ХИЛ ВОКРУГ СЕБЯ (вокруг карты типа, на 1 слот) ИЛИ НА СЕБЯ ЕСЛИ НИКОГО НЕТ
func CastHealAdjacentSkill(m *MatchState, a Action, caster *UnitState) error {
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
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	casterSlot := -1
	for slot := 0; slot < TableSize; slot++ {
		u := owner.Table[slot]
		if u != nil && u.InstanceID == caster.InstanceID {
			casterSlot = slot
			break
		}
	}
	if casterSlot == -1 {
		return ErrCardSkillBadTarget
	}
	targets := make([]EventTarget, 0, 2)
	heal := caster.Skill.Power
	healTarget := func(target *UnitState) {
		if target == nil {
			return
		}
		beforeHP := target.HP
		target.HP += heal
		if target.HP > target.MaxHP {
			target.HP = target.MaxHP
		}
		actualHeal := target.HP - beforeHP
		targets = append(targets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     actualHeal,
			Died:       false,
			NewHP:      target.HP,
		})
	}
	left := casterSlot - 1
	right := casterSlot + 1
	if left >= 0 {
		healTarget(owner.Table[left])
	}
	if right < TableSize {
		healTarget(owner.Table[right])
	}
	if len(targets) == 0 {
		if caster.HP >= caster.MaxHP {
			return errors.New("cant cast this skill while HP full and noones near the healer")
		}
		beforeHP := caster.HP
		caster.HP += caster.Skill.Power
		if caster.HP > caster.MaxHP {
			caster.HP = caster.MaxHP
		}
		actualHeal := caster.HP - beforeHP
		targets = append(targets, EventTarget{
			InstanceID: caster.InstanceID,
			TemplateID: caster.TemplateID,
			Amount:     actualHeal,
			Died:       false,
			NewHP:      caster.HP,
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
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

//ХИЛИМ ВООБЩЕ ВСЕХ И СЕБЯ В ТОМ ЧИСЛЕ
func CastHealAllAlliesSkill(m *MatchState, a Action, caster *UnitState) error {
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
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	heal := caster.Skill.Power
	targets := make([]EventTarget, 0, TableSize)
	hasTargets := false
	for i := 0; i < TableSize; i++ {
		u := owner.Table[i]
		if u == nil {
			continue
		}
		hasTargets = true
		beforeHP := u.HP
		u.HP += heal
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
		actualHeal := u.HP - beforeHP
		targets = append(targets, EventTarget{
			InstanceID: u.InstanceID,
			TemplateID: u.TemplateID,
			Amount:     actualHeal,
			Died:       false,
			NewHP:      u.HP,
		})
	}
	if !hasTargets {
		return ErrCardSkillBadTarget
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

// ХИЛИМ САМОГО СЛАБОГО ПО ХП БОЙЦА
func CastHealLowestHPSkill(m *MatchState, a Action, caster *UnitState) error {
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
	if a.AttackHero || a.TargetInstanceID != "" {
		return ErrCardSkillBadTarget
	}
	var target *UnitState
	for slot := 0; slot < TableSize; slot++ {
		u := owner.Table[slot]
		if u == nil {
			continue
		}
		if target == nil || u.HP < target.HP {
			target = u
		}
	}
	if target == nil {
		return ErrCardSkillBadTarget
	}
	beforeHP := target.HP
	heal := caster.Skill.Power
	inst := target.InstanceID
	tplID := target.TemplateID
	if target.HP >= target.MaxHP {
		return errors.New("cant cast this skill while HP full")
	}
	target.HP += heal
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	newHP := target.HP
	actualHeal := newHP - beforeHP
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets: []EventTarget{
			{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     actualHeal,
				Died:       false,
				NewHP:      newHP,
			},
		},
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// ХИЛИМ САМОГО СИЛЬНОГО ПО АТАКЕ БОЙЦА
func CastHealHighestAttackSkill(m *MatchState, a Action, caster *UnitState) error {
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
	if a.AttackHero || a.TargetInstanceID != "" {
		return ErrCardSkillBadTarget
	}
	var target *UnitState
	for slot := 0; slot < TableSize; slot++ {
		u := owner.Table[slot]
		if u == nil {
			continue
		}
		if target == nil || u.Attack > target.Attack {
			target = u
		}
	}
	if target == nil {
		return ErrCardSkillBadTarget
	}
	beforeHP := target.HP
	heal := caster.Skill.Power
	inst := target.InstanceID
	tplID := target.TemplateID
	if target.HP >= target.MaxHP {
		return errors.New("cant cast this skill while HP full")
	}
	target.HP += heal
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	newHP := target.HP
	actualHeal := newHP - beforeHP
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets: []EventTarget{
			{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     actualHeal,
				Died:       false,
				NewHP:      newHP,
			},
		},
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}
