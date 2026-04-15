package game

//НАБРОСКИ СКИЛОВ

// import (
// 	"TheWar/internal/domain/cards"
// 	"TheWar/internal/domain/heroes"
// 	"errors"
// )

// type HeroAbility interface {
// 	Apply(m *MatchState, a Action, spec heroes.AbilitySpec) error
// }

// type HeroBuffAttackAbility struct{}

// func (h HeroBuffAttackAbility) Apply(m *MatchState, a Action, spec heroes.AbilitySpec) error {
// 	if m == nil {
// 		return errors.New("nil match state")
// 	}
// 	p := m.Players[a.PlayerIndex]
// 	if p == nil {
// 		return errors.New("bad player index")
// 	}
// 	_, u := p.FindSlot(a.TargetInstanceID)
// 	if u == nil {
// 		return ErrTargetNotFound
// 	}
// 	e := UnitEffect{
// 		EffectType:  cards.BuffEffectAttack,
// 		TurnsLeft:   spec.Duration,
// 		Value:       spec.Value,
// 		Dispellable: true,
// 		SourceType:  string(SourceHero),
// 		Polarity:    "buff",
// 	}
// 	if err := AddEffect(u, e); err != nil {
// 		return err
// 	}
// 	return nil
// }

// type HeroBuffMakeTank struct{}

// func (h HeroBuffMakeTank) Apply(m *MatchState, a Action, spec heroes.AbilitySpec) error {
// 	if m == nil {
// 		return errors.New("nil match state")
// 	}
// 	p := m.Players[a.PlayerIndex]
// 	if p == nil {
// 		return errors.New("bad player index")
// 	}
// 	_, u := p.FindSlot(a.TargetInstanceID)
// 	if u == nil {
// 		return ErrTargetNotFound
// 	}
// 	e := UnitEffect{
// 		EffectType:  cards.BuffEffectMakeTank,
// 		TurnsLeft:   spec.Duration,
// 		Value:       0,
// 		Dispellable: true,
// 		SourceType:  string(SourceHero),
// 		Polarity:    "buff",
// 	}
// 	if err := AddEffect(u, e); err != nil {
// 		return err
// 	}
// 	return nil
// }

// type HeroHealUnitAbility struct{}

// func (h HeroHealUnitAbility) Apply(m *MatchState, a Action, spec heroes.AbilitySpec) error {
// 	if m == nil {
// 		return errors.New("nil match state")
// 	}
// 	p := m.Players[a.PlayerIndex]
// 	if p == nil {
// 		return errors.New("bad player index")
// 	}
// 	_, u := p.FindSlot(a.TargetInstanceID)
// 	if u == nil {
// 		return ErrTargetNotFound
// 	}
// 	return nil
// }
