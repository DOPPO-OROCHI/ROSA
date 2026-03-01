package bootstrap

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/db"
	"TheWar/internal/infra/repository"
)

//Полностью посвящен миграциям нужных структур внутрь БД, а так же заполнению их же начальными данными

func Init() error {
	if err := db.DB.AutoMigrate(&player.TelegramUser{},
		&cards.BattleCardTemplate{},
		&cards.BuffCardsTemplate{},
		&heroes.CharacterTemplate{},
		&repository.GamerBattleCards{},
		&repository.GamerBuffCards{},
		&repository.GamerCharacter{},
		&repository.GamerDeckEntry{},
		&repository.Match{},
	); err != nil {
		return err
	}
	return db.SeedEverything(db.DB)
}
