package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type CreateMatchRequest struct {
	OpponentUserID uint `json:"opponent_user_id"`
}

type GetMatchHandlerDeps struct {
	DB *gorm.DB
}

type CreateMatchHandlerDeps struct {
	DB *gorm.DB
}

type MathesListHandlerDeps struct {
	DB *gorm.DB
}

func NewCreateMatchHandler(d CreateMatchHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		var req CreateMatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong")
			return
		}
		if req.OpponentUserID == 0 {
			middleware.WriteErr(w, http.StatusBadRequest, "opponent_user_id is required")
			return
		}
		if req.OpponentUserID == userID {
			middleware.WriteErr(w, http.StatusBadRequest, "cannot play against yourself")
			return
		}
		var p1 player.TelegramUser
		if err := d.DB.Select("id", "selected_hero_template_id").
			Where("id = ?", userID).
			First(&p1).Error; err != nil {
			middleware.WriteErr(w, http.StatusUnauthorized, "user not found")
			return
		}
		if p1.SelectedHeroTemplateID == nil {
			middleware.WriteErr(w, http.StatusBadRequest, "select hero first")
			return
		}
		var p1Tpl heroes.CharacterTemplate
		if err := d.DB.Select("id", "character_code").
			Where("id = ?", *p1.SelectedHeroTemplateID).
			First(&p1Tpl).Error; err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "selected hero template not found")
			return
		}
		p1HeroCode := p1Tpl.CharacterCode
		// p2: выбранный герой оппонента
		var p2 player.TelegramUser
		if err := d.DB.Select("id", "selected_hero_template_id").
			Where("id = ?", req.OpponentUserID).
			First(&p2).Error; err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "opponent not found")
			return
		}
		if p2.SelectedHeroTemplateID == nil {
			middleware.WriteErr(w, http.StatusBadRequest, "opponent has no selected hero")
			return
		}
		var p2Tpl heroes.CharacterTemplate
		if err := d.DB.Select("id", "character_code").
			Where("id = ?", *p2.SelectedHeroTemplateID).
			First(&p2Tpl).Error; err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "opponent selected hero template not found")
			return
		}
		p2HeroCode := p2Tpl.CharacterCode
		st, err := repository.CreateMatchTX(
			d.DB,
			userID,
			req.OpponentUserID,
			p1HeroCode,
			p2HeroCode,
		)
		if err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(st, userID))
	}
}

func NewGetMatchHandler(d GetMatchHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil || tail != "" {
			middleware.WriteErr(w, http.StatusNotFound, "not found")
			return
		}
		var row repository.Match
		if err := d.DB.First(&row, matchID).Error; err != nil {
			middleware.WriteErr(w, http.StatusNotFound, "match not found")
			return
		}
		if row.PlayerID1 != userID && row.PlayerID2 != userID {
			middleware.WriteErr(w, http.StatusForbidden, "not a participant")
			return
		}
		var st game.MatchState
		if err := json.Unmarshal(row.State, &st); err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "bad match state json")
			return
		}
		st.Version = row.Version
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(&st, userID))
	}
}

func NewMathesListHandler(d MathesListHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		var rows []repository.Match
		if err := d.DB.Where("player_id1 = ? OR player_id2 = ?", userID, userID).
			Order("updated_at DESC").Limit(50).Find(&rows).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		out := make([]*dto.MaskedMatchState, 0, len(rows))
		for _, row := range rows {
			var st game.MatchState
			if err := json.Unmarshal(row.State, &st); err != nil {
				middleware.WriteErr(w, http.StatusInternalServerError, "bad match state JSON")
				return
			}
			st.Version = row.Version
			out = append(out, maskMatchStateForUser(&st, userID))
		}
		middleware.WriteJSON(w, http.StatusOK, out)
	}
}
