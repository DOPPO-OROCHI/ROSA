package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

func TickerEffects(p *PlayerState) {
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		if u == nil {
			continue
		}
		out := u.Effects[:0]
		for _, e := range u.Effects {
			if e.TurnsLeft == 0 {
				out = append(out, e)
				continue
			}
			e.TurnsLeft--
			if e.TurnsLeft <= 0 {
				RemoveEffect(u, e)
				continue
			}
			out = append(out, e)
		}
		u.Effects = out
	}
}

func AddEffect(u *UnitState, e UnitEffect) {
	u.Effects = append(u.Effects, e)
	ApplyEffect(u, e)
}

func RemoveEffect(u *UnitState, e UnitEffect) {
	switch e.EffectType {
	case cards.DamageUpdate:
		u.Attack -= e.Value
	case cards.HealthPointsUpdate:
		u.HP -= e.Value
		if u.HP < 0 {
			u.HP = 0
		}
	case cards.CoolDownUpdate:
		u.Cooldown += e.Value
	case cards.MakeTankUpdate:
		u.IsTank = false
	}
}

func ApplyEffect(u *UnitState, buff UnitEffect) error {
	switch buff.EffectType {
	case cards.DamageUpdate:
		u.Attack += buff.Value
	case cards.HealthPointsUpdate:
		u.HP += buff.Value
	case cards.CoolDownUpdate:
		u.Cooldown -= buff.Value
		if u.Cooldown < 0 {
			u.Cooldown = 0
		}
	case cards.MakeTankUpdate:
		if u.IsTank == true {
			return errors.New("card is tank type already")
		}
		u.IsTank = true
	}
	return nil
}
