package player

import "gorm.io/gorm"

type TelegramUser struct {
	gorm.Model
	TGID                   int    `gorm:"not null;uniqueIndex"`
	Username               string `gorm:"not null"`
	FirstName              string `gorm:"not null"`
	LastName               string `gorm:"not null"`
	Language               string `gorm:"not null"`
	Rating                 int    `gorm:"not null;default:1500"`
	XP                     int    `gorm:"not null;default:0"`
	SelectedHeroTemplateID *uint  `gorm:"index"`
}
