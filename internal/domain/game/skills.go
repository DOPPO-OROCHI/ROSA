package game

import (
	"errors"
	"fmt"
)

type SkillHandler func(m *MatchState, a Action, caster *UnitState) error

var SkillHandlers = map[string]SkillHandler{
	"fragmentation_grenades": CastSplashDamageSkill,
}

func CastSingleDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match of caster")
	}
	owner := m.Players[a.PlayerIndex]
	enemy := m.Players[1-a.PlayerIndex]
	if owner == nil || enemy == nil {
		return errors.New("nil player state")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if a.AttackHero {
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) {
			return ErrCardSkillTargetTankBlocked
		}
		damage := caster.Skill.Power
		enemy.HeroHP -= damage
		heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
		m.Events = append(m.Events, Event{
			Targets: []EventTarget{
				{
					InstanceID: heroID,
					Amount:     damage,
					Died:       enemy.HeroHP <= 0,
					NewHP:      enemy.HeroHP,
				},
			},
		})
		if enemy.HeroHP <= 0 {
			m.Finished = true
			if a.PlayerIndex == 0 {
				m.Result = MatchWinP1
			} else {
				m.Result = MatchWinP2
			}
		}
	} else {
		targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
		if target == nil || targetSlot == -1 {
			return ErrTargetNotFound
		}
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
			return ErrCardSkillTargetTankBlocked
		}
		damage := caster.Skill.Power
		target.HP -= damage
		died := target.HP <= 0
		newHP := target.HP
		if died {
			if err := killUnitAt(m, 1-a.PlayerIndex, targetSlot, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
		}
		m.Events = append(m.Events, Event{
			Targets: []EventTarget{
				{
					InstanceID: target.InstanceID,
					TemplateID: target.TemplateID,
					Amount:     damage,
					Died:       died,
					NewHP:      newHP,
				},
			},
		})
	}
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

func CastSplashDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	return nil
}

// ищем танка в руке противника
func enemyHasTank(p *PlayerState) bool {
	if p == nil {
		return false
	}
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		if u != nil && u.IsTank {
			return true
		}
	}
	return false
}
