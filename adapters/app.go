package adapters

import (
	"TheWar/adapters/httpme/middleware"
	"net/http"
)

type App struct {
	CreateMatch http.HandlerFunc
	GetMatch    http.HandlerFunc
	ApplyAction http.HandlerFunc

	GetMe        http.HandlerFunc
	UpdateAvatar http.HandlerFunc

	SaveDeck http.HandlerFunc
	GetDeck  http.HandlerFunc

	CardsList   http.HandlerFunc
	HeroesList  http.HandlerFunc
	MatchesList http.HandlerFunc

	SelectHero   http.HandlerFunc
	StreamMatch  http.HandlerFunc
	AuthTelegram http.HandlerFunc
}

func NewMux(app App) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/matches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			app.CreateMatch(w, r)
		case http.MethodGet:
			app.MatchesList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/matches/", func(w http.ResponseWriter, r *http.Request) {
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_ = matchID
		switch {
		case r.Method == http.MethodGet && tail == "":
			app.GetMatch(w, r)
			return
		case r.Method == http.MethodPost && tail == "actions":
			app.ApplyAction(w, r)
			return
		case r.Method == http.MethodGet && tail == "stream":
			app.StreamMatch(w, r)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetMe(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/deck", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetDeck(w, r)
		case http.MethodPost:
			app.SaveDeck(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.CardsList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/heroes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.HeroesList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/heroes/select", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.SelectHero(w, r)
	})
	mux.HandleFunc("/auth/telegram", func(w http.ResponseWriter, r *http.Request) {
		app.AuthTelegram(w, r)
	})
	return mux
}
