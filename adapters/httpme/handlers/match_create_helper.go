package handlers

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"
	"errors"

	"gorm.io/gorm"
)

func createMatchForUsers(db *gorm.DB, userID1, userID2 uint) (*game.MatchState, error) {
	var p1 player.TelegramUser
	if err := db.Select("id", "selected_hero_template_id").
		Where("id = ?", userID1).First(&p1).Error; err != nil {
		return nil, err
	}
	var p2 player.TelegramUser
	if err := db.Select("id", "selected_hero_template_id").
		Where("id = ?", userID2).First(&p2).Error; err != nil {
		return nil, err
	}
	if p1.SelectedHeroTemplateID == nil {
		return nil, errors.New("player 1 has no selected hero")
	}
	if p2.SelectedHeroTemplateID == nil {
		return nil, errors.New("player 2 has no selected hero")
	}
	var p1Tpl heroes.CharacterTemplate
	if err := db.Select("id", "character_code").
		Where("id = ?", *p1.SelectedHeroTemplateID).First(&p1Tpl).Error; err != nil {
		return nil, err
	}
	p1HeroCode := p1Tpl.CharacterCode
	var p2Tpl heroes.CharacterTemplate
	if err := db.Select("id", "character_code").
		Where("id = ?", *p2.SelectedHeroTemplateID).First(&p2Tpl).Error; err != nil {
		return nil, err
	}
	p2HeroCode := p2Tpl.CharacterCode
	return repository.CreateMatchTX(db, userID1, userID2, p1HeroCode, p2HeroCode)
}
