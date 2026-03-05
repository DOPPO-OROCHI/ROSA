package cache

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"sync"

	"gorm.io/gorm"
)

type ResolverCache struct {
	mu   sync.RWMutex
	res  game.Resolvers
	init bool
}

func (c *ResolverCache) Reload(db *gorm.DB) error {
	var battles []cards.BattleCardTemplate
	if err := db.Find(&battles).Error; err != nil {
		return err
	}
	battleMap := make(map[string]game.BattleTemplate, len(battles))
	for _, t := range battles {
		battleMap[t.CodeString] = game.BattleTemplate{
			TemplateID:    t.CodeString,
			HealthPoints:  t.HealthPoints,
			Attack:        t.Attack,
			SplashRadius:  t.SplashRadius,
			Cooldown:      t.CoolDown,
			Manacost:      t.ManaCost,
			IsTank:        t.IsTank,
			CardType:      t.CardType,
			CanBeUpgraded: t.BuffSlot,
			ImageKey:      t.ImageKey,
			AssetBaseKey:  t.AssetBaseKey,
		}
	}
	var buffs []cards.BuffCardsTemplate
	if err := db.Find(&buffs).Error; err != nil {
		return err
	}
	buffMap := make(map[string]game.BuffTemplate, len(buffs))
	for _, t := range buffs {
		buffMap[t.CodeString] = game.BuffTemplate{
			TemplateID:   t.CodeString,
			ManaCost:     t.ManaCost,
			BuffType:     t.BuffType,
			BuffValue:    t.BuffValue,
			OnlyFor:      t.OnlyFor,
			Duration:     t.Duration,
			ImageKey:     t.ImageKey,
			AssetBaseKey: t.AssetBaseKey,
		}
	}
	res := game.Resolvers{
		HeroAbility: func(herocode string) (game.HeroAbility, bool) {
			switch herocode {
			case "suprime_lider":
				return game.SupremeLiderAbilitySpec{}, true
			case "karn":
				return game.KarnAbilitySpec{}, true
			case "the_system":
				return game.TheSystemAbilitySpec{}, true
			case "imperial_commander":
				return game.ImperialCommanderAbilitySpec{}, true
			case "black_cell":
				return game.BlackCellAbilitySpec{}, true
			case "slavic_priest":
				return game.SlavicPriestAbilitySpec{}, true
			default:
				return nil, false
			}
		},
		Battle: game.BattleMapResolver{M: battleMap},
		Buff:   game.BuffMapResolver{M: buffMap},
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = res
	c.init = true
	return nil
}

func (c *ResolverCache) Get() (game.Resolvers, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.res, c.init
}
