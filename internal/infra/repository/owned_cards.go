package repository

import (
	"TheWar/internal/domain/cards"

	"gorm.io/gorm"
)

type OwnnedBattleCardRow struct {
	OwnedID uint
	Copies  int
	Level   int
	XP      int
	Tpl     cards.BattleCardTemplate
}
type OwnedBuffCardRow struct {
	OwnedID uint
	Copies  int
	Level   int
	XP      int
	Tpl     cards.BuffCardsTemplate
}

func LoadOwnedBattleCardsRows(tx *gorm.DB, userID uint) ([]OwnnedBattleCardRow, error) {
	var rows []GamerBattleCards
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]OwnnedBattleCardRow, 0, len(rows))
	for _, t := range rows {
		out = append(out, OwnnedBattleCardRow{
			OwnedID: t.ID,
			Copies:  t.Copies,
			Level:   t.CardLevel,
			XP:      t.CardXP,
			Tpl:     t.CardTemplate,
		})
	}
	return out, nil
}

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
