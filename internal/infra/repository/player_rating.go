package repository

import (
	"TheWar/internal/domain/player"
	"errors"

	"gorm.io/gorm"
)

const (
	RatingNumber = 25
)

func RatingUp(userID uint, db *gorm.DB) error {
	res := db.Model(&player.TelegramUser{}).Where("id = ?", userID).
		UpdateColumn("rating", gorm.Expr("rating + ?", RatingNumber))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("cant find this id")
	}
	return nil
}

func RatingDown(userID uint, db *gorm.DB) error {
	res := db.Model(&player.TelegramUser{}).Where("id = ?", userID).
		UpdateColumn("rating", gorm.Expr("rating - ?", RatingNumber))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("cant find this id")
	}
	return nil
}
