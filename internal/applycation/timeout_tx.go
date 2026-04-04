package applycation

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/repository"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

//файл полностью посвящен для функции, которая применяет таймаут к матчу в рамках транзакции БД.

var (
	ErrNotParticipant      = errors.New("not a participant")
	ErrCorruptedMatchState = errors.New("corrupted match state")
)

func ApplyTimeOutToMatchTX(db *gorm.DB, matchID uint) (st *game.MatchState, changed bool, err error) {
	now := time.Now().Unix()
	err = db.Transaction(func(tx *gorm.DB) error {
		var row repository.Match
		if err := tx.First(&row, matchID).Error; err != nil {
			return err
		}
		expected := row.Version
		var state game.MatchState
		if err := json.Unmarshal(row.State, &state); err != nil {
			return ErrCorruptedMatchState
		}
		state.Version = expected
		ch, err := game.ForceTimeOut(&state, now)
		if err != nil {
			return err
		}
		if !ch {
			changed = false
			return nil
		}
		state.Version++
		justFinished := !row.Finished && state.Finished
		if justFinished {
			switch state.Result {
			case game.MatchWinP1:
				winnerID := row.PlayerID1
				loserID := row.PlayerID2
				if err := repository.RatingUp(winnerID, tx); err != nil {
					return err
				}
				if err := repository.RatingDown(loserID, tx); err != nil {
					return err
				}
			case game.MatchWinP2:
				winnerID := row.PlayerID2
				loserID := row.PlayerID1
				if err := repository.RatingUp(winnerID, tx); err != nil {
					return err
				}
				if err := repository.RatingDown(loserID, tx); err != nil {
					return err
				}
			case game.MatchDraw:
			}
		}
		newJSON, err := json.Marshal(&state)
		if err != nil {
			return ErrCorruptedMatchState
		}
		if err := repository.SaveMatchState(tx, row.ID, expected, newJSON, state.Version, state.Finished, state.TurnDeadline); err != nil {
			return err
		}
		st = &state
		changed = true
		return nil
	})
	return st, changed, err
}
