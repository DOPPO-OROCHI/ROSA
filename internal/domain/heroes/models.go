package heroes

import "gorm.io/gorm"

type CharacterTemplate struct {
	gorm.Model
	Name           string      `gorm:"not null"`
	CharacterCode  string      `gorm:"not null;uniqueIndex"`
	AttackPower    int         `gorm:"not null;default:0"`
	HealthPoints   int         `gorm:"not null;default:0"`
	AttackCooldown int         `gorm:"not null;default:0"`
	SplashRadius   int         `gorm:"not null;default:0"`
	Ability        AbilitySpec `gorm:"embedded;embeddedPrefix:ability_"`
	Description    string      `gorm:"text;not null"`
	ImageKey       string
	AssetBaseKey   string
}

type AbilitySpec struct {
	Code     string `gorm:"not null"`
	Target   string `gorm:"not null"`
	CoolDown int    `gorm:"not null;default:0"`
	ManaCost int    `gorm:"not null;default:0"`
	Value    int    `gorm:"not null;default:0"`
	Duration int    `gorm:"not null;default:0"`
}
