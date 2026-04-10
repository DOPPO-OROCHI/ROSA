package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/player"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
)

type AuthDevDeps struct {
	DB    *gorm.DB
	Store *middleware.TokenStore
}

type authDevReq struct {
	Username string `json:"username"`
}

func NewAuthDevHandler(d AuthDevDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.WriteErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if d.Store == nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "auth store is not configured")
			return
		}

		var req authDevReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "bad json")
			return
		}

		username := strings.TrimSpace(req.Username)
		if username == "" {
			username = "dev_player"
		}

		var user player.TelegramUser
		err := d.DB.Where("username = ?", username).First(&user).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				middleware.WriteErr(w, http.StatusInternalServerError, "failed to load user")
				return
			}

			user = player.TelegramUser{
				TGID:      -int64(len(username) + 1000),
				Username:  username,
				FirstName: "Dev",
				LastName:  "User",
				Language:  "ru",
			}
			if createErr := d.DB.Create(&user).Error; createErr != nil {
				middleware.WriteErr(w, http.StatusInternalServerError, "failed to create user")
				return
			}
		}

		token, exp, err := d.Store.Issue(user.ID, int(user.TGID))
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "failed to create session")
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
			"user_id":  user.ID,
			"username": user.Username,
		})
	}
}
