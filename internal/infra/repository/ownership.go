package repository

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/player"

	"gorm.io/gorm"
)

type GamerBattleCards struct {
	gorm.Model
	GamerID        uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_battle_card"`
	Gamer          player.TelegramUser      `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` //ключевая вещь, связывающая юзера и карты
	CardTemplateID uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_battle_card"`
	CardTemplate   cards.BattleCardTemplate `gorm:"foreignKey:CardTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Copies         int                      `gorm:"not null;default:0"`
	CardLevel      int                      `gorm:"not null;default:1"`
	CardXP         int                      `gorm:"not null;default:0"`
}

type GamerBuffCards struct {
	gorm.Model
	GamerID        uint                    `gorm:"not null;index;uniqueIndex:ux_gamer_buff_card"`
	Gamer          player.TelegramUser     `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` //ключевая вещь, связывающая юзера и карты
	CardTemplateID uint                    `gorm:"not null;index;uniqueIndex:ux_gamer_buff_card"`
	CardTemplate   cards.BuffCardsTemplate `gorm:"foreignKey:CardTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Copies         int                     `gorm:"not null;default:0"`
	CardLevel      int                     `gorm:"not null;default:1"`
	CardXP         int                     `gorm:"not null;default:0"`
}

func LoadOwnedBattleCards(tx *gorm.DB, userID uint) (map[string]game.OwnedCardInfo, map[string]int, error) {
	var rows []GamerBattleCards
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	info := make(map[string]game.OwnedCardInfo, len(rows))
	copies := make(map[string]int, len(rows))
	for _, r := range rows {
		code := r.CardTemplate.CodeString
		info[code] = game.OwnedCardInfo{GamerCardID: r.ID, Copies: r.Copies, Level: r.CardLevel}
		copies[code] = r.Copies
	}
	return info, copies, nil
}

func LoadOwnedBuff(tx *gorm.DB, userID uint) (map[string]game.OwnedCardInfo, map[string]int, error) {
	var rows []GamerBuffCards
	if err := tx.Preload("CardTemplate").Where("gamer_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	info := make(map[string]game.OwnedCardInfo, len(rows))
	copies := make(map[string]int, len(rows))
	for _, r := range rows {
		code := r.CardTemplate.CodeString
		info[code] = game.OwnedCardInfo{GamerCardID: r.ID, Copies: r.Copies, Level: r.CardLevel}
		copies[code] = r.Copies
	}
	return info, copies, nil
}
