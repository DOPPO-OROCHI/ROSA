package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"testing"
)

func testMatch() *MatchState {
	return &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, HeroCode: "hero0", HeroHP: 50, Mana: 10, Turns: 1},
			{PlayerID: 1, HeroCode: "hero1", HeroHP: 50, Mana: 10, Turns: 1},
		},
	}
}

func testUnit(id string, hp int) *UnitState {
	return &UnitState{
		InstanceID:   id,
		TemplateID:   id + "_tpl",
		HP:           hp,
		MaxHP:        hp,
		Attack:       2,
		BaseCooldown: 1,
		CardType:     cards.Human,
		AssetBaseKey: id,
	}
}

func TestPassive_OnEnterAndOnEnemyPlay_FromPlayBattleCard(t *testing.T) {
	m := testMatch()
	m.Players[0].Hand = []CardsInMatch{{
		InstanceID:  "enter_card",
		TemplateID:  "ancient_tpl",
		CardLevel:   1,
		GamerCardID: 10,
	}}
	reactor := testUnit("enemy_reactor", 12)
	reactor.Passive = cards.PassiveSpec{
		Code:        "humanity_revenge",
		Kind:        cards.PassiveKindReactive,
		Trigger:     cards.PassiveTriggerOnEnemyPlay,
		EffectKind:  cards.PassiveEffectDamage,
		Target:      cards.SkillTargetEnemyRandom,
		Power:       3,
		EventFilter: cards.PassiveEventFilterCardPlayed,
		IgnoreTank:  true,
	}
	m.Players[1].Table[0] = reactor

	r := BattleMapResolver{M: map[string]BattleTemplate{
		"ancient_tpl": {
			TemplateID:    "ancient_tpl",
			HealthPoints:  20,
			Attack:        1,
			BaseCooldown:  1,
			ManaCost:      1,
			CardType:      cards.Human,
			AssetBaseKey:  "ancient",
			SkillImageKey: "ancient_skill",
			Passive: cards.PassiveSpec{
				Code:       "ancient_enemy",
				Kind:       cards.PassiveKindReactive,
				Trigger:    cards.PassiveTriggerOnEnter,
				EffectKind: cards.PassiveEffectDamage,
				Target:     cards.SkillTargetEnemyAll,
				Power:      4,
				IgnoreTank: true,
			},
		},
	}}

	if err := PlayBattleCard(m, 0, "enter_card", 0, r); err != nil {
		t.Fatalf("PlayBattleCard returned error: %v", err)
	}
	summoned := m.Players[0].Table[0]
	if summoned == nil {
		t.Fatalf("expected summoned unit on table")
	}
	if reactor.HP != 8 {
		t.Fatalf("expected OnEnter to damage enemy reactor to 8 HP, got %d", reactor.HP)
	}
	if summoned.HP != 17 {
		t.Fatalf("expected OnEnemyPlay passive to damage summoned unit to 17 HP, got %d", summoned.HP)
	}
}

func TestPassive_OnDamaged_FromApplyDamageToUnit(t *testing.T) {
	m := testMatch()
	target := testUnit("damaged", 10)
	target.Attack = 2
	target.Passive = cards.PassiveSpec{
		Code:       "to_the_last_drop",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerOnDamaged,
		EffectKind: cards.PassiveEffectBuff,
		Target:     cards.SkillTargetSelf,
		Power:      1,
		BuffEffect: cards.BuffEffectAttack,
		IgnoreTank: true,
	}
	m.Players[0].Table[0] = target
	m.Players[1].Table[0] = testUnit("attacker", 10)

	res, err := applyDamageToUnit(m, 0, 0, target, 2, "attacker", 1, false)
	if err != nil {
		t.Fatalf("applyDamageToUnit returned error: %v", err)
	}
	if res.DamageToHP != 2 {
		t.Fatalf("expected 2 damage to HP, got %d", res.DamageToHP)
	}
	if target.HP != 8 {
		t.Fatalf("expected target HP 8, got %d", target.HP)
	}
	if target.Attack != 3 {
		t.Fatalf("expected OnDamaged passive to buff attack to 3, got %d", target.Attack)
	}
}

func TestPassive_OnAttack_FromCardAttack(t *testing.T) {
	m := testMatch()
	attacker := testUnit("attacker", 10)
	attacker.Attack = 2
	attacker.SummonedInTurn = 0
	attacker.Passive = cards.PassiveSpec{
		Code:       "claw_evolution",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerOnAttack,
		EffectKind: cards.PassiveEffectBuff,
		Target:     cards.SkillTargetSelf,
		Power:      1,
		BuffEffect: cards.BuffEffectAttack,
		IgnoreTank: true,
	}
	defender := testUnit("defender", 10)
	m.Players[0].Table[0] = attacker
	m.Players[1].Table[0] = defender

	r := BattleMapResolver{M: map[string]BattleTemplate{
		attacker.TemplateID: {TemplateID: attacker.TemplateID, AssetBaseKey: "attacker"},
	}}

	if err := CardAttack(m, 0, attacker.InstanceID, defender.InstanceID, false, r); err != nil {
		t.Fatalf("CardAttack returned error: %v", err)
	}
	if defender.HP != 8 {
		t.Fatalf("expected defender HP 8, got %d", defender.HP)
	}
	if attacker.Attack != 3 {
		t.Fatalf("expected OnAttack passive to buff attack to 3, got %d", attacker.Attack)
	}
}

func TestPassive_TurnStartAndTurnEnd(t *testing.T) {
	m := testMatch()
	startUnit := testUnit("start", 10)
	startUnit.Cooldown = 2
	startUnit.Skill.BaseCooldown = 3
	startUnit.Skill.CooldownLeft = 2
	startUnit.Passive = cards.PassiveSpec{
		Code:       "trench_warfare",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerTurnStart,
		EffectKind: cards.PassiveEffectBuff,
		Target:     cards.SkillTargetSelf,
		Power:      1,
		BuffEffect: cards.BuffEffectAttack,
		IgnoreTank: true,
	}
	endUnit := testUnit("end", 10)
	endUnit.Passive = cards.PassiveSpec{
		Code:       "constant_training",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerTurnEnd,
		EffectKind: cards.PassiveEffectBuff,
		Target:     cards.SkillTargetSelfAndAdjacent,
		Power:      1,
		BuffEffect: cards.BuffEffectAttack,
		IgnoreTank: true,
	}
	m.Players[0].Table[0] = startUnit
	m.Players[0].Table[1] = endUnit

	if err := StartTurn(m, 100); err != nil {
		t.Fatalf("StartTurn returned error: %v", err)
	}
	if startUnit.Attack != 3 {
		t.Fatalf("expected TurnStart passive to buff attack to 3, got %d", startUnit.Attack)
	}
	if startUnit.Cooldown != 1 {
		t.Fatalf("expected cooldown to tick once to 1, got %d", startUnit.Cooldown)
	}
	if startUnit.Skill.CooldownLeft != 1 {
		t.Fatalf("expected skill cooldown to tick once to 1, got %d", startUnit.Skill.CooldownLeft)
	}

	m.ActivePlayer = 0
	m.Phase = PhaseMain
	if err := EndTurn(m); err != nil {
		t.Fatalf("EndTurn returned error: %v", err)
	}
	if endUnit.Attack != 3 {
		t.Fatalf("expected TurnEnd passive to buff end unit attack to 3, got %d", endUnit.Attack)
	}
	if startUnit.Attack != 4 {
		t.Fatalf("expected adjacent TurnEnd passive to buff start unit attack to 4, got %d", startUnit.Attack)
	}
}

func TestPassive_CardSkillAndHeroSkillTriggers(t *testing.T) {
	m := testMatch()
	caster := testUnit("skill_caster", 10)
	caster.HasSkill = true
	caster.Skill = cards.UnitSkillState{
		Code:         "continuous_battle",
		Target:       cards.SkillTargetSelf,
		Power:        1,
		BuffEffect:   cards.BuffEffectAttack,
		BaseCooldown: 1,
	}
	reactor := testUnit("skill_reactor", 20)
	reactor.Passive = cards.PassiveSpec{
		Code:        "countermeasures",
		Kind:        cards.PassiveKindReactive,
		Trigger:     cards.PassiveTriggerOnEnemySkill,
		EffectKind:  cards.PassiveEffectDamage,
		Target:      cards.SkillTargetEnemyRandom,
		Power:       2,
		EventFilter: cards.PassiveEventFilterSkillUsed,
		IgnoreTank:  true,
	}
	m.Players[0].Table[0] = caster
	m.Players[1].Table[0] = reactor

	if err := PlayCardSkill(m, Action{PlayerIndex: 0, CardInstanceID: caster.InstanceID}); err != nil {
		t.Fatalf("PlayCardSkill returned error: %v", err)
	}
	if caster.HP != 8 {
		t.Fatalf("expected enemy OnSkill passive to damage caster to 8 HP, got %d", caster.HP)
	}

	m.Events = nil
	heroReactor := reactor
	heroReactor.Passive = cards.PassiveSpec{
		Code:       "instant_reaction",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerOnEnemyHeroSkill,
		EffectKind: cards.PassiveEffectDamage,
		Target:     cards.SkillTargetEnemyRandom,
		Power:      2,
		IgnoreTank: true,
	}
	heroTarget := testUnit("hero_target", 20)
	m.Players[1].Table[1] = heroTarget
	m.Players[0].HeroCode = "marksman_hero"
	m.Players[0].Mana = 10

	res := Resolvers{HeroTemplate: &HeroTemplateMapResolver{M: map[string]heroes.CharacterTemplate{
		"marksman_hero": {
			CharacterCode: "marksman_hero",
			Ability: heroes.AbilitySpec{
				Code:       "outstanding_marksman",
				Kind:       cards.SkillKindDamage,
				Target:     cards.SkillTargetEnemySingle,
				Power:      3,
				CoolDown:   1,
				ManaCost:   0,
				IgnoreTank: true,
			},
		},
	}}}

	if err := PlayHeroSpell(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: heroTarget.InstanceID,
	}, res); err != nil {
		t.Fatalf("PlayHeroSpell returned error: %v", err)
	}
	if heroTarget.HP != 17 {
		t.Fatalf("expected hero skill to damage target to 17 HP, got %d", heroTarget.HP)
	}
	if caster.HP != 6 {
		t.Fatalf("expected enemy hero skill passive to damage caster to 6 HP, got %d", caster.HP)
	}
}

func TestPassive_OnLeaveAndDeathTriggers(t *testing.T) {
	m := testMatch()
	dead := testUnit("dead", 1)
	dead.Passive = cards.PassiveSpec{
		Code:         "fatal_infection",
		Kind:         cards.PassiveKindReactive,
		Trigger:      cards.PassiveTriggerOnLeave,
		EffectKind:   cards.PassiveEffectDebuff,
		Target:       cards.SkillTargetEnemyAll,
		Power:        5,
		Duration:     3,
		DebuffEffect: cards.DebuffEffectDamageOverTime,
		IgnoreTank:   true,
	}
	enemy := testUnit("enemy", 10)
	allyReactor := testUnit("ally_reactor", 10)
	allyReactor.Passive = cards.PassiveSpec{
		Code:       "perfection_of_tactics",
		Kind:       cards.PassiveKindReactive,
		Trigger:    cards.PassiveTriggerOnAllyDeath,
		EffectKind: cards.PassiveEffectBuff,
		Target:     cards.SkillTargetSelf,
		Power:      1,
		BuffEffect: cards.BuffEffectAttack,
		IgnoreTank: true,
	}
	m.Players[0].Table[0] = dead
	m.Players[0].Table[1] = allyReactor
	m.Players[1].Table[0] = enemy

	if _, err := applyDamageToUnit(m, 0, 0, dead, 2, "killer", 1, false); err != nil {
		t.Fatalf("applyDamageToUnit returned error: %v", err)
	}
	if m.Players[0].Table[0] != nil {
		t.Fatalf("expected dead unit removed from table")
	}
	if len(enemy.Effects) != 1 || enemy.Effects[0].EffectType != cards.DebuffEffectDamageOverTime {
		t.Fatalf("expected OnLeave fatal infection debuff on enemy, got %+v", enemy.Effects)
	}
	if allyReactor.Attack != 3 {
		t.Fatalf("expected ally death passive to buff attack to 3, got %d", allyReactor.Attack)
	}
}
