package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/queue"
	"net/http"
)

type DeclineQueueHandlerDeps struct {
	Queue *queue.Queue
}

type DeclineQueueHandler struct {
	queue *queue.Queue
}

type DeclineQueueResponse struct {
	State string `json:"state"`
}

func NewDeclineQueueHandler(deps DeclineQueueHandlerDeps) DeclineQueueHandler {
	return DeclineQueueHandler{queue: deps.Queue}
}

func DeclineQueue(q DeclineQueueHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err := q.queue.DeclinePendingMatch(int(user.UserID)); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "bad request")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, DeclineQueueResponse{
			State: queue.MatchMakingStatePenalty,
		})
	}
}
