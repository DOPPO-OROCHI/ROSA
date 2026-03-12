package applycation

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/internal/domain/game"
	"TheWar/internal/infra/repository"
	"encoding/json"

	"gorm.io/gorm"
)

/*Данный файл целиком и полностью посвящен изменению состояния матча в зависимости от дейсвтия игрока внутри
строки этого же матча в БД. Смысл в чем? Каждый матч хранится в БД в реальном времени. В нем есть теущие игроки,
состояния матча, карты игроко, персонажи и все то, что необходимо в целом для игры. Но самое главное что в нем
есть это версия конкретного действия. Эта версия нужна для того, чтобы избежать повторные нажатия, абузы и всего
того, что может расцениться как нечестная игра. Ну и для реконнекта да. Но не суть. Перейдем к функции*/

/*
Функция для обновления состояния матча транзакцией. Почему транзакцией думаю объяснять не надо (потому что либо
меняем все, либо не меняем ничего). Во входящих аргументах принимаем переменную БД матч, айдишник юзера, дейсвтие
игрока, который он послал в наш сервер и резолверы карт. А теперь к делу
*/
func ApplyActionToMatchTx(db *gorm.DB,
	matchID uint,
	userID uint,
	req dto.ApplyActionRequest,
	r game.Resolvers) (*game.MatchState, error) {
	//вводим переменную состояния матча
	var out *game.MatchState
	//начинаем транзакцию, где мы...
	err := db.Transaction(func(tx *gorm.DB) error {
		var row repository.Match //<-вводим переменную матча, чтобы записать в нее текущее состояние конкретного
		if err := tx.First(&row, matchID).Error; err != nil {
			return err
		}
		playerIndex := -1
		switch userID { //<-смотрим на пользователей и если что вдруг...
		case row.PlayerID1:
			playerIndex = 0
		case row.PlayerID2:
			playerIndex = 1
		default:
			return ErrNotParticipant //<-говорим что тот, кто пытается обновить состояние не участник мачта
		}
		expectedDBVersion := row.Version                       //<-обязательно записываем версию, чтобы у нас был optimistic locking
		var st game.MatchState                                 //<-а сюда будем десериализировать состояние матча
		if err := json.Unmarshal(row.State, &st); err != nil { //<-ищем в БД текущее состояние матча
			return ErrCorruptedMatchState
		}
		st.Version = expectedDBVersion //<-синхронизирую доменную версию с БД
		act := game.Action{            //<-а здесь описывает вообще что за действие. Над каким юнитом, кого атакуем, кем и так далее
			PlayerIndex:      playerIndex,
			Type:             req.Type,
			CardInstanceID:   req.CardInstanceID,
			TargetInstanceID: req.TargetInstanceID,
			AttackHero:       req.AttackHero,
			ExpectedVersion:  req.ExpectedVersion,
			TargetSlot:       req.TargetSlot,
		}
		if err := game.ApplyAction(&st, act, r); err != nil { //<-применяем действие
			return err
		}
		newJSON, err := json.Marshal(&st) //<-маршалим новое состояние обратно в JSON
		if err != nil {
			return err
		}
		if err := repository.SaveMatchState(tx, row.ID, expectedDBVersion, newJSON, st.Version, st.Finished, st.TurnDeadLineAt); err != nil {
			return err
		} //<-обновляем строку конкретного матча атомарно
		out = &st //<-назначаем аут
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil //<-возвращаем новое состояние
}

/*Таким образом производится изменение матча. Любые действия игрока отражаются здесь. Ключевой момент в этой функции
-атомарное обновление БД. Естественно перед этим надо прочитать вводные, который юзер отправляет. Я читаю JSON и обновляю
собственно состояние матча, описывая сначала само действие, потом вызывая ApplyAction, в котором уже и лежат все обработчики
со ввсеми возможными в рамках игры сценариями*/
