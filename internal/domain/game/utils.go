package game

import (
	"TheWar/internal/domain/cards"
	"errors"
)

//ПЕРЕЛОПАТИТЬ

/*Файл посвящен функциям хелперам, которые так или иначе влияют на геймплей. Но я сейчас так подумал, наверное это
уничижительное определение, поскольку данные функции ебать как помогают... Ну вот к примеру*/

/*
Функция тикера. Именно она определеяет то, когда должен быть снят баф с конкретной карты. Так как бафы
(в основном) вещь временная, должен быть механизм, который снимает эффекты по истечению определенного времени.
Во входящих аргументах принимаем состояние плеера, откуда и будем брать инфу о количестве ходов.
*/
func TickerEffects(m *MatchState, ownerIdx int) error {
	if m == nil || ownerIdx < 0 || ownerIdx > 1 {
		return errors.New("nil match or bad owner index")
	}
	player := m.Players[ownerIdx]
	if player == nil {
		return errors.New("bad player index")
	}
	for i := 0; i < TableSize; i++ {
		u := player.Table[i]
		if u == nil {
			continue
		}
		out := u.Effects[:0]
		unitDied := false
		for _, e := range u.Effects {
			if e.TurnsLeft == 0 {
				out = append(out, e)
				continue
			}
			switch e.EffectType {
			case cards.BuffEffectHealPerTurn:
				u.HP += e.Value
				if u.HP > u.MaxHP {
					u.HP = u.MaxHP
				}
			case cards.DebuffEffectDamageOverTime:
				u.HP -= e.Value
				if u.HP <= 0 {
					if err := killUnitAt(m, ownerIdx, i, "", ownerIdx); err != nil {
						return err
					}
					unitDied = true
				}
			}
			if unitDied {
				break
			}
			e.TurnsLeft--
			if e.TurnsLeft <= 0 {
				switch e.EffectType {
				case cards.BuffEffectAttack,
					cards.BuffEffectHP,
					cards.BuffEffectAttackCooldown,
					cards.BuffEffectSkillCooldown,
					cards.BuffEffectMakeTank,
					cards.DebuffEffectAttackDown,
					cards.DebuffEffectCooldownUp,
					cards.DebuffEffectSkillCooldownUp:
					_ = RemoveEffect(u, e)
				}
				continue
			}
			out = append(out, e)
		}
		if unitDied {
			continue
		}
		u.Effects = out
	}
	return nil
}

/*Таким образом данная функция реализует собой простой тикер, который считает счетчик TurnsLeft, каждый раз перезаписывая
баф с новым, уже обновленным счетчиком. Умнее не придумал. Сори*/

/*
Функция добавления эффекта на карту. По сути тут все просто. Ок, разжую.
Данная функция сохраняет эффект в UnitState, тем самым бафая определенные
характеристики. Ура
*/
func AddEffect(u *UnitState, e UnitEffect) error {
	if u == nil {
		return errors.New("nil unit state")
	}
	if err := ApplyEffect(u, e); err != nil {
		return err
	}
	u.Effects = append(u.Effects, e)
	return nil
}

/*Функция удаления эжффекта из UnitState. Принимаем собственно UnitState и эффект. Круто*/
func RemoveEffect(u *UnitState, e UnitEffect) error {
	if u == nil { //<-проверяем, чтобы избежать паник
		return errors.New("nil unit state")
	}
	switch e.EffectType {
	case cards.BuffEffectAttack:
		u.Attack -= e.Value
		if u.Attack < 0 {
			u.Attack = 0
		}
	case cards.BuffEffectHP:
		u.MaxHP -= e.Value
		if u.MaxHP < 1 {
			u.MaxHP = 1
		}
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
	case cards.BuffEffectAttackCooldown:
		u.BaseCooldown += e.Value
		if u.BaseCooldown < 1 {
			u.BaseCooldown = 1
		}
		if u.Cooldown > u.BaseCooldown {
			u.Cooldown = u.BaseCooldown
		}
	case cards.BuffEffectSkillCooldown:
		u.Skill.BaseCooldown += e.Value
		if u.Skill.BaseCooldown < 1 {
			u.Skill.BaseCooldown = 1
		}
		if u.Skill.CooldownLeft > u.Skill.BaseCooldown {
			u.Skill.CooldownLeft = u.Skill.BaseCooldown
		}
	case cards.BuffEffectMakeTank:
		u.IsTank = false
	case cards.DebuffEffectAttackDown:
		u.Attack += e.Value
	case cards.DebuffEffectCooldownUp:
		u.BaseCooldown -= e.Value
		if u.BaseCooldown < 1 {
			u.BaseCooldown = 1
		}
		if u.Cooldown > u.BaseCooldown {
			u.Cooldown = u.BaseCooldown
		}
		if u.Cooldown < 0 {
			u.Cooldown = 0
		}
	case cards.DebuffEffectSkillCooldownUp:
		u.Skill.BaseCooldown -= e.Value
		if u.Skill.BaseCooldown < 1 {
			u.Skill.BaseCooldown = 1
		}
		if u.Skill.CooldownLeft > u.Skill.BaseCooldown {
			u.Skill.CooldownLeft = u.Skill.BaseCooldown
		}
		if u.Skill.CooldownLeft < 0 {
			u.Skill.CooldownLeft = 0
		}
	case cards.BuffEffectHealPerTurn,
		cards.BuffEffectShield,
		cards.BuffEffectReflectShield,
		cards.BuffEffectRedirectDamage,
		cards.BuffEffectDamageReduction,
		cards.BuffEffectOverdrive,
		cards.BuffEffectMulticast,
		cards.BuffEffectVampiricStrike,
		cards.BuffEffectChainAttack,
		cards.BuffEffectDeathExplosion,
		cards.BuffEffectDeathMassHeal,
		cards.BuffEffectCounterattack,
		cards.BuffEffectLifeOnHit,
		cards.BuffEffectBonusAfterAttack,
		cards.DebuffEffectDamageOverTime,
		cards.DebuffEffectSilence,
		cards.DebuffEffectNoHeal,
		cards.DebuffEffectVulnerable,
		cards.DebuffEffectDisarm,
		cards.DebuffEffectStun:
		return nil
	default:
		return nil
	}
	return nil
}

/*Место для применения моментальных жффектов, так или иначе менябщих статы сразу, на месте. Могут быть откатаны через Remove*/
func ApplyEffect(u *UnitState, buff UnitEffect) error {
	if u == nil {
		return errors.New("nil unit state")
	}
	switch buff.EffectType {
	case cards.BuffEffectAttack:
		u.Attack += buff.Value
	case cards.BuffEffectHP:
		u.MaxHP += buff.Value
		if u.MaxHP < 1 {
			u.MaxHP = 1
		}
		u.HP += buff.Value
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
	case cards.BuffEffectAttackCooldown:
		u.BaseCooldown -= buff.Value
		if u.BaseCooldown < 1 {
			u.BaseCooldown = 1
		}
		if u.Cooldown > u.BaseCooldown {
			u.Cooldown = u.BaseCooldown
		}
		if u.Cooldown < 0 {
			u.Cooldown = 0
		}
	case cards.BuffEffectSkillCooldown:
		u.Skill.BaseCooldown -= buff.Value
		if u.Skill.BaseCooldown < 1 {
			u.Skill.BaseCooldown = 1
		}
		if u.Skill.CooldownLeft > u.Skill.BaseCooldown {
			u.Skill.CooldownLeft = u.Skill.BaseCooldown
		}
		if u.Skill.CooldownLeft < 0 {
			u.Skill.CooldownLeft = 0
		}
	case cards.BuffEffectMakeTank:
		if u.IsTank {
			return errors.New("card is tank already")
		}
		u.IsTank = true
	case cards.DebuffEffectAttackDown:
		u.Attack -= buff.Value
	case cards.DebuffEffectCooldownUp:
		u.BaseCooldown += buff.Value
		if u.BaseCooldown < 1 {
			u.BaseCooldown = 1
		}
		if u.Cooldown > u.BaseCooldown {
			u.Cooldown = u.BaseCooldown
		}
	case cards.DebuffEffectSkillCooldownUp:
		u.Skill.BaseCooldown += buff.Value
		if u.Skill.BaseCooldown < 1 {
			u.Skill.BaseCooldown = 1
		}
		if u.Skill.CooldownLeft > u.Skill.BaseCooldown {
			u.Skill.CooldownLeft = u.Skill.BaseCooldown
		}
	case cards.BuffEffectHealPerTurn,
		cards.BuffEffectShield,
		cards.BuffEffectReflectShield,
		cards.BuffEffectRedirectDamage,
		cards.BuffEffectDamageReduction,
		cards.BuffEffectOverdrive,
		cards.BuffEffectMulticast,
		cards.BuffEffectVampiricStrike,
		cards.BuffEffectChainAttack,
		cards.BuffEffectDeathExplosion,
		cards.BuffEffectDeathMassHeal,
		cards.BuffEffectCounterattack,
		cards.BuffEffectLifeOnHit,
		cards.BuffEffectBonusAfterAttack,
		cards.DebuffEffectDamageOverTime,
		cards.DebuffEffectSilence,
		cards.DebuffEffectNoHeal,
		cards.DebuffEffectVulnerable,
		cards.DebuffEffectDisarm,
		cards.DebuffEffectStun:
		return nil
	default:
		return nil
	}
	return nil
}

// Хелпер, который позволяет стабилизировать КД карты
func clampUnitCooldown(u *UnitState) {
	if u == nil {
		return
	}
	if u.BaseCooldown < 1 {
		u.BaseCooldown = 1
	}
	if u.Cooldown < 1 {
		u.Cooldown = 1
	}
}

/*Таким образом работают хелп функции вокруг основных функций (которые мы написали в turn.go). Тут происходит
вся движуха и естественно то, что коли я решу добавить какой-либо баф, он неминуемо должен быть описан тут. К
примеру -ебануть весь стол танками на Х ходов. Звучит бредово, да, но что делать...*/
