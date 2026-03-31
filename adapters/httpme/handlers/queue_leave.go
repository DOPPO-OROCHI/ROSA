package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/queue"
	"net/http"
)

type LeaveQueueHandler struct {
	queue *queue.Queue
}

type LeaveQueueHandlerDeps struct {
	Queue *queue.Queue
}

type LeaveQueueResponse struct {
	State string `json:"state"`
}

func NewLeaveQueueHandler(deps LeaveQueueHandlerDeps) LeaveQueueHandler {
	return LeaveQueueHandler{queue: deps.Queue}
}

func LeaveQueue(h LeaveQueueHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userInQueue := queue.UserInQueue{UserID: int(user.UserID)}
		_ = h.queue.RemoveUserFromQueue(&userInQueue)
		middleware.WriteJSON(w, http.StatusOK, LeaveQueueResponse{State: queue.MatchMakingStateIdle})
	}
}
