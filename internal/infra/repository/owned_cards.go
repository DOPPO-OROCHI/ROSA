package repository

import (
	"TheWar/internal/domain/cards"

	"gorm.io/gorm"
)

/*Файл посвящен функциям, которые так или иначе описывают карты игрока
Данные функции никак не участвуют в доменной логике и служат исключительно
для того, чтобы удобно складывать карты владения для DTO приколов*/

// структура описания боевых карт во владении
type OwnnedBattleCardRow struct {
	OwnedID uint
	Copies  int
	Level   int
	XP      int
	Tpl     cards.BattleCardTemplate
}

// структура описания баф карт во владении
type OwnedBuffCardRow struct {
	OwnedID uint
	Copies  int
	Level   int
	XP      int
	Tpl     cards.BuffCardsTemplate
}

/*
Функция служащая для выгрузки боевых карт владения. Здесь принимаем БД (блять заебался уже) и айдишник
пользователя, который и запрашивает свои карты чтобы посмотреть. Возвращаем же все в удобном массиве
*/
func LoadOwnedBattleCardsRows(tx *gorm.DB, userID uint) ([]OwnnedBattleCardRow, error) {
	var rows []GamerBattleCards //<-сюда будем все складывать
	/*Важно отметить Preload. Че он делает? Сама по себе функция подтягивает шаблон карты, притом делает
	это в рамках одного ORM прохода (добавляет к Find связанные таблицы шаблонов)*/
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil { //<-ищем карты
		return nil, err
	}
	out := make([]OwnnedBattleCardRow, 0, len(rows)) //<-а сюда добавляем все необходимое для DTO чтения
	for _, t := range rows {
		out = append(out, OwnnedBattleCardRow{
			OwnedID: t.ID,
			Copies:  t.Copies,
			Level:   t.CardLevel,
			XP:      t.CardXP,
			Tpl:     t.CardTemplate,
		})
	}
	return out, nil //<-отдаем всю хуйню
}

// Аналогичная муря и в загрузке баф карт
func LoadOwnedBuffCardsRows(tx *gorm.DB, userID uint) ([]OwnedBuffCardRow, error) {
	var rows []GamerBuffCards
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]OwnedBuffCardRow, 0, len(rows))
	for _, t := range rows {
		out = append(out, OwnedBuffCardRow{
			OwnedID: t.ID,
			Copies:  t.Copies,
			Level:   t.CardLevel,
			XP:      t.CardXP,
			Tpl:     t.CardTemplate,
		})
	}
	return out, nil
}

/*Таким образом реализована рид онли концепция выгрузки владения картами игрока. Че, читаем все из БД вместе с
шаблонами, с помощью Preload, потом все это трансформируем в удобный и приятный массив, откуда уже мапится вся
тема для HTTP/DTO. Нужно это, чтобы комфортно реализовать запрос /cards, без участия матчевой логики и изменения
состояния*/
