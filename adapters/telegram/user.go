package telegram

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/*Идемпотентное создание пользователя в БД при его отправке сообщения в бот. Без этого процесс регистрации
не будет реализован никак.*/

func AddNewUser(db *gorm.DB, update tgbotapi.Update) error {
	if update.Message == nil || update.Message.From == nil {
		return nil
	}
	from := update.Message.From
	return db.Transaction(func(tx *gorm.DB) error {
		user := player.TelegramUser{
			TGID:      from.ID,
			Username:  from.UserName,
			FirstName: from.FirstName,
			LastName:  from.LastName,
			Language:  from.LanguageCode,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tg_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"username", "first_name", "last_name", "language",
			}),
		}).Create(&user).Error; err != nil {
			return err
		}
		var battleTemplates []cards.BattleCardTemplate
		if err := tx.Select("id", "max_copies").Find(&battleTemplates).
			Error; err != nil {
			return err
		}
		grants := make([]repository.GamerBattleCards, 0, len(battleTemplates))
		for _, t := range battleTemplates {
			grants = append(grants, repository.GamerBattleCards{
				GamerID:        user.ID,
				CardTemplateID: t.ID,
				Copies:         t.MaxCopies,
				CardLevel:      1,
				CardXP:         0,
			})
		}
		if len(grants) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "gamer_id"}, {Name: "card_template_id"}},
				DoNothing: true,
			}).CreateInBatches(&grants, 500).Error; err != nil {
				return err
			}
		}
		{
			var buffTemplates []cards.BuffCardsTemplate
			if err := tx.Select("id", "max_copies").Find(&buffTemplates).Error; err != nil {
				return err
			}
			grants := make([]repository.GamerBuffCards, 0, len(buffTemplates))
			for _, t := range buffTemplates {
				grants = append(grants, repository.GamerBuffCards{
					GamerID:        user.ID,
					CardTemplateID: t.ID,
					Copies:         t.MaxCopies,
					CardLevel:      1,
					CardXP:         0,
				})
			}
			if len(grants) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "gamer_id"}, {Name: "card_template_id"}},
					DoNothing: true,
				}).CreateInBatches(&grants, 500).Error; err != nil {
					return err
				}
			}
		}
		{
			var heroTemplates []heroes.CharacterTemplate
			if err := tx.Select("id").Find(&heroTemplates).Error; err != nil {
				return err
			}
			grants := make([]repository.GamerCharacter, 0, len(heroTemplates))
			for _, t := range heroTemplates {
				grants = append(grants, repository.GamerCharacter{
					GamerID:             user.ID,
					CharacterTemplateID: t.ID,
					CharacterLevel:      1,
					CharacterXP:         0,
				})
			}
			if len(grants) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "gamer_id"}, {Name: "character_template_id"}},
					DoNothing: true,
				}).CreateInBatches(&grants, 500).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
