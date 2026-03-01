package handlers

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/transport"
	"encoding/json"
)

func PublishMatchToSSE(hub *transport.Hub, st *game.MatchState) {
	if hub == nil || st == nil {
		return
	}
	for i := 0; i < 2; i++ {
		p := st.Players[i]
		if p == nil || p.UserID == 0 {
			continue
		}
		masked := maskMatchStateForUser(st, p.UserID)
		b, err := json.Marshal(masked)
		if err != nil {
			continue
		}
		hub.Publish(transport.StreamKey{MatchID: st.MatchID, ViewerUserID: p.UserID}, b)
	}
}
