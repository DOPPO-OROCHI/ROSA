package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

//Здесь будут располагаться боевые хуки, типа life_on_heat, counerattack и так далее

// ХИЛИМСЯ ЕСЛИ НАС ПИЗДЯТ
func applyLifeOnHit(target *UnitState) int {
	if target == nil {
		return 0
	}
	if HasEffect(target, cards.DebuffEffectNoHeal) {
		return 0
	}
	heal := 0
	for _, e := range target.Effects {
		if e.EffectType == cards.BuffEffectLifeOnHit {
			heal += e.Value
		}
	}
	if heal <= 0 {
		return 0
	}
	beforeHP := target.HP
	target.HP += heal
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	return target.HP - beforeHP
}

func applyCounterattack(m *MatchState, defenderOwnerIdx int,
	defender *UnitState, attackerOwnerIdx int, attackerSlot int,
	attacker *UnitState) (UnitDamageResult, error) {
	result := UnitDamageResult{}
	if m == nil {
		return result, errors.New("nil match state")
	}
	if defender == nil || attacker == nil {
		return result, nil
	}
	if attackerSlot < 0 || attackerSlot >= TableSize {
		return result, nil
	}
	counterDamage := 0
	for _, e := range defender.Effects {
		if e.EffectType == cards.BuffEffectCounterattack {
			counterDamage += e.Value
		}
	}
	if counterDamage <= 0 {
		return result, nil
	}
	res, err := applyDamageToUnit(m, attackerOwnerIdx,
		attackerSlot, attacker, counterDamage,
		defender.InstanceID, defenderOwnerIdx, false)
	if err != nil {
		return result, err
	}
	return res, nil
}

func applyVampiricOnHit(attacker *UnitState, dealtToHP int) int {
	return 0
}
