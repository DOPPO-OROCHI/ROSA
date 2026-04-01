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

/*Ну че, подходим к основным хендлерам по созданию матча. По аналогии с предыдущими файлами, здесь есть все структуры
с зависимостями, которые нам нужны (в будущем смогу расширяться, к примеру, когда я решу добавить reddis). Так и вот...*/

// пока что поиск матча происходит по айди противника. Пока что это работает ультра криво, поскольку пока что матчи не учитывают
// такую ненужную на этапе сырой разработки вещь как очередь, или хз, подбор по рейтингу. Поэтому ПОКА ЧТО реализовано все так.
type CreateMatchRequest struct {
	OpponentUserID uint `json:"opponent_user_id"`
}

// Все матчи игрока хранятся в БД, соответственно нам нужен доступ к БД
type GetMatchHandlerDeps struct {
	DB *gorm.DB
}

// аналогично
type CreateMatchHandlerDeps struct {
	DB *gorm.DB
}

// аналогично
type MathesListHandlerDeps struct {
	DB *gorm.DB
}

/*
Хендлер, который служит для того, чтобы собственно начать матч. Здесь мы принимаем айди противника, отдавая
хендлер фанк в наш мукс. В целом все...
*/
func NewCreateMatchHandler(d CreateMatchHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //<-проверка на аутентификацию
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID //<- грузим игрока в память
		var req CreateMatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { //<-декодим джисонку, которую отправил игрок
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong")
			return
		}
		if req.OpponentUserID == 0 { //<-если айдишник оппонента равен нулю, то отдаем ошибку
			middleware.WriteErr(w, http.StatusBadRequest, "opponent_user_id is required")
			return
		}
		if req.OpponentUserID == userID { //<-нельзя создать матч против самого себя
			middleware.WriteErr(w, http.StatusBadRequest, "cannot play against yourself")
			return
		}
		st, err := createMatchForUsers(d.DB, userID, req.OpponentUserID)
		if err != nil {
			if errors.Is(err, repository.ErrActiveMatchExists) {
				middleware.WriteErr(w, http.StatusConflict, err.Error())
				return
			}
			middleware.WriteErr(w, http.StatusBadRequest, err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(st, userID))
	}
}

/*Таким образом, в этом хендлере реализована логика создания матча, которая учитывает двух игроков и главным
образом валидирует непосредственно персонажей, с которыми пришел игрок, доставая данные из запроса в память
все о тех же персонажах (в структуру heroes.CharacterTemplate), чтобы иметь возможность работать с ними. По
существу говоря, этой проверки здесь быть не должно и она должна быть вынесена в тот же CreateMatchTX, но...
первый шаг к решению проблемы это ее принятие. Верно ?...*/

/*
Необходимый хендлер, когда вопросы доходят до дисконнектов и прочей херни, которая может выбить из ровновесия
пользователя. В чем прикол ? Он проверяет айдищник пользователя на предмет того, является ли он вообще участником
конкретного матча. Если нет то...
*/
func NewGetMatchHandler(d GetMatchHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-база
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID                                         //<-тоже база только про память
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path) //<-парсим патч на предмет нахождения нужного айди
		if err != nil || tail != "" {
			middleware.WriteErr(w, http.StatusNotFound, "not found")
			return
		}
		var row repository.Match //<-нужная переменная чтобы записать в нее текущее состояние матча
		if err := d.DB.First(&row, matchID).Error; err != nil {
			middleware.WriteErr(w, http.StatusNotFound, "match not found")
			return
		}
		if row.PlayerID1 != userID && row.PlayerID2 != userID { //<-вот кстати для этого и нужная (как один из факторов)
			//если игрок не является участником матча -нахуй
			middleware.WriteErr(w, http.StatusForbidden, "not a participant")
			return
		}
		var st game.MatchState                                 //<-сюда будем записывать все то, что нужно для отображения текущего состояния матча
		if err := json.Unmarshal(row.State, &st); err != nil { //<-а здесь мы десериализируем состояние матча в наш ST
			middleware.WriteErr(w, http.StatusInternalServerError, "bad match state json") //<-понятно
			return
		}
		st.Version = row.Version //<-обязательно отдаем версию, потому что от нее зависит буквально
		// все, что касается случайных нажатий, обрывов и тд
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(&st, userID)) //<-отдаем пользователю замаскированное состояние матча
	}
}

/*Логика такова, что при отключении от матча, игрок все равно должен иметь возможность вернуться в него
Для того, чтобы сделать это безболезненно, я должен пойти в конкретный матч, который находится в репо, достать
из конкретного матча все состояние на актуальный момент, а потом, вернуть это состояние пользователю. Для этого
я десериализирую состояние мачта в переменную ST, отдавая ее при помощи функции маскировщика игроку. И получается,
что при такой конфигурации игрок, который по каким то причинам ливнул с карты до фактического завершения матча
может в него вернуться без никаких проблем*/

/*Хендлер получения списка 50 последних мачтей игрока. Нужен главнм образом просто для просмотра истории.*/
func NewMathesListHandler(d MathesListHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context())
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID
		/*И тут встала моральная диллема. Вот в чем рофл. Это очень опасная вещь под абьюз. В чем она заключается ?
		Созданный матч уже отображается в БД, так ? А теперь смотри мем, если чувак просто выйдет, нажмет на историю
		матчей и все такое, то он сможет увидеть вообще все состояние, если мы об этом не позаботимся. Это значит, что
		он потенциально увидит деку, руку и все то, что может помешгать игре в истории. Этого быть не должно, поэтому
		далее я буду маскировать и историю матчей. Похуй, пока так.
		Окей, в будущем я добавлю отдельную БД под уже завершенные матчи с их полным состоянием, но не сейчас.*/
		var rows []repository.Match //<-вводим слайс матчей в память, чтобы его заполнять
		if err := d.DB.Where("player_id1 = ? OR player_id2 = ?", userID, userID).
			Order("updated_at DESC").Limit(50).Find(&rows).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		out := make([]*dto.MaskedMatchState, 0, len(rows)) //<-это нужно для того, чтобы
		for _, row := range rows {
			var st game.MatchState
			if err := json.Unmarshal(row.State, &st); err != nil { //<-десериализировать состояние в переменную
				middleware.WriteErr(w, http.StatusInternalServerError, "bad match state JSON")
				return
			}
			st.Version = row.Version                              //<-записать версию
			out = append(out, maskMatchStateForUser(&st, userID)) //<-и добавить спрятанные данные в этот же out
		}
		middleware.WriteJSON(w, http.StatusOK, out) //<-после чего отдать матчи пользователю
	}
}

/*Таким образом логика такая. Достаем последние 50 матчей, отдаем пользователю. Ничего сложного*/
