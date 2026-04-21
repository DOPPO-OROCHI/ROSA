package game

import "TheWar/internal/domain/heroes"

type HeroAbilityHandler = func(m *MatchState, a Action, owner *PlayerState, spec heroes.AbilitySpec) error

var HeroAbilityHandlers = map[string]HeroAbilityHandler{
	"war_machine":          CastHeroAttackAndHPBuffAbility,
	"primal_rage":          CastHeroDebuffAbility,
	"swarm_defense":        CastHeroBuffAbility,
	"outstanding_marksman": CastHeroDamageAbility,
	"lightning_mastery":    CastHeroDebuffAbility,
	"organic_change":       CastHeroHybridAbility,
}
