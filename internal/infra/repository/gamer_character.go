package repository

import (
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"errors"

	"gorm.io/gorm"
)

/*Данный файл посвящен выборке героя игрока (не владения, а лишь чертежа). Суть в чем? Перед игрой чувак должен
выбрать персонажа. Сам персонаж хранится в БД, вместе с его особенностями. Какого рода? Персонаж привязан к
конкретному игроку, чтобы иметь возможность так или иначе прокачивать его уровень, который влияет на здоровье
перса и силу его атаки. Так же у него есть XP, которые влияют на уровень прогрессии. Приступим к коду*/

// структура описывающая персонажа из владения, вместе с его айди, геймерайди (типа привязанный к игроку) и так далее
type GamerCharacter struct {
	gorm.Model
	//айди игрока
	GamerID uint `gorm:"not null;index;uniqueIndex:ux_gamer_character"`
	//сам игрок, вместе со всеми его полями
	Gamer player.TelegramUser `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	//айди чертежа персонажа
	CharacterTemplateID uint `gorm:"not null;index;uniqueIndex:ux_gamer_character"`
	//сам чертеж персонажа
	CharacterTemplate heroes.CharacterTemplate `gorm:"foreignKey:CharacterTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	//уровень персонажа
	CharacterLevel int `gorm:"not null;default:1"`
	//уровень прогрессии персонажа
	CharacterXP int `gorm:"not null;default:0"`
}

// Данная функция занята тем, что достает айди выбранного персонажа из таблицы игрока
func LoadSelectedHeroCodeTx(tx *gorm.DB, userID uint) (string, error) {
	var u player.TelegramUser                               //<-вводим переменную игрока
	if err := tx.Select("id", "selected_hero_template_id"). //<-а здесь будем доставать именно его персонажа
								Where("id = ?", userID).      //<-по айди пользователя
								First(&u).Error; err != nil { //<-записывая все в переменную
		return "", err
	}
	if u.SelectedHeroTemplateID == nil { //<-если персонаж не найден
		return "", errors.New("selected hero is not set") //<-то возвращаем ошибку
	}
	var tpl heroes.CharacterTemplate //<-а сюда будем записывать айдишник выбранного персонажа
	if err := tx.Select("character_code").
		Where("id = ?", *u.SelectedHeroTemplateID).
		First(&tpl).Error; err != nil {
		return "", err
	}
	//и если все круто, то возвращаем айди и пустоту, как сигнал о том, что все заебись
	return tpl.CharacterCode, nil
}
