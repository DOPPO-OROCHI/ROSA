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

/*Помним о том, что для расширяемости системы, зависимости надо засовывать в структуры, чтобы потом иметь возможность
добавить собственно зависимости, без раздувания входящих аргументов*
Так и вот. Данный файл представляет собой набор хендлеров для различных операций с деками. К примеру -посмотреть свою деку-
и -собрать новую деку-. В целом ничего нового я тут не делаю. давай разбираться*/

// база
type DeckHandlerDeps struct {
	DB *gorm.DB
}

// базированный прием аргументов и отдача HandlerFunc (к слову, я долго не мог понять, в чем разница Handler и Handle)
func NewGetDeckHandler(d DeckHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-все так же валидируем авторизацию
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		entries, err := repository.LoadDeckTx(d.DB, au.UserID) //<-достаем карты именно нашего пользователя
		//(кстати, вот пример хорошего хендлера, который сам ничего не валидирует)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				middleware.WriteJSON(w, http.StatusOK, dto.DeckResponce{Entries: []game.DeckEntry{}}) //<-в случае
				//ошибки отдаем пустую деку. Так то могли просто написать типа -Дека пустая, иди нахуй. Но мы порядочные
				return
			}
			//но бывает что клиент может в целом послать нахуй и меня. Так что вот так...
			middleware.WriteErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		/*А здесь важная херня. Смотри в чем прикол (сам бы я до этого не додумался). Есть два состояния
		деки. 1-деки вообще нет, физически. ErrRecordNotFound. Этот косяк мы уже отработали и в таком случае
		мы вернем пустую деку (не NIL). 2-дека есть, но она пустая по факту. В корне заеба лежит то простое
		обстоятельство, что это два разных состояния. Проблема в том, что для нормального пацана пишущего на ГО
		разницы между nil и [] по факту нет, но вот для JSON есть. Почему ? Потому что Nil может сказать о том, что
		значения как будто вообще нет, а вот [] говорит о том, что список есть, но в нем ничего нет, что заочно
		говорит клиенту о том, что список неплохо бы составить. Это не обязательная часть, но считается что
		это типа дружелюбный код!!!!1*/
		if entries == nil {
			entries = []game.DeckEntry{}
		}
		//отдаем клиенту его деку
		middleware.WriteJSON(w, http.StatusOK, dto.DeckResponce{Entries: entries})
	}
}

// А это хендлер, который позволяет записать деку из имеющихся у игрока карт
func NewSaveDeckHandler(d DeckHandlerDeps) http.HandlerFunc { //<-классическая преамбула
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-уже ставшая классикой проверка на ауф
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if r.Method != http.MethodPost { //<-проверяем метод
			middleware.WriteErr(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		var req dto.SaveDeckRequest //<-вводим переменную типа структуры запроса на деку, чтобы в нее декодировать тело запроса
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "something went wrong")
			return
		}
		if len(req.Entries) == 0 { //<-проверяем на то, реально ли челик заполнил деку или просто скинул херню (не критичная
			//секция, но проверить стоит. Потому что игра все таки WebApp, а значит кто-то да решит порофлить, скидывая кривые JSON)
			middleware.WriteErr(w, http.StatusBadRequest, "empty deck")
			return
		}
		/*После того, как клиент прислал нам валидный JSON с намерением сохранить деку, нам нужно эту деку обновить.
		Делается это в моем случае с помощью транзакций (поскольку я по другому не умею). Почему ? Потому что помимо
		прочих проверок на чистоту JSON, так же необходимо проверить и сохранить всю колоду как единое целое, перед этим
		проверяя а вообще есть такие карты или нет.*/
		err := d.DB.Transaction(func(tx *gorm.DB) error {
			//ключевая проверка. Дело в том, что у карт есть лимиты (сколько максимум можно с собой взять)
			//если клиент прислал JSON, который прошел по правилам и там лимиты больше (что с точки зрения кода
			//не является проблемой), я должен обработать этот сценарий и дать клиенту по рукам раньше, чем он
			//сможет сохранить хоть что либо.
			battleMax, buffMax, err := repository.LoadTemplateLimits(tx)
			if err != nil {
				return err
			}
			//грузим карты именно с владения. Там к слову и левелы, и копии, и прочее прочее что характерно для игрока
			_, ownedBattleCopies, err := repository.LoadOwnedBattleCards(tx, au.UserID)
			if err != nil {
				return err
			}
			//то же самое. По сути эти куски отвечают на вопрос, реально ли владеет игрок тем, что прислал
			_, ownedBuffCopies, err := repository.LoadOwnedBuff(tx, au.UserID)
			if err != nil {
				return err
			}
			//валидируем всю деку, которую прислал клиент, используя данные из БД как источник данных для ограничений
			if err := game.ValidateDeckList(req.Entries, battleMax, buffMax, ownedBattleCopies, ownedBuffCopies); err != nil {
				return err
			}
			//сохраняем колоду
			return repository.SaveDeckTx(tx, au.UserID, req.Entries)
		})
		//классическая обработка ошибок
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
	/*Что мы имеем? Мы точно так же в начале пошли декодить клиентский запрос. Точно так же проверили все
	на ошибки и все такое. Но потом, в дело вступила транзакция, где я как крутой пацан обязан был понять,
	владеет ли реально чувак картами или нет. То есть схема такая ЗАПРОС -> ДЕКОДИРОВАНИЕ ЗАПРОСА -> ПРОВЕРКА
	ДАННЫХ ИЗ БД -> ЕСЛИ ВСЕ ПЕРФЕКТ, СОХРАНЯЕМ ДЕКУ.*/
}

/*
Хелпер для оперативного перехвата серверных ошибок, которые могут возникнуть
Работает схожим образом с картой ошибок, но чисто на деки. Так то можно было
отдать и карту, но в рамках текущей задачи такое больше подходит
*/
func isDeckValidationErr(err error) bool {
	switch err { //<-смотрим на ошибку и если что отдаем трушку
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
