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
		return nil //<-к слову. То, что в дефолтах нет данных не является ошибкой именно этой функции
		//почему ? Потому что единственная задача сида - выгрузить данные в БД. Она не обязана валидировать темплейты
	}
	FillBattleKeys(cards.DefaultBattleCards) //<-вызываем функцию заполнения ключей
	return db.Clauses(clause.OnConflict{     //<-вызываем безопасное добавление записей
		//к слову, раньше я не понимал почему мы это можем вызвать через ретерн. Но как по мне все супер понятно,
		//даже с языковой точки зрения. По сути, по-русски это звучало бы так -верни по выполнению функции ошибку,
		//если она имело место быть. Прикольный синтаксис. Это такой же прикол как if err := ... ; err != nil {...}
		Columns: []clause.Column{{Name: "code_string"}}, //<-но тем не менее, указываем по какому ключу мы если что будем апдейтить карты
		DoUpdates: clause.AssignmentColumns([]string{ //<-и если произошел конфликт с code_string, то мы просто обновляем следующие поля:
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
			"back_pic",
		}),
	}).CreateInBatches(&cards.DefaultBattleCards, 200).Error //<-вот это тоже интересно, до сих пор не виданно
	/*Здесь стоит представить, что такое массовый сидинг данных внутрь БД. По сути, отдельный сид = отдельный SQL
	запрос, из чего становится ясно, что в рамках больших массивов данных, это очень быстро станет неподъемным для
	стабильной работы (хоть и вряд ли это когда нибудь будет актуальным в рамках моей игры). CreateInBatches решает
	этот вопрос. Что здесь происходит ? Мы передаем массив с нашими картами, а вторым аргументом целое число, которое
	отражает то, сколько данных мы будем сидить в рамках одного SQL запроса. Таким образом за раз я буду сгружать
	200 сущностей, что решает проблему с нагрузкой на БД*/
}

/*Аналогичная ситуация и в случае бафф карт. Идентичная схема, идентичный сидинг*/
func SeedBuffCardTemplate(db *gorm.DB) error {
	if len(cards.DefaultBuffCards) == 0 {
		return nil
	}
	FillBuffKeys(cards.DefaultBuffCards) //<-так же вызываем автозаполнение ключей
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
			"back_pic",
		}),
	}).CreateInBatches(&cards.DefaultBuffCards, 200).Error
}

func SeedCharacterTemplate(db *gorm.DB) error {
	if len(heroes.DefaultHeroTemplate) == 0 {
		return nil
	}
	FillHeroKeys(heroes.DefaultHeroTemplate)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "character_code"}},
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
			"back_pic",
		}),
	}).CreateInBatches(&heroes.DefaultHeroTemplate, 200).Error
}

/*
Для того, чтобы отдельно не вызывать каждую из сущностей, нужно как то загрузить все это добно в одно.
Этим и занимается SeedEverything, последовательно вызывая каждую из вышенаписанных функций. В случае провала
хоть одной, вовзращаем ошибку
*/
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

/*Таким образом реализована система сидинга данных, со всеми вытекающими. Вызывается штука на самом старте
приложения (как кэш резолверов), чтобы обеспечить данными БД, что в последствии обеспечит рантайм компонент.*/
