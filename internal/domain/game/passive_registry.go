package game

import "TheWar/internal/domain/cards"

type PassiveContext struct {
	Trigger             string
	SourcePlayerIndex   int
	AttackTargetID      string
	AttackerInstanceID  string
	DamagedByInstanceID string
	PlayedCardCode      string
	PlayedCardType      string
	PlayedCardIsTank    bool
	PlayedSkillCode     string
	HeroSkillUsed       bool
}

type CardPassiveHandler func(m *MatchState, source *UnitState, owner *PlayerState,
	enemy *PlayerState, spec cards.PassiveSpec, ctx PassiveContext) error

var cardPassiveHandlers = map[string]CardPassiveHandler{
	"coordinated_actions":   HandlePassiveAuraBuff,
	"perfection_of_tactics": HandlePassiveReactiveBuff,
}
