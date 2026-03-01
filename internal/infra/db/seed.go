package db

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedBattleCardTemplate(db *gorm.DB) error {
	if len(cards.DefaultBattleCards) == 0 {
		return nil
	}
	fillBattleKeys(cards.DefaultBattleCards)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code_string"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"health_points",
			"attack",
			"splash_radius",
			"is_tank",
			"card_type",
			"cool_down",
			"mana_cost",
			"buff_slot",
			"max_copies",
			"description",
			"image_key",
			"asset_base_key",
		}),
	}).CreateInBatches(&cards.DefaultBattleCards, 200).Error
}

func SeedBuffCardTemplate(db *gorm.DB) error {
	if len(cards.DefaultBuffCards) == 0 {
		return nil
	}
	fillBuffKeys(cards.DefaultBuffCards)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code_string"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"mana_cost",
			"buff_type",
			"buff_value",
			"only_for",
			"max_copies",
			"duration",
			"description",
			"image_key",
			"asset_base_key",
		}),
	}).CreateInBatches(&cards.DefaultBuffCards, 200).Error
}

func SeedCharacterTemplate(db *gorm.DB) error { //<-принимаем ДБ как аргумент
	if len(heroes.DefaultHeroTemplate) == 0 { //<-если у нас никого нет в темлейте, ниче не возвращаем, это не ошибка
		return nil
	}
	fillHeroKeys(heroes.DefaultHeroTemplate)
	return db.Clauses(clause.OnConflict{ //<-начинаем вечеринку
		Columns: []clause.Column{{Name: "character_code"}}, //<-проверяем, если код персонажа уже есть, ничего не делаем (херово под балансные правки)
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"attack_power",
			"health_points",
			"attack_cooldown",
			"splash_radius",
			"ability_code",
			"ability_target",
			"ability_cool_down",
			"ability_mana_cost",
			"ability_value",
			"ability_duration",
			"description",
			"image_key",
			"asset_base_key",
		}),
	}).CreateInBatches(&heroes.DefaultHeroTemplate, 200).Error
}

func SeedEverything(db *gorm.DB) error {
	if err := SeedCharacterTemplate(db); err != nil {
		return err
	}
	if err := SeedBattleCardTemplate(db); err != nil {
		return err
	}
	if err := SeedBuffCardTemplate(db); err != nil {
		return err
	}
	return nil
}
