package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/player"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

type AuthTelegramDeps struct {
	DB    *gorm.DB
	Store *middleware.TokenStore
}

type authTelegramReq struct {
	InitData string `json:"initData"`
}

func NewAuthTelegramHandler(d AuthTelegramDeps) http.HandlerFunc {
	botToken := os.Getenv("BOT_API")
	if botToken == "" {
		panic("BOT_API empty")
	}
	initDataMaxAge := 5 * time.Minute
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req authTelegramReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.InitData == "" {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		tgID, err := middleware.ValidateTelegramInitData(req.InitData, botToken, initDataMaxAge)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var dbUser player.TelegramUser
		if err := d.DB.Where("tg_id = ?", tgID).First(&dbUser).Error; err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token, exp, err := d.Store.Issue(dbUser.ID, tgID)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		secure := os.Getenv("COOKIE_SECURE") != "0"
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  exp,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteNoneMode,
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id": dbUser.ID,
			"tg_id":   tgID,
		})
	}
}
