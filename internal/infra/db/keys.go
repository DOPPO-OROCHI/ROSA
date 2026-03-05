package db

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
)

func fillBattleKeys(ts []cards.BattleCardTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.BattleCardBaseKey(ts[i].CodeString)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
	}
}

func fillBuffKeys(ts []cards.BuffCardsTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.BuffCardBaseKey(ts[i].CodeString)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
	}
}
func fillHeroKeys(ts []heroes.CharacterTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = game.HeroBaseKey(ts[i].CharacterCode)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = game.ImageKey(ts[i].AssetBaseKey)
		}
	}
}
