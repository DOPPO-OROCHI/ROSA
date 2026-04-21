package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"testing"
)

func testHeroResolvers() Resolvers {
	heroMap := make(map[string]heroes.CharacterTemplate, len(heroes.DefaultHeroTemplate))
	for _, tpl := range heroes.DefaultHeroTemplate {
		heroMap[tpl.CharacterCode] = tpl
	}
	return Resolvers{
		HeroTemplate: &HeroTemplateMapResolver{M: heroMap},
	}
}

func testHeroMatch(ownerHeroCode string) *MatchState {
	return &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{
				PlayerID: 0,
				HeroCode: ownerHeroCode,
				HeroHP:   60,
				Mana:     10,
			},
			{
				PlayerID: 1,
				HeroCode: "enemy_hero",
				HeroHP:   60,
				Mana:     10,
			},
		},
	}
}

func TestPlayHeroSpell_DamageSingle_OK(t *testing.T) {
	m := testHeroMatch("black_cell")
	enemy := &UnitState{
		InstanceID: "enemy_1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     3,
	}
	m.Players[1].Table[0] = enemy

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: enemy.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if enemy.HP != 7 {
		t.Fatalf("expected enemy HP=7, got %d", enemy.HP)
	}
	if m.Players[0].Mana != 8 {
		t.Fatalf("expected owner mana=8, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 1 {
		t.Fatalf("expected hero ability cooldown=1, got %d", m.Players[0].HeroAbilityCooldown)
	}
	if len(m.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(m.Events))
	}
	if m.Events[0].Type != string(EventHeroSpell) {
		t.Fatalf("expected hero spell event, got %s", m.Events[0].Type)
	}
}

func TestPlayHeroSpell_BuffAttackAndHP_OK(t *testing.T) {
	m := testHeroMatch("imperial_commander")
	ally := &UnitState{
		InstanceID: "ally_1",
		TemplateID: "ally_tpl",
		HP:         5,
		MaxHP:      5,
		Attack:     2,
	}
	m.Players[0].Table[0] = ally

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: ally.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if ally.Attack != 5 {
		t.Fatalf("expected ally attack=5, got %d", ally.Attack)
	}
	if ally.HP != 8 {
		t.Fatalf("expected ally HP=8, got %d", ally.HP)
	}
	if ally.MaxHP != 8 {
		t.Fatalf("expected ally MaxHP=8, got %d", ally.MaxHP)
	}
	if len(ally.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(ally.Effects))
	}
	if ally.Effects[0].EffectType != cards.BuffEffectAttackAndHP {
		t.Fatalf("unexpected effect type: %s", ally.Effects[0].EffectType)
	}
	if m.Players[0].Mana != 7 {
		t.Fatalf("expected owner mana=7, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 3 {
		t.Fatalf("expected hero ability cooldown=3, got %d", m.Players[0].HeroAbilityCooldown)
	}
}

func TestPlayHeroSpell_DebuffSplash_OK(t *testing.T) {
	m := testHeroMatch("karn")
	left := &UnitState{InstanceID: "enemy_left", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	center := &UnitState{InstanceID: "enemy_center", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	right := &UnitState{InstanceID: "enemy_right", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	m.Players[1].Table[0] = left
	m.Players[1].Table[1] = center
	m.Players[1].Table[2] = right

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: center.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	for _, u := range []*UnitState{left, center, right} {
		if len(u.Effects) != 1 {
			t.Fatalf("expected 1 effect on %s, got %d", u.InstanceID, len(u.Effects))
		}
		if u.Effects[0].EffectType != cards.DebuffEffectDamageOverTime {
			t.Fatalf("unexpected effect type on %s: %s", u.InstanceID, u.Effects[0].EffectType)
		}
		if u.Effects[0].TurnsLeft != 3 {
			t.Fatalf("expected turns_left=3 on %s, got %d", u.InstanceID, u.Effects[0].TurnsLeft)
		}
		if u.Effects[0].Value != 3 {
			t.Fatalf("expected value=3 on %s, got %d", u.InstanceID, u.Effects[0].Value)
		}
	}
	if m.Players[0].Mana != 8 {
		t.Fatalf("expected owner mana=8, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 3 {
		t.Fatalf("expected hero ability cooldown=3, got %d", m.Players[0].HeroAbilityCooldown)
	}
}

func TestPlayHeroSpell_Hybrid_OK(t *testing.T) {
	m := testHeroMatch("suprime_lider")
	ally := &UnitState{
		InstanceID: "ally_1",
		TemplateID: "ally_tpl",
		HP:         5,
		MaxHP:      5,
		Attack:     2,
	}
	m.Players[0].Table[0] = ally

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: ally.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if ally.Attack != 9 {
		t.Fatalf("expected ally attack=9, got %d", ally.Attack)
	}
	if ally.HP != 1 {
		t.Fatalf("expected ally HP=1, got %d", ally.HP)
	}
	if ally.MaxHP != 1 {
		t.Fatalf("expected ally MaxHP=1, got %d", ally.MaxHP)
	}
	if len(ally.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(ally.Effects))
	}
	if ally.Effects[0].EffectType != cards.BuffEffectAttack {
		t.Fatalf("unexpected effect type: %s", ally.Effects[0].EffectType)
	}
	if m.Players[0].Mana != 8 {
		t.Fatalf("expected owner mana=8, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 3 {
		t.Fatalf("expected hero ability cooldown=3, got %d", m.Players[0].HeroAbilityCooldown)
	}
}

func TestPlayHeroSpell_OnCooldown_ReturnsErrHeroAbilityOnCooldown(t *testing.T) {
	m := testHeroMatch("black_cell")
	m.Players[0].HeroAbilityCooldown = 1
	m.Players[1].Table[0] = &UnitState{
		InstanceID: "enemy_1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
	}

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: "enemy_1",
	}, testHeroResolvers())
	if err != ErrHeroAbilityOnCooldown {
		t.Fatalf("expected ErrHeroAbilityOnCooldown, got %v", err)
	}
}

func TestPlayHeroSpell_DamageHero_KillsMatch_OK(t *testing.T) {
	m := testHeroMatch("black_cell")
	m.Players[1].HeroHP = 3

	err := PlayHeroSpell(m, Action{
		PlayerIndex: 0,
		AttackHero:  true,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if m.Players[1].HeroHP != 0 {
		t.Fatalf("expected enemy hero HP=0, got %d", m.Players[1].HeroHP)
	}
	if !m.Finished {
		t.Fatalf("expected match to be finished")
	}
	if m.Result != MatchWinP1 {
		t.Fatalf("expected MatchWinP1, got %v", m.Result)
	}
}

func TestPlayHeroSpell_NotEnoughMana_ReturnsErrNotEnoughMana(t *testing.T) {
	m := testHeroMatch("black_cell")
	m.Players[0].Mana = 1
	m.Players[1].Table[0] = &UnitState{
		InstanceID: "enemy_1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
	}

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: "enemy_1",
	}, testHeroResolvers())
	if err != ErrNotEnoughMana {
		t.Fatalf("expected ErrNotEnoughMana, got %v", err)
	}
}

func TestPlayHeroSpell_BadTarget_ReturnsErrHeroAbilityBadTarget(t *testing.T) {
	m := testHeroMatch("black_cell")

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: "missing_target",
	}, testHeroResolvers())
	if err != ErrHeroAbilityBadTarget {
		t.Fatalf("expected ErrHeroAbilityBadTarget, got %v", err)
	}
}

func TestPlayHeroSpell_MakeTank_OK(t *testing.T) {
	m := testHeroMatch("the_system")
	ally := &UnitState{
		InstanceID: "ally_1",
		TemplateID: "ally_tpl",
		HP:         6,
		MaxHP:      6,
		Attack:     2,
		IsTank:     false,
	}
	m.Players[0].Table[0] = ally

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: ally.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if !ally.IsTank {
		t.Fatalf("expected ally to become tank")
	}
	if len(ally.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(ally.Effects))
	}
	if ally.Effects[0].EffectType != cards.BuffEffectMakeTank {
		t.Fatalf("unexpected effect type: %s", ally.Effects[0].EffectType)
	}
	if m.Players[0].Mana != 7 {
		t.Fatalf("expected owner mana=7, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 3 {
		t.Fatalf("expected hero ability cooldown=3, got %d", m.Players[0].HeroAbilityCooldown)
	}
}

func TestPlayHeroSpell_LightningMastery_StunsSplash_OK(t *testing.T) {
	m := testHeroMatch("slavic_priest")
	left := &UnitState{InstanceID: "enemy_left", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	center := &UnitState{InstanceID: "enemy_center", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	right := &UnitState{InstanceID: "enemy_right", TemplateID: "enemy_tpl", HP: 10, MaxHP: 10, Attack: 2}
	m.Players[1].Table[0] = left
	m.Players[1].Table[1] = center
	m.Players[1].Table[2] = right

	err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: center.InstanceID,
	}, testHeroResolvers())
	if err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	for _, u := range []*UnitState{left, center, right} {
		if len(u.Effects) != 1 {
			t.Fatalf("expected 1 effect on %s, got %d", u.InstanceID, len(u.Effects))
		}
		if u.Effects[0].EffectType != cards.DebuffEffectStun {
			t.Fatalf("unexpected effect type on %s: %s", u.InstanceID, u.Effects[0].EffectType)
		}
		if u.Effects[0].TurnsLeft != 1 {
			t.Fatalf("expected turns_left=1 on %s, got %d", u.InstanceID, u.Effects[0].TurnsLeft)
		}
	}
	if m.Players[0].Mana != 6 {
		t.Fatalf("expected owner mana=6, got %d", m.Players[0].Mana)
	}
	if m.Players[0].HeroAbilityCooldown != 5 {
		t.Fatalf("expected hero ability cooldown=5, got %d", m.Players[0].HeroAbilityCooldown)
	}
}

func TestTickerEffects_HeroStunExpires_OK(t *testing.T) {
	u := &UnitState{
		InstanceID: "enemy_1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     2,
		Effects: []UnitEffect{
			{
				EffectType: cards.DebuffEffectStun,
				TurnsLeft:  1,
				Value:      0,
				SourceType: string(SourceHero),
			},
		},
	}
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0},
			{PlayerID: 1, Table: [TableSize]*UnitState{u}},
		},
	}

	err := TickerEffects(m, 1)
	if err != nil {
		t.Fatalf("TickerEffects returned error: %v", err)
	}
	if len(u.Effects) != 0 {
		t.Fatalf("expected stun effect to expire, got %d effects", len(u.Effects))
	}
	if m.Players[1].Table[0] == nil {
		t.Fatalf("expected unit to remain on table")
	}
}
