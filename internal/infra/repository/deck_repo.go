package repository

import (
	"TheWar/internal/domain/game"

	"gorm.io/gorm"
)

func LoadDeckTx(tx *gorm.DB, userID uint) ([]game.DeckEntry, error) {
	var rows []GamerDeckEntry
	if err := tx.Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]game.DeckEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, game.DeckEntry{
			Kind:       r.Kind,
			TemplateID: r.TemplateID,
			Count:      r.Count,
		})
	}
	return out, nil
}

func SaveDeckTx(tx *gorm.DB, userID uint, entries []game.DeckEntry) error {
	if err := tx.Where("gamer_id = ?", userID).Delete(&GamerDeckEntry{}).Error; err != nil {
		return err
	}
	rows := make([]GamerDeckEntry, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, GamerDeckEntry{
			GamerID:    userID,
			Kind:       e.Kind,
			TemplateID: e.TemplateID,
			Count:      e.Count,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.CreateInBatches(&rows, 200).Error
}
