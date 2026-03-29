package queue

import "time"

const (
	QueueStatusSearching   = "searching"     //<-стоит в очереди
	QueueStatusPendingMatch = "pending_match" //<-найден кандидат
)

// ПРАВИЛА ПОИСКА
const (
	DefaultSearchRange    = 50               //<-границы поиска (плюс минус 50 рейтинга)
	PenaltyDefaultMinutes = 3 * time.Minute  //<-время, на которе чувак отлетает в бан если не принял игру
	AcceptTimeoutDefault  = 10 * time.Second //<-время, которое дается на принятие матча
)
