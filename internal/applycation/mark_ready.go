package applycation

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/repository"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

func MarkMatchReadyTX(db *gorm.DB, matchID uint, userID uint) (*game.MatchState, error) {
	for attempt := 0; attempt < 3; attempt++ {
		var out *game.MatchState
		err := db.Transaction(func(tx *gorm.DB) error {
			var row repository.Match
			if err := tx.First(&row, matchID).Error; err != nil {
				return err
			}
			playerIndex := -1
			switch userID {
			case row.PlayerID1:
				playerIndex = 0
			case row.PlayerID2:
				playerIndex = 1
			default:
				return ErrNotParticipant
			}
			expectedDBVersion := row.Version
			var st game.MatchState
			if err := json.Unmarshal(row.State, &st); err != nil {
				return ErrCorruptedMatchState
			}
			st.Version = expectedDBVersion
			if _, err := game.MarkLoadingReady(&st, playerIndex, time.Now().Unix()); err != nil {
				return err
			}
			if st.Version == expectedDBVersion {
				out = &st
				return nil
			}
			newJSON, err := json.Marshal(&st)
			if err != nil {
				return err
			}
			if err := repository.SaveMatchState(tx, row.ID, expectedDBVersion,
				newJSON, st.Version, st.Finished, st.TurnDeadline); err != nil {
				return err
			}
			out = &st
			return nil
		})
		if errors.Is(err, game.ErrStaleAction) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, game.ErrStaleAction
}
