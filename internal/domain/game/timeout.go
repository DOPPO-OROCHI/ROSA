package game

//Функция, которая отвечает за проверку таймаута хода и если да, принудительно завершает ход
func ForceTimeOut(m *MatchState, nowUnix int64) (bool, error) {
	if m.Finished { //<-проверяем матч на завршение
		return false, ErrMatchFinished
	}
	if m.Phase == PhaseStart {
		return false, nil
	}
	if m.Phase != PhaseMain || m.TurnDeadline <= 0 || nowUnix < m.TurnDeadline { //<-если таймаут еще не наступил
		return false, nil //<-возвращаем фолс, который в смысле функции означает -не делаем ничего
	}
	m.Events = m.Events[:0]            //<-чистим ивентный массив, чтобы не засирать пользователя лишними анимациями
	m.Events = append(m.Events, Event{ //<-добавляем ивент завершения хода по истечению времени
		Type:        "turn_timeout",
		PlayerIndex: m.ActivePlayer,
	})
	EndTurn(m) //<-заканчиваем ход
	return true, nil
}
