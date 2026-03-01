package repository

import (
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"errors"

	"gorm.io/gorm"
)

type GamerCharacter struct {
	gorm.Model
	GamerID             uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_character"`
	Gamer               player.TelegramUser      `gorm:"foreignKey:GamerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CharacterTemplateID uint                     `gorm:"not null;index;uniqueIndex:ux_gamer_character"`
	CharacterTemplate   heroes.CharacterTemplate `gorm:"foreignKey:CharacterTemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CharacterLevel      int                      `gorm:"not null;default:1"`
	CharacterXP         int                      `gorm:"not null;default:0"`
}

func LoadSelectedHeroCodeTx(tx *gorm.DB, userID uint) (string, error) {
	var u player.TelegramUser
	if err := tx.Select("id", "selected_hero_template_id").
		Where("id = ?", userID).First(&u).Error; err != nil {
		return "", err
	}
	if u.SelectedHeroTemplateID == nil {
		return "", errors.New("selected hero is not set")
	}
	var tpl heroes.CharacterTemplate
	if err := tx.Select("character_code").
		Where("id = ?", *u.SelectedHeroTemplateID).
		First(&tpl).Error; err != nil {
		return "", err
	}
	return tpl.CharacterCode, nil
}
