package middleware

import (
	"TheWar/internal/domain/game"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteErr(w http.ResponseWriter, code int, msg string) {
	WriteJSON(w, code, map[string]any{"error": msg})
}

func ParceMatchPath(path string) (matchID uint, tail string, err error) {
	s := strings.TrimPrefix(path, "/matches/")
	if s == path {
		return 0, "", errors.New("bad path")
	}
	parts := strings.Split(strings.Trim(s, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", errors.New("missing match id")
	}
	id64, e := strconv.ParseUint(parts[0], 10, 64)
	if e != nil || id64 == 0 {
		return 0, "", errors.New("bad match id")
	}
	matchID = uint(id64)
	if len(parts) == 1 {
		return matchID, "", nil
	}
	return matchID, parts[1], nil
}

func PlayerIndex(st *game.MatchState, userID uint) int {
	if st == nil {
		return -1
	}
	if st.Players[0] != nil && st.Players[0].UserID == userID {
		return 0
	}
	if st.Players[1] != nil && st.Players[1].UserID == userID {
		return 1
	}
	return -1
}

func FromContext(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(AuthKey).(AuthUser)
	return u, ok
}
