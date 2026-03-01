package applycation

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/repository"
	"encoding/json"

	"gorm.io/gorm"
)

//файл полностью посвящен для функции, которая применяет действие к матчу в рамках транзакции БД.

func ApplyActionToMatchTx(db *gorm.DB,
	matchID uint,
	userID uint,
	req dto.ApplyActionReplace,
	r game.Resolvers) (*game.MatchState, error) {
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
		act := game.Action{
			PlayerIndex:      playerIndex,
			Type:             req.Type,
			CardInstanceID:   req.CardInstanceID,
			TargetInstanceID: req.TargetInstanceID,
			AttackHero:       req.AttackHero,
			ExpectedVersion:  req.ExpectedVersion,
			TargetSlot:       req.TargetSlot,
		}
		if err := game.ApplyAction(&st, act, r); err != nil {
			return err
		}
		newJSON, err := json.Marshal(&st)
		if err != nil {
			return err
		}
		if err := repository.SaveMatchState(tx, row.ID, expectedDBVersion, newJSON, st.Version, st.TurnDeadLineAt); err != nil {
			return err
		}
		stCopy := st
		out = &stCopy
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
