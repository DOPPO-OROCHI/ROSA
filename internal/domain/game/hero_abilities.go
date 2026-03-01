package game

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"errors"
)

type HeroAbility interface {
	Spec() heroes.AbilitySpec
	Apply(st *MatchState, a Action) error
}

type SupremeLiderAbilitySpec struct{}

type KarnAbilitySpec struct{}

type TheSystemAbilitySpec struct{}

type ImperialCommanderAbilitySpec struct{}

type BlackCellAbilitySpec struct{}

type SlavicPriestAbilitySpec struct{}

// СЛАВЯНСКИЙ СВЯЩЕННИК
func (SlavicPriestAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.HEAL_UNIT,
		Target:   heroes.OWN_UNIT,
		ManaCost: 3,
		Value:    15,
		Duration: 0,
	}
}

// СЛАВЯНСКИЙ СВЯЩЕННИК
func (ab SlavicPriestAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	p := st.Players[a.PlayerIndex]
	if p == nil {
		return errors.New("nil playe state")
	}
	slot, u := p.FindSlot(a.TargetInstanceID)
	if u == nil || slot < 0 {
		return ErrTargetNotFound
	}
	u.HP += spec.Value
	return nil
}

// ТЕРРОРИСТ
func (BlackCellAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.BUFF_ATK_PERM,
		Target:   heroes.OWN_UNIT,
		CoolDown: 2,
		ManaCost: 2,
		Value:    10,
		Duration: 0,
	}
}

// ТЕРРОРИСТ
func (ab BlackCellAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	p := st.Players[a.PlayerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	slot, u := p.FindSlot(a.TargetInstanceID)
	if u == nil || slot < 0 {
		return ErrTargetNotFound
	}
	e := UnitEffect{
		EffectType: heroes.BUFF_ATK_PERM,
		TurnsLeft:  spec.Duration,
		Value:      spec.Value,
	}
	u.HP = 1
	AddEffect(u, e)
	return nil
}

// КАРН
func (KarnAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.MAKE_TANK,
		Target:   heroes.OWN_UNIT,
		CoolDown: 5,
		Value:    0,
		Duration: 3,
	}
}

// ИМПЕРСКИЙ КОММАНДИР
func (ImperialCommanderAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.BUFF_ATK,
		CoolDown: 3,
		ManaCost: 2,
		Value:    10,
		Duration: 3,
	}
}

// ИМПЕРСКИЙ КОММАНДИР
func (ab ImperialCommanderAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	p := st.Players[a.PlayerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	slot, u := p.FindSlot(a.TargetInstanceID)
	if u == nil || slot < 0 {
		return ErrTargetNotFound
	}
	e := UnitEffect{
		EffectType: heroes.BUFF_ATK,
		TurnsLeft:  spec.Duration,
		Value:      10,
	}
	ApplyEffect(u, e)
	return nil
}

// КАРН
func (ab KarnAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	p := st.Players[a.PlayerIndex]
	if p == nil {
		return errors.New("nil player state")
	}
	slot, u := p.FindSlot(a.TargetInstanceID)
	if u == nil || slot < 0 {
		return ErrTargetNotFound
	}
	if u.IsTank {
		return errors.New("card is tank type already")
	}
	e := UnitEffect{
		EffectType: cards.MakeTankUpdate,
		TurnsLeft:  spec.Duration,
		Value:      0,
	}
	AddEffect(u, e)
	return nil
}

// ВЕРХОВНЫЙ ЛИДЕР
func (SupremeLiderAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.ATTACK_ANY,
		Target:   heroes.ENEMY_ANY,
		CoolDown: 3,
		ManaCost: 5,
		Value:    30,
		Duration: 0,
	}
}

// ВЕРХОВНЫЙ ЛИДЕР
func (ab SupremeLiderAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	def := st.Players[1-a.PlayerIndex]
	if def == nil {
		return ErrTargetNotFound
	}
	dmg := spec.Value
	if a.AttackHero {
		def.HeroHP -= dmg
		if def.HeroHP <= 0 {
			st.Finished = true
			if a.PlayerIndex == 0 {
				st.Result = MatchWinP1
			} else {
				st.Result = MatchWinP2
			}
		}
		return nil
	}
	slot, u := def.FindSlot(a.TargetInstanceID)
	if u == nil || slot < 0 {
		return ErrDefenderNotFound
	}
	u.HP -= dmg
	if u.HP <= 0 {
		def.RemoveAt(slot)
	}
	return nil
}

// СИСТЕМА
func (TheSystemAbilitySpec) Spec() heroes.AbilitySpec {
	return heroes.AbilitySpec{
		Code:     heroes.ATTACK_SPLASH,
		Target:   heroes.ENEMY_UNIT,
		CoolDown: 3,
		ManaCost: 3,
		Value:    4,
		Duration: 0,
	}
}

// СИСТЕМА
func (ab TheSystemAbilitySpec) Apply(st *MatchState, a Action) error {
	spec := ab.Spec()
	atk := st.Players[a.PlayerIndex]
	def := st.Players[1-a.PlayerIndex]
	if atk == nil || def == nil {
		return errors.New("nil player state")
	}
	if a.AttackHero {
		return ErrHeroAbilityCannotAttackHero
	}
	if a.CardInstanceID == "" {
		return ErrTargetNotFound
	}
	centerSlot, center := def.FindSlot(a.TargetInstanceID)
	if center == nil || centerSlot < 0 {
		return ErrDefenderNotFound
	}
	dmg := spec.Value
	targetSlot := make([]int, 0, 3)
	targetSlot = append(targetSlot, centerSlot)
	left, right := centerSlot-1, centerSlot+1
	if left >= 0 && def.Table[left] != nil {
		targetSlot = append(targetSlot, left)
	}
	if right < TableSize && def.Table[right] != nil {
		targetSlot = append(targetSlot, right)
	}
	for _, r := range targetSlot {
		u := def.Table[r]
		if u == nil {
			continue
		}
		u.HP -= dmg
		if u.HP <= 0 {
			def.RemoveAt(r)
		}
	}
	return nil
}
