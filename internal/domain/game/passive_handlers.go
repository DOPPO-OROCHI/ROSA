package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

func HandlePassiveAuraBuff(
	m *MatchState,
	source *UnitState,
	owner *PlayerState,
	enemy *PlayerState,
	spec cards.PassiveSpec,
	ctx PassiveContext,
) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if source == nil {
		return errors.New("nil source unit")
	}
	if owner == nil {
		return errors.New("nil owner state")
	}
	if spec.Kind != cards.PassiveKindAura {
		return errors.New("bad passive kind aura")
	}
	if spec.Trigger != cards.PassiveTriggerWhileAlive {
		return errors.New("bad passive trigger")
	}
	if spec.BuffEffect == "" || spec.BuffEffect == cards.BuffEffectNone {
		return errors.New("bad passive buff effect")
	}
	if spec.ScaleMode != "" && spec.ScaleMode != cards.PassiveScaleNone {
		return errors.New("scaling aura must use dedicated handler")
	}
	switch spec.Target {
	case cards.SkillTargetSelf,
		cards.SkillTargetAllyAdjacent,
		cards.SkillTargetSelfAndAdjacent,
		cards.SkillTargetAllyAll:
	default:
		return errors.New("bad aura buff target")
	}
	clearPassiveAuraEffect(owner, source.InstanceID, spec)
	_, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targets := resolvePassiveTargets(owner, enemy, aliveSource, spec, ctx)
	if len(targets) == 0 {
		return nil
	}
	for _, target := range targets {
		if target.Unit == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.BuffEffect,
			TurnsLeft:        0,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceUnit),
			EffectLayer:      cards.EffectLayerPassive,
			Polarity:         "buff",
			SourceInstanceID: aliveSource.InstanceID,
			Dispellable:      false,
			Targeting:        spec.Target,
		}
		if err := AddEffect(target.Unit, e); err != nil {
			return err
		}
	}
	_ = ctx
	return nil
}

func HandlePassiveReactiveBuff(
	m *MatchState,
	source *UnitState,
	owner *PlayerState,
	enemy *PlayerState,
	spec cards.PassiveSpec,
	ctx PassiveContext,
) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if source == nil {
		return errors.New("nil source unit")
	}
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
	}
	if spec.Kind != cards.PassiveKindReactive {
		return errors.New("bad passive kind for reactive buff")
	}
	if spec.BuffEffect == "" || spec.BuffEffect == cards.BuffEffectNone {
		return errors.New("bad passive buff effect")
	}
	if spec.DebuffEffect != "" && spec.DebuffEffect != cards.DebuffEffectNone {
		return errors.New("reactive buff cant use debuff effect")
	}
	if spec.ScaleMode != "" && spec.ScaleMode != cards.PassiveScaleNone {
		return errors.New("scaling passive must use dedicated handler")
	}
	if ctx.Trigger != spec.Trigger {
		return nil
	}
	switch spec.Target {
	case cards.SkillTargetSelf,
		cards.SkillTargetAllyAll,
		cards.SkillTargetAllyAdjacent,
		cards.SkillTargetSelfAndAdjacent,
		cards.SkillTargetAllyLowestHP,
		cards.SkillTargetAllyHighestAttack:
	default:
		return errors.New("bad reactive buff target")
	}
	_, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targets := resolvePassiveTargets(owner, enemy, aliveSource, spec, ctx)
	if len(targets) == 0 {
		return nil
	}
	for _, target := range targets {
		if target.Unit == nil {
			continue
		}
		if spec.TargetRace != "" && target.Unit.CardType != spec.TargetRace {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.BuffEffect,
			TurnsLeft:        spec.Duration,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceUnit),
			EffectLayer:      cards.EffectLayerPassive,
			Polarity:         "buff",
			SourceInstanceID: aliveSource.InstanceID,
			Dispellable:      false,
			Targeting:        spec.Target,
		}
		if err := AddEffect(target.Unit, e); err != nil {
			return err
		}
	}
	_ = ctx
	return nil
}

func HandlePassiveReactiveDamage(
	m *MatchState,
	source *UnitState,
	owner *PlayerState,
	enemy *PlayerState,
	spec cards.PassiveSpec,
	ctx PassiveContext,
) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if source == nil {
		return errors.New("nil source unit")
	}
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
	}
	if ctx.Trigger != spec.Trigger {
		return nil
	}
	if spec.Kind != cards.PassiveKindReactive {
		return errors.New("bad passive kind for reactive damage")
	}
	if spec.BuffEffect != "" && spec.BuffEffect != cards.BuffEffectNone {
		return errors.New("reactive damage cant use buff effect")
	}
	if spec.DebuffEffect != "" && spec.DebuffEffect != cards.DebuffEffectNone {
		return errors.New("reactive damage cant use debuff effect")
	}
	if spec.ScaleMode != "" && spec.ScaleMode != cards.PassiveScaleNone {
		return errors.New("scaling passive must use dedicated handler")
	}
	switch spec.Target {
	case cards.SkillTargetEnemySingle,
		cards.SkillTargetEnemyAll,
		cards.SkillTargetEnemySplash,
		cards.SkillTargetEnemyRandom:
	default:
		return errors.New("bad reactive damage target")
	}
	_, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targetRefs := resolvePassiveTargets(owner, enemy, aliveSource, spec, ctx)
	if len(targetRefs) == 0 {
		return nil
	}
	eventTargets := make([]EventTarget, 0, len(targetRefs))
	for _, ref := range targetRefs {
		if ref.Unit == nil {
			continue
		}
		result, err := applyDamageToUnit(
			m,
			ref.OwnerIndex,
			ref.Slot,
			ref.Unit,
			spec.Power,
			aliveSource.InstanceID,
			ctx.SourcePlayerIndex,
			true,
		)
		if err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: ref.Unit.InstanceID,
			TemplateID: ref.Unit.TemplateID,
			Amount:     result.TotalDamage,
			Died:       result.Died,
			NewHP:      result.NewHP,
		})
	}
	if len(eventTargets) == 0 {
		return nil
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      ctx.SourcePlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: aliveSource.InstanceID,
		SourceTemplateID: aliveSource.TemplateID,
		VFXKey:           BuildVFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		SFXKey:           BuildSFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		ImpactVFXKey:     BuildVFXKey(aliveSource.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(aliveSource.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	return nil
}

func HandlePassiveReactiveDebuff(
	m *MatchState,
	source *UnitState,
	owner *PlayerState,
	enemy *PlayerState,
	spec cards.PassiveSpec,
	ctx PassiveContext,
) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if source == nil {
		return errors.New("nil source unit")
	}
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
	}
	if ctx.Trigger != spec.Trigger {
		return nil
	}
	if spec.Kind != cards.PassiveKindReactive {
		return errors.New("bad passive kind for reactive debuff")
	}
	if spec.BuffEffect != "" && spec.BuffEffect != cards.BuffEffectNone {
		return errors.New("reactive debuff cant use buff effect")
	}
	if spec.DebuffEffect == "" || spec.DebuffEffect == cards.DebuffEffectNone {
		return errors.New("bad passive debuff effect")
	}
	if spec.ScaleMode != "" && spec.ScaleMode != cards.PassiveScaleNone {
		return errors.New("scaling passive must use dedicated handler")
	}
	switch spec.Target {
	case cards.SkillTargetEnemySingle,
		cards.SkillTargetEnemyAll,
		cards.SkillTargetEnemySplash,
		cards.SkillTargetEnemyRandom,
		cards.SkillTargetAttackTarget:
	default:
		return errors.New("bad reactive debuff target")
	}
	_, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targetRefs := resolvePassiveTargets(owner, enemy, aliveSource, spec, ctx)
	if len(targetRefs) == 0 {
		return nil
	}
	eventTargets := make([]EventTarget, 0, len(targetRefs))
	for _, ref := range targetRefs {
		if ref.Unit == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.DebuffEffect,
			TurnsLeft:        spec.Duration,
			Value:            spec.Power,
			ExtraValue:       spec.ExtraValue,
			SourceType:       string(SourceUnit),
			EffectLayer:      cards.EffectLayerPassive,
			Polarity:         "debuff",
			SourceInstanceID: aliveSource.InstanceID,
			Dispellable:      false,
			Targeting:        spec.Target,
		}
		if err := AddEffect(ref.Unit, e); err != nil {
			return err
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: ref.Unit.InstanceID,
			TemplateID: ref.Unit.TemplateID,
			Amount:     spec.Power,
			Died:       false,
			NewHP:      ref.Unit.HP,
		})
	}
	if len(eventTargets) == 0 {
		return nil
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      ctx.SourcePlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: aliveSource.InstanceID,
		SourceTemplateID: aliveSource.TemplateID,
		VFXKey:           BuildVFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		SFXKey:           BuildSFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		ImpactVFXKey:     BuildVFXKey(aliveSource.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(aliveSource.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	return nil
}
