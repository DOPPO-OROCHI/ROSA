package game

import (
	"TheWar/internal/domain/cards"
	"errors"
	"math/rand/v2"
)

// Здесь все о бафах-дебафах

// ПРИМЕНЯЕМ НА СОЮЗНУЮ ЦЕЛЬ БАФ, В ЗАВИСИМОСТИ ОТ ТИПА БАФА И ТД
func CastBuffSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if HasEffect(caster, cards.DebuffEffectStun) {
		return errors.New("caster is stunned")
	}
	if HasEffect(caster, cards.DebuffEffectSilence) {
		return errors.New("caster is silenced")
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
		_, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, target)
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
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
	case cards.SkillTargetAllyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
		//ПОТОМ ДОПИСАТЬ КАСТ БАФА НА РАНДОМНОГО СОЮЗНИКА
	default:
		return ErrCardSkillUnsupported
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	for _, target := range targets {
		if target == nil {
			continue
		}
		if caster.Skill.BuffEffect == cards.BuffEffectMulticast && target.InstanceID == caster.InstanceID {
			return errors.New("cant cast multicast on self")
		}
		if caster.Skill.BuffEffect == cards.BuffEffectMakeTank && target.IsTank {
			return errors.New("target already tank")
		}
		e := UnitEffect{
			EffectType:       caster.Skill.BuffEffect,
			TurnsLeft:        caster.Skill.Duration,
			Value:            caster.Skill.Power,
			ExtraValue:       caster.Skill.ExtraValue,
			SourceType:       "skill",
			Polarity:         "buff",
			SourceInstanceID: caster.InstanceID,
			Dispellable:      true,
			Targeting:        caster.Skill.Target,
		}
		if err := AddEffect(target, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     caster.Skill.Power,
			Died:       false,
			NewHP:      target.HP,
		})
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
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// КАСТ ДЕБАФА НА ПРОТИВНИКА
func CastDebuffSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if HasEffect(caster, cards.DebuffEffectStun) {
		return errors.New("caster is stunned")
	}
	if HasEffect(caster, cards.DebuffEffectSilence) {
		return errors.New("caster is silenced")
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	if caster.Skill.DebuffEffect == "" {
		return ErrCardSkillUnsupported
	}
	var targets []*UnitState
	switch caster.Skill.Target {
	case cards.SkillTargetEnemySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		_, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrTargetNotFound
		}
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return ErrCardSkillTargetTankBlocked
		}
		targets = append(targets, target)
	case cards.SkillTargetEnemyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
	case cards.SkillTargetEnemySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot < 0 {
			return ErrCardSkillBadTarget
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
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
	case cards.SkillTargetEnemyLowestHP:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		var target *UnitState
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
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
		targets = append(targets, target)
	case cards.SkillTargetEnemyHighestHP:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		var target *UnitState
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			if target == nil || u.HP > target.HP {
				target = u
			}
		}
		if target == nil {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, target)
	case cards.SkillTargetEnemyHighestAttack:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		var target *UnitState
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
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
		targets = append(targets, target)
	case cards.SkillTargetEnemyLowestAttack:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		var target *UnitState
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			if target == nil || u.Attack < target.Attack {
				target = u
			}
		}
		if target == nil {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, target)
	case cards.SkillTargetEnemyRandom:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		pool := make([]*UnitState, 0, TableSize)
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !u.IsTank {
				continue
			}
			pool = append(pool, u)
		}
		if len(pool) == 0 {
			return ErrCardSkillBadTarget
		}
		pick := rand.IntN(len(pool))
		targets = append(targets, pool[pick])
	case cards.SkillTargetEnemyRandomMulti:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		pool := make([]*UnitState, 0, TableSize)
		for i := 0; i < TableSize; i++ {
			u := enemy.Table[i]
			if u == nil {
				continue
			}
			if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !u.IsTank {
				continue
			}
			pool = append(pool, u)
		}
		if len(pool) == 0 {
			return ErrCardSkillBadTarget
		}
		hits := caster.Skill.ApplyCount
		if hits <= 0 {
			hits = 1
		}
		for i := 0; i < hits; i++ {
			pick := rand.IntN(len(pool))
			targets = append(targets, pool[pick])
		}
	default:
		return ErrCardSkillUnsupported
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	for _, target := range targets {
		if target == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       caster.Skill.DebuffEffect,
			TurnsLeft:        caster.Skill.Duration,
			Value:            caster.Skill.Power,
			ExtraValue:       caster.Skill.ExtraValue,
			SourceType:       "skill",
			Polarity:         "debuff",
			SourceInstanceID: caster.InstanceID,
			Dispellable:      true,
			Targeting:        caster.Skill.Target,
		}
		if err := AddEffect(target, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     caster.Skill.Power,
			Died:       false,
			NewHP:      target.HP,
		})
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
		Targets:          eventTargets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// ДИСПЕЛЛ ПРИКОЛ. ПО СУТИ СНИМАЕМ ЭФФЕКТ С СОЮЗНОЙ КАРТЫ.
func CastDispelDebuffsFromAllySkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if HasEffect(caster, cards.DebuffEffectStun) {
		return errors.New("caster is stunned")
	}
	if HasEffect(caster, cards.DebuffEffectSilence) {
		return errors.New("caster is silenced")
	}
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	var targets []*UnitState
	switch caster.Skill.Target {
	//кастуем диспелл на себя
	case cards.SkillTargetSelf:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, caster)
		//кастуем диспел на соло союзника
	case cards.SkillTargetAllySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		_, target := owner.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrTargetNotFound
		}
		targets = append(targets, target)
		//кастуем диспелл на рядом стоящих союзников
	case cards.SkillTargetAllyAdjacent:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		casterSlot := -1
		for slot := 0; slot < TableSize; slot++ {
			//здесь ищу самого кастера
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
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
		//кастуем на всех союзников диспелл
	case cards.SkillTargetAllyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
	default:
		return ErrCardSkillUnsupported
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	totalRemoved := 0
	for _, target := range targets {
		if target == nil {
			continue
		}
		toRemoveIdx := make([]int, 0)
		switch caster.Skill.CleanseMode {
		//снимаем один дебаф с цели
		case cards.CleanseModeRemoveDebuff:
			for i, e := range target.Effects {
				if e.Dispellable && e.Polarity == "debuff" {
					toRemoveIdx = append(toRemoveIdx, i)
					break
				}
			}
			//снимаем вообще все дебафы с цели
		case cards.CleanseModeRemoveAllDebuffs:
			for i, e := range target.Effects {
				if e.Dispellable && e.Polarity == "debuff" {
					toRemoveIdx = append(toRemoveIdx, i)
				}
			}
			//снимаем вообще все эффекты с цели
		case cards.CleanseModeRemoveAllEffects:
			for i, e := range target.Effects {
				if e.Dispellable {
					toRemoveIdx = append(toRemoveIdx, i)
				}
			}
		default:
			return ErrCardSkillUnsupported
		}
		if len(toRemoveIdx) == 0 {
			continue
		}
		drop := make(map[int]struct{}, len(toRemoveIdx))
		for _, idx := range toRemoveIdx {
			drop[idx] = struct{}{}
			if err := RemoveEffect(target, target.Effects[idx]); err != nil {
				return err
			}
		}
		kept := make([]UnitEffect, 0, len(target.Effects)-len(toRemoveIdx))
		for i, e := range target.Effects {
			if _, ok := drop[i]; ok {
				continue
			}
			kept = append(kept, e)
		}
		target.Effects = kept
		totalRemoved += len(toRemoveIdx)
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     len(toRemoveIdx),
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if totalRemoved == 0 {
		return errors.New("nothing to dispel")
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

// ТО ЖЕ САМОЕ, ТОЛЬКО УЖЕ СНИМАЕМ ПОЛОЖИТЕЛЬНЫЕ ЭФФЕКТЫ С ПРОТИВНИКА
func CastDispelBuffsFromEnemySkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if HasEffect(caster, cards.DebuffEffectStun) {
		return errors.New("caster is stunned")
	}
	if HasEffect(caster, cards.DebuffEffectSilence) {
		return errors.New("caster is silenced")
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	var targets []*UnitState
	switch caster.Skill.Target {
	//кастуем диспел на соло противника
	case cards.SkillTargetEnemySingle:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		_, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil {
			return ErrTargetNotFound
		}
		targets = append(targets, target)
		//кастуем на всех противников диспелл
	case cards.SkillTargetEnemyAll:
		if a.AttackHero || a.TargetInstanceID != "" {
			return ErrCardSkillBadTarget
		}
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u != nil {
				targets = append(targets, u)
			}
		}
		if len(targets) == 0 {
			return ErrCardSkillBadTarget
		}
	case cards.SkillTargetEnemySplash:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrCardSkillBadTarget
		}
		targetSlot := -1
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u != nil && u.InstanceID == a.TargetInstanceID {
				targetSlot = slot
				break
			}
		}
		if targetSlot == -1 {
			return ErrCardSkillBadTarget
		}
		targets = append(targets, enemy.Table[targetSlot])
		left := targetSlot - 1
		right := targetSlot + 1
		if left >= 0 && enemy.Table[left] != nil {
			targets = append(targets, enemy.Table[left])
		}
		if right < TableSize && enemy.Table[right] != nil {
			targets = append(targets, enemy.Table[right])
		}
	default:
		return ErrCardSkillUnsupported
	}
	eventTargets := make([]EventTarget, 0, len(targets))
	totalRemoved := 0
	for _, target := range targets {
		if target == nil {
			continue
		}
		toRemoveIdx := make([]int, 0)
		switch caster.Skill.CleanseMode {
		//снимаем один баф с цели
		case cards.CleanseModeRemoveBuff:
			for i, e := range target.Effects {
				if e.Dispellable && e.Polarity == "buff" {
					toRemoveIdx = append(toRemoveIdx, i)
					break
				}
			}
			//снимаем вообще все бафы с цели
		case cards.CleanseModeRemoveAllBuffs:
			for i, e := range target.Effects {
				if e.Dispellable && e.Polarity == "buff" {
					toRemoveIdx = append(toRemoveIdx, i)
				}
			}
			//снимаем вообще все эффекты с цели
		case cards.CleanseModeRemoveAllEffects:
			for i, e := range target.Effects {
				if e.Dispellable {
					toRemoveIdx = append(toRemoveIdx, i)
				}
			}
		default:
			return ErrCardSkillUnsupported
		}
		if len(toRemoveIdx) == 0 {
			continue
		}
		drop := make(map[int]struct{}, len(toRemoveIdx))
		for _, idx := range toRemoveIdx {
			drop[idx] = struct{}{}
			if err := RemoveEffect(target, target.Effects[idx]); err != nil {
				return err
			}
		}
		kept := make([]UnitEffect, 0, len(target.Effects)-len(toRemoveIdx))
		for i, e := range target.Effects {
			if _, ok := drop[i]; ok {
				continue
			}
			kept = append(kept, e)
		}
		target.Effects = kept
		totalRemoved += len(toRemoveIdx)
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     len(toRemoveIdx),
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if totalRemoved == 0 {
		return errors.New("nothing to dispel")
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
