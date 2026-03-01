package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/repository"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type DeckHandlerDeps struct {
	DB *gorm.DB
}

func NewGetDeckHandler(d DeckHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		entries, err := repository.LoadDeckTx(d.DB, au.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				middleware.WriteJSON(w, http.StatusOK, dto.DeckResponce{Entries: []game.DeckEntry{}})
				return
			}
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		if entries == nil {
			entries = []game.DeckEntry{}
		}
		middleware.WriteJSON(w, http.StatusOK, dto.DeckResponce{Entries: entries})
	}
}
func NewSaveDeckHandler(d DeckHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var req dto.SaveDeckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong")
			return
		}
		if len(req.Entries) == 0 {
			middleware.WriteErr(w, http.StatusBadRequest, "empty deck")
			return
		}
		err := d.DB.Transaction(func(tx *gorm.DB) error {
			battleMax, buffMax, err := repository.LoadTemplateLimits(tx)
			if err != nil {
				return err
			}
			_, ownedBattleCopies, err := repository.LoadOwnedBattleCards(tx, au.UserID)
			if err != nil {
				return err
			}
			_, ownedBuffCopies, err := repository.LoadOwnedBuff(tx, au.UserID)
			if err != nil {
				return err
			}
			if err := game.ValidateDeckList(req.Entries, battleMax, buffMax, ownedBattleCopies, ownedBuffCopies); err != nil {
				return err
			}
			return repository.SaveDeckTx(tx, au.UserID, req.Entries)
		})
		if err != nil {
			code := http.StatusInternalServerError
			if isDeckValidationErr(err) {
				code = http.StatusBadRequest
			}
			middleware.WriteErr(w, code, "something went wrong")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "deck saved"})
	}
}
func isDeckValidationErr(err error) bool {
	switch err {
	case game.ErrDeckSizeNot20,
		game.ErrDeckCountInvalid,
		game.ErrDeckTooManyCopies,
		game.ErrDeckNotOwnedEnough,
		game.ErrDeckUnknownCard,
		game.ErrDeckUnknownKind:
		return true
	default:
		return false
	}
}
