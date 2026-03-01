package db

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/heroes"
	"strings"
)

func BaseKey(kind, code string) string {
	code = strings.TrimSpace(code)
	kind = strings.Trim(kind, "/")
	return kind + "/" + code
}

func ImageKey(base string) string {
	base = strings.Trim(base, "/")
	return base + "/image"
}

func fillBattleKeys(ts []cards.BattleCardTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = BaseKey("card", ts[i].CodeString)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = ImageKey(ts[i].AssetBaseKey)
		}
	}
}

func fillBuffKeys(ts []cards.BuffCardsTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = BaseKey("buff", ts[i].CodeString)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = ImageKey(ts[i].AssetBaseKey)
		}
	}
}
func fillHeroKeys(ts []heroes.CharacterTemplate) {
	for i := range ts {
		if ts[i].AssetBaseKey == "" {
			ts[i].AssetBaseKey = BaseKey("hero", ts[i].CharacterCode)
		}
		if ts[i].ImageKey == "" {
			ts[i].ImageKey = ImageKey(ts[i].AssetBaseKey)
		}
	}
}
