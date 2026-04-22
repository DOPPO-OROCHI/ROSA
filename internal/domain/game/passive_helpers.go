package game

import (
	"TheWar/internal/domain/cards"
	"math/rand/v2"
)

type PassiveTargetRef struct {
	OwnerIndex int
	Slot       int
	Unit       *UnitState
}

func passiveConditionSatisfied(owner *PlayerState, enemy *PlayerState, spec cards.PassiveSpec) bool {
	switch spec.Condition {
	case "":
		return true
	case cards.PassiveConditionOwnerHasTank:
		return playerHasTank(owner)
	case cards.PassiveConditionEnemyHasTank:
		return playerHasTank(enemy)
	case cards.PassiveConditionOwnerRaceCount:
		return countRaceTable(owner, spec.ConditionRace) >= spec.ConditionValue
	case cards.PassiveConditionEnemyRaceCount:
		return countRaceTable(enemy, spec.ConditionRace) >= spec.ConditionValue
	case cards.PassiveConditionOwnerAllRace:
		return allUnitsMatchRace(owner, spec.ConditionRace)
	case cards.PassiveConditionEnemyAllRace:
		return allUnitsMatchRace(enemy, spec.ConditionRace)
	default:
		return false
	}
}

// УБИРАЕМ ПАССИВКУ С КАРТЫ
func clearPassiveAuraEffect(p *PlayerState, sourceInstanceID string, spec cards.PassiveSpec) {
	if p == nil {
		return
	}
	for slot := 0; slot < TableSize; slot++ {
		u := p.Table[slot]
		if u == nil || len(u.Effects) == 0 {
			continue
		}
		kept := make([]UnitEffect, 0, len(u.Effects))
		for _, e := range u.Effects {
			if e.EffectLayer == cards.EffectLayerPassive && e.SourceInstanceID == sourceInstanceID && e.EffectType == spec.BuffEffect {
				_ = RemoveEffect(u, e)
				continue
			}
			kept = append(kept, e)
		}
		u.Effects = kept
	}
}

func resolvePassiveTargets(
	owner *PlayerState,
	enemy *PlayerState,
	source *UnitState,
	spec cards.PassiveSpec,
	ctx PassiveContext,
) []PassiveTargetRef {
	targets := make([]PassiveTargetRef, 0, TableSize)
	if owner == nil || enemy == nil || source == nil {
		return targets
	}
	ownerIndex := ctx.SourcePlayerIndex
	enemyIndex := 1 - ctx.SourcePlayerIndex
	appendTarget := func(player *PlayerState, playerIndex int, slot int, unit *UnitState) {
		if unit == nil {
			return
		}
		if spec.TargetRace != "" && unit.CardType != spec.TargetRace {
			return
		}
		targets = append(targets, PassiveTargetRef{
			OwnerIndex: playerIndex,
			Slot:       slot,
			Unit:       unit,
		})
	}
	switch spec.Target {
	case cards.SkillTargetSelf:
		sourceSlot, _ := owner.FindSlot(source.InstanceID)
		if sourceSlot >= 0 {
			appendTarget(owner, ownerIndex, sourceSlot, source)
		}
	case cards.SkillTargetAllyAll:
		for slot := 0; slot < TableSize; slot++ {
			appendTarget(owner, ownerIndex, slot, owner.Table[slot])
		}
	case cards.SkillTargetEnemyAll:
		for slot := 0; slot < TableSize; slot++ {
			appendTarget(enemy, enemyIndex, slot, enemy.Table[slot])
		}
	case cards.SkillTargetAllyAdjacent:
		sourceSlot, _ := owner.FindSlot(source.InstanceID)
		if sourceSlot >= 0 {
			left := sourceSlot - 1
			right := sourceSlot + 1
			if left >= 0 {
				appendTarget(owner, ownerIndex, left, owner.Table[left])
			}
			if right < TableSize {
				appendTarget(owner, ownerIndex, right, owner.Table[right])
			}
		}
	case cards.SkillTargetSelfAndAdjacent:
		sourceSlot, _ := owner.FindSlot(source.InstanceID)
		if sourceSlot >= 0 {
			appendTarget(owner, ownerIndex, sourceSlot, source)
			left := sourceSlot - 1
			right := sourceSlot + 1
			if left >= 0 {
				appendTarget(owner, ownerIndex, left, owner.Table[left])
			}
			if right < TableSize {
				appendTarget(owner, ownerIndex, right, owner.Table[right])
			}
		}
	case cards.SkillTargetAllyLowestHP:
		var best *UnitState
		bestSlot := -1
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u == nil {
				continue
			}
			if spec.TargetRace != "" && u.CardType != spec.TargetRace {
				continue
			}
			if best == nil || u.HP < best.HP {
				best = u
				bestSlot = slot
			}
		}
		if best != nil {
			appendTarget(owner, ownerIndex, bestSlot, best)
		}
	case cards.SkillTargetAllyHighestAttack:
		var best *UnitState
		bestSlot := -1
		for slot := 0; slot < TableSize; slot++ {
			u := owner.Table[slot]
			if u == nil {
				continue
			}
			if spec.TargetRace != "" && u.CardType != spec.TargetRace {
				continue
			}
			if best == nil || u.Attack > best.Attack {
				best = u
				bestSlot = slot
			}
		}
		if best != nil {
			appendTarget(owner, ownerIndex, bestSlot, best)
		}
	case cards.SkillTargetEnemySingle:
		if ctx.AttackTargetID == "" {
			return targets
		}
		slot, target := enemy.FindSlot(ctx.AttackTargetID)
		if target == nil || slot < 0 {
			return targets
		}
		if !spec.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return targets
		}
		appendTarget(enemy, enemyIndex, slot, target)
	case cards.SkillTargetEnemySplash:
		if ctx.AttackTargetID == "" {
			return targets
		}
		slot, target := enemy.FindSlot(ctx.AttackTargetID)
		if target == nil || slot < 0 {
			return targets
		}
		if !spec.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return targets
		}
		appendTarget(enemy, enemyIndex, slot, target)
		left := slot - 1
		right := slot + 1
		if left >= 0 {
			appendTarget(enemy, enemyIndex, left, enemy.Table[left])
		}
		if right < TableSize {
			appendTarget(enemy, enemyIndex, right, enemy.Table[right])
		}
	case cards.SkillTargetEnemyRandom:
		pool := make([]PassiveTargetRef, 0, TableSize)
		for slot := 0; slot < TableSize; slot++ {
			u := enemy.Table[slot]
			if u == nil {
				continue
			}
			if spec.TargetRace != "" && u.CardType != spec.TargetRace {
				continue
			}
			if !spec.IgnoreTank && enemyHasTank(enemy) && !u.IsTank {
				continue
			}
			pool = append(pool, PassiveTargetRef{
				OwnerIndex: enemyIndex,
				Slot:       slot,
				Unit:       u,
			})
		}
		if len(pool) == 0 {
			return targets
		}
		pick := rand.IntN(len(pool))
		targets = append(targets, pool[pick])
	case cards.SkillTargetAttackTarget:
		if ctx.DamagedByInstanceID == "" {
			return targets
		}
		slot, target := enemy.FindSlot(ctx.DamagedByInstanceID)
		if target == nil || slot < 0 {
			return targets
		}
		appendTarget(enemy, enemyIndex, slot, target)
	}
	return targets
}

// ищем танка на столе
func playerHasTank(p *PlayerState) bool {
	if p == nil {
		return false
	}
	for slot := 0; slot < TableSize; slot++ {
		u := p.Table[slot]
		if u != nil && u.IsTank {
			return true
		}
	}
	return false
}

// ищем по расе
func countRaceTable(p *PlayerState, race string) int {
	if p == nil || race == "" {
		return 0
	}
	n := 0
	for slot := 0; slot < TableSize; slot++ {
		u := p.Table[slot]
		if u != nil && u.CardType == race {
			n++
		}
	}
	return n
}

func allUnitsMatchRace(p *PlayerState, race string) bool {
	if p == nil || race == "" {
		return false
	}
	hasAny := false
	for slot := 0; slot < TableSize; slot++ {
		u := p.Table[slot]
		if u == nil {
			continue
		}
		hasAny = true
		if u.CardType != race {
			return false
		}
	}
	return hasAny
}
