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
func TickerEffects(m *MatchState, ownerIdx int) {
	if m == nil || ownerIdx < 0 || ownerIdx > 1 {
		return
	}
	p := m.Players[ownerIdx]
	if p == nil {
		return
	}
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
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
			if u.HP <= 0 {
				_ = killUnitAt(m, ownerIdx, i, "", ownerIdx)
				unitDied = true
				break
			}
			e.TurnsLeft--
			if e.TurnsLeft <= 0 {
				switch e.EffectType {
				case cards.DamageUpdate,
					cards.HealthPointsUpdate,
					cards.MaxHealthPointsUpdate,
					cards.CoolDownUpdate,
					cards.MakeTankUpdate,
					cards.SkillDamageUpdate,
					cards.SkillCooldownUpdate:
					RemoveEffect(u, e)
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
}

/*Таким образом данная функция реализует собой простой тикер, который считает счетчик TurnsLeft, каждый раз перезаписывая
баф с новым, уже обновленным счетчиком. Умнее не придумал. Сори*/

/*
Функция добавления эффекта на карту. По сути тут все просто. Ок, разжую.
Данная функция сохраняет эффект в UnitState, тем самым бафая определенные
характеристики. Ура
*/
func AddEffect(u *UnitState, e UnitEffect) {
	u.Effects = append(u.Effects, e)
	ApplyEffect(u, e) //<-ниже пояснение
}

/*Функция удаления эжффекта из UnitState. Принимаем собственно UnitState и эффект. Круто*/
func RemoveEffect(u *UnitState, e UnitEffect) {
	if u == nil { //<-проверяем, чтобы избежать паник
		return
	}
	switch e.EffectType { //<-свичим тип эффекта, который есть уже на карте
	case cards.DamageUpdate: //<-кейс с апдейтом дамага
		u.Attack -= e.Value //<-и здесь просто минусуем то, что прибавляли ранее
	case cards.HealthPointsUpdate:
		u.HP -= e.Value //<-а тут тонкий момент. Дело в том, что юнит может здохнуть изза этой темы. Но такова механика...
		if u.HP < 0 {
			u.HP = 0 //<-и да, обработка чтобы не уйти в минус
		}
	case cards.MaxHealthPointsUpdate:
		u.MaxHP -= e.Value
		if u.MaxHP < 1 {
			u.MaxHP = 1
		}
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
	case cards.CoolDownUpdate:
		u.BaseCooldown += e.Value //<-возвращаем базовый кд, а не раздуваем текущий remaining cooldown.
		clampUnitCooldown(u)
	case cards.SkillDamageUpdate:
		u.Skill.Power -= e.Value
		if u.Skill.Power < 0 {
			u.Skill.Power = 0
		}
	case cards.SkillCooldownUpdate:
		u.Skill.BaseCooldown += e.Value
	case cards.MakeTankUpdate:
		u.IsTank = false //<-снимаем маркер танка с карты
	}
}

// а тут мы просто добавляем эффект на наш всеми любимый UntiState. НО! Тут нужно вернуть ошибку, поскольку бафы иногда не для всех
func ApplyEffect(u *UnitState, buff UnitEffect) error {
	if u == nil {
		return errors.New("nil unit state")
	}
	switch buff.EffectType { //<-свичим тип бафа
	case cards.DamageUpdate:
		u.Attack += buff.Value //<-в случае если баф на атаку, поднимаем атаку на значение бафа
	case cards.HealthPointsUpdate:
		u.HP += buff.Value //<-та же тема только с хп
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
	case cards.MaxHealthPointsUpdate:
		u.MaxHP += buff.Value
		if u.MaxHP < 1 {
			u.MaxHP = 1
		}
		u.HP += buff.Value
		if u.HP > u.MaxHP {
			u.HP = u.MaxHP
		}
	case cards.CoolDownUpdate:
		u.BaseCooldown -= buff.Value //<-на кд баф
		if u.Cooldown > 0 {
			u.Cooldown -= buff.Value
		}
		clampUnitCooldown(u)
	case cards.SkillDamageUpdate:
		u.Skill.Power += buff.Value
	case cards.SkillCooldownUpdate:
		u.Skill.BaseCooldown -= buff.Value
		if u.Skill.BaseCooldown < 0 {
			u.Skill.BaseCooldown = 0
		}
		if u.Skill.CooldownLeft > u.Skill.BaseCooldown {
			u.Skill.CooldownLeft = u.Skill.BaseCooldown
		}
	case cards.MakeTankUpdate: //<-а тут весело, поскольку на танк карту нельзя нанести баф мэйк танк
		if u.IsTank == true {
			return errors.New("card is tank type already") //<-это мы и обрабатываем
		}
		u.IsTank = true //<-а если все круто, делаем из карты танка
	}
	return nil
}

// Хелпер, который позволяет стабилизировать КД карты
func clampUnitCooldown(u *UnitState) {
	if u == nil {
		return
	}
	if u.BaseCooldown < 0 {
		u.BaseCooldown = 0
	}
	if u.Cooldown < 0 {
		u.Cooldown = 0
	}
}

/*Таким образом работают хелп функции вокруг основных функций (которые мы написали в turn.go). Тут происходит
вся движуха и естественно то, что коли я решу добавить какой-либо баф, он неминуемо должен быть описан тут. К
примеру -ебануть весь стол танками на Х ходов. Звучит бредово, да, но что делать...*/
