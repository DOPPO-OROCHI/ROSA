package game

import (
	"TheWar/internal/domain/cards"
	"fmt"
	"testing"
)

func TestCastDebuffSkill_EnemySingle_OK(t *testing.T) {
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0},
			{PlayerID: 1},
		},
	}

	caster := &UnitState{
		InstanceID:   "c1",
		TemplateID:   "caster_tpl",
		HP:           10,
		MaxHP:        10,
		Attack:       3,
		AssetBaseKey: "caster",
		Skill: cards.UnitSkillState{
			Code:         "burning",
			Target:       cards.SkillTargetEnemySingle,
			DebuffEffect: cards.DebuffEffectDamageOverTime,
			Power:        2,
			Duration:     3,
			BaseCooldown: 2,
			CooldownLeft: 0,
		},
	}

	target := &UnitState{
		InstanceID: "e1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     2,
	}

	m.Players[0].Table[0] = caster
	m.Players[1].Table[0] = target

	a := Action{
		PlayerIndex:      0,
		TargetInstanceID: "e1",
	}

	err := CastDebuffSkill(m, a, caster)
	if err != nil {
		t.Fatalf("CastDebuffSkill returned error: %v", err)
	}
	if len(target.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(target.Effects))
	}
	if target.Effects[0].EffectType != cards.DebuffEffectDamageOverTime {
		t.Fatalf("unexpected effect type: %s", target.Effects[0].EffectType)
	}
	if caster.Skill.CooldownLeft != caster.Skill.BaseCooldown {
		t.Fatalf("cooldown not set: got %d want %d", caster.Skill.CooldownLeft, caster.Skill.BaseCooldown)
	}
	if len(m.Events) == 0 {
		t.Fatalf("expected event to be appended")
	}
}

func TestCastDebuffSkill_OnCooldown_ReturnsError(t *testing.T) {
	m := &MatchState{}
	caster := &UnitState{
		Skill: cards.UnitSkillState{
			Code:         "burning",
			CooldownLeft: 1,
		},
	}

	err := CastDebuffSkill(m, Action{PlayerIndex: 0}, caster)
	if err != ErrCardSkillOnCooldown {
		t.Fatalf("expected ErrCardSkillOnCooldown, got %v", err)
	}
}

func TestCastBuffSkill_Self_OK(t *testing.T) {
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0},
			{PlayerID: 1},
		},
	}
	caster := &UnitState{
		InstanceID:   "b1",
		TemplateID:   "buff_tpl",
		HP:           10,
		MaxHP:        10,
		Attack:       3,
		AssetBaseKey: "buff_unit",
		Skill: cards.UnitSkillState{
			Code:         "expansive_projectiles",
			Target:       cards.SkillTargetSelf,
			BuffEffect:   cards.BuffEffectAttack,
			Power:        2,
			Duration:     2,
			BaseCooldown: 3,
		},
	}
	m.Players[0].Table[0] = caster

	err := CastBuffSkill(m, Action{PlayerIndex: 0}, caster)
	if err != nil {
		t.Fatalf("CastBuffSkill returned error: %v", err)
	}
	if caster.Attack != 5 {
		t.Fatalf("expected attack=5, got %d", caster.Attack)
	}
	if len(caster.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(caster.Effects))
	}
	if caster.Skill.CooldownLeft != caster.Skill.BaseCooldown {
		t.Fatalf("cooldown not set: got %d want %d", caster.Skill.CooldownLeft, caster.Skill.BaseCooldown)
	}
	if len(m.Events) == 0 {
		t.Fatalf("expected event to be appended")
	}
}

func TestCastBuffSkill_OnCooldown_ReturnsError(t *testing.T) {
	m := &MatchState{}
	caster := &UnitState{
		Skill: cards.UnitSkillState{
			Code:         "expansive_projectiles",
			CooldownLeft: 1,
		},
	}
	err := CastBuffSkill(m, Action{PlayerIndex: 0}, caster)
	if err != ErrCardSkillOnCooldown {
		t.Fatalf("expected ErrCardSkillOnCooldown, got %v", err)
	}
}

func TestOverdrive_AllowsSecondAttackAfterBuffSkill(t *testing.T) {
	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 2, HeroHP: 30},
			{PlayerID: 1, Turns: 2, HeroHP: 30},
		},
	}
	attacker := &UnitState{
		InstanceID:      "attacker",
		TemplateID:      "overdrive_unit",
		HP:              10,
		MaxHP:           10,
		Attack:          2,
		BaseCooldown:    2,
		Cooldown:        2,
		AttacksThisTurn: 0,
		SummonedInTurn:  1,
		Skill: cards.UnitSkillState{
			Code:         "suppression",
			Target:       cards.SkillTargetSelf,
			BuffEffect:   cards.BuffEffectOverdrive,
			Duration:     1,
			BaseCooldown: 4,
			CooldownLeft: 0,
		},
	}
	firstTarget := &UnitState{
		InstanceID:   "target_1",
		TemplateID:   "target_unit",
		HP:           10,
		MaxHP:        10,
		Attack:       1,
		BaseCooldown: 1,
	}
	secondTarget := &UnitState{
		InstanceID:   "target_2",
		TemplateID:   "target_unit",
		HP:           10,
		MaxHP:        10,
		Attack:       1,
		BaseCooldown: 1,
	}
	m.Players[0].Table[0] = attacker
	m.Players[1].Table[0] = firstTarget
	m.Players[1].Table[1] = secondTarget

	if err := PlayCardSkill(m, Action{PlayerIndex: 0, CardInstanceID: attacker.InstanceID}); err != nil {
		t.Fatalf("PlayCardSkill returned error: %v", err)
	}
	if !hasOverdrive(attacker) {
		t.Fatalf("expected overdrive effect after skill")
	}
	if attacker.Cooldown != 0 {
		t.Fatalf("expected overdrive skill to reset attack cooldown to 0, got %d", attacker.Cooldown)
	}

	resolver := BattleMapResolver{M: map[string]BattleTemplate{
		"overdrive_unit": {
			TemplateID:   "overdrive_unit",
			Attack:       2,
			BaseCooldown: 2,
			AssetBaseKey: "overdrive_unit",
		},
	}}

	if err := CardAttack(m, 0, attacker.InstanceID, firstTarget.InstanceID, false, resolver); err != nil {
		t.Fatalf("first CardAttack returned error: %v", err)
	}
	if attacker.AttacksThisTurn != 1 {
		t.Fatalf("expected attacks_this_turn=1 after first attack, got %d", attacker.AttacksThisTurn)
	}
	if attacker.Cooldown != 0 {
		t.Fatalf("expected cooldown=0 after first overdrive attack, got %d", attacker.Cooldown)
	}

	if err := CardAttack(m, 0, attacker.InstanceID, secondTarget.InstanceID, false, resolver); err != nil {
		t.Fatalf("second CardAttack returned error: %v", err)
	}
	if attacker.AttacksThisTurn != 2 {
		t.Fatalf("expected attacks_this_turn=2 after second attack, got %d", attacker.AttacksThisTurn)
	}
	if attacker.Cooldown != attacker.BaseCooldown {
		t.Fatalf("expected cooldown=%d after second attack, got %d", attacker.BaseCooldown, attacker.Cooldown)
	}
}

func TestCastSummonSelfCopySkill_OK(t *testing.T) {
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 3},
			{PlayerID: 1},
		},
	}
	caster := &UnitState{
		InstanceID:   "boar_1",
		TemplateID:   "boar",
		HP:           5,
		MaxHP:        5,
		Attack:       3,
		BaseCooldown: 2,
		CardType:     cards.Mechanical,
		IsTank:       true,
		HasSkill:     true,
		Skill: cards.UnitSkillState{
			Code:         "group_work",
			Target:       cards.SkillTargetSelf,
			ApplyCount:   1,
			BaseCooldown: 8,
		},
	}
	m.Players[0].Table[0] = caster

	err := CastSummonSelfCopySkill(m, Action{PlayerIndex: 0}, caster)
	if err != nil {
		t.Fatalf("CastSummonSelfCopySkill returned error: %v", err)
	}

	copies := 0
	for i := 0; i < TableSize; i++ {
		u := m.Players[0].Table[i]
		if u == nil || u.InstanceID == caster.InstanceID {
			continue
		}
		copies++
		if u.TemplateID != caster.TemplateID {
			t.Fatalf("copy template mismatch: got %s want %s", u.TemplateID, caster.TemplateID)
		}
		if u.HasSkill {
			t.Fatalf("copy must not have skill")
		}
	}
	if copies != 1 {
		t.Fatalf("expected 1 copy, got %d", copies)
	}
	if caster.Skill.CooldownLeft != caster.Skill.BaseCooldown {
		t.Fatalf("cooldown not set: got %d want %d", caster.Skill.CooldownLeft, caster.Skill.BaseCooldown)
	}
	if len(m.Events) == 0 {
		t.Fatalf("expected event to be appended")
	}
}

func TestCastSummonSelfCopySkill_NoFreeSlot_ReturnsErrSlotOccupied(t *testing.T) {
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1},
			{PlayerID: 1},
		},
	}
	caster := &UnitState{
		InstanceID: "boar_1",
		TemplateID: "boar",
		HP:         5,
		MaxHP:      5,
		Attack:     3,
		Skill: cards.UnitSkillState{
			Code:       "group_work",
			Target:     cards.SkillTargetSelf,
			ApplyCount: 1,
		},
	}
	m.Players[0].Table[0] = caster
	for i := 1; i < TableSize; i++ {
		m.Players[0].Table[i] = &UnitState{
			InstanceID: fmt.Sprintf("ally_%d", i),
			TemplateID: "x",
			HP:         1,
			MaxHP:      1,
			Attack:     1,
		}
	}

	err := CastSummonSelfCopySkill(m, Action{PlayerIndex: 0}, caster)
	if err != ErrSlotOccupied {
		t.Fatalf("expected ErrSlotOccupied, got %v", err)
	}
}

// ///////////////////////////////////////////////////////////////////
func TestPlayCardSkill_Stun_Blocked(t *testing.T) {
	const skillCode = "__test_stun_blocked"

	// временный хендлер
	old, had := SkillHandlers[skillCode]
	SkillHandlers[skillCode] = func(m *MatchState, a Action, caster *UnitState) error { return nil }
	t.Cleanup(func() {
		if had {
			SkillHandlers[skillCode] = old
		} else {
			delete(SkillHandlers, skillCode)
		}
	})

	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     2,
		Skill: cards.UnitSkillState{
			Code:         skillCode,
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
		Effects: []UnitEffect{
			{EffectType: cards.DebuffEffectStun},
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err == nil || err.Error() != "caster is stunned" {
		t.Fatalf("expected 'caster is stunned', got %v", err)
	}
}

func TestPlayCardSkill_Silence_Blocked(t *testing.T) {
	const skillCode = "__test_silence_blocked"

	old, had := SkillHandlers[skillCode]
	SkillHandlers[skillCode] = func(m *MatchState, a Action, caster *UnitState) error { return nil }
	t.Cleanup(func() {
		if had {
			SkillHandlers[skillCode] = old
		} else {
			delete(SkillHandlers, skillCode)
		}
	})

	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     2,
		Skill: cards.UnitSkillState{
			Code:         skillCode,
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
		Effects: []UnitEffect{
			{EffectType: cards.DebuffEffectSilence},
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err == nil || err.Error() != "caster is silenced" {
		t.Fatalf("expected 'caster is silenced', got %v", err)
	}
}

func TestPlayCardSkill_Multicast_CallsHandlerTwice(t *testing.T) {
	const skillCode = "__test_multicast"

	old, had := SkillHandlers[skillCode]
	calls := 0
	SkillHandlers[skillCode] = func(m *MatchState, a Action, caster *UnitState) error {
		calls++
		return nil
	}
	t.Cleanup(func() {
		if had {
			SkillHandlers[skillCode] = old
		} else {
			delete(SkillHandlers, skillCode)
		}
	})

	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     2,
		Skill: cards.UnitSkillState{
			Code:         skillCode,
			BaseCooldown: 5,
			CooldownLeft: 0,
		},
		Effects: []UnitEffect{
			{EffectType: cards.BuffEffectMulticast},
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err != nil {
		t.Fatalf("PlayCardSkill returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected handler to be called 2 times, got %d", calls)
	}
	if caster.Skill.CooldownLeft != caster.Skill.BaseCooldown {
		t.Fatalf("cooldown mismatch: got %d want %d", caster.Skill.CooldownLeft, caster.Skill.BaseCooldown)
	}
}

// ////////////////////////////////////////////////////////////////////////////
func TestPlayCardSkill_OnCooldown_ReturnsErrCardSkillOnCooldown(t *testing.T) {
	const skillCode = "__test_on_cd"

	// даже если хендлер есть, до него не дойдет
	old, had := SkillHandlers[skillCode]
	SkillHandlers[skillCode] = func(m *MatchState, a Action, caster *UnitState) error { return nil }
	t.Cleanup(func() {
		if had {
			SkillHandlers[skillCode] = old
		} else {
			delete(SkillHandlers, skillCode)
		}
	})

	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		Skill: cards.UnitSkillState{
			Code:         skillCode,
			BaseCooldown: 5,
			CooldownLeft: 1,
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err != ErrCardSkillOnCooldown {
		t.Fatalf("expected ErrCardSkillOnCooldown, got %v", err)
	}
}

func TestPlayCardSkill_UnsupportedCode_ReturnsErrCardSkillUnsupported(t *testing.T) {
	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		Skill: cards.UnitSkillState{
			Code:         "__definitely_not_registered__",
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err != ErrCardSkillUnsupported {
		t.Fatalf("expected ErrCardSkillUnsupported, got %v", err)
	}
}

// ///////////////////////////////////////////////////////////////////////////
func TestPlayCardSkill_NoSkillCode_ReturnsErrCardSkillNotFound(t *testing.T) {
	caster := &UnitState{
		InstanceID: "c1",
		TemplateID: "tpl",
		Skill: cards.UnitSkillState{
			Code:         "",
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{Table: [TableSize]*UnitState{caster}},
			{},
		},
	}

	err := PlayCardSkill(m, Action{
		PlayerIndex:    0,
		CardInstanceID: caster.InstanceID,
	})
	if err != ErrCardSkillNotFound {
		t.Fatalf("expected ErrCardSkillNotFound, got %v", err)
	}
}

///////////////////////////////////////////////////////////////////////////

func TestCardAttack_ReflectShield_DamagesAttacker(t *testing.T) {
	atk := &UnitState{
		InstanceID:     "atk_1",
		TemplateID:     "atk_tpl",
		HP:             10,
		MaxHP:          10,
		Attack:         5,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
	}
	def := &UnitState{
		InstanceID:     "def_1",
		TemplateID:     "def_tpl",
		HP:             10,
		MaxHP:          10,
		Attack:         2,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
		Effects: []UnitEffect{
			{
				EffectType: cards.BuffEffectReflectShield,
				Value:      2, // щит поглотит 2
				ExtraValue: 2, // отражение 2
			},
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{atk}},
			{PlayerID: 1, Turns: 1, Table: [TableSize]*UnitState{def}},
		},
	}

	r := BattleMapResolver{
		M: map[string]BattleTemplate{
			"atk_tpl": {TemplateID: "atk_tpl", AssetBaseKey: "atk_asset"},
		},
	}

	err := CardAttack(m, 0, atk.InstanceID, def.InstanceID, false, r)
	if err != nil {
		t.Fatalf("CardAttack returned error: %v", err)
	}

	// 5 урона - 2 в щит = 3 в HP
	if def.HP != 7 {
		t.Fatalf("expected defender HP=7, got %d", def.HP)
	}
	// отражение 2 в атакера
	if atk.HP != 8 {
		t.Fatalf("expected attacker HP=8, got %d", atk.HP)
	}
	if len(m.Events) == 0 {
		t.Fatalf("expected attack event")
	}
}

func TestCardAttack_VampiricStrike_HealsAttacker(t *testing.T) {
	atk := &UnitState{
		InstanceID:     "atk_1",
		TemplateID:     "atk_tpl",
		HP:             5, // не фулл, чтобы увидеть хил
		MaxHP:          10,
		Attack:         4,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
		Effects: []UnitEffect{
			{
				EffectType: cards.BuffEffectVampiricStrike,
				Value:      50, // 50% от урона по HP
			},
		},
	}
	def := &UnitState{
		InstanceID:     "def_1",
		TemplateID:     "def_tpl",
		HP:             10,
		MaxHP:          10,
		Attack:         2,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{atk}},
			{PlayerID: 1, Turns: 1, Table: [TableSize]*UnitState{def}},
		},
	}

	r := BattleMapResolver{
		M: map[string]BattleTemplate{
			"atk_tpl": {TemplateID: "atk_tpl", AssetBaseKey: "atk_asset"},
		},
	}

	err := CardAttack(m, 0, atk.InstanceID, def.InstanceID, false, r)
	if err != nil {
		t.Fatalf("CardAttack returned error: %v", err)
	}

	// 4 урона в HP цели
	if def.HP != 6 {
		t.Fatalf("expected defender HP=6, got %d", def.HP)
	}
	// вампирик 50% от 4 = 2, было 5 -> стало 7
	if atk.HP != 7 {
		t.Fatalf("expected attacker HP=7, got %d", atk.HP)
	}
	if len(m.Events) == 0 {
		t.Fatalf("expected attack event")
	}
}

// /////////////////////////////////////////////////////////////////////////
func TestCardAttack_VampiricStrike_BlockedByNoHeal(t *testing.T) {
	atk := &UnitState{
		InstanceID:     "atk_1",
		TemplateID:     "atk_tpl",
		HP:             5,
		MaxHP:          10,
		Attack:         4,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
		Effects: []UnitEffect{
			{EffectType: cards.BuffEffectVampiricStrike, Value: 50},
			{EffectType: cards.DebuffEffectNoHeal},
		},
	}
	def := &UnitState{
		InstanceID:     "def_1",
		TemplateID:     "def_tpl",
		HP:             10,
		MaxHP:          10,
		Attack:         2,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{atk}},
			{PlayerID: 1, Turns: 1, Table: [TableSize]*UnitState{def}},
		},
	}

	r := BattleMapResolver{
		M: map[string]BattleTemplate{
			"atk_tpl": {TemplateID: "atk_tpl", AssetBaseKey: "atk_asset"},
		},
	}

	err := CardAttack(m, 0, atk.InstanceID, def.InstanceID, false, r)
	if err != nil {
		t.Fatalf("CardAttack returned error: %v", err)
	}

	// Урон по цели проходит
	if def.HP != 6 {
		t.Fatalf("expected defender HP=6, got %d", def.HP)
	}
	// Но хил от вампирика заблокирован no_heal
	if atk.HP != 5 {
		t.Fatalf("expected attacker HP=5 (no heal), got %d", atk.HP)
	}
}

// /////////////////////////////////////////////////////////////////////
func TestTickerEffects_DOT_KillsUnit(t *testing.T) {
	u := &UnitState{
		InstanceID: "u1",
		TemplateID: "tpl1",
		HP:         2,
		MaxHP:      5,
		Attack:     2,
		Effects: []UnitEffect{
			{
				EffectType: cards.DebuffEffectDamageOverTime,
				TurnsLeft:  1,
				Value:      2,
			},
		},
	}
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0, Table: [TableSize]*UnitState{u}},
			{PlayerID: 1},
		},
	}

	err := TickerEffects(m, 0)
	if err != nil {
		t.Fatalf("TickerEffects returned error: %v", err)
	}
	if m.Players[0].Table[0] != nil {
		t.Fatalf("expected unit to die and be removed from table")
	}
	if len(m.Players[0].GraveYard) != 1 {
		t.Fatalf("expected 1 unit in graveyard, got %d", len(m.Players[0].GraveYard))
	}
}

func TestTickerEffects_HealPerTurn_CappedByMaxHP(t *testing.T) {
	u := &UnitState{
		InstanceID: "u1",
		TemplateID: "tpl1",
		HP:         9,
		MaxHP:      10,
		Attack:     2,
		Effects: []UnitEffect{
			{
				EffectType: cards.BuffEffectHealPerTurn,
				TurnsLeft:  1,
				Value:      3,
			},
		},
	}
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0, Table: [TableSize]*UnitState{u}},
			{PlayerID: 1},
		},
	}

	err := TickerEffects(m, 0)
	if err != nil {
		t.Fatalf("TickerEffects returned error: %v", err)
	}
	if u.HP != 10 {
		t.Fatalf("expected HP=10, got %d", u.HP)
	}
	if len(u.Effects) != 0 {
		t.Fatalf("expected HoT effect to expire and be removed")
	}
}

func TestTickerEffects_ExpiresAttackBuffAndRollsBackStats(t *testing.T) {
	u := &UnitState{
		InstanceID: "u1",
		TemplateID: "tpl1",
		HP:         5,
		MaxHP:      5,
		Attack:     7, // база 5 + баф 2 уже применен
		Effects: []UnitEffect{
			{
				EffectType: cards.BuffEffectAttack,
				TurnsLeft:  1,
				Value:      2,
			},
		},
	}
	m := &MatchState{
		Players: [2]*PlayerState{
			{PlayerID: 0, Table: [TableSize]*UnitState{u}},
			{PlayerID: 1},
		},
	}

	err := TickerEffects(m, 0)
	if err != nil {
		t.Fatalf("TickerEffects returned error: %v", err)
	}
	if u.Attack != 5 {
		t.Fatalf("expected attack to rollback to 5, got %d", u.Attack)
	}
	if len(u.Effects) != 0 {
		t.Fatalf("expected expired buff to be removed")
	}
}

// ///////////////////////////////////////////////////////////////////////
func TestIntegration_Summon_Attack_Death(t *testing.T) {
	// caster умеет призывать копию
	caster := &UnitState{
		InstanceID:     "boar_1",
		TemplateID:     "boar_tpl",
		HP:             5,
		MaxHP:          5,
		Attack:         3,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
		CardType:       cards.Mechanical,
		IsTank:         true,
		HasSkill:       true,
		AssetBaseKey:   "boar_asset",
		Skill: cards.UnitSkillState{
			Code:         "group_work",
			Target:       cards.SkillTargetSelf,
			ApplyCount:   1,
			BaseCooldown: 8,
			CooldownLeft: 0,
		},
	}

	// цель, которая должна умереть от атаки
	def := &UnitState{
		InstanceID:     "enemy_1",
		TemplateID:     "enemy_tpl",
		HP:             2,
		MaxHP:          2,
		Attack:         1,
		BaseCooldown:   1,
		Cooldown:       0,
		SummonedInTurn: 0,
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{caster}},
			{PlayerID: 1, Turns: 1, Table: [TableSize]*UnitState{def}},
		},
	}

	// 1) summon
	err := CastSummonSelfCopySkill(m, Action{
		PlayerIndex: 0,
	}, caster)
	if err != nil {
		t.Fatalf("CastSummonSelfCopySkill returned error: %v", err)
	}

	// проверяем, что копия реально появилась
	copies := 0
	for i := 0; i < TableSize; i++ {
		u := m.Players[0].Table[i]
		if u != nil && u.InstanceID != caster.InstanceID {
			copies++
		}
	}
	if copies != 1 {
		t.Fatalf("expected 1 summoned copy, got %d", copies)
	}

	// 2) attack (оригинал атакует и убивает цель)
	r := BattleMapResolver{
		M: map[string]BattleTemplate{
			"boar_tpl": {TemplateID: "boar_tpl", AssetBaseKey: "boar_asset"},
		},
	}
	err = CardAttack(m, 0, caster.InstanceID, def.InstanceID, false, r)
	if err != nil {
		t.Fatalf("CardAttack returned error: %v", err)
	}

	// 3) death проверка
	if m.Players[1].Table[0] != nil {
		t.Fatalf("expected defender to be removed from table")
	}
	if len(m.Players[1].GraveYard) != 1 {
		t.Fatalf("expected 1 unit in enemy graveyard, got %d", len(m.Players[1].GraveYard))
	}

	// проверяем, что в событиях есть и summon, и death, и attack
	hasSummonSkillEvent := false
	hasAttackEvent := false
	hasDeathEvent := false
	for _, ev := range m.Events {
		if ev.Type == string(EventCardSkill) && ev.SourceInstanceID == caster.InstanceID {
			hasSummonSkillEvent = true
		}
		if ev.Type == string(EventAttack) && ev.SourceInstanceID == caster.InstanceID {
			hasAttackEvent = true
		}
		if ev.Type == string(EventDeath) {
			hasDeathEvent = true
		}
	}
	if !hasSummonSkillEvent {
		t.Fatalf("expected summon skill event")
	}
	if !hasAttackEvent {
		t.Fatalf("expected attack event")
	}
	if !hasDeathEvent {
		t.Fatalf("expected death event")
	}
}

// /////////////////////////////////////////////////////////////////////////
func TestIntegration_BuffDurationTwoTurns_RollbackStats(t *testing.T) {
	u := &UnitState{
		InstanceID: "u1",
		TemplateID: "tpl1",
		HP:         5,
		MaxHP:      5,
		Attack:     3,
		Skill: cards.UnitSkillState{
			Code: "dummy",
		},
	}
	caster := &UnitState{
		InstanceID:   "caster_1",
		TemplateID:   "caster_tpl",
		AssetBaseKey: "caster_asset",
		Skill: cards.UnitSkillState{
			Code:         "expansive_projectiles",
			Target:       cards.SkillTargetAllySingle,
			BuffEffect:   cards.BuffEffectAttack,
			Power:        2,
			Duration:     2,
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{caster, u}},
			{PlayerID: 1},
		},
	}

	// Каст бафа на u
	err := CastBuffSkill(m, Action{
		PlayerIndex:      0,
		TargetInstanceID: u.InstanceID,
	}, caster)
	if err != nil {
		t.Fatalf("CastBuffSkill returned error: %v", err)
	}
	if u.Attack != 5 {
		t.Fatalf("after buff expected attack=5, got %d", u.Attack)
	}

	// 1-й тик: баф еще должен висеть
	err = TickerEffects(m, 0)
	if err != nil {
		t.Fatalf("TickerEffects(1) returned error: %v", err)
	}
	if u.Attack != 5 {
		t.Fatalf("after tick #1 expected attack=5, got %d", u.Attack)
	}
	if len(u.Effects) != 1 {
		t.Fatalf("after tick #1 expected 1 effect, got %d", len(u.Effects))
	}

	// 2-й тик: баф должен сняться и откатить стат
	err = TickerEffects(m, 0)
	if err != nil {
		t.Fatalf("TickerEffects(2) returned error: %v", err)
	}
	if u.Attack != 3 {
		t.Fatalf("after tick #2 expected attack rollback to 3, got %d", u.Attack)
	}
	if len(u.Effects) != 0 {
		t.Fatalf("after tick #2 expected no effects, got %d", len(u.Effects))
	}
}

// //////////////////////////////////////////////////////////////////////////

func TestIntegration_MulticastDebuff_ThenCleanse(t *testing.T) {
	enemy := &UnitState{
		InstanceID: "enemy_1",
		TemplateID: "enemy_tpl",
		HP:         10,
		MaxHP:      10,
		Attack:     6,
	}

	debuffer := &UnitState{
		InstanceID:   "debuffer_1",
		TemplateID:   "debuffer_tpl",
		AssetBaseKey: "debuffer_asset",
		HP:           8,
		MaxHP:        8,
		Attack:       2,
		Skill: cards.UnitSkillState{
			Code:         "deep_wounds",
			Target:       cards.SkillTargetEnemySingle,
			DebuffEffect: cards.DebuffEffectAttackDown,
			Power:        1,
			Duration:     2,
			BaseCooldown: 5,
			CooldownLeft: 0,
		},
		Effects: []UnitEffect{
			{EffectType: cards.BuffEffectMulticast},
		},
	}

	cleanser := &UnitState{
		InstanceID:   "cleanser_1",
		TemplateID:   "cleanser_tpl",
		AssetBaseKey: "cleanser_asset",
		HP:           7,
		MaxHP:        7,
		Attack:       1,
		Skill: cards.UnitSkillState{
			Code:         "enemy_dispel",
			Target:       cards.SkillTargetEnemySingle,
			CleanseMode:  cards.CleanseModeRemoveAllEffects, // снимет и дебаффы тоже
			BaseCooldown: 3,
			CooldownLeft: 0,
		},
	}

	m := &MatchState{
		ActivePlayer: 0,
		Phase:        PhaseMain,
		Players: [2]*PlayerState{
			{PlayerID: 0, Turns: 1, Table: [TableSize]*UnitState{debuffer, cleanser}},
			{PlayerID: 1, Turns: 1, Table: [TableSize]*UnitState{enemy}},
		},
	}

	// 1) Мультикаст-дебафф (через общий пайплайн)
	err := PlayCardSkill(m, Action{
		PlayerIndex:      0,
		CardInstanceID:   debuffer.InstanceID,
		TargetInstanceID: enemy.InstanceID,
	})
	if err != nil {
		t.Fatalf("PlayCardSkill(debuffer) returned error: %v", err)
	}

	if enemy.Attack != 4 {
		t.Fatalf("expected enemy attack=4 after multicast debuff, got %d", enemy.Attack)
	}

	// 2) Снимаем эффекты с противника прямым вызовом твоей функции
	err = CastDispelBuffsFromEnemySkill(m, Action{
		PlayerIndex:      0,
		CardInstanceID:   cleanser.InstanceID,
		TargetInstanceID: enemy.InstanceID,
	}, cleanser)
	if err != nil {
		t.Fatalf("CastDispelBuffsFromEnemySkill returned error: %v", err)
	}

	if enemy.Attack != 6 {
		t.Fatalf("expected enemy attack restored to 6 after dispel, got %d", enemy.Attack)
	}
	for _, e := range enemy.Effects {
		if e.EffectType == cards.DebuffEffectAttackDown {
			t.Fatalf("expected no attack_down effects after dispel")
		}
	}
}
