package db

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/*Файл целиком и полностью посвящен сидингу. Здесь описаны все функции для того, чтобы загрузить файлы в БД.
Почему это важно? По началу я не знал даже что такое сидинг. Вопрос плотно встал тогда, когда стало понятно
что выгрузить объемные массивы данных в миграции не вариант) да, я стажер... Тем не менее этот массив есть,
соответственно надо его выгрузить в БД. Делается это с помощью концепции сидинга данных, чем и занимается этот
файл в рамках моей игры. Перейдем к функциям*/

/*
Данная функция выгружает в БД весь массив боевых карт делая это идемпотентно. В случае, если карты поменяли
какую либо характеристику внутри памяти, мы просто обновляем это поле внутри БД с помощью OnConflict
*/
func SeedBattleCardTemplate(db *gorm.DB) error {
	if len(cards.DefaultBattleCards) == 0 { //<-просто проверка
		return nil
	}
	FillBattleKeys(cards.DefaultBattleCards)
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
	FillBuffKeys(cards.DefaultBuffCards)
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
	FillHeroKeys(heroes.DefaultHeroTemplate)
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
