package game

// import "TheWar/internal/domain/cards"

// // здесь все про пассивные ауры (не скиллы)

// // Аура под пассивное увеличение дамага
// func applyPassiveAuraAttackBuff(targets []*UnitState, value int) {
// 	for _, u := range targets {
// 		if u == nil {
// 			continue
// 		}
// 		AddEffect(u, UnitEffect{
// 			EffectType: cards.AuraDamageUpdate,
// 			TurnsLeft:  0,
// 			Value:      value,
// 		})
// 	}
// }

// // Аура под пассивное увеличение максимального кол-ва здоровья
// func applyPassiveAuraMaxHealthPointsUpdate(targets []*UnitState, value int) {
// 	for _, u := range targets {
// 		if u == nil {
// 			continue
// 		}
// 		AddEffect(u, UnitEffect{
// 			EffectType: cards.AuraMaxHealthPointsUpdate,
// 			TurnsLeft:  0,
// 			Value:      value,
// 		})
// 	}
// }

// // Аура под пассивное уменьшение КД атаки
// func applyPassiveAuraCooldownUpdate(targets []*UnitState, value int) {
// 	for _, u := range targets {
// 		if u == nil {
// 			continue
// 		}
// 		AddEffect(u, UnitEffect{
// 			EffectType: cards.AuraCooldownUpdate,
// 			TurnsLeft:  0,
// 			Value:      value,
// 		})
// 	}
// }

// // Аура под пассивное увеличение урона скилла
// func applyPassiveAuraSkillDamageUpdate(targets []*UnitState, value int) {
// 	for _, u := range targets {
// 		if u == nil {
// 			continue
// 		}
// 		AddEffect(u, UnitEffect{
// 			EffectType: cards.AuraSkillDamageUpdate,
// 			TurnsLeft:  0,
// 			Value:      value,
// 		})
// 	}
// }

// // Аура под пассивное снижение КД скилла
// func applyPassiveAuraSkillCooldownUpdate(targets []*UnitState, value int) {
// 	for _, u := range targets {
// 		if u == nil {
// 			continue
// 		}
// 		AddEffect(u, UnitEffect{
// 			EffectType: cards.AuraSkillCooldownUpdate,
// 			TurnsLeft:  0,
// 			Value:      value,
// 		})
// 	}
// }

// /*
// ------ХЕЛПЕРЫ К ПАССИВНЫМ АУРАМ------
// */
// func clearContinuousPassiveEffects(m *MatchState) {
// 	if m == nil {
// 		return
// 	}
// 	for playerIdx := 0; playerIdx < 2; playerIdx++ {
// 		p := m.Players[playerIdx]
// 		if p == nil {
// 			continue
// 		}
// 		for i := 0; i < TableSize; i++ {
// 			u := p.Table[i]
// 			if u == nil {
// 				continue
// 			}
// 			kept := make([]UnitEffect, 0, len(u.Effects))
// 			for _, e := range u.Effects {
// 				if isContinuousAuraEffect(e.EffectType) {
// 					RemoveEffect(u, e)
// 					continue
// 				}
// 				kept = append(kept, e)
// 			}
// 			u.Effects = kept
// 		}
// 	}
// }

// func isContinuousAuraEffect(effectType string) bool {
// 	switch effectType {
// 	case cards.AuraDamageUpdate:
// 		return true
// 	case cards.AuraMaxHealthPointsUpdate:
// 		return true
// 	case cards.AuraCooldownUpdate:
// 		return true
// 	case cards.AuraSkillDamageUpdate:
// 		return true
// 	case cards.AuraSkillCooldownUpdate:
// 		return true
// 	default:
// 		return false
// 	}
// }

// func triggerContinuousPassive(m *MatchState, ownerIdx int, source *UnitState) error {
// 	if m == nil || source == nil {
// 		return nil
// 	}
// 	if source.PassiveCode != "" {
// 		if source.PassiveTrigger == cards.PassiveTriggerContinuous {
// 			h, err := getPassiveSkillHandler(source.PassiveCode)
// 			if err != nil {
// 				return err
// 			}
// 			return h(m, ownerIdx, source, cards.PassiveTriggerContinuous, PassiveTriggerContext{})
// 		}
// 	}
// 	return nil
// }

// func prepareContinuousPassive(m *MatchState, ownerIdx int, source *UnitState) (passivePrepared, bool) {
// 	if m == nil || ownerIdx > 1 || ownerIdx < 0 || source == nil {
// 		return passivePrepared{}, false
// 	}
// 	if source.PassiveTrigger == cards.PassiveTriggerContinuous {
// 		matched := passiveMatchedCount(m, ownerIdx, source)
// 		if !passiveConditionOK(m, ownerIdx, source, matched) {
// 			return passivePrepared{}, false
// 		}
// 		value := calcPassiveFinalValue(source, matched)
// 		if value == 0 {
// 			return passivePrepared{}, false
// 		}
// 		targets := passiveTargets(m, ownerIdx, source, PassiveTriggerContext{})
// 		if len(targets) == 0 {
// 			return passivePrepared{}, false
// 		}
// 		return passivePrepared{targets: targets, value: value}, true
// 	}
// 	return passivePrepared{}, false
// }
