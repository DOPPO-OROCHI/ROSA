package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/infra/repository"
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type HeroListHandler struct {
	DB *gorm.DB
}

func NewHeroesListHandler(d HeroListHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if r.Method != http.MethodGet {
			middleware.WriteErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		var owned []repository.GamerCharacter
		if err := d.DB.Preload("CharacterTemplate").Where("gamer_id = ?", userID).Find(&owned).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		out := make([]dto.OwnedHeroDTO, 0, len(owned))
		for _, g := range owned {
			t := g.CharacterTemplate
			out = append(out, dto.OwnedHeroDTO{
				HeroID:         t.ID,
				HeroCode:       t.CharacterCode,
				Name:           t.Name,
				Level:          g.CharacterLevel,
				HealthPoints:   t.HealthPoints,
				AttackPower:    t.AttackPower,
				AttackCooldown: t.AttackCooldown,
				SplashRadius:   t.SplashRadius,
				Description:    t.Description,
				ImageKey:       t.ImageKey,
				AssetBaseKey:   t.AssetBaseKey,
			})
		}
		resp := dto.HeroesListResponce{Heroes: out}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
