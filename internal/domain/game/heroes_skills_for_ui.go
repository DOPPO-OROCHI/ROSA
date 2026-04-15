package game

// import (
// 	"TheWar/internal/domain/heroes"
// 	"errors"
// 	"fmt"
// )

// // логика под UI, для того, чтобы понять, какие юниты были затронуты способностью
// type HeroSellSnap struct {
// 	playerIndex int
// 	isHero      bool
// 	slot        int
// 	instID      string
// 	tplID       string
// 	hpBefore    int
// }

// func buildHeroSpellShanpsBefore(m *MatchState, a Action, spec heroes.AbilitySpec) ([]HeroSellSnap, error) {
// 	if m == nil {
// 		return nil, errors.New("nil match")
// 	}
// 	snaps := make([]HeroSellSnap, 0, 3)
// 	snapUnit := func(pi int, inst string) (HeroSellSnap, bool) {
// 		p := m.Players[pi]
// 		if p == nil {
// 			return HeroSellSnap{}, false
// 		}
// 		slot, u := p.FindSlot(inst)
// 		if u == nil || slot < 0 {
// 			return HeroSellSnap{}, false
// 		}
// 		return HeroSellSnap{
// 			playerIndex: pi,
// 			isHero:      false,
// 			slot:        slot,
// 			instID:      u.InstanceID,
// 			tplID:       u.TemplateID,
// 			hpBefore:    u.HP,
// 		}, true
// 	}
// 	snapHero := func(pi int) HeroSellSnap {
// 		p := m.Players[pi]
// 		hp := 0
// 		if p != nil {
// 			hp = p.HeroHP
// 		}
// 		return HeroSellSnap{
// 			playerIndex: pi,
// 			isHero:      true,
// 			slot:        -1,
// 			instID:      fmt.Sprintf("hero:p%d", pi),
// 			tplID:       "",
// 			hpBefore:    hp,
// 		}
// 	}
// 	switch spec.Code {
// 	case heroes.ATTACK_ANY:
// 		defPI := 1 - a.PlayerIndex
// 		if a.AttackHero {
// 			snaps = append(snaps, snapHero(defPI))
// 			return snaps, nil
// 		}
// 		s, ok := snapUnit(defPI, a.TargetInstanceID)
// 		if !ok {
// 			return nil, ErrDefenderNotFound
// 		}
// 		snaps = append(snaps, s)
// 		return snaps, nil
// 	case heroes.ATTACK_SPLASH:
// 		defPI := 1 - a.PlayerIndex
// 		p := m.Players[defPI]
// 		if p == nil {
// 			return nil, errors.New("nil enemy player state")
// 		}
// 		centerSlot, center := p.FindSlot(a.TargetInstanceID)
// 		if center == nil || centerSlot < 0 {
// 			return nil, ErrDefenderNotFound
// 		}
// 		snaps = append(snaps, HeroSellSnap{
// 			playerIndex: defPI, isHero: false, slot: centerSlot,
// 			instID: center.InstanceID, tplID: center.TemplateID,
// 			hpBefore: center.HP,
// 		})
// 		left, right := centerSlot-1, centerSlot+1
// 		if left >= 0 && p.Table[left] != nil {
// 			u := p.Table[left]
// 			snaps = append(snaps, HeroSellSnap{
// 				playerIndex: defPI, isHero: false, slot: left,
// 				instID: u.InstanceID, tplID: u.TemplateID,
// 				hpBefore: u.HP,
// 			})
// 		}
// 		if right < TableSize && p.Table[right] != nil {
// 			u := p.Table[right]
// 			snaps = append(snaps, HeroSellSnap{
// 				playerIndex: defPI, isHero: false, slot: right,
// 				instID: u.InstanceID, tplID: u.TemplateID,
// 				hpBefore: u.HP,
// 			})
// 		}
// 		return snaps, nil
// 	case heroes.HEAL_UNIT:
// 		s, ok := snapUnit(a.PlayerIndex, a.TargetInstanceID)
// 		if !ok {
// 			return nil, ErrTargetNotFound
// 		}
// 		snaps = append(snaps, s)
// 		return snaps, nil
// 	default:
// 		if a.AttackHero {
// 			snaps = append(snaps, snapHero(1-a.PlayerIndex))
// 			return snaps, nil
// 		}
// 		if s, ok := snapUnit(1-a.PlayerIndex, a.TargetInstanceID); ok {
// 			snaps = append(snaps, s)
// 			return snaps, nil
// 		}
// 		if s, ok := snapUnit(a.PlayerIndex, a.TargetInstanceID); ok {
// 			snaps = append(snaps, s)
// 			return snaps, nil
// 		}
// 		return nil, ErrHeroAbilityBadTarget
// 	}
// }

// func buildHeroSpellTargetsAfter(m *MatchState, spec heroes.AbilitySpec, snaps []HeroSellSnap) []EventTarget {
// 	out := make([]EventTarget, 0, len(snaps))
// 	for _, s := range snaps {
// 		p := m.Players[s.playerIndex]
// 		if s.isHero {
// 			newHP := 0
// 			if p != nil {
// 				newHP = p.HeroHP
// 			}
// 			amt := spec.Value
// 			if spec.Code == heroes.ATTACK_ANY || spec.Code == heroes.ATTACK_SPLASH {
// 				d := s.hpBefore - newHP
// 				if d < 0 {
// 					d = -d
// 				}
// 				amt = d
// 			}
// 			if spec.Code == heroes.HEAL_UNIT {
// 				d := newHP - s.hpBefore
// 				if d < 0 {
// 					d = -d
// 				}
// 				amt = d
// 			}
// 			out = append(out, EventTarget{
// 				InstanceID: s.instID,
// 				Amount:     amt,
// 				Died:       newHP <= 0,
// 				NewHP:      newHP,
// 			})
// 			continue
// 		}
// 		newHP := 0
// 		died := true
// 		tplID := s.tplID
// 		if p != nil {
// 			_, u := p.FindSlot(s.instID)
// 			if u != nil {
// 				newHP = u.HP
// 				died = newHP <= 0
// 				tplID = u.TemplateID
// 			}
// 		}
// 		amt := spec.Value
// 		if spec.Code == heroes.ATTACK_ANY || spec.Code == heroes.ATTACK_SPLASH {
// 			d := s.hpBefore - newHP
// 			if d < 0 {
// 				d = -d
// 			}
// 			amt = d
// 		}
// 		if spec.Code == heroes.HEAL_UNIT {
// 			d := newHP - s.hpBefore
// 			if d < 0 {
// 				d = -d
// 			}
// 			amt = d
// 		}
// 		if !died && newHP > 0 {
// 			died = false
// 		}
// 		out = append(out, EventTarget{
// 			InstanceID: s.instID,
// 			TemplateID: tplID,
// 			Amount:     amt,
// 			Died:       died,
// 			NewHP:      newHP,
// 		})
// 	}
// 	return out
// }

// func findTargetSlotForHeroSpell(m *MatchState, a Action, spec heroes.AbilitySpec) int {
// 	if a.AttackHero || a.TargetInstanceID == "" {
// 		return -1
// 	}
// 	switch spec.Target {
// 	case heroes.OWN_UNIT:
// 		p := m.Players[a.PlayerIndex]
// 		if p == nil {
// 			return -1
// 		}
// 		slot, _ := p.FindSlot(a.TargetInstanceID)
// 		return slot
// 	case heroes.ENEMY_UNIT, heroes.ENEMY_ANY:
// 		ep := m.Players[1-a.PlayerIndex]
// 		if ep == nil {
// 			return -1
// 		}
// 		slot, _ := ep.FindSlot(a.TargetInstanceID)
// 		return slot
// 	default:
// 		return -1
// 	}
// }
