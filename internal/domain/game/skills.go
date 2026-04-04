package game

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

/*Данный файл целиком и полностью описывает функциональную сторону скиллов в моей игре. Здесь представлены
хендлеры по их поведению, например -ебнуть одну цель, ебнуть несколько случайных, ебнуть всех и тд*/

type SkillHandler func(m *MatchState, a Action, caster *UnitState) error

var SkillHandlers = map[string]SkillHandler{
	"fragmentation_grenades": CastSplashDamageSkill,
}

// ХЕНДЛЕР ПОД ПРЯМОЙ УРОН В СОЛО ТАРГЕТ
func CastSingleDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or caster")
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
	if !a.AttackHero && a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	if a.AttackHero {
		if !caster.Skill.IgnoreTank && enemyHasTank(enemy) {
			return ErrCardSkillTargetTankBlocked
		}
		damage := caster.Skill.Power
		enemy.HeroHP -= damage
		heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
		m.Events = append(m.Events, Event{
			Type:             string(EventCardSkill),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
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
			Type:             string(EventCardSkill),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
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

// ХЕНДЛЕР ПОД УРОН СПЛЕШЕМ ПО СТОЛУ. НЕЛЬЗЯ ИМ БИТЬ ГЕРОЯ
func CastSplashDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or caster")
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if a.AttackHero {
		return fmt.Errorf("%s cant attack hero", caster.Skill.Name)
	}
	if a.TargetInstanceID == "" {
		return ErrCardSkillBadTarget
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	targetSlot, target := enemy.FindSlot(a.TargetInstanceID)
	if target == nil || targetSlot < 0 {
		return ErrCardSkillBadTarget
	}
	if !caster.Skill.IgnoreTank && enemyHasTank(enemy) && !target.IsTank {
		return ErrCardSkillTargetTankBlocked
	}
	targetSlots := make([]int, 0, 3)
	targetSlots = append(targetSlots, targetSlot)
	left := targetSlot - 1
	right := targetSlot + 1
	if left >= 0 && enemy.Table[left] != nil {
		targetSlots = append(targetSlots, left)
	}
	if right < TableSize && enemy.Table[right] != nil {
		targetSlots = append(targetSlots, right)
	}
	targets := make([]EventTarget, 0, len(targetSlots))
	baseDamage := caster.Skill.Power
	for _, slot := range targetSlots {
		u := enemy.Table[slot]
		if u == nil {
			continue
		}
		damage := baseDamage
		if slot != targetSlot {
			damage = damage / 2
		}
		inst := u.InstanceID
		tplID := u.TemplateID
		u.HP -= damage
		died := u.HP <= 0
		newHP := u.HP
		if died {
			if err := killUnitAt(m, 1-a.PlayerIndex, slot, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
		}
		targets = append(targets, EventTarget{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     damage,
			Died:       died,
			NewHP:      newHP,
		})
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// ХЕНДЛЕР ПО УРОНУ ВСЕМУ СТОЛУ НАХУЙ. НЕЛЬЗЯ БИТЬ ГЕРОЯ!
func CastAllEnemiesDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or caster")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	if a.AttackHero {
		return ErrCardSkillBadTarget
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	targets := make([]EventTarget, 0, TableSize)
	damage := caster.Skill.Power
	hasTarget := false
	for slot := 0; slot < TableSize; slot++ {
		u := enemy.Table[slot]
		if u == nil {
			continue
		}
		hasTarget = true
		inst := u.InstanceID
		tplID := u.TemplateID
		u.HP -= damage
		died := u.HP <= 0
		newHP := u.HP
		if died {
			if err := killUnitAt(m, 1-a.PlayerIndex, slot, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
		}
		targets = append(targets, EventTarget{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     damage,
			Died:       died,
			NewHP:      newHP,
		})
	}
	if !hasTarget {
		return ErrCardSkillBadTarget
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// КАСТУЕМ УРОН ПО СЛУЧАЙНОМУ ПРОТИВНИКУ (МБ И ГЕРОЮ)
func CastRandomSingleEnemyDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	targetSlots := make([]int, 0, TableSize)
	for slot := 0; slot < TableSize; slot++ {
		if enemy.Table[slot] != nil {
			targetSlots = append(targetSlots, slot)
		}
	}
	canHitHero := caster.Skill.IgnoreTank || !enemyHasTank(enemy)
	targetPoolSize := len(targetSlots)
	if canHitHero {
		targetPoolSize++
	}
	if targetPoolSize == 0 {
		return ErrCardSkillBadTarget
	}
	roll := rand.IntN(targetPoolSize)
	damage := caster.Skill.Power
	if canHitHero && roll == len(targetSlots) {
		enemy.HeroHP -= damage
		heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
		m.Events = append(m.Events, Event{
			Type:             string(EventCardSkill),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
			VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
			SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
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
		targetSlot := targetSlots[roll]
		target := enemy.Table[targetSlot]
		if target == nil {
			return nil
		}
		inst := target.InstanceID
		tplID := target.TemplateID
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
			Type:             string(EventCardSkill),
			PlayerIndex:      a.PlayerIndex,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: caster.InstanceID,
			SourceTemplateID: caster.TemplateID,
			VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
			SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
			Targets: []EventTarget{
				{
					InstanceID: inst,
					TemplateID: tplID,
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

// КАСТУЕМ СКИЛЛ НЕСКОЛЬКО РАЗ ПО СЛУЧАЙНЫМ ЦЕЛЯМ
func CastRandomMultiplEnemyDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	type randomTarget struct {
		isHero bool
		slot   int
	}
	pool := make([]randomTarget, 0, TableSize+1)
	for slot := 0; slot < TableSize; slot++ {
		if enemy.Table[slot] != nil {
			pool = append(pool, randomTarget{slot: slot})
		}
	}
	canHitHero := caster.Skill.IgnoreTank || !enemyHasTank(enemy)
	if canHitHero {
		pool = append(pool, randomTarget{isHero: true})
	}
	if len(pool) == 0 {
		return ErrCardSkillBadTarget
	}
	hits := caster.Skill.ApplyCount
	if hits <= 0 {
		hits = 1
	}
	damage := caster.Skill.Power
	targets := make([]EventTarget, 0, hits)
	for i := 0; i < hits; i++ {
		if len(pool) == 0 {
			break
		}
		pick := rand.IntN(len(pool))
		rt := pool[pick]
		if rt.isHero {
			enemy.HeroHP -= damage
			heroID := fmt.Sprintf("hero:p%d", 1-a.PlayerIndex)
			targets = append(targets, EventTarget{
				InstanceID: heroID,
				Amount:     damage,
				Died:       enemy.HeroHP <= 0,
				NewHP:      enemy.HeroHP,
			})
			if enemy.HeroHP <= 0 {
				m.Finished = true
				if a.PlayerIndex == 0 {
					m.Result = MatchWinP1
				} else {
					m.Result = MatchWinP2
				}
				pool = append(pool[:pick], pool[pick+1:]...)
				break
			}
			continue
		}
		target := enemy.Table[rt.slot]
		if target == nil {
			pool = append(pool[:pick], pool[pick+1:]...)
			i--
			continue
		}
		inst := target.InstanceID
		tplID := target.TemplateID
		target.HP -= damage
		died := target.HP <= 0
		newHP := target.HP
		if died {
			if err := killUnitAt(m, 1-a.PlayerIndex, rt.slot, caster.InstanceID, a.PlayerIndex); err != nil {
				return err
			}
			newHP = 0
			pool = append(pool[:pick], pool[pick+1:]...)
		}
		targets = append(targets, EventTarget{
			InstanceID: inst,
			TemplateID: tplID,
			Amount:     damage,
			Died:       died,
			NewHP:      newHP,
		})
	}
	if len(targets) == 0 {
		return ErrCardSkillBadTarget
	}
	m.Events = append(m.Events, Event{
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets:          targets,
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
	return nil
}

// КАСТУЕМ СКИЛЛ ПО САМОЙ СЛАБОЙ ПО ХП ЦЕЛИ
func CastLowestHPDamageSkill(m *MatchState, a Action, caster *UnitState) error {
	if m == nil || caster == nil {
		return errors.New("nil match or casters")
	}
	if caster.Skill.Code == "" {
		return ErrCardSkillNotFound
	}
	if caster.Skill.CooldownLeft > 0 {
		return ErrCardSkillOnCooldown
	}
	enemy := m.Players[1-a.PlayerIndex]
	if enemy == nil {
		return errors.New("nil enemy state")
	}
	if a.AttackHero {
		return ErrCardSkillBadTarget
	}
	targetSlot := -1
	var target *UnitState
	for slot := 0; slot < TableSize; slot++ {
		u := enemy.Table[slot]
		if u == nil {
			continue
		}
		if target == nil || u.HP < target.HP {
			target = u
			targetSlot = slot
		}
	}
	if target == nil || targetSlot == -1 {
		return ErrCardSkillBadTarget
	}
	damage := caster.Skill.Power
	inst := target.InstanceID
	tplID := target.TemplateID
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
		Type:             string(EventCardSkill),
		PlayerIndex:      a.PlayerIndex,
		SourceKind:       string(SourceUnit),
		SourceInstanceID: caster.InstanceID,
		SourceTemplateID: caster.TemplateID,
		VFXKey:           BuildVFXKey(caster.AssetBaseKey, "spell"),
		SFXKey:           BuildSFXKey(caster.AssetBaseKey, "spell"),
		Targets: []EventTarget{
			{
				InstanceID: inst,
				TemplateID: tplID,
				Amount:     damage,
				Died:       died,
				NewHP:      newHP,
			},
		},
	})
	caster.Skill.CooldownLeft = caster.Skill.BaseCooldown
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
