package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/applycation"
	"TheWar/internal/transport"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type ReadyMatchHandlersDeps struct {
	DB  *gorm.DB
	Hub *transport.Hub
}

func NewReadyMatchHandler(d ReadyMatchHandlersDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil || tail != "ready" {
			middleware.WriteErr(w, http.StatusNotFound, "not found")
			return
		}
		st, err := applycation.MarkMatchReadyTX(d.DB, matchID, au.UserID)
		if err != nil {
			switch {
			case errors.Is(err, applycation.ErrNotParticipant):
				middleware.WriteErr(w, http.StatusForbidden, "forbidden")
				return
			case errors.Is(err, applycation.ErrCorruptedMatchState):
				middleware.WriteErr(w, http.StatusInternalServerError, "something went worng")
				return
			default:
				middleware.WriteErr(w, MapEngineErr(err), err.Error())
				return
			}
		}
		PublishMatchToSSE(d.Hub, st)
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(st, au.UserID))
	}
}
