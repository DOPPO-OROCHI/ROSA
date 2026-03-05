package game

type MatchResult string
type TurnPhase string

const (
	MatchOnGoing MatchResult = "ON_GOING"
	MatchWinP1   MatchResult = "P1_WIN"
	MatchWinP2   MatchResult = "P2_WIN"
	MatchDraw    MatchResult = "DRAW"
	PhaseStart   TurnPhase   = "START"
	PhaseMain    TurnPhase   = "MAIN"
)

const TableSize = 5

const MaxCardLevel = 30

type ActionType string

const (
	ActionEndTurn       ActionType = "end_turn"
	ActionPlayBattle    ActionType = "play_battle_card"
	ActionPlayBuff      ActionType = "play_buff_card"
	ActionCardAttack    ActionType = "card_attack"
	ActionHeroAttack    ActionType = "hero_attack"
	ActionPlayHeroSpell ActionType = "hero_spell"
	ActionLeaveMatch    ActionType = "leave_match"
)

type SourceKind string

const (
	SourceUnit   SourceKind = "unit"
	SourceHero   SourceKind = "hero"
	SourceCard   SourceKind = "card"
	SourceSystem SourceKind = "system"
)

type EventType string

const (
	EventSummon     EventType = "summon"
	EventAttack     EventType = "attack"
	EventHeal       EventType = "heal"
	EventBuff       EventType = "buff"
	EventHeroSpell  EventType = "hero_spell"
	EventDeath      EventType = "death"
	EventTurn       EventType = "turn"
	EventHeroAttack EventType = "hero_attack"
)
