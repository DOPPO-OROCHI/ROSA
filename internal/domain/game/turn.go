package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"errors"
	"fmt"
	"time"
)

func StartTurn(m *MatchState, nowUnix int64) error {
	p := m.Players[m.ActivePlayer]
	if p == nil || m.Finished {
		return errors.New("bad player index or match finished")
	}
	p.Turns++
	if p.Turns >= 10 {
		p.Mana = 10
	} else {
		p.Mana = p.Turns
	}
	if p.HeroAttackCooldown > 0 {
		p.HeroAttackCooldown--
	}
	if p.HeroAbilityCooldown > 0 {
		p.HeroAbilityCooldown--
	}
	// _ = processPendingResurections(m, m.ActivePlayer)
	for i := range p.Table {
		u := p.Table[i]
		if u == nil {
			continue
		}
		if u.Cooldown > 0 {
			u.Cooldown--
		}
		if u.Skill.CooldownLeft > 0 {
			u.Skill.CooldownLeft--
		}
	}
	if err := TickerEffects(m, m.ActivePlayer); err != nil {
		return err
	}
	// _ = DispathContinuousPassives(m)
	// _ = DispathPassives(m, m.ActivePlayer, cards.PassiveTriggerTurnStart, PassiveTriggerContext{
	// 	SourceOwnerIdx: m.ActivePlayer,
	// })
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		if u == nil {
			continue
		}
		// _ = triggerCardSkillByTrigger(m, m.ActivePlayer, u, cards.TriggerTurnStart, Action{
		// 	PlayerIndex:    m.ActivePlayer,
		// 	CardInstanceID: u.InstanceID,
		// })
	}
	draw := 1
	if len(p.Deck) < draw {
		draw = len(p.Deck)
	}
	for i := 0; i < draw; i++ {
		card := p.Deck[0]
		p.Deck = p.Deck[1:]
		p.Hand = append(p.Hand, card)
	}
	if p.Turns >= 31 && p.Turns <= 35 {
		p.HeroHP -= 10
	} else if p.Turns >= 36 && p.Turns <= 40 {
		p.HeroHP -= 20
	} else if p.Turns >= 41 {
		p.HeroHP -= 40
	}
	if p.HeroHP <= 0 {
		m.Finished = true
		if m.ActivePlayer == 0 {
			m.Result = MatchWinP2
		} else {
			m.Result = MatchWinP1
		}
		return nil
	}
	if m.TurnTimeSec <= 0 {
		m.TurnTimeSec = 45
	}
	m.TurnStartedAt = nowUnix
	m.TurnDeadline = nowUnix + int64(m.TurnTimeSec)
	m.Phase = PhaseMain
	return nil
}

func EndTurn(m *MatchState) {
	if m.Finished {
		return
	}
	if m.Phase != PhaseMain {
		return
	}
	// _ = DispathPassives(m, m.ActivePlayer, cards.PassiveTriggerTurnEnd, PassiveTriggerContext{
	// 	SourceOwnerIdx: m.ActivePlayer,
	// })
	m.ActivePlayer = 1 - m.ActivePlayer
	if err := StartTurn(m, time.Now().Unix()); err != nil {
		return
	}
}

func PlayBattleCard(m *MatchState,
	playerIndex int,
	cardInctanceID string,
	targetSlot int,
	r BattleTemplateResolver) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if playerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	p := m.Players[playerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	if targetSlot < 0 || targetSlot >= TableSize {
		return errors.New("bad target slot")
	}
	if p.Table[targetSlot] != nil {
		return ErrSlotOccupied
	}
	handIdx := -1
	var c CardsInMatch
	for i := range p.Hand {
		if p.Hand[i].InstanceID == cardInctanceID {
			handIdx = i
			c = p.Hand[i]
			break
		}
	}
	if handIdx == -1 {
		return ErrHandCardNotFound
	}
	tpl, ok := r.GetBattleTemplate(c.TemplateID)
	if !ok {
		return errors.New("unknown battle template: " + c.TemplateID)
	}
	if p.Mana < tpl.ManaCost {
		return ErrNotEnoughMana
	}
	p.Mana -= tpl.ManaCost
	last := len(p.Hand) - 1
	p.Hand[handIdx] = p.Hand[last]
	p.Hand = p.Hand[:last]
	hp, atk := ScaleBattleStats(tpl.HealthPoints, tpl.Attack, c.CardLevel)
	u := &UnitState{
		InstanceID:      c.InstanceID,
		TemplateID:      c.TemplateID,
		GamerCardID:     c.GamerCardID,
		CardLevel:       c.CardLevel,
		HP:              hp,
		MaxHP:           hp,
		Attack:          atk,
		SplashRadius:    tpl.SplashRadius,
		IsTank:          tpl.IsTank,
		CardType:        tpl.CardType,
		BaseCooldown:    tpl.BaseCooldown,
		Cooldown:        0,
		SummonedInTurn:  p.Turns,
		ImageKey:        tpl.ImageKey,
		AssetBaseKey:    tpl.AssetBaseKey,
		HasSkill:        tpl.HasSkill,
		SkillImageKey:   tpl.SkillImageKey,
		Effects:         nil,
		ResurrectedUsed: false,
	}
	if tpl.HasSkill {
		u.Skill = cards.UnitSkillState{
			Code:         tpl.Skill.Code,
			Kind:         tpl.Skill.Kind,
			Target:       tpl.Skill.Target,
			Power:        tpl.Skill.Power,
			BaseCooldown: tpl.Skill.BaseCooldown,
			CooldownLeft: 0,
			Duration:     tpl.Skill.Duration,
			ExtraValue:   tpl.Skill.ExtraValue,
			IgnoreTank:   tpl.Skill.IgnoreTank,
			ApplyCount:   tpl.Skill.HitCount,
		}
	}
	p.Table[targetSlot] = u
	m.Events = append(m.Events, Event{
		Type:             string(EventSummon),
		PlayerIndex:      playerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: u.InstanceID,
		SourceTemplateID: u.TemplateID,
		VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "summon"),
		SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "summon"),
		TargetSlot:       targetSlot,
	})
	// _ = triggerPassiveByTrigger(m, playerIndex, u, cards.PassiveTriggerOnPlay, PassiveTriggerContext{
	// 	TargetInstanceID: u.InstanceID,
	// 	SourceOwnerIdx:   playerIndex,
	// 	TargetOwnerIdx:   playerIndex,
	// 	TargetSlot:       targetSlot,
	// })
	// _ = DispathContinuousPassives(m)
	// _ = triggerCardSkillByTrigger(m, playerIndex, u, cards.TriggerOnPlay, Action{
	// 	PlayerIndex:    playerIndex,
	// 	CardInstanceID: u.InstanceID,
	// 	TargetSlot:     targetSlot,
	// })
	return nil
}

func PlayBuffCard(m *MatchState,
	playerIndex int,
	buffInstanceID string,
	tatgetInstanceID string,
	r BuffTemplateResolver) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if playerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	p := m.Players[playerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	handIdx := -1
	var buffCard CardsInMatch
	for i := range p.Hand {
		if p.Hand[i].InstanceID == buffInstanceID {
			handIdx = i
			buffCard = p.Hand[i]
			break
		}
	}
	if handIdx == -1 {
		return ErrBuffCardNotFound
	}
	tpl, ok := r.GetBuffTemplate(buffCard.TemplateID)
	if !ok {
		return errors.New("unknown buff template: " + buffCard.TemplateID)
	}
	_, target := p.FindSlot(tatgetInstanceID)
	if target == nil {
		return ErrTargetNotFound
	}
	if tpl.OnlyFor != "" && target.CardType != tpl.OnlyFor {
		return ErrWrongTargetType
	}
	if p.Mana < tpl.ManaCost {
		return ErrNotEnoughMana
	}
	p.Mana -= tpl.ManaCost
	e := UnitEffect{
		EffectType: tpl.BuffType,
		TurnsLeft:  tpl.Duration,
		Value:      tpl.BuffValue,
	}
	beforeHP := target.HP
	if err := AddEffect(target, e); err != nil {
		return err
	}
	afterHP := target.HP
	deltaHP := afterHP - beforeHP
	last := len(p.Hand) - 1
	p.Hand[handIdx] = p.Hand[last]
	p.Hand = p.Hand[:last]
	p.Discard = append(p.Discard, buffCard)
	ev := Event{
		Type:                 string(EventBuff),
		PlayerIndex:          playerIndex,
		SourceKind:           string(SourceCard),
		SourceCardTemplateID: buffCard.TemplateID,
		VFXKey:               BuildVFXKey(tpl.AssetBaseKey, "cast"),
		SFXKey:               BuildSFXKey(tpl.AssetBaseKey, "cast"),
		Targets: []EventTarget{
			{
				InstanceID: target.InstanceID,
				TemplateID: target.TemplateID,
				Amount:     tpl.BuffValue,
				NewHP:      afterHP,
			},
		},
	}
	_ = deltaHP
	m.Events = append(m.Events, ev)
	return nil
}

func CardAttack(m *MatchState,
	playerIndex int,
	attackerInstanceID string,
	defenderInstanceID string,
	attackHero bool,
	r BattleTemplateResolver) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if playerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	atkPlayer := m.Players[playerIndex]
	defPlayer := m.Players[1-playerIndex]
	if atkPlayer == nil || defPlayer == nil {
		return errors.New("nil player state")
	}
	atkIdx, atk := atkPlayer.FindSlot(attackerInstanceID)
	if atkIdx == -1 || atk == nil {
		return ErrAttackerNotFound
	}
	if atk.SummonedInTurn == atkPlayer.Turns {
		return ErrAttackerSummoneddThisTurn
	}
	if HasEffect(atk, cards.DebuffEffectStun) {
		return errors.New("attacker is stunned")
	}
	if HasEffect(atk, cards.DebuffEffectDisarm) {
		return errors.New("attacker is disarmed")
	}
	targets := make([]EventTarget, 0, 3)
	if atk.CardType == cards.Healer {
		if attackHero {
			return ErrHealerCannotAttackHero
		}
		ti, tu := atkPlayer.FindSlot(defenderInstanceID)
		if ti == -1 || tu == nil {
			return ErrTargetNotFound
		}
		heal := atk.Attack
		if HasEffect(tu, cards.DebuffEffectNoHeal) {
			return errors.New("target cannot be healed")
		}
		beforeHP := tu.HP
		tu.HP += heal
		if tu.HP > tu.MaxHP {
			tu.HP = tu.MaxHP
		}
		actualHeal := tu.HP - beforeHP
		tpl, ok := r.GetBattleTemplate(atk.TemplateID)
		if !ok {
			return errors.New("unknown battle template: " + atk.TemplateID)
		}
		atk.Cooldown = atk.BaseCooldown
		targets = append(targets, EventTarget{
			InstanceID: tu.InstanceID,
			TemplateID: tu.TemplateID,
			Amount:     actualHeal,
			Died:       false,
			NewHP:      tu.HP,
		})
		m.Events = append(m.Events, Event{
			Type:             string(EventHeal),
			PlayerIndex:      playerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: atk.InstanceID,
			SourceTemplateID: atk.TemplateID,
			VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "attack"),
			SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "attack"),
			Targets:          targets,
		})
		return nil
	}
	if atk.Cooldown > 0 {
		return ErrAttackerOnCooldown
	}
	defenderHasTank := false
	for i := 0; i < TableSize; i++ {
		u := defPlayer.Table[i]
		if u != nil && u.IsTank {
			defenderHasTank = true
			break
		}
	}
	if attackHero {
		if defenderHasTank {
			return ErrCannotHitHeroWhileTanks
		}
		dmg := atk.Attack
		defPlayer.HeroHP -= dmg
		heroID := fmt.Sprintf("hero:p%d", 1-playerIndex)
		targets = append(targets, EventTarget{
			InstanceID: heroID,
			Amount:     dmg,
			Died:       defPlayer.HeroHP <= 0,
			NewHP:      defPlayer.HeroHP,
		})
	} else {
		di, du := defPlayer.FindSlot(defenderInstanceID)
		if di == -1 || du == nil {
			return ErrDefenderNotFound
		}
		if defenderHasTank && !du.IsTank {
			return ErrMustAttackTank
		}
		targetSlots := make([]int, 0, 3)
		targetSlots = append(targetSlots, di)
		if atk.SplashRadius > 0 {
			left, right := di-1, di+1
			if left >= 0 && defPlayer.Table[left] != nil {
				targetSlots = append(targetSlots, left)
			}
			if right < TableSize && defPlayer.Table[right] != nil {
				targetSlots = append(targetSlots, right)
			}
		}
		baseDamage := atk.Attack
		for _, s := range targetSlots {
			u := defPlayer.Table[s]
			if u == nil {
				continue
			}
			dmg := baseDamage
			inst := u.InstanceID
			tplID := u.TemplateID
			if s != di {
				dmg = baseDamage / 2
			}
			res, err := applyDamageToUnit(m, 1-playerIndex, s, u, dmg, atk.InstanceID, playerIndex, true)
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
				atkSlot, aliveAtk := atkPlayer.FindSlot(attackerInstanceID)
				if aliveAtk != nil && atkSlot >= 0 {
					reflectRes, err := applyDamageToUnit(m, playerIndex, atkSlot, aliveAtk, res.ReflectedDamage,
						u.InstanceID, 1-playerIndex, false)
					if err != nil {
						return err
					}
					targets = append(targets, EventTarget{
						InstanceID: aliveAtk.InstanceID,
						TemplateID: aliveAtk.TemplateID,
						Amount:     reflectRes.DamageToHP,
						Died:       reflectRes.Died,
						NewHP:      reflectRes.NewHP,
					})
				}
			}
		}
	}
	tpl, ok := r.GetBattleTemplate(atk.TemplateID)
	if !ok {
		return errors.New("unknown Battle Template: " + atk.TemplateID)
	}
	atk.Cooldown = atk.BaseCooldown
	m.Events = append(m.Events, Event{
		Type:             string(EventAttack),
		PlayerIndex:      playerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: atk.InstanceID,
		SourceTemplateID: atk.TemplateID,
		VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "attack"),
		SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "attack"),
		Targets:          targets,
	})
	if defPlayer.HeroHP <= 0 {
		m.Finished = true
		if playerIndex == 0 {
			m.Result = MatchWinP1
		} else {
			m.Result = MatchWinP2
		}
	}
	return nil
}

func HeroAttack(m *MatchState,
	playerIndex int,
	defenderUnitInstanceID string,
	attackHero bool) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if playerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	atkPlayer := m.Players[playerIndex]
	defPlayer := m.Players[1-playerIndex]
	if atkPlayer == nil || defPlayer == nil {
		return errors.New("nill player state")
	}
	if atkPlayer.HeroAttackCooldown > 0 {
		return ErrHeroOnCooldown
	}
	if atkPlayer.HeroAttackPower <= 0 {
		return ErrHeroAttackIsZero
	}
	defenderHasTank := false
	for i := 0; i < TableSize; i++ {
		u := defPlayer.Table[i]
		if u != nil && u.IsTank {
			defenderHasTank = true
			break
		}
	}
	targets := make([]EventTarget, 0, 3)
	if attackHero {
		if defenderHasTank {
			return ErrCannotHitHeroWhileTanks
		}
		dmg := atkPlayer.HeroAttackPower
		defPlayer.HeroHP -= dmg
		heroID := fmt.Sprintf("hero:p%d", 1-playerIndex)
		targets = append(targets, EventTarget{
			InstanceID: heroID,
			Amount:     dmg,
			Died:       defPlayer.HeroHP <= 0,
			NewHP:      defPlayer.HeroHP,
		})
	} else {
		di, du := defPlayer.FindSlot(defenderUnitInstanceID)
		if di == -1 || du == nil {
			return ErrDefenderNotFound
		}
		if defenderHasTank && !du.IsTank {
			return ErrMustAttackTank
		}
		targetSlots := make([]int, 0, 3)
		targetSlots = append(targetSlots, di)
		if atkPlayer.HeroSplashRadius > 0 {
			left, right := di-1, di+1
			if left >= 0 && defPlayer.Table[left] != nil {
				targetSlots = append(targetSlots, left)
			}
			if right < TableSize && defPlayer.Table[right] != nil {
				targetSlots = append(targetSlots, right)
			}
		}
		baseDamage := atkPlayer.HeroAttackPower
		for _, s := range targetSlots {
			u := defPlayer.Table[s]
			if u == nil {
				continue
			}
			dmg := baseDamage
			if s != di {
				dmg = baseDamage / 2
			}
			inst := u.InstanceID
			tplID := u.TemplateID
			res, err := applyDamageToUnit(m, 1-playerIndex, s, u, dmg, "", playerIndex, true)
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
				heroID := fmt.Sprintf("hero:p%d", playerIndex)
				atkPlayer.HeroHP -= res.ReflectedDamage
				targets = append(targets, EventTarget{
					InstanceID: heroID,
					Amount:     res.ReflectedDamage,
					Died:       atkPlayer.HeroHP <= 0,
					NewHP:      atkPlayer.HeroHP,
				})
			}
		}
	}
	atkPlayer.HeroAttackCooldown = atkPlayer.HeroAttackBaseCooldown
	heroBase := HeroBaseKey(atkPlayer.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroAttack),
		PlayerIndex:    playerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: atkPlayer.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "attack"),
		SFXKey:         BuildSFXKey(heroBase, "attack"),
		Targets:        targets,
	})
	if atkPlayer.HeroHP <= 0 {
		m.Finished = true
		if playerIndex == 0 {
			m.Result = MatchWinP2
		} else {
			m.Result = MatchWinP1
		}
		return nil
	}
	if defPlayer.HeroHP <= 0 {
		m.Finished = true
		if playerIndex == 0 {
			m.Result = MatchWinP1
		} else {
			m.Result = MatchWinP2
		}
		return nil
	}
	return nil
}

func PlayHeroSpell(m *MatchState, a Action, r Resolvers) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if a.PlayerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	p := m.Players[a.PlayerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	if r.HeroAbility == nil {
		return errors.New("hero ability resolver is nil")
	}
	ability, ok := r.HeroAbility(p.HeroCode)
	if !ok || ability == nil {
		return ErrHeroAbilityUnknown
	}
	spec := ability.Spec()
	if p.HeroAbilityCooldown > 0 {
		return ErrHeroAbilityOnCooldown
	}
	if p.Mana < spec.ManaCost {
		return ErrNotEnoughMana
	}
	switch spec.Target {
	case heroes.ENEMY_ANY:
		if !a.AttackHero && a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
	case heroes.OWN_UNIT:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		if _, u := p.FindSlot(a.TargetInstanceID); u == nil {
			return ErrHeroAbilityBadTarget
		}
	case heroes.ENEMY_UNIT:
		if a.AttackHero || a.TargetInstanceID == "" {
			return ErrHeroAbilityBadTarget
		}
		ep := m.Players[1-a.PlayerIndex]
		if ep == nil {
			return errors.New("nil enemy player state")
		}
		if _, u := ep.FindSlot(a.TargetInstanceID); u == nil {
			return ErrHeroAbilityBadTarget
		}
	default:
		return ErrHeroAbilityBadTarget
	}
	targetSlot := findTargetSlotForHeroSpell(m, a, spec)
	snaps, err := buildHeroSpellShanpsBefore(m, a, spec)
	if err != nil {
		return err
	}
	if err := ability.Apply(m, a); err != nil {
		return err
	}
	p.Mana -= spec.ManaCost
	p.HeroAbilityCooldown = spec.CoolDown
	targets := buildHeroSpellTargetsAfter(m, spec, snaps)
	heroBase := HeroBaseKey(p.HeroCode)
	m.Events = append(m.Events, Event{
		Type:           string(EventHeroSpell),
		PlayerIndex:    a.PlayerIndex,
		SourceKind:     string(SourceHero),
		SourceHeroCode: p.HeroCode,
		VFXKey:         BuildVFXKey(heroBase, "spell"),
		SFXKey:         BuildSFXKey(heroBase, "spell"),
		TargetSlot:     targetSlot,
		Targets:        targets,
	})
	return nil
}

// менеджер вызова определенных скилов карт
// func PlayCardSkill(m *MatchState, a Action, r BattleTemplateResolver) error {
// 	if m.Finished {
// 		return ErrMatchFinished
// 	}
// 	if a.PlayerIndex != m.ActivePlayer {
// 		return ErrNotYourTurn
// 	}
// 	if m.Phase != PhaseMain {
// 		return ErrWrongPhase
// 	}
// 	owner := m.Players[a.PlayerIndex]
// 	enemy := m.Players[1-a.PlayerIndex]
// 	if owner == nil || enemy == nil {
// 		return errors.New("nil player state")
// 	}
// 	_, caster := owner.FindSlot(a.CardInstanceID)
// 	if caster == nil || caster.SkillCode == "" {
// 		return ErrCardSkillNotFound
// 	}
// 	if caster.SkillTrigger != cards.TriggerActive {
// 		return ErrCardSkillNotActive
// 	}
// 	if caster.SkillCooldownLeft > 0 {
// 		return ErrCardSkillOnCooldown
// 	}
// 	h, err := getCardSkillHandler(caster.SkillCode)
// 	if err != nil {
// 		return err
// 	}
// 	if err := h(m, a, caster, owner, enemy); err != nil {
// 		return err
// 	}
// 	caster.SkillCooldownLeft = caster.SkillCooldown
// 	return nil
// }

// func DispathContinuousPassives(m *MatchState) error {
// 	if m == nil {
// 		return nil
// 	}
// 	clearContinuousPassiveEffects(m)
// 	for playerIdx := 0; playerIdx < 2; playerIdx++ {
// 		p := m.Players[playerIdx]
// 		if p == nil {
// 			continue
// 		}
// 		for i := 0; i < TableSize; i++ {
// 			u := p.Table[i]
// 			if u == nil {
// 				continue
// 			}
// 			if err := triggerContinuousPassive(m, playerIdx, u); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func DispathPassives(m *MatchState, ownerIdx int, trigger string, ctx PassiveTriggerContext) error {
// 	if m == nil || ownerIdx < 0 || ownerIdx > 1 || trigger == "" {
// 		return nil
// 	}
// 	p := m.Players[ownerIdx]
// 	if p == nil {
// 		return nil
// 	}
// 	for i := 0; i < TableSize; i++ {
// 		u := p.Table[i]
// 		if u == nil {
// 			continue
// 		}
// 		if err := triggerPassiveByTrigger(m, ownerIdx, u, trigger, ctx); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// ХЕЛПЕР СМЕРТИ //
func killUnitAt(m *MatchState, ownerIdx int, slot int, killerInstanceID string, killerOwnerIdx int) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if ownerIdx < 0 || ownerIdx > 1 {
		return errors.New("bad owner index")
	}
	if slot < 0 || slot >= TableSize {
		return errors.New("bad slot target")
	}
	owner := m.Players[ownerIdx]
	if owner == nil {
		return errors.New("nil owner state")
	}
	u := owner.Table[slot]
	if u == nil {
		return nil
	}
	dead := *u
	// _ = triggerPassiveByTrigger(m, ownerIdx, &dead, cards.PassiveTriggerOnDeath, PassiveTriggerContext{
	// 	DeadInstanceID:   dead.InstanceID,
	// 	TargetInstanceID: dead.InstanceID,
	// 	SourceOwnerIdx:   ownerIdx,
	// 	TargetOwnerIdx:   ownerIdx,
	// 	TargetSlot:       slot,
	// })
	// if dead.SkillCode == cards.SkillResurrectNextTurn && dead.SkillTrigger == cards.TriggerOnDeath && !dead.ResurrectedUsed {
	// 	owner.PendingRes = append(owner.PendingRes, PendingResurrected{
	// 		InstanceID: dead.InstanceID,
	// 		DueTurn:    owner.Turns + 1,
	// 	})
	// }
	owner.GraveYard = append(owner.GraveYard, GraveEntry{
		Unit:       dead,
		DiedAtTurn: owner.Turns,
	})
	// enemyIdx := 1 - ownerIdx
	// _ = triggerOnDeathSkill(m, &dead, ownerIdx, killerInstanceID, killerOwnerIdx)
	owner.RemoveAt(slot)
	// _ = DispathPassives(m, ownerIdx, cards.PassiveTriggerOnAllyDead, PassiveTriggerContext{
	// 	DeadInstanceID: dead.InstanceID,
	// 	SourceOwnerIdx: ownerIdx,
	// 	TargetOwnerIdx: ownerIdx,
	// 	TargetSlot:     slot,
	// })
	// _ = DispathPassives(m, enemyIdx, cards.PassiveTriggerOnEnemyDead, PassiveTriggerContext{
	// 	DeadInstanceID: dead.InstanceID,
	// 	SourceOwnerIdx: ownerIdx,
	// 	TargetOwnerIdx: ownerIdx,
	// 	TargetSlot:     slot,
	// })
	// _ = DispathContinuousPassives(m)
	m.Events = append(m.Events, Event{
		Type:             string(EventDeath),
		SourceKind:       string(SourceUnit),
		SourceInstanceID: dead.InstanceID,
		SourceTemplateID: dead.TemplateID,
		TargetSlot:       slot,
	})
	return nil
}

// Хелмер чисто для проверки эффектов стана, сайленса и так далее
func HasEffect(u *UnitState, effectType string) bool {
	if u == nil || effectType == "" {
		return false
	}
	for _, e := range u.Effects {
		if e.EffectType == effectType {
			return true
		}
	}
	return false
}
