package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/infra/repository"
	"net/http"

	"gorm.io/gorm"
)

type CardListHandlerDeps struct {
	DB *gorm.DB
}

func NewCardsListHandler(d CardListHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		battleRows, err := repository.LoadOwnedBattleCardsRows(d.DB, userID)
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		buffRows, err := repository.LoadOwnedBuffCardsRows(d.DB, userID)
		if err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		out := dto.CardsListResponse{
			Battle: make([]dto.OwnedBattleCardsDTO, 0, len(battleRows)),
			Buff:   make([]dto.OwnedBuffCardsDTO, 0, len(buffRows)),
		}
		for _, r := range battleRows {
			t := r.Tpl
			out.Battle = append(out.Battle, dto.OwnedBattleCardsDTO{
				Kind:         dto.CardKindBattle,
				TemplateID:   t.CodeString,
				Name:         t.Name,
				CardType:     t.CardType,
				ManaCost:     t.ManaCost,
				HealthPoints: t.HealthPoints,
				Attack:       t.Attack,
				SplashRadius: t.SplashRadius,
				Cooldown:     t.CoolDown,
				IsTank:       t.IsTank,
				BuffSlot:     t.BuffSlot,
				MaxCopies:    t.MaxCopies,
				OwnedCardID:  r.OwnedID,
				Copies:       r.Copies,
				Level:        r.Level,
				XP:           r.XP,
				ImageKey:     t.ImageKey,
				AssetBaseKey: t.AssetBaseKey,
			})
		}
		for _, r := range buffRows {
			t := r.Tpl
			out.Buff = append(out.Buff, dto.OwnedBuffCardsDTO{
				Kind:        dto.CardKindBuff,
				TemplateID:  t.CodeString,
				Name:        t.Name,
				ManaCost:    t.ManaCost,
				BuffType:    t.BuffType,
				BuffValue:   t.BuffValue,
				OnlyFor:     t.OnlyFor,
				Duration:    t.Duration,
				MaxCopies:   t.MaxCopies,
				OwnedCardID: r.OwnedID,
				Copies:      r.Copies,
				Level:       r.Level,
				XP:          r.XP,
			})
		}
		middleware.WriteJSON(w, http.StatusOK, out)
	}
}
