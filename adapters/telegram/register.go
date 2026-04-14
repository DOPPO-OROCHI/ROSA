package telegram

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Profile struct {
	TGID      int64
	Username  string
	FirstName string
	LastName  string
	Language  string
}

func EnsureUser(db *gorm.DB, profile Profile) error {
	return db.Transaction(func(tx *gorm.DB) error {
		user := player.TelegramUser{
			TGID:      profile.TGID,
			Username:  profile.Username,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Language:  profile.Language,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tg_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"username":   user.Username,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"language":   user.Language,
			}),
		}).Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Where("tg_id = ?", profile.TGID).First(&user).Error; err != nil {
			return err
		}

		var battleTemplates []cards.BattleCardTemplate
		if err := tx.Select("id", "max_copies").Find(&battleTemplates).Error; err != nil {
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

		var buffTemplates []cards.BuffCardsTemplate
		if err := tx.Select("id", "max_copies").Find(&buffTemplates).Error; err != nil {
			return err
		}
		buffGrants := make([]repository.GamerBuffCards, 0, len(buffTemplates))
		for _, t := range buffTemplates {
			buffGrants = append(buffGrants, repository.GamerBuffCards{
				GamerID:        user.ID,
				CardTemplateID: t.ID,
				Copies:         t.MaxCopies,
				CardLevel:      1,
				CardXP:         0,
			})
		}
		if len(buffGrants) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "gamer_id"}, {Name: "card_template_id"}},
				DoNothing: true,
			}).CreateInBatches(&buffGrants, 500).Error; err != nil {
				return err
			}
		}

		var heroTemplates []heroes.CharacterTemplate
		if err := tx.Select("id").Find(&heroTemplates).Error; err != nil {
			return err
		}
		heroGrants := make([]repository.GamerCharacter, 0, len(heroTemplates))
		for _, t := range heroTemplates {
			heroGrants = append(heroGrants, repository.GamerCharacter{
				GamerID:             user.ID,
				CharacterTemplateID: t.ID,
				CharacterLevel:      1,
				CharacterXP:         0,
			})
		}
		if len(heroGrants) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "gamer_id"}, {Name: "character_template_id"}},
				DoNothing: true,
			}).CreateInBatches(&heroGrants, 500).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
