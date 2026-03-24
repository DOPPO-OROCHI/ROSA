package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

// Хелпер для тех карт, смысл которых заключается в тригере скила при смерти
func triggerOnDeathSkill(m *MatchState, dead *UnitState, deadOwnedIndex int) error {
	if m == nil || dead == nil {
		return nil
	}
	if dead.SkillCode == "" || dead.SkillTrigger != cards.TriggerOnDeath {
		return nil
	}
	owner := m.Players[deadOwnedIndex]
	enemy := m.Players[1-deadOwnedIndex]
	if owner == nil || enemy == nil {
		return nil
	}
	h, err := getCardSkillHandler(dead.SkillCode)
	if err != nil {
		return err
	}
	a := Action{
		PlayerIndex:    deadOwnedIndex,
		CardInstanceID: dead.InstanceID,
	}
	return h(m, a, dead, owner, enemy)
}

func triggerCardSkillByTrigger(m *MatchState, ownerIdx int, caster *UnitState, trigger string, a Action) error {
	if m == nil {
		return nil
	}
	if caster == nil {
		return nil
	}
	if caster.SkillCode == "" || caster.SkillTrigger != trigger {
		return nil
	}
	owner := m.Players[ownerIdx]
	enemy := m.Players[1-ownerIdx]
	if owner == nil || enemy == nil {
		return nil
	}
	if a.PlayerIndex != ownerIdx {
		a.PlayerIndex = ownerIdx
	}
	if a.CardInstanceID == "" {
		a.CardInstanceID = caster.InstanceID
	}
	h, err := getCardSkillHandler(caster.SkillCode)
	if err != nil {
		return nil
	}
	return h(m, a, caster, owner, enemy)
}

// ФУНККЦИЯ ОБРАБОТКИ СМЕРТИ (ПАССИВКА, ПОДНИМАЕТ КАРТЫ ИЗ МОГИЛЫ НА СЛЕДУЮЩИЙ ХОД)
func processPendingResurections(m *MatchState, ownerIdx int) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if ownerIdx < 0 || ownerIdx > 1 {
		return errors.New("bad owner index")
	}
	owner := m.Players[ownerIdx]
	if owner == nil {
		return errors.New("bad owner")
	}
	nextPending := make([]PendingResurrected, 0, len(owner.PendingRes))
	for i := range owner.PendingRes {
		pr := owner.PendingRes[i]
		if pr.DueTurn > owner.Turns {
			nextPending = append(nextPending, pr)
			continue
		}
		graveIdx := -1
		for j := range owner.GraveYard {
			if owner.GraveYard[j].Unit.InstanceID == pr.InstanceID {
				graveIdx = j
				break
			}
		}
		if graveIdx < 0 {
			continue
		}
		slot := firstFreeSlot(owner)
		if slot < 0 {
			nextPending = append(nextPending, pr)
			continue
		}
		revived := owner.GraveYard[graveIdx].Unit
		revived.HP /= 2
		revived.ResurrectedUsed = true
		revived.SummonedInTurn = owner.Turns
		owner.Table[slot] = &revived
		last := len(owner.GraveYard) - 1
		owner.GraveYard[graveIdx] = owner.GraveYard[last]
		owner.GraveYard = owner.GraveYard[:last]
		m.Events = append(m.Events, Event{
			Type:             string(EventResurrect),
			PlayerIndex:      ownerIdx,
			SourceKind:       string(SourceUnit),
			SourceInstanceID: revived.InstanceID,
			SourceTemplateID: revived.TemplateID,
			TargetSlot:       slot,
			Targets: []EventTarget{
				{
					InstanceID: revived.InstanceID,
					TemplateID: revived.TemplateID,
					Died:       false,
					NewHP:      revived.HP,
				},
			},
		})
	}
	owner.PendingRes = nextPending
	return nil
}
