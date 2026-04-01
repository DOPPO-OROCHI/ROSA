package handlers

import "TheWar/internal/domain/queue"

type DeclineQueueHandlerDeps struct {
	Queue *queue.Queue
}

type DeclineQueueHandler struct {
	queue *queue.Queue
}

func NewDeclineQueueHandler(deps DeclineQueueHandlerDeps) DeclineQueueHandler {
	return DeclineQueueHandler{queue: deps.Queue}
}
