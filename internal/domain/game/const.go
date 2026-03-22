package game

/*Файл целиком и полностью посвящен описанием констант, которые нужны для
доменной логики. Мысль в том, чтобы привести разные аспекты игры к единому
стандарту. Так вот...*/

/*
Я ввел два эти типа не случайно, а во имя того, чтобы избежать путаницы при
обработке специализированных именно под (имена говорят сами за себя) эти аспекты
игры. MathResult определяет результат матча (обычная строка не подойдет, страшно).
TurnPhase о том, какая стадия сейчас в игре. Существуют две стадии -Старт (системная),
мэйн (игровая) именно в матче, и результат.
*/
type MatchResult string
type TurnPhase string

const (
	//матч еще играется
	MatchOnGoing MatchResult = "ON_GOING"
	//победил игрок 1
	MatchWinP1 MatchResult = "P1_WIN"
	//победил игрок 2
	MatchWinP2 MatchResult = "P2_WIN"
	//ничья
	MatchDraw MatchResult = "DRAW"
	//системные действия (подготовка перед мэйном)
	PhaseStart TurnPhase = "START"
	//игровые действия (когда игрок именно играет)
	PhaseMain TurnPhase = "MAIN"
)

// макимальная длина стола. Не вариант поставить больше 5 карт на стол
const TableSize = 5

// дедлайн хода
const DeadLineTurn = 30

// максимальный уровень прокачки карт
const MaxCardLevel = 30

// тип действия
type ActionType string

// константы типов действий
const (
	//закончить действие (считай закончить ход)
	ActionEndTurn ActionType = "end_turn"
	//сыграть боевую карту (ту, которая может стрелять, гасить и так далее)
	ActionPlayBattle ActionType = "play_battle_card"
	//сыграть баф карту
	ActionPlayBuff ActionType = "play_buff_card"
	//действие, когда игрок атакует картой
	ActionCardAttack ActionType = "card_attack"
	//действие атаки персонажем игрока
	ActionHeroAttack ActionType = "hero_attack"
	//сгрытьа спелл персонажа
	ActionPlayHeroSpell ActionType = "hero_spell"
	//ливнуть из матча
	ActionLeaveMatch ActionType = "leave_match"
	//сыграть скилл карты
	ActionPlayCardSkill ActionType = "card_skill"
)

// а эта штука нужна для UI. Здесь описывается источник анимаций. Будь то карта, баф и все такое
type SourceKind string

// а это возможные источники анимаций
const (
	//атакующая карта (юнит, типа юнит на столе...)
	SourceUnit SourceKind = "unit"
	//персонаж
	SourceHero SourceKind = "hero"
	//карта из руки (баф)
	SourceCard SourceKind = "card"
	//системные анимации
	SourceSystem SourceKind = "system"
)

/*
тоже штука нужная чисто для UI. Здесь, в зависимости от типа ивента будет
определяться какую анимацию нужно отдать клиенту
*/
type EventType string

// а здесь описание ивентов
const (
	//поставить карту на стол
	EventSummon EventType = "summon"
	//атаковать картой
	EventAttack EventType = "attack"
	//хилить
	EventHeal EventType = "heal"
	//бафнуть
	EventBuff EventType = "buff"
	//геройская способность
	EventHeroSpell EventType = "hero_spell"
	//анимация уничтожения
	EventDeath EventType = "death"
	//анимация смены хода
	EventTurn EventType = "turn"
	//анимация атаки героя
	EventHeroAttack EventType = "hero_attack"
	//анимация применения скилла карты
	EventCardSkill EventType = "card_skill"
)
