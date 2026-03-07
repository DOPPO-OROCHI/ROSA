package repository

import (
	"TheWar/internal/domain/player"

	"gorm.io/gorm"
)

/*А это структура входящей в матч деки. Здесь представлены айди персонажа, который владеет декой, сам пользователь,
тип карты, шаблон карты а так же, сколько копий одной карты есть в деке. Это грубо говоря таблица сохраненной деки
игрока в БД. Через нее мы читаем, сохраняем деку а так же используем это при создании рантайм матча. Такие дела*/

type GamerDeckEntry struct {
	gorm.Model
	GamerID    uint                `gorm:"not null;index;uniqueIndex:ux_gamer_deck_entry"`
	Gamer      player.TelegramUser `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Kind       string              `gorm:"not null;uniqueIndex:ux_gamer_deck_entry"`
	TemplateID string              `gorm:"not null;uniqueIndex:ux_gamer_deck_entry"`
	Count      int                 `gorm:"not null"`
}
