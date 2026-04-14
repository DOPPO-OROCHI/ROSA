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

/*Короче пиздеть не буду, это полностью спизженный концепт у ИИ терминатора. Я вообще не ебу че здесь происходит.
Причина по которой я это говорю заключается в том, что этот блок будет переписываться раз наверное 150, поскольку
текущая авторизация не устраивает ни меня, ни ИИ который это придумал. НО) Она рабочая блять. На основе нее пока
что и работает аутентификация) Поэтому пока что это есть. Но этого вскоре не будет... Это да...*/

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
			middleware.WriteErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if d.Store == nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "auth store is not configured")
			return
		}
		var req authTelegramReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.InitData == "" {
			middleware.WriteErr(w, http.StatusBadRequest, "bad json")
			return
		}
		tgID, err := middleware.ValidateTelegramInitData(req.InitData, botToken, initDataMaxAge)
		if err != nil {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var dbUser player.TelegramUser
		if err := d.DB.Where("tg_id = ?", tgID).First(&dbUser).Error; err != nil {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token, exp, err := d.Store.Issue(dbUser.ID, tgID)
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		secure := os.Getenv("COOKIE_SECURE") != "0"
		sameSite := http.SameSiteLaxMode
		if secure {
			sameSite = http.SameSiteNoneMode
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  exp,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})
		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"user_id": dbUser.ID,
			"tg_id":   tgID,
		})
	}
}
