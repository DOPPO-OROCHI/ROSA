package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/player"
	"TheWar/internal/domain/queue"
	"net/http"

	"gorm.io/gorm"
)

type JoinQueueHandlerDeps struct {
	DB    *gorm.DB
	Queue *queue.Queue
}

type JoinQueueHandler struct {
	db    *gorm.DB
	queue *queue.Queue
}

type JoinQueueResponse struct {
	State          string `gorm:"state"`
	OpponentUserID int    `gorm:"opponent_user_id,omitempty"`
}

func NewJoinHandler(deps JoinQueueHandlerDeps) JoinQueueHandler {
	return JoinQueueHandler{
		db:    deps.DB,
		queue: deps.Queue,
	}
}

func JoinQueue(q JoinQueueHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		user := player.TelegramUser{}
		if err := q.db.Where("id = ?", userID).First(&user).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "server error")
			return
		}
		userInQueue := queue.UserInQueue{
			UserID: int(userID),
			Rating: user.Rating,
		}
		if err := q.queue.AddUserToQueue(&userInQueue); err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "server error")
			return
		}
		_, opponent, reserved, err := q.queue.ReserveMatchForUser(int(userID))
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "server error")
			return
		}
		if !reserved {
			middleware.WriteJSON(w, http.StatusOK, JoinQueueResponse{State: queue.MatchMakingStateSearching})
			return
		}
		middleware.WriteJSON(w, http.StatusOK, JoinQueueResponse{
			State:          queue.MatchMakingStatePendingMatch,
			OpponentUserID: opponent.UserID,
		})
	}
}
