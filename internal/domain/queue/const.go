package queue

const (
	QueueStatusSearching   = "searching"     //<-стоит в очереди
	QueueStatusPendingMath = "pending_match" //<-найден кандидат
	QueueStatusPenalty     = "penalty"       //<-временно заблокирован от поиска
)

// ПРАВИЛА ПОИСКА
const (
	DefaultSearchRange = 50 //<-границы поиска (плюс минус 50 рейтинга)
	PenaltyMinutes     = 3  //<-время, на которе чувак отлетает в бан если не принял игру
	AcceptTimeoutSec   = 10 //<-время, которое дается на принятие матча
)
