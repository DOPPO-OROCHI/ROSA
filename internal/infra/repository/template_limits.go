package repository

import (
	"TheWar/internal/domain/cards"

	"gorm.io/gorm"
)

func LoadTemplateLimits(tx *gorm.DB) (battleMax map[string]int, buffMax map[string]int, err error) {
	var battleTpl []cards.BattleCardTemplate
	if err := tx.Select("code_string", "max_copies").Find(&battleTpl).Error; err != nil {
		return nil, nil, err
	}
	battleMax = make(map[string]int, len(battleTpl))
	for _, t := range battleTpl {
		battleMax[t.CodeString] = t.MaxCopies
	}
	var buffTpl []cards.BuffCardsTemplate
	if err := tx.Select("code_string", "max_copies").Find(&buffTpl).Error; err != nil {
		return nil, nil, err
	}
	buffMax = make(map[string]int, len(buffTpl))
	for _, t := range buffTpl {
		buffMax[t.CodeString] = t.MaxCopies
	}
	return battleMax, buffMax, nil
}
