package game

func ForceTimeOut(m *MatchState, nowUnix int64) (bool, error) {
	if m.Finished {
		return false, ErrMatchFinished
	}
	if m.Phase == PhaseStart {
		StartTurn(m, nowUnix)
	}
	if m.Phase != PhaseMain || m.TurnDeadLineAt <= 0 || nowUnix <= m.TurnDeadLineAt {
		return false, nil
	}
	m.Events = m.Events[:0]
	m.Events = append(m.Events, Event{
		Type:        "turn_timeout",
		PlayerIndex: m.ActivePlayer,
	})
	EndTurn(m)
	if !m.Finished {
		StartTurn(m, nowUnix)
	}
	return true, nil
}
