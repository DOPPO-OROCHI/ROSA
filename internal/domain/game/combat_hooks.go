package game

import (
	"TheWar/internal/domain/cards"
	"errors"
	"math/rand/v2"
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

/*ДАРУЕМ ЦЕЛИ АУРУ КОНТРАТТАК НА Х ХОДОВ*/
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

/*ДАРУЕМ ВАМПИРИЗМ ЦЕЛИ*/
func applyVampiricOnHit(attacker *UnitState, dealtToHP int) int {
	if attacker == nil || dealtToHP <= 0 {
		return 0
	}
	if HasEffect(attacker, cards.DebuffEffectNoHeal) {
		return 0
	}
	total := 0
	for _, e := range attacker.Effects {
		if e.EffectType == cards.BuffEffectVampiricStrike {
			total += e.Value
		}
	}
	if total <= 0 {
		return 0
	}
	heal := (dealtToHP * total) / 100
	if heal <= 0 {
		heal = 1
	}
	beforeHP := attacker.HP
	attacker.HP += heal
	if attacker.HP > attacker.MaxHP {
		attacker.HP = attacker.MaxHP
	}
	return attacker.HP - beforeHP
}

/*
БОНУС ПОСЛЕ АТАКИ
Смысл в том, чтобы чувак получил какой-либо бонус после атаки. Бонусов есть 2 вида:
1-Бонус к атаке.
2-Бонус к КД атаки.
Атака регулируется основной силой скилла (в дефолтах карт, SkillPower, а так же в эффектах Value),
в то время как КД регулируется ExtraValue. Функция должна работать только на атакующем юните.
*/
func applyBonusAfterAttack(attacker *UnitState) {
	if attacker == nil {
		return
	}
	attackBonus := 0
	cdBonus := 0
	for _, e := range attacker.Effects {
		if e.EffectType != cards.BuffEffectBonusAfterAttack {
			continue
		}
		attackBonus += e.Value
		cdBonus += e.ExtraValue
	}
	if attackBonus != 0 {
		attacker.Attack += attackBonus
		if attacker.Attack < 1 {
			attacker.Attack = 1
		}
	}
	if cdBonus != 0 {
		attacker.BaseCooldown -= cdBonus
		if attacker.BaseCooldown < 1 {
			attacker.BaseCooldown = 1
		}
		if attacker.Cooldown > attacker.BaseCooldown {
			attacker.Cooldown = attacker.BaseCooldown
		}
		if attacker.Cooldown < 0 {
			attacker.Cooldown = 0
		}
	}
}

/*
Функция посвященная цепной атаке. В чем смысл ? Это баф на цель, который гарантирует ей то, что во время ее
атаки, сам удар отскочит в случайных противников, число которых регулируется с помощью ExtraValue. Сам урон
определяется базовой атакой карты, на которую наложили эффект. То есть при описании карты, не нужно указывать
value(skillPower), нужно только SkillExtraValue, которое отвечает за кол-во целей после чейна
*/
func applyChainAttack(m *MatchState, attackerOwnerIdx int, attacker *UnitState) ([]EventTarget, error) {
	out := make([]EventTarget, 0)
	if m == nil || attacker == nil {
		return out, nil
	}
	if attackerOwnerIdx < 0 || attackerOwnerIdx > 1 {
		return out, nil
	}
	enemyIdx := 1 - attackerOwnerIdx
	enemy := m.Players[enemyIdx]
	if enemy == nil {
		return out, nil
	}
	hits := 0
	for _, e := range attacker.Effects {
		if e.EffectType != cards.BuffEffectChainAttack {
			continue
		}
		if e.ExtraValue > 0 {
			hits += e.ExtraValue
		}
	}
	if hits <= 0 {
		return out, nil
	}
	damage := attacker.Attack / 2
	if damage <= 0 {
		return out, nil
	}
	for i := 0; i < hits; i++ {
		pool := make([]int, 0, TableSize)
		for s := 0; s < TableSize; s++ {
			if enemy.Table[s] != nil {
				pool = append(pool, s)
			}
		}
		if len(pool) == 0 {
			break
		}
		slot := pool[rand.IntN(len(pool))]
		target := enemy.Table[slot]
		if target == nil {
			continue
		}
		inst := target.InstanceID
		tplID := target.TemplateID
		res, err := applyDamageToUnit(m, enemyIdx, slot,
			target, damage, attacker.InstanceID, attackerOwnerIdx, true)
		if err != nil {
			return out, err
		}
		out = append(out, EventTarget{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     res.DamageToHP,
			Died:       res.Died,
			NewHP:      res.NewHP,
		})
	}
	return out, nil
}

/*
Хелпер на овердрайв. Так как это лишь статус о том, что юнит теперь имеет даблхит
эффект, отдельного хендлера писать не надо. Овердрайв всегда и всегда только о
даблхите. Больше хитов не предусмотрено.
*/
func hasOverdrive(u *UnitState) bool {
	if u == nil {
		return false
	}
	for _, e := range u.Effects {
		if e.EffectType == cards.BuffEffectOverdrive {
			return true
		}
	}
	return false
}

func maxAttacksPerTurn(u *UnitState) int {
	if hasOverdrive(u) {
		return 2
	}
	return 1
}

/*
Хелпер под мультикаст. Такая же логика как и в случае овердрайва
*/
func hasMulticast(u *UnitState) bool {
	if u == nil {
		return false
	}
	for _, e := range u.Effects {
		if e.EffectType == cards.BuffEffectMulticast {
			return true
		}
	}
	return false
}
