package game

import (
	"TheWar/internal/domain/cards"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
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
		ImpactVFXKey:          BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:          BuildSFXKey(caster.AssetBaseKey, "impact"),
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// СПЕЦИАЛЬНЫЙ ЭФФЕКТ. ТУТ МЫ МЕНЯЕМ МЕСТАМИ ХП И АТАКУ СОЮЗНОЙ ЦЕЛИ
func CastEqualizeAllyHPSkill(m *MatchState, a Action, caster *UnitState) error {
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
			return ErrCardSkillBadTarget
		}
		_, t := owner.FindSlot(a.TargetInstanceID)
		if t == nil {
			return ErrTargetNotFound
		}
		targets = append(targets, t)
	case cards.SkillTargetAllyAdjacent:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
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
		left := casterSlot - 1
		right := casterSlot + 1
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
	default:
		return ErrCardSkillUnsupported
	}
	if len(targets) == 0 {
		return ErrCardSkillBadTarget
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	for i := 0; i < len(targets); i++ {
		t := targets[i]
		if t == nil {
			continue
		}
		if t.Attack <= 0 {
			return ErrCardSkillUnsupported
		}
		beforeHP := t.HP
		newHP := t.Attack
		t.HP = newHP
		t.MaxHP = newHP
		t.Attack = beforeHP
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// ФУНКЦИЯ ПОД ХИЛ И УРОН
func CastHybridHealDamageSkill(m *MatchState, a Action, caster *UnitState) error {
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
	enemy := m.Players[1-a.PlayerIndex]
	if owner == nil || enemy == nil {
		return errors.New("nil owner or enemy state")
	}
	eventTargets := make([]EventTarget, 0, 2)
	healAmount := caster.Skill.Power
	damageAmount := caster.Skill.ExtraValue
	if healAmount > 0 && caster.HP < caster.MaxHP {
		beforeHP := caster.HP
		caster.HP += healAmount
		if caster.HP > caster.MaxHP {
			caster.HP = caster.MaxHP
		}
		actualHP := caster.HP - beforeHP
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: caster.InstanceID,
			TemplateID: caster.TemplateID,
			Amount:     actualHP,
			Died:       false,
			NewHP:      caster.HP,
		})
	}
	if damageAmount > 0 {
		targetSlots := make([]int, 0, TableSize)
		for slot := 0; slot < TableSize; slot++ {
			if enemy.Table[slot] != nil {
				targetSlots = append(targetSlots, slot)
			}
		}
		canHitHero := caster.Skill.IgnoreTank || !enemyHasTank(enemy)
		targetPollSize := len(targetSlots)
		if canHitHero {
			targetPollSize++
		}
		if targetPollSize > 0 {
			roll := rand.IntN(targetPollSize)
			if canHitHero && roll == len(targetSlots) {
				enemy.HeroHP -= damageAmount
				heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
				eventTargets = append(eventTargets, EventTarget{
					InstanceID: heroID,
					Amount:     damageAmount,
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
				}
			} else {
				targetSlot := targetSlots[roll]
				target := enemy.Table[targetSlot]
				if target != nil {
					inst := target.InstanceID
					tplID := target.TemplateID
					target.HP -= damageAmount
					died := target.HP <= 0
					newHP := target.HP
					if died {
						if err := killUnitAt(m, 1-a.PlayerIndex, targetSlot, caster.InstanceID, a.PlayerIndex); err != nil {
							return err
						}
						newHP = 0
					}
					eventTargets = append(eventTargets, EventTarget{
						InstanceID: inst,
						TemplateID: tplID,
						Amount:     damageAmount,
						Died:       died,
						NewHP:      newHP,
					})
				}
			}
		}
	}
	if len(eventTargets) == 0 {
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

/*
Суммоним свою копию на стол
Здесь количество копий определяется тем, сколько в описании карты SkillApplyCount
*/
func CastSummonSelfCopySkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or caster state")
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
	need := caster.Skill.ApplyCount
	if need <= 0 {
		need = 1
	}
	targets := make([]EventTarget, 0, need)
	summoned := 0
	for i := 0; i < need; i++ {
		slot := -1
		for s := 0; s < TableSize; s++ {
			if owner.Table[s] == nil {
				slot = s
				break
			}
		}
		if slot < 0 {
			break
		}
		newID := fmt.Sprintf("%s_copy_%d_%d", caster.InstanceID, time.Now().UnixNano(), i)
		copyUnit := &UnitState{
			InstanceID:      newID,
			TemplateID:      caster.TemplateID,
			GamerCardID:     0,
			CardLevel:       caster.CardLevel,
			HP:              caster.HP,
			MaxHP:           caster.MaxHP,
			Attack:          caster.Attack,
			SplashRadius:    caster.SplashRadius,
			IsTank:          caster.IsTank,
			CardType:        caster.CardType,
			BaseCooldown:    caster.BaseCooldown,
			Cooldown:        1,
			SummonedInTurn:  owner.Turns,
			ImageKey:        caster.ImageKey,
			AssetBaseKey:    caster.AssetBaseKey,
			HasSkill:        false,
			Effects:         nil,
			ResurrectedUsed: false,
			Passive:         caster.Passive,
		}
		owner.Table[slot] = copyUnit
		summoned++
		targets = append(targets, EventTarget{
			InstanceID: copyUnit.InstanceID,
			TemplateID: copyUnit.TemplateID,
		})
	}
	if summoned == 0 {
		return ErrSlotOccupied
	}
	if err := RefreshPassiveAuras(m); err != nil {
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
		ImpactVFXKey:     BuildVFXKey(caster.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(caster.AssetBaseKey, "impact"),
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}
