package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/queue"
	"TheWar/internal/infra/repository"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type AcceptQueueHandlerDeps struct {
	DB    *gorm.DB
	Queue *queue.Queue
}

type AcceptQueueHandler struct {
	db    *gorm.DB
	queue *queue.Queue
}

type AcceptQueueResponse struct {
	State string `json:"state"`
}

func NewAcceptQueueHandler(deps AcceptQueueHandlerDeps) AcceptQueueHandler {
	return AcceptQueueHandler{
		db:    deps.DB,
		queue: deps.Queue,
	}
}

func AcceptQueue(q AcceptQueueHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		pending, bothAccepted, err := q.queue.AcceptPendingMatch(int(user.UserID))
		if err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "")
			return
		}
		if !bothAccepted {
			res := AcceptQueueResponse{State: queue.MatchMakingStatePendingMatch}
			middleware.WriteJSON(w, http.StatusOK, res)
			return
		}
		st, err := createMatchForUsers(q.db, uint(pending.UserID1), uint(pending.UserID2))
		if err != nil {
			if errors.Is(err, repository.ErrActiveMatchExists) {
				middleware.WriteErr(w, http.StatusConflict, "active match exist")
				return
			}
			middleware.WriteErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := q.queue.FinalizeAcceptedMatch(pending.UserID1, pending.UserID2); err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "failed to finalize accepted match")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(st, user.UserID))
	}
}
