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
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
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
	eventTargets := make([]EventTarget, 0, len(targets))
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
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.Unit.InstanceID,
			TemplateID: target.Unit.TemplateID,
			Amount:     spec.Power,
			Died:       false,
			NewHP:      target.Unit.HP,
		})
	}
	if len(eventTargets) == 0 {
		return nil
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventPassive),
		EffectKind:       spec.EffectKind,
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
		Type:             string(EventPassive),
		EffectKind:       spec.EffectKind,
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
	passiveSource := resolvePassiveSourceForTrigger(owner, source, spec)
	if passiveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targetRefs := resolvePassiveTargets(owner, enemy, passiveSource, spec, ctx)
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
			SourceInstanceID: passiveSource.InstanceID,
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
		Type:             string(EventPassive),
		EffectKind:       spec.EffectKind,
		PlayerIndex:      ctx.SourcePlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: passiveSource.InstanceID,
		SourceTemplateID: passiveSource.TemplateID,
		VFXKey:           BuildVFXKey(passiveSource.AssetBaseKey, string(EventPassive)),
		SFXKey:           BuildSFXKey(passiveSource.AssetBaseKey, string(EventPassive)),
		ImpactVFXKey:     BuildVFXKey(passiveSource.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(passiveSource.AssetBaseKey, "impact"),
		Targets:          eventTargets,
	})
	return nil
}

func HandlePassiveScalingAuraBuff(m *MatchState,
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
	if spec.Kind != cards.PassiveKindAura {
		return errors.New("bad passive kind for scaling aura buff")
	}
	if spec.Trigger != cards.PassiveTriggerWhileAlive {
		return errors.New("bad passive trigger for scaling aura buff")
	}
	if spec.BuffEffect == "" || spec.BuffEffect == cards.BuffEffectNone {
		return errors.New("bad passive buff effect")
	}
	if spec.DebuffEffect != "" && spec.DebuffEffect != cards.DebuffEffectNone {
		return errors.New("scaling aura buff cant use debuff effect")
	}
	if spec.ScaleMode == "" || spec.ScaleMode == cards.PassiveScaleNone {
		return errors.New("bad passive scale mode")
	}
	clearPassiveAuraEffect(owner, source.InstanceID, spec)
	_, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	scale := resolvePassiveScale(owner, enemy, spec)
	if scale <= 0 {
		return nil
	}
	value := spec.Power * scale
	extraValue := spec.ExtraValue
	if extraValue > 0 {
		extraValue *= scale
	}
	targets := resolvePassiveTargets(owner, enemy, aliveSource, spec, ctx)
	if len(targets) == 0 {
		return nil
	}
	for _, ref := range targets {
		if ref.Unit == nil {
			continue
		}
		e := UnitEffect{
			EffectType:       spec.BuffEffect,
			TurnsLeft:        0,
			Value:            value,
			ExtraValue:       extraValue,
			SourceType:       string(SourceUnit),
			EffectLayer:      cards.EffectLayerPassive,
			Polarity:         "buff",
			SourceInstanceID: aliveSource.InstanceID,
			Dispellable:      false,
			Targeting:        spec.Target,
		}
		if err := AddEffect(ref.Unit, e); err != nil {
			return err
		}
	}
	return nil
}

func HandlePassiveReactiveHeal(
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
		return errors.New("bad passive kind for reactive heal")
	}
	if spec.EffectKind != cards.PassiveEffectHeal {
		return errors.New("bad passive effect kind for heal")
	}
	if spec.BuffEffect != "" && spec.BuffEffect != cards.BuffEffectNone {
		return errors.New("reactive heal cant use buff effect")
	}
	if spec.DebuffEffect != "" && spec.DebuffEffect != cards.DebuffEffectNone {
		return errors.New("reactive heal cant use debuff effect")
	}
	if spec.ScaleMode != "" && spec.ScaleMode != cards.PassiveScaleNone {
		return errors.New("scaling passive must use dedicated handler")
	}
	if spec.Power <= 0 {
		return errors.New("bad passive heal power")
	}
	switch spec.Target {
	case cards.SkillTargetSelf,
		cards.SkillTargetAllyAll,
		cards.SkillTargetAllyAdjacent,
		cards.SkillTargetSelfAndAdjacent,
		cards.SkillTargetAllyLowestHP:
	default:
		return errors.New("bad reactive heal target")
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
	eventTargets := make([]EventTarget, 0, len(targets))
	for _, ref := range targets {
		target := ref.Unit
		if target == nil {
			continue
		}
		if target.HP >= target.MaxHP {
			continue
		}
		before := target.HP
		target.HP += spec.Power
		if target.HP > target.MaxHP {
			target.HP = target.MaxHP
		}
		healed := target.HP - before
		if healed <= 0 {
			continue
		}
		eventTargets = append(eventTargets, EventTarget{
			InstanceID: target.InstanceID,
			TemplateID: target.TemplateID,
			Amount:     healed,
			Died:       false,
			NewHP:      target.HP,
		})
	}
	if len(eventTargets) == 0 {
		return nil
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventPassive),
		EffectKind:       spec.EffectKind,
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

func HandlePassiveCounterattack(
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
		return errors.New("bad passive kind for counterattack")
	}
	if spec.Trigger != cards.PassiveTriggerOnDamaged {
		return errors.New("bad passive trigger for counterattack")
	}
	if spec.Target != cards.SkillTargetAttackTarget {
		return errors.New("bad passive target for counterattack")
	}
	if spec.BuffEffect != cards.BuffEffectCounterattack {
		return errors.New("bad passive buff effect for counterattack")
	}
	if ctx.DamagedByInstanceID == "" {
		return nil
	}
	souceSlot, aliveSource := owner.FindSlot(source.InstanceID)
	if aliveSource == nil || souceSlot < 0 {
		return nil
	}
	if aliveSource.Attack <= 0 {
		return nil
	}
	if !passiveConditionSatisfied(owner, enemy, spec) {
		return nil
	}
	targetSlot, target := enemy.FindSlot(ctx.DamagedByInstanceID)
	if target == nil || targetSlot < 0 {
		return nil
	}
	result, err := applyDamageToUnit(
		m,
		1-ctx.SourcePlayerIndex,
		targetSlot,
		target,
		aliveSource.Attack,
		aliveSource.InstanceID,
		ctx.SourcePlayerIndex,
		false,
	)
	if err != nil {
		return err
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventPassive),
		EffectKind:       spec.EffectKind,
		PlayerIndex:      ctx.SourcePlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: aliveSource.InstanceID,
		SourceTemplateID: aliveSource.TemplateID,
		VFXKey:           BuildVFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		SFXKey:           BuildSFXKey(aliveSource.AssetBaseKey, string(EventPassive)),
		ImpactVFXKey:     BuildVFXKey(aliveSource.AssetBaseKey, "impact"),
		ImpactSFXKey:     BuildSFXKey(aliveSource.AssetBaseKey, "impact"),
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     result.TotalDamage,
				Died:       result.Died,
				NewHP:      result.NewHP,
			},
		},
	})
	return nil
}
