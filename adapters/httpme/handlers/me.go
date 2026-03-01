package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"net/http"

	"gorm.io/gorm"
)

type GetMeHandler struct {
	DB *gorm.DB
}

func NewGetMeHandler(d GetMeHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var u player.TelegramUser
		if err := d.DB.First(&u, au.UserID).Error; err != nil {
			middleware.WriteErr(w, http.StatusUnauthorized, "user not found")
			return
		}
		resp := dto.MeResponse{
			UserID:              u.ID,
			TGID:                u.TGID,
			Username:            u.Username,
			FirstName:           u.FirstName,
			Rating:              u.Rating,
			XP:                  u.XP,
			SelectedHeroRequest: u.SelectedHeroTemplateID,
		}
		if u.SelectedHeroTemplateID != nil {
			var tpl heroes.CharacterTemplate
			if err := d.DB.Select("id", "character_code", "name").First(&tpl, *u.SelectedHeroTemplateID).Error; err == nil {
				resp.SelectedHeroCode = tpl.CharacterCode
				resp.SelectedHeroName = tpl.Name
			}
		}
		middleware.WriteJSON(w, http.StatusOK, resp)
	}
}
