package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/applycation"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"
	"TheWar/internal/transport"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type ApplyActionHandlerDeps struct {
	DB        *gorm.DB
	Resolvers game.Resolvers
	Hub       *transport.Hub
}

func NewApplyActionHandler(a ApplyActionHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil || tail != "actions" {
			middleware.WriteErr(w, http.StatusNotFound, "not found")
			return
		}
		var req dto.ApplyActionReplace
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "bad json")
			return
		}
		newState, err := applycation.ApplyActionToMatchTx(a.DB, matchID, userID, req, a.Resolvers)
		if err != nil {
			switch {
			case errors.Is(err, applycation.ErrNotParticipant):
				middleware.WriteErr(w, http.StatusForbidden, "forbidden")
				return
			case errors.Is(err, applycation.ErrCorruptedMatchState):
				middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
				return
			default:
				middleware.WriteErr(w, MapEngineErr(err), "something went wrong")
				return
			}
		}
		PublishMatchToSSE(a.Hub, newState)
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(newState, userID))
	}
}
func MapEngineErr(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	//если челик борщанул с кликами
	case errors.Is(err, game.ErrStaleAction):
		return http.StatusConflict
		//право на действие
	case errors.Is(err, game.ErrNotYourTurn):
		return http.StatusForbidden
		//ошибка фазы
	case errors.Is(err, game.ErrMatchFinished):
		return http.StatusConflict
		//ресурсы
	case errors.Is(err, game.ErrNotEnoughMana):
		return http.StatusBadRequest
		//ошибки стола
	case errors.Is(err, game.ErrTablesFull):
		return http.StatusBadRequest
	case errors.Is(err, game.ErrSlotOccupied):
		return http.StatusBadRequest
		//правила боя
	case errors.Is(err, game.ErrAttackerOnCooldown),
		errors.Is(err, game.ErrAttackerSummoneddThisTurn),
		errors.Is(err, game.ErrMustAttackTank),
		errors.Is(err, game.ErrCannotAttackHeroWithTanks),
		errors.Is(err, game.ErrCannotHitHeroWhileTanks),
		errors.Is(err, game.ErrHeroOnCooldown),
		errors.Is(err, game.ErrHeroAttackIsZero),
		errors.Is(err, game.ErrHealerCannotAttack):
		return http.StatusBadRequest
	//дека
	case errors.Is(err, game.ErrDeckSizeNot20),
		errors.Is(err, game.ErrDeckCountInvalid),
		errors.Is(err, game.ErrDeckTooManyCopies),
		errors.Is(err, game.ErrDeckNotOwnedEnough),
		errors.Is(err, game.ErrDeckUnknownCard),
		errors.Is(err, game.ErrDeckUnknownKind):
		return http.StatusBadRequest
	//способности героя
	case errors.Is(err, game.ErrHeroAbilityOnCooldown),
		errors.Is(err, game.ErrHeroAbilityBadTarget),
		errors.Is(err, game.ErrHeroAbilityUnknown):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
func NewSelectedHeroHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		var req dto.SelectedHeroRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong") //<-пока тестово отдаю такое, во избежание раскрытия архитектуры БД
			return
		}
		var tpl heroes.CharacterTemplate
		if err := db.Where("character_code = ?", req.HeroCode).First(&tpl).Error; err != nil {
			middleware.WriteErr(w, http.StatusNotFound, "something went wrong")
			return
		}
		var owned repository.GamerCharacter
		if err := db.Where("gamer_id = ? AND character_template_id = ?", userID, tpl.ID).First(&owned).Error; err != nil {
			middleware.WriteErr(w, http.StatusForbidden, "something went wrong")
			return
		}
		if err := db.Model(&player.TelegramUser{}).Where("id = ?", userID).Update("selected_hero_template_id", tpl.ID).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "hero selected",
		})
	}
}
