package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/queue"
	"net/http"
	"time"
)

type QueueStatusResponse struct {
	State              string    `json:"state"`
	OpponentUserID     int       `json:"opponent_user_id,omitempty"`
	SearchDurationSec  int       `json:"search_duration_sec,omitempty"`
	PenaltyUntil       time.Time `json:"penalty_until,omitempty"`
	AcceptDeadlineAt   time.Time `json:"accept_deadline_at,omitempty"`
	AcceptedByMe       bool      `json:"accepted_by_me,omitempty"`
	AcceptedByOpponent bool      `json:"accepted_by_opponent,omitempty"`
}

type QueueStatusHandlerDeps struct {
	Queue *queue.Queue
}

type QueueStatusHandler struct {
	queue *queue.Queue
}

func NewQueueStatusHandler(deps QueueStatusHandlerDeps) QueueStatusHandler {
	return QueueStatusHandler{
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
			}
		}
		middleware.WriteJSON(w, http.StatusOK, resp)
	}
}
