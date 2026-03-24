package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"errors"
	"fmt"
	"time"
)

func StartTurn(m *MatchState, nowUnix int64) {
	p := m.Players[m.ActivePlayer]
	if p == nil || m.Finished {
		return
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
	_ = processPendingResurections(m, m.ActivePlayer)
	for i := range p.Table {
		u := p.Table[i]
		if u == nil {
			continue
		}
		if u.Cooldown > 0 {
			u.Cooldown--
		}
		if u.SkillCooldownLeft > 0 {
			u.SkillCooldownLeft--
		}
	}
	TickerEffects(m, m.ActivePlayer)
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		if u == nil {
			continue
		}
		_ = triggerCardSkillByTrigger(m, m.ActivePlayer, u, cards.TriggerTurnStart, Action{
			PlayerIndex:    m.ActivePlayer,
			CardInstanceID: u.InstanceID,
		})
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
		return
	}
	if m.TurnTimeSec <= 0 {
		m.TurnTimeSec = 45
	}
	m.TurnStartedAt = nowUnix
	m.TurnDeadline = nowUnix + int64(m.TurnTimeSec)
	m.Phase = PhaseMain
}

func EndTurn(m *MatchState) {
	if m.Finished {
		return
	}
	if m.Phase != PhaseMain {
		return
	}
	m.ActivePlayer = 1 - m.ActivePlayer
	StartTurn(m, time.Now().Unix())
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
	if p.Mana < tpl.Manacost {
		return ErrNotEnoughMana
	}
	p.Mana -= tpl.Manacost
	last := len(p.Hand) - 1
	p.Hand[handIdx] = p.Hand[last]
	p.Hand = p.Hand[:last]
	hp, atk := ScaleBattleStats(tpl.HealthPoints, tpl.Attack, c.CardLevel)
	u := &UnitState{
		InstanceID:        c.InstanceID,
		TemplateID:        c.TemplateID,
		GamerCardID:       c.GamerCardID,
		CardLevel:         c.CardLevel,
		HP:                hp,
		MaxHP:             hp,
		Attack:            atk,
		SplashRadius:      tpl.SplashRadius,
		CanBeUpgraded:     tpl.CanBeUpgraded,
		Cooldown:          0,
		IsTank:            tpl.IsTank,
		Effects:           nil,
		CardType:          tpl.CardType,
		SummonedInTurn:    p.Turns,
		SkillImageKey:     tpl.SkillImageKey,
		SkillName:         tpl.SkillName,
		SkillCode:         tpl.SkillCode,
		SkillTrigger:      tpl.SkillTrigger,
		SkillTarget:       tpl.SkillTarget,
		SkillValue:        tpl.SkillValue,
		SkillDuration:     tpl.SkillDuration,
		SkillCooldown:     tpl.SkillCooldown,
		SkillCooldownLeft: 0,
		SkillParamsJSON:   tpl.SkillParams,
		ResurrectedUsed:   false,
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
	_ = triggerCardSkillByTrigger(m, playerIndex, u, cards.TriggerOnPlay, Action{
		PlayerIndex:    playerIndex,
		CardInstanceID: u.InstanceID,
		TargetSlot:     targetSlot,
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
	if !target.CanBeUpgraded {
		return errors.New("card cannot be upgraded")
	}
	if tpl.OnlyFor != "" && tpl.OnlyFor != cards.All && target.CardType != tpl.OnlyFor {
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
	AddEffect(target, e)
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
	ai, a := atkPlayer.FindSlot(attackerInstanceID)
	if ai == -1 || a == nil {
		return ErrAttackerNotFound
	}
	if a.SummonedInTurn == atkPlayer.Turns {
		return ErrAttackerSummoneddThisTurn
	}
	targets := make([]EventTarget, 0, 3)
	if a.CardType == cards.HealerCard {
		if attackHero {
			return ErrHealerCannotAttackHero
		}
		ti, tu := atkPlayer.FindSlot(defenderInstanceID)
		if ti == -1 || tu == nil {
			return ErrTargetNotFound
		}
		heal := a.Attack
		tu.HP += heal
		if tu.HP > tu.MaxHP {
			tu.HP = tu.MaxHP
		}
		tpl, ok := r.GetBattleTemplate(a.TemplateID)
		if !ok {
			return errors.New("unknown battle template: " + a.TemplateID)
		}
		a.Cooldown = tpl.Cooldown
		targets = append(targets, EventTarget{
			InstanceID: tu.InstanceID,
			TemplateID: tu.TemplateID,
			Amount:     heal,
			Died:       false,
			NewHP:      tu.HP,
		})
		m.Events = append(m.Events, Event{
			Type:             string(EventHeal),
			PlayerIndex:      playerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: a.InstanceID,
			SourceTemplateID: a.TemplateID,
			VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "attack"), //бессмыслица, добавил чисто для красоты
			SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "attack"), //ок не бессмыслица, хил с точки зрения движка это атака по своим)))
			Targets:          targets,
		})
		return nil
	}
	if a.Cooldown > 0 {
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
		dmg := a.Attack
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
		if a.SplashRadius > 0 {
			left, right := di-1, di+1
			if left >= 0 && defPlayer.Table[left] != nil {
				targetSlots = append(targetSlots, left)
			}
			if right < TableSize && defPlayer.Table[right] != nil {
				targetSlots = append(targetSlots, right)
			}
		}
		dmg := a.Attack
		for _, s := range targetSlots {
			u := defPlayer.Table[s]
			if u == nil {
				continue
			}
			inst := u.InstanceID
			tplID := u.TemplateID
			u.HP -= dmg
			died := u.HP <= 0
			newHP := u.HP
			if died {
				if err := killUnitAt(m, 1-playerIndex, s); err != nil {
					return err
				}
				newHP = 0
			}
			targets = append(targets, EventTarget{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     dmg,
				Died:       died,
				NewHP:      newHP,
			})
		}
	}
	tpl, ok := r.GetBattleTemplate(a.TemplateID)
	if !ok {
		return errors.New("unknown Battle Template: " + a.TemplateID)
	}
	a.Cooldown = tpl.Cooldown
	m.Events = append(m.Events, Event{
		Type:             string(EventAttack),
		PlayerIndex:      playerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: a.InstanceID,
		SourceTemplateID: a.TemplateID,
		VFXKey:           BuildVFXKey(tpl.AssetBaseKey, "attack"),
		SFXKey:           BuildSFXKey(tpl.AssetBaseKey, "attack"),
		Targets:          targets,
	})
	_, aliveAttacker := atkPlayer.FindSlot(attackerInstanceID)
	if aliveAttacker != nil {
		_ = triggerCardSkillByTrigger(m, playerIndex, aliveAttacker, cards.TriggerOnAttack, Action{
			PlayerIndex:      playerIndex,
			CardInstanceID:   aliveAttacker.InstanceID,
			TargetInstanceID: defenderInstanceID,
			AttackHero:       attackHero,
		})
	}
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
		dmg := atkPlayer.HeroAttackPower
		for _, s := range targetSlots {
			u := defPlayer.Table[s]
			if u == nil {
				continue
			}
			inst := u.InstanceID
			tplID := u.TemplateID
			u.HP -= dmg
			died := u.HP <= 0
			newHP := u.HP
			if died {
				if err := killUnitAt(m, 1-playerIndex, s); err != nil {
					return err
				}
				newHP = 0
			}
			targets = append(targets, EventTarget{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     dmg,
				Died:       died,
				NewHP:      newHP,
			})
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
func PlayCardSkill(m *MatchState, a Action, r BattleTemplateResolver) error {
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
	enemy := m.Players[1-a.PlayerIndex]
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
	}
	_, caster := owner.FindSlot(a.CardInstanceID)
	if caster == nil || caster.SkillCode == "" {
		return ErrCardSkillNotFound
	}
	if caster.SkillTrigger != cards.TriggerActive {
		return ErrCardSkillNotActive
	}
	if caster.SkillCooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	h, err := getCardSkillHandler(caster.SkillCode)
	if err != nil {
		return err
	}
	if err := h(m, a, caster, owner, enemy); err != nil {
		return err
	}
	caster.SkillCooldownLeft = caster.SkillCooldown
	return nil
}

// Хелпер для тех карт, смысл которых заключается в тригере скила при смерти
func triggerOnDeathSkill(m *MatchState, dead *UnitState, deadOwnedIndex int) error {
	if m == nil || dead == nil {
		return nil
	}
	if dead.SkillCode == "" || dead.SkillTrigger != cards.TriggerOnDeath {
		return nil
	}
	owner := m.Players[deadOwnedIndex]
	enemy := m.Players[1-deadOwnedIndex]
	if owner == nil || enemy == nil {
		return nil
	}
	h, err := getCardSkillHandler(dead.SkillCode)
	if err != nil {
		return err
	}
	a := Action{
		PlayerIndex:    deadOwnedIndex,
		CardInstanceID: dead.InstanceID,
	}
	return h(m, a, dead, owner, enemy)
}

// логика под UI, для того, чтобы понять, какие юниты были затронуты способностью
type HeroSellSnap struct {
	playerIndex int
	isHero      bool
	slot        int
	instID      string
	tplID       string
	hpBefore    int
}

func buildHeroSpellShanpsBefore(m *MatchState, a Action, spec heroes.AbilitySpec) ([]HeroSellSnap, error) {
	if m == nil {
		return nil, errors.New("nil match")
	}
	snaps := make([]HeroSellSnap, 0, 3)
	snapUnit := func(pi int, inst string) (HeroSellSnap, bool) {
		p := m.Players[pi]
		if p == nil {
			return HeroSellSnap{}, false
		}
		slot, u := p.FindSlot(inst)
		if u == nil || slot < 0 {
			return HeroSellSnap{}, false
		}
		return HeroSellSnap{
			playerIndex: pi,
			isHero:      false,
			slot:        slot,
			instID:      u.InstanceID,
			tplID:       u.TemplateID,
			hpBefore:    u.HP,
		}, true
	}
	snapHero := func(pi int) HeroSellSnap {
		p := m.Players[pi]
		hp := 0
		if p != nil {
			hp = p.HeroHP
		}
		return HeroSellSnap{
			playerIndex: pi,
			isHero:      true,
			slot:        -1,
			instID:      fmt.Sprintf("hero:p%d", pi),
			tplID:       "",
			hpBefore:    hp,
		}
	}
	switch spec.Code {
	case heroes.ATTACK_ANY:
		defPI := 1 - a.PlayerIndex
		if a.AttackHero {
			snaps = append(snaps, snapHero(defPI))
			return snaps, nil
		}
		s, ok := snapUnit(defPI, a.TargetInstanceID)
		if !ok {
			return nil, ErrDefenderNotFound
		}
		snaps = append(snaps, s)
		return snaps, nil
	case heroes.ATTACK_SPLASH:
		defPI := 1 - a.PlayerIndex
		p := m.Players[defPI]
		if p == nil {
			return nil, errors.New("nil enemy player state")
		}
		centerSlot, center := p.FindSlot(a.TargetInstanceID)
		if center == nil || centerSlot < 0 {
			return nil, ErrDefenderNotFound
		}
		snaps = append(snaps, HeroSellSnap{
			playerIndex: defPI, isHero: false, slot: centerSlot,
			instID: center.InstanceID, tplID: center.TemplateID,
			hpBefore: center.HP,
		})
		left, right := centerSlot-1, centerSlot+1
		if left >= 0 && p.Table[left] != nil {
			u := p.Table[left]
			snaps = append(snaps, HeroSellSnap{
				playerIndex: defPI, isHero: false, slot: left,
				instID: u.InstanceID, tplID: u.TemplateID,
				hpBefore: u.HP,
			})
		}
		if right < TableSize && p.Table[right] != nil {
			u := p.Table[right]
			snaps = append(snaps, HeroSellSnap{
				playerIndex: defPI, isHero: false, slot: right,
				instID: u.InstanceID, tplID: u.TemplateID,
				hpBefore: u.HP,
			})
		}
		return snaps, nil
	case heroes.HEAL_UNIT:
		s, ok := snapUnit(a.PlayerIndex, a.TargetInstanceID)
		if !ok {
			return nil, ErrTargetNotFound
		}
		snaps = append(snaps, s)
		return snaps, nil
	default:
		if a.AttackHero {
			snaps = append(snaps, snapHero(1-a.PlayerIndex))
			return snaps, nil
		}
		if s, ok := snapUnit(1-a.PlayerIndex, a.TargetInstanceID); ok {
			snaps = append(snaps, s)
			return snaps, nil
		}
		if s, ok := snapUnit(a.PlayerIndex, a.TargetInstanceID); ok {
			snaps = append(snaps, s)
			return snaps, nil
		}
		return nil, ErrHeroAbilityBadTarget
	}
}

func buildHeroSpellTargetsAfter(m *MatchState, spec heroes.AbilitySpec, snaps []HeroSellSnap) []EventTarget {
	out := make([]EventTarget, 0, len(snaps))
	for _, s := range snaps {
		p := m.Players[s.playerIndex]
		if s.isHero {
			newHP := 0
			if p != nil {
				newHP = p.HeroHP
			}
			amt := spec.Value
			if spec.Code == heroes.ATTACK_ANY || spec.Code == heroes.ATTACK_SPLASH {
				d := s.hpBefore - newHP
				if d < 0 {
					d = -d
				}
				amt = d
			}
			if spec.Code == heroes.HEAL_UNIT {
				d := newHP - s.hpBefore
				if d < 0 {
					d = -d
				}
				amt = d
			}
			out = append(out, EventTarget{
				InstanceID: s.instID,
				Amount:     amt,
				Died:       newHP <= 0,
				NewHP:      newHP,
			})
			continue
		}
		newHP := 0
		died := true
		tplID := s.tplID
		if p != nil {
			_, u := p.FindSlot(s.instID)
			if u != nil {
				newHP = u.HP
				died = newHP <= 0
				tplID = u.TemplateID
			}
		}
		amt := spec.Value
		if spec.Code == heroes.ATTACK_ANY || spec.Code == heroes.ATTACK_SPLASH {
			d := s.hpBefore - newHP
			if d < 0 {
				d = -d
			}
			amt = d
		}
		if spec.Code == heroes.HEAL_UNIT {
			d := newHP - s.hpBefore
			if d < 0 {
				d = -d
			}
			amt = d
		}
		if !died && newHP > 0 {
			died = false
		}
		out = append(out, EventTarget{
			InstanceID: s.instID,
			TemplateID: tplID,
			Amount:     amt,
			Died:       died,
			NewHP:      newHP,
		})
	}
	return out
}

func ScaleBattleStats(baseHP, baseAttack, level int) (int, int) {
	if level < 1 {
		level = 1
	}
	if level > MaxCardLevel {
		level = MaxCardLevel
	}
	bonus := level - 1
	return baseHP + bonus, baseAttack + bonus
}

func ScaleBuffStats(baseValue, level int) int {
	if level < 1 {
		level = 1
	}
	if level > MaxCardLevel {
		level = MaxCardLevel
	}
	bonus := level - 1
	return baseValue + bonus
}

func findTargetSlotForHeroSpell(m *MatchState, a Action, spec heroes.AbilitySpec) int {
	if a.AttackHero || a.TargetInstanceID == "" {
		return -1
	}
	switch spec.Target {
	case heroes.OWN_UNIT:
		p := m.Players[a.PlayerIndex]
		if p == nil {
			return -1
		}
		slot, _ := p.FindSlot(a.TargetInstanceID)
		return slot
	case heroes.ENEMY_UNIT, heroes.ENEMY_ANY:
		ep := m.Players[1-a.PlayerIndex]
		if ep == nil {
			return -1
		}
		slot, _ := ep.FindSlot(a.TargetInstanceID)
		return slot
	default:
		return -1
	}
}

// ХЕЛПЕР СМЕРТИ
func killUnitAt(m *MatchState, ownerIdx int, slot int) error {
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
	if dead.SkillCode == cards.SkillResurrectNextTurn && dead.SkillTrigger == cards.TriggerOnDeath && !dead.ResurrectedUsed {
		owner.PendingRes = append(owner.PendingRes, PendingResurrected{
			InstanceID: dead.InstanceID,
			DueTurn:    owner.Turns + 1,
		})
	}
	owner.GraveYard = append(owner.GraveYard, GraveEntry{
		Unit:       dead,
		DiedAtTurn: owner.Turns,
	})
	_ = triggerOnDeathSkill(m, &dead, ownerIdx)
	owner.RemoveAt(slot)
	m.Events = append(m.Events, Event{
		Type:             string(EventDeath),
		SourceKind:       string(SourceUnit),
		SourceInstanceID: dead.InstanceID,
		SourceTemplateID: dead.TemplateID,
		TargetSlot:       slot,
	})
	return nil
}

// ФУНККЦИЯ ОБРАБОТКИ СМЕРТИ (ПАССИВКА, ПОДНИМАЕТ КАРТЫ ИЗ МОГИЛЫ НА СЛЕДУЮЩИЙ ХОД)
func processPendingResurections(m *MatchState, ownerIdx int) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if ownerIdx < 0 || ownerIdx > 1 {
		return errors.New("bad owner index")
	}
	owner := m.Players[ownerIdx]
	if owner == nil {
		return errors.New("bad owner")
	}
	nextPending := make([]PendingResurrected, 0, len(owner.PendingRes))
	for i := range owner.PendingRes {
		pr := owner.PendingRes[i]
		if pr.DueTurn > owner.Turns {
			nextPending = append(nextPending, pr)
			continue
		}
		graveIdx := -1
		for j := range owner.GraveYard {
			if owner.GraveYard[j].Unit.InstanceID == pr.InstanceID {
				graveIdx = j
				break
			}
		}
		if graveIdx < 0 {
			continue
		}
		slot := firstFreeSlot(owner)
		if slot < 0 {
			nextPending = append(nextPending, pr)
			continue
		}
		revived := owner.GraveYard[graveIdx].Unit
		revived.HP /= 2
		revived.ResurrectedUsed = true
		revived.SummonedInTurn = owner.Turns
		owner.Table[slot] = &revived
		last := len(owner.GraveYard) - 1
		owner.GraveYard[graveIdx] = owner.GraveYard[last]
		owner.GraveYard = owner.GraveYard[:last]
		m.Events = append(m.Events, Event{
			Type:             string(EventResurrect),
			PlayerIndex:      ownerIdx,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: revived.InstanceID,
			SourceTemplateID: revived.TemplateID,
			TargetSlot:       slot,
			Targets: []EventTarget{
				{
					InstanceID: revived.InstanceID,
					TemplateID: revived.TemplateID,
					Died:       false,
					NewHP:      revived.HP,
				},
			},
		})
	}
	owner.PendingRes = nextPending
	return nil
}

func triggerCardSkillByTrigger(m *MatchState, ownerIdx int, caster *UnitState, trigger string, a Action) error {
	if m == nil {
		return nil
	}
	if caster == nil {
		return nil
	}
	if caster.SkillCode == "" || caster.SkillTrigger != trigger {
		return nil
	}
	owner := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	if owner == nil || enemy == nil {
		return nil
	}
	if a.PlayerIndex != ownerIdx {
		a.PlayerIndex = ownerIdx
	}
	if a.CardInstanceID == "" {
		a.CardInstanceID = caster.InstanceID
	}
	h, err := getCardSkillHandler(caster.SkillCode)
	if err != nil {
		return nil
	}
	return h(m, a, caster, owner, enemy)
}
