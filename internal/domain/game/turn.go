package game

import (
	"TheWar/internal/domain/cards"
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
	for i := range p.Table {
		u := p.Table[i]
		if u == nil {
			continue
		}
		u.AttacksThisTurn = 0
		if u.Cooldown > 0 {
			u.Cooldown--
		}
		if u.Skill.CooldownLeft > 0 {
			u.Skill.CooldownLeft--
		}
	}
	if err := DispatchPassives(m, PassiveContext{
		Trigger:          cards.PassiveTriggerTurnStart,
		ActorPlayerIndex: m.ActivePlayer,
	}); err != nil {
		return err
	}
	if err := TickerEffects(m, m.ActivePlayer); err != nil {
		return err
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

func EndTurn(m *MatchState) error {
	if m.Finished {
		return nil
	}
	if m.Phase != PhaseMain {
		return nil
	}
	endingPlayer := m.ActivePlayer
	finalizeOverdriveCooldowns(m.Players[endingPlayer])
	if err := DispatchPassives(m, PassiveContext{
		Trigger:          cards.PassiveTriggerTurnEnd,
		ActorPlayerIndex: endingPlayer,
	}); err != nil {
		return err
	}
	m.ActivePlayer = 1 - m.ActivePlayer
	if err := StartTurn(m, time.Now().Unix()); err != nil {
		return err
	}
	return nil
}

func finalizeOverdriveCooldowns(p *PlayerState) {
	if p == nil {
		return
	}
	for _, u := range p.Table {
		if u == nil {
			continue
		}
		if u.AttacksThisTurn > 0 && u.Cooldown == 0 {
			u.Cooldown = u.BaseCooldown
		}
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
		AttacksThisTurn: 0,
		SummonedInTurn:  p.Turns,
		ImageKey:        tpl.ImageKey,
		AssetBaseKey:    tpl.AssetBaseKey,
		HasSkill:        tpl.HasSkill,
		Passive:         tpl.Passive,
		SkillImageKey:   tpl.SkillImageKey,
		Effects:         nil,
		ResurrectedUsed: false,
	}
	if tpl.HasSkill {
		u.Skill = cards.UnitSkillState{
			Name:         tpl.Skill.Name,
			Code:         tpl.Skill.Code,
			Kind:         tpl.Skill.Kind,
			Target:       tpl.Skill.Target,
			Power:        tpl.Skill.Power,
			BaseCooldown: tpl.Skill.BaseCooldown,
			CooldownLeft: 0,
			Duration:     tpl.Skill.Duration,
			ExtraValue:   tpl.Skill.ExtraValue,
			BuffEffect:   tpl.Skill.BuffEffect,
			DebuffEffect: tpl.Skill.DeBuffEffect,
			CleanseMode:  tpl.Skill.CleanseMode,
			IgnoreTank:   tpl.Skill.IgnoreTank,
			ApplyCount:   tpl.Skill.HitCount,
		}
	}
	p.Table[targetSlot] = u
	if err := DispatchPassives(m, PassiveContext{
		Trigger:             cards.PassiveTriggerOnEnter,
		ActorPlayerIndex:    playerIndex,
		EventUnitInstanceID: u.InstanceID,
		PlayedCardCode:      u.TemplateID,
		PlayedCardType:      u.CardType,
		PlayedCardIsTank:    u.IsTank,
	}); err != nil {
		return err
	}
	if err := DispatchPassives(m, PassiveContext{
		Trigger:             cards.PassiveTriggerOnAllyPlay,
		ActorPlayerIndex:    playerIndex,
		EventUnitInstanceID: u.InstanceID,
		PlayedCardCode:      u.TemplateID,
		PlayedCardType:      u.CardType,
		PlayedCardIsTank:    u.IsTank,
	}); err != nil {
		return err
	}
	if err := RefreshPassiveAuras(m); err != nil {
		return err
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventSummon),
		PlayerIndex:      playerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: u.InstanceID,
		SourceTemplateID: u.TemplateID,
		VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "summon"),
		SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "summon"),
		ImpactVFXKey:     BuildVFXKey(tpl.AssetBaseKey, "summon"),
		ImpactSFXKey:     BuildSFXKey(tpl.AssetBaseKey, "summon"),
		TargetSlot:       targetSlot,
	})
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
		ImpactVFXKey:         BuildVFXKey(tpl.AssetBaseKey, "cast"),
		ImpactSFXKey:         BuildSFXKey(tpl.AssetBaseKey, "cast"),
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
	if atk.AttacksThisTurn >= maxAttacksPerTurn(atk) {
		return ErrAttackerOnCooldown
	}
	targets := make([]EventTarget, 0, 3)
	isHealAttack := false
	attackRepeats := 1
	for rep := 0; rep < attackRepeats; rep++ {
		if atk == nil || atk.HP <= 0 {
			break
		}
		if atk.Cooldown > 0 {
			return ErrAttackerOnCooldown
		}
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
			targets = append(targets, EventTarget{
				InstanceID: tu.InstanceID,
				TemplateID: tu.TemplateID,
				Amount:     actualHeal,
				Died:       false,
				NewHP:      tu.HP,
			})
			isHealAttack = true
			continue
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
			beforeHP := defPlayer.HeroHP
			dmg := atk.Attack
			defPlayer.HeroHP -= dmg
			if defPlayer.HeroHP < 0 {
				defPlayer.HeroHP = 0
			}
			dealtToHeroHP := beforeHP - defPlayer.HeroHP
			healed := applyVampiricOnHit(atk, dealtToHeroHP)
			if healed > 0 {
				targets = append(targets, EventTarget{
					InstanceID: atk.InstanceID,
					TemplateID: atk.TemplateID,
					Amount:     healed,
					Died:       false,
					NewHP:      atk.HP,
				})
			}
			heroID := fmt.Sprintf("hero:p%d", 1-playerIndex)
			targets = append(targets, EventTarget{
				InstanceID: heroID,
				Amount:     dealtToHeroHP,
				Died:       defPlayer.HeroHP <= 0,
				NewHP:      defPlayer.HeroHP,
			})
		} else {
			di, du := defPlayer.FindSlot(defenderInstanceID)
			if di == -1 || du == nil {
				if rep > 0 {
					break
				}
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
				if atk.HP > 0 {
					healed := applyVampiricOnHit(atk, res.DamageToHP)
					if healed > 0 {
						targets = append(targets, EventTarget{
							InstanceID: atk.InstanceID,
							TemplateID: atk.TemplateID,
							Amount:     healed,
							Died:       false,
							NewHP:      atk.HP,
						})
					}
				}
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
				if u.HP > 0 && atk.HP > 0 {
					counterRes, err := applyCounterattack(m, 1-playerIndex, u, playerIndex, atkIdx, atk)
					if err != nil {
						return err
					}
					if counterRes.TotalDamage > 0 {
						targets = append(targets, EventTarget{
							InstanceID: atk.InstanceID,
							TemplateID: atk.TemplateID,
							Amount:     counterRes.DamageToHP,
							Died:       counterRes.Died,
							NewHP:      counterRes.NewHP,
						})
					}
				}
			}
		}
	}
	if atk.HP > 0 && !isHealAttack {
		if err := DispatchPassives(m, PassiveContext{
			Trigger:             cards.PassiveTriggerOnAttack,
			ActorPlayerIndex:    playerIndex,
			EventUnitInstanceID: atk.InstanceID,
			AttackTargetID:      defenderInstanceID,
		}); err != nil {
			return err
		}
	}
	tpl, ok := r.GetBattleTemplate(atk.TemplateID)
	if !ok {
		return errors.New("unknown Battle Template: " + atk.TemplateID)
	}
	if atk.HP > 0 && !isHealAttack {
		chainTargets, err := applyChainAttack(m, playerIndex, atk)
		if err != nil {
			return err
		}
		if len(chainTargets) > 0 {
			targets = append(targets, chainTargets...)
		}
	}
	if atk.HP > 0 && !isHealAttack {
		applyBonusAfterAttack(atk)
	}
	eventType := string(EventAttack)
	if isHealAttack {
		eventType = string(EventHeal)
	}
	if atk.HP > 0 {
		atk.AttacksThisTurn++
		if hasOverdrive(atk) && atk.AttacksThisTurn < maxAttacksPerTurn(atk) {
			atk.Cooldown = 0
		} else {
			atk.Cooldown = atk.BaseCooldown
		}
	}
	m.Events = append(m.Events, Event{
		Type:             eventType,
		PlayerIndex:      playerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: atk.InstanceID,
		SourceTemplateID: atk.TemplateID,
		VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "attack"),
		SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "attack"),
		ImpactVFXKey:     BuildVFXKey(tpl.AssetBaseKey, "attack"),
		ImpactSFXKey:     BuildSFXKey(tpl.AssetBaseKey, "attack"),
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
		ImpactVFXKey:   BuildVFXKey(heroBase, "attack"),
		ImpactSFXKey:   BuildSFXKey(heroBase, "attack"),
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
	if m == nil {
		return errors.New("nil match state")
	}
	if m.Finished {
		return ErrMatchFinished
	}
	if a.PlayerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	if r.HeroTemplate == nil {
		return errors.New("hero template resolver is nil")
	}
	tpl, ok := r.HeroTemplate.GetHeroTemplate(owner.HeroCode)
	if !ok {
		return ErrUnknownHeroTemplate
	}
	spec := tpl.Ability
	if spec.Code == "" {
		return ErrHeroAbilityUnknown
	}
	h, ok := HeroAbilityHandlers[spec.Code]
	if !ok {
		return ErrHeroAbilityUnknown
	}
	if err := h(m, a, owner, spec); err != nil {
		return err
	}
	if err := DispatchPassives(m, PassiveContext{
		Trigger:          cards.PassiveTriggerOnEnemyHeroSkill,
		ActorPlayerIndex: a.PlayerIndex,
		PlayedSkillCode:  spec.Code,
		HeroSkillUsed:    true,
	}); err != nil {
		return err
	}
	return nil
}

func PlayCardSkill(m *MatchState, a Action) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if m.Finished {
		return ErrMatchFinished
	}
	if a.PlayerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	if m.Phase != PhaseMain {
		return ErrWrongPhase
	}
	owner := m.Players[a.PlayerIndex]
	if owner == nil {
		return errors.New("nil owner state")
	}
	_, caster := owner.FindSlot(a.CardInstanceID)
	if caster == nil || caster.Skill.Code == "" {
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
	h, ok := SkillHandlers[caster.Skill.Code]
	if !ok {
		return ErrCardSkillUnsupported
	}
	repeates := 1
	if hasMulticast(caster) {
		repeates = 2
	}
	for i := 0; i < repeates; i++ {
		if caster.HP <= 0 {
			break
		}
		if i > 0 {
			caster.Skill.CooldownLeft = 0
		}
		if err := h(m, a, caster); err != nil {
			return err
		}
	}
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	if err := DispatchPassives(m, PassiveContext{
		Trigger:             cards.PassiveTriggerOnAllySkill,
		ActorPlayerIndex:    a.PlayerIndex,
		EventUnitInstanceID: caster.InstanceID,
		PlayedSkillCode:     caster.Skill.Code,
	}); err != nil {
		return err
	}
	return nil
}

// ДИСПЕТЧЕР ПАССИВОК
func DispatchPassives(m *MatchState, ctx PassiveContext) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if ctx.Trigger == "" {
		return nil
	}
	for playerIndex := 0; playerIndex < len(m.Players); playerIndex++ {
		owner := m.Players[playerIndex]
		enemy := m.Players[1-playerIndex]
		if owner == nil || enemy == nil {
			continue
		}
		if (ctx.Trigger == cards.PassiveTriggerTurnStart || ctx.Trigger == cards.PassiveTriggerTurnEnd) &&
			playerIndex != ctx.ActorPlayerIndex {
			continue
		}
		sourceTrigger := resolvePassiveTriggerForSource(ctx.Trigger, playerIndex, ctx.ActorPlayerIndex)
		for slot := 0; slot < TableSize; slot++ {
			source := owner.Table[slot]
			if source == nil {
				continue
			}
			spec := source.Passive
			if spec.Code == "" {
				continue
			}
			if spec.Trigger != sourceTrigger {
				continue
			}
			if !passiveSourceMatchesEvent(source, sourceTrigger, ctx) {
				continue
			}
			if !passiveEventMatches(spec, ctx) {
				continue
			}
			handler, ok := getCardPassiveHandler(spec.Code)
			if !ok {
				return errors.New("passive handler not found: " + spec.Code)
			}
			localCtx := ctx
			localCtx.Trigger = sourceTrigger
			localCtx.SourcePlayerIndex = playerIndex

			if err := handler(m, source, owner, enemy, spec, localCtx); err != nil {
				return err
			}
		}
	}
	return nil
}

func RefreshPassiveAuras(m *MatchState) error {
	return DispatchPassives(m, PassiveContext{
		Trigger: cards.PassiveTriggerWhileAlive,
	})
}

// ХЕЛПЕР СМЕРТИ
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
	if err := triggerEnemyOnDeathExplosion(m, ownerIdx, &dead, slot); err != nil {
		return err
	}
	if err := triggerAllyOnDeathMassHeal(m, ownerIdx, &dead, slot); err != nil {
		return err
	}
	owner.GraveYard = append(owner.GraveYard, GraveEntry{
		Unit:       dead,
		DiedAtTurn: owner.Turns,
	})
	if err := DispatchPassives(m, PassiveContext{
		Trigger:             cards.PassiveTriggerOnLeave,
		ActorPlayerIndex:    ownerIdx,
		EventUnitInstanceID: u.InstanceID,
		PlayedCardCode:      u.TemplateID,
		PlayedCardType:      u.CardType,
		PlayedCardIsTank:    u.IsTank,
	}); err != nil {
		return err
	}
	if err := DispatchPassives(m, PassiveContext{
		Trigger:             cards.PassiveTriggerOnAllyDeath,
		ActorPlayerIndex:    ownerIdx,
		EventUnitInstanceID: u.InstanceID,
		PlayedCardCode:      u.TemplateID,
		PlayedCardType:      u.CardType,
		PlayedCardIsTank:    u.IsTank,
	}); err != nil {
		return err
	}
	needRefreshAuras := dead.Passive.Kind == cards.PassiveKindAura && dead.Passive.Trigger == cards.PassiveTriggerWhileAlive
	if needRefreshAuras {
		clearPassiveAuraEffect(owner, dead.InstanceID, dead.Passive)
	}
	owner.RemoveAt(slot)
	if needRefreshAuras {
		if err := RefreshPassiveAuras(m); err != nil {
			return err
		}
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventDeath),
		SourceKind:       string(SourceUnit),
		SourceInstanceID: dead.InstanceID,
		SourceTemplateID: dead.TemplateID,
		SFXKey:           BuildSFXKey(dead.AssetBaseKey, "death"),
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
