package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"net/http"
	"os"
	"strconv"
)

// тестовая тема, для проверки внутри браузера
func DevAuthRofl(stroe *middleware.TokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_AUTH") != "1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uidStr := r.URL.Query().Get("user_id")
		uid64, err := strconv.ParseUint(uidStr, 10, 64)
		if err != nil || uid64 == 0 {
			http.Error(w, "bad user_id", http.StatusBadRequest)
			return
		}
		token, exp, err := stroe.Issue(uint(uid64), 0)
		if err != nil {
			http.Error(w, "somethin went wrong", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  exp,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusNoContent)
	}
}
