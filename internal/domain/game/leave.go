package game

import "errors"

/*А этот файл отвечает онли за то, чтобы здесь написать функцию лива из матча. Добавил ее слишком поздно,
но добавил... Тем не менее, го разберем*/

// принимаем состояние матча и игрока. В случае чего отдаем ошибку
func LeaveMatch(m *MatchState, playerIndex int) error {
	//если матч пуст, отдаем ошибку
	if m == nil {
		return errors.New("nil match state")
	}
	//если матч закончен-тоже
	if m.Finished {
		return ErrMatchFinished
	}
	//если игрока нет в матче, то тоже отдаем ошибку. Считай такого игрока не существует
	if playerIndex != 0 && playerIndex != 1 {
		return errors.New("bad player index")
	}
	//завершаем матч, путем присвоения статуса Finished=true
	m.Finished = true
	//и отдаем победу тому игроку, который НЕ инициировал лив
	if playerIndex == 0 {
		m.Result = MatchWinP2
	} else {
		m.Result = MatchWinP1
	}
	//пишем ивент, который и будет отвечать за анимацию выхода из матча
	m.Events = append(m.Events, Event{
		Type:        string(EventTurn),
		PlayerIndex: playerIndex,
		SourceKind:  string(SourceSystem),
	})
	return nil
}
