package game

import "errors"

func LeaveMatch(m *MatchState, playerIndex int) error {
	if m == nil {
		return errors.New("nil match state")
	}
	if m.Finished {
		return ErrMatchFinished
	}
	if playerIndex != 0 && playerIndex != 1 {
		return errors.New("bad player index")
	}

	m.Finished = true
	if playerIndex == 0 {
		m.Result = MatchWinP2
	} else {
		m.Result = MatchWinP1
	}

	m.Events = append(m.Events, Event{
		Type:        string(EventTurn),
		PlayerIndex: playerIndex,
		SourceKind:  string(SourceSystem),
	})

	return nil
}
