package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/player"
	"TheWar/internal/domain/queue"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type QueueStatusResponse struct {
	State              string    `json:"state"`
	OpponentUserID     int       `json:"opponent_user_id,omitempty"`
	OpponentRating     int       `json:"opponent_rating,omitempty"`
	SearchDurationSec  int       `json:"search_duration_sec,omitempty"`
	PenaltyUntil       time.Time `json:"penalty_until,omitempty"`
	AcceptDeadlineAt   time.Time `json:"accept_deadline_at,omitempty"`
	AcceptedByMe       bool      `json:"accepted_by_me,omitempty"`
	AcceptedByOpponent bool      `json:"accepted_by_opponent,omitempty"`
}

type QueueStatusHandlerDeps struct {
	DB    *gorm.DB
	Queue *queue.Queue
}

type QueueStatusHandler struct {
	db    *gorm.DB
	queue *queue.Queue
}

func NewQueueStatusHandler(deps QueueStatusHandlerDeps) QueueStatusHandler {
	return QueueStatusHandler{
		db:    deps.DB,
		queue: deps.Queue,
	}
}

func QueueStatus(h QueueStatusHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := user.UserID
		state := h.queue.GetUserMatchmakingState(int(userID))
		resp := QueueStatusResponse{
			State: state,
		}
		if state == queue.MatchMakingStateSearching {
			dur, ok := h.queue.GetSearchDuration(int(user.UserID))
			if ok {
				resp.SearchDurationSec = int(dur.Seconds())
			}
		}
		if state == queue.MatchMakingStatePenalty {
			penaltyUntil, ok := h.queue.GetPenaltyUntil(int(user.UserID))
			if ok {
				resp.PenaltyUntil = penaltyUntil
			}
		}
		if state == queue.MatchMakingStatePendingMatch {
			acceptedByMe, acceptedByOpponent, expiresAt, opponentUserID, ok := h.queue.GetPendingAcceptanceState(int(user.UserID))
			if ok {
				resp.OpponentUserID = opponentUserID
				resp.AcceptDeadlineAt = expiresAt
				resp.AcceptedByMe = acceptedByMe
				resp.AcceptedByOpponent = acceptedByOpponent
				if h.db != nil && opponentUserID > 0 {
					var opponent player.TelegramUser
					if err := h.db.Select("rating").Where("id = ?", opponentUserID).First(&opponent).Error; err == nil {
						resp.OpponentRating = opponent.Rating
					}
				}
			}
		}
		middleware.WriteJSON(w, http.StatusOK, resp)
	}
}
