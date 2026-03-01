package game

type MatchResult string
type TurnPhase string

const (
	MatchOnGoing MatchResult = "ON_GOING" //<-матч в процесса
	MatchWinP1   MatchResult = "P1_WIN"   //<-выиграл первый игрок
	MatchWinP2   MatchResult = "P2_WIN"   //<-выиграл второй игрок
	MatchDraw    MatchResult = "DRAW"     //<-ничья
	PhaseStart   TurnPhase   = "START"    //<-старт. На этой фазе сервер делает все необходимые приготовления к ходу игрока
	PhaseMain    TurnPhase   = "MAIN"     //<-мэйн. На этой фазе игрок уже сам принимает решения. Кого ударить, ливнуть и тд..
)

const TableSize = 5

const MaxCardLevel = 30

type ActionType string

const (
	ActionEndTurn       ActionType = "end_turn"         //<-закончить ход
	ActionPlayBattle    ActionType = "play_battle_card" //<-поставить карту на стол
	ActionPlayBuff      ActionType = "play_buff_card"   //<-поставить бафф карту на стол
	ActionCardAttack    ActionType = "card_attack"      //<-атаковать картой
	ActionHeroAttack    ActionType = "hero_attack"      //<-атаковать героя картой
	ActionPlayHeroSpell ActionType = "hero_spell"       //<-использовать геройскую способность
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
