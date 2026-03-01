package repository

import (
	"TheWar/internal/domain/player"

	"gorm.io/gorm"
)

type GamerDeckEntry struct {
	gorm.Model
	GamerID    uint                `gorm:"not null;index;uniqueIndex:ux_gamer_deck_entry"`
	Gamer      player.TelegramUser `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Kind       string              `gorm:"not null;uniqueIndex:ux_gamer_deck_entry"`
	TemplateID string              `gorm:"not null;uniqueIndex:ux_gamer_deck_entry"`
	Count      int                 `gorm:"not null"`
}
