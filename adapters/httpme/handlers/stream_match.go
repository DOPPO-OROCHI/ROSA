package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/transport"
	"fmt"
	"io"
	"net/http"
	"time"
)

type StreamMatchDeps struct {
	Hub   *transport.Hub
	Store *middleware.TokenStore
}

func NewStreamMatchHandler(deps StreamMatchDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil || tail != "stream" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		au, ok := middleware.FromContext(r.Context())
		if !ok || au.UserID == 0 {
			tok := r.URL.Query().Get("token")
			if tok == "" || deps.Store == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			sess, valid := deps.Store.Validate(tok)
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			au = middleware.AuthUser{UserID: sess.UserID, TGID: sess.TGID}
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		key := transport.StreamKey{MatchID: matchID, ViewerUserID: au.UserID}
		ch := deps.Hub.Subscribe(key)
		defer deps.Hub.Unsubscribe(key, ch)
		keepAlive := time.NewTicker(15 * time.Second)
		defer keepAlive.Stop()
		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				writeSSE(w, "state", msg)
				flusher.Flush()
			case <-keepAlive.C:
				io.WriteString(w, ": ping\n\n")
				flusher.Flush()
			}
		}
	}
}

func writeSSE(w io.Writer, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}
