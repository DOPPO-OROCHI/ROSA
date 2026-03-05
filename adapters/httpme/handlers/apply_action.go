package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/applycation"
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"TheWar/internal/domain/player"
	"TheWar/internal/infra/repository"
	"TheWar/internal/transport"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

/*Данный файл содержит хендлеры для обработки действий игрока. В нем реализованый хендлеры,
которые служат для обработки запросов клиента на применение действия в рамках матча. */

/*
Важнейшая часть сего файла. Почему ? Потому что это по сути фабрика хендлера. Она принимает
зависимости (DB -> потому что все операции должны идти транзакционно, через БД (версионирование
например), reslovers -> это зависимости игрового движка от шаблонов карт-героев и так далее)
HUB -> это чтобы пушнуть новое состояние после его изменения. SSE тема. Здесь мы собственно
и принимаем этот SSE.
Сами хендлеры ничего не придумывают а принимают эту структуру в качестве аргумента, абсолютно
не заботясь о том, откуда пришли данные
*/
type ApplyActionHandlerDeps struct {
	DB        *gorm.DB
	Resolvers game.Resolvers
	Hub       *transport.Hub
}

/*
А здесь мы собственно и описываем хендлер, который в будущем будем передавать в наш MUX (app.go).
Что здесь происходит? Типичный паттерн, мы снаружи передаем зависимости в виде нашей структуры, отдавая
при этом готовый хендлер.
*/
func NewApplyActionHandler(a ApplyActionHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //<-это и есть сам хендлер, который будет обрабатывать запросы
		au, ok := middleware.FromContext(r.Context()) //<-получаем юзера из контекста. Так мы сразу фильтруем неавторизованных
		if !ok {
			//обращаю внимание на то, что хендлер не читает куки сам, за это отвечает отдельный middleware
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized") //<-если юзера нет, то отдаем 401
			return
		}
		userID := au.UserID                                         //<-получаем юзера в нашу память, для дальнейших операций
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path) //<-пасрим путь, чтобы понять что игрок хочет сделать.
		//в данном случае мы ожидаем только путь вида /matches/{id}/actions (типа применение к матчу), если он не такой, то отдаем 404
		if err != nil || tail != "actions" {
			middleware.WriteErr(w, http.StatusNotFound, "not found")
			return
		}
		//после проведенных операций мы готовы к тому, чтобы применить действие. Для этого нам сугубо необходимо декодировать
		//тело запроса, в котором будут описаны намерения игрока. И для этого, нам нужна наша DTO...
		var req dto.ApplyActionRequest //<-создаем структуру DTO, в которую будем декодировать тело запроса (JSON)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, "bad json") //<-если JSON кривой, то отдаем 400
			return
		}
		//здесь хендлер перестает заниматься HTTP деталями (проверки) и передает управление дальше, в application (БД) слой
		//а уходить будет как раз DB, потому что все операции изменения это транзакции. ID матча, юзера и резолверы
		//для игрового движка. Почему мы передаем резолверы ? Потому что движок заранее не значет с какими картами пришел
		//игрок, мы их именно что смотрим по тому, с чем он пришел. Нужно это для валидации дейсвтий и прочих бэк приколов
		newState, err := applycation.ApplyActionToMatchTx(a.DB, matchID, userID, req, a.Resolvers)
		if err != nil {
			switch {
			//маппим ошибки
			case errors.Is(err, applycation.ErrNotParticipant):
				middleware.WriteErr(w, http.StatusForbidden, "forbidden")
				return
			case errors.Is(err, applycation.ErrCorruptedMatchState):
				middleware.WriteErr(w, http.StatusInternalServerError, "something went wrong")
				return
			default:
				middleware.WriteErr(w, MapEngineErr(err), err.Error()) //<-пока что отдаем внутреннюю ошибку. Похуй, потом поменяю
				return
			}
		}
		//а здесь мы заняты как раз пушингом нового состояния клиенту. Если все проверки прошли успешно-> матч изменился. Соответственно
		//мы должны опубликовать эти изменения в SSE, всем подписчикам и фанатам...
		PublishMatchToSSE(a.Hub, newState)
		//завершаем наш познавательный хендлер отправкой ответа клиенту о том, что у нас все хорошо и новым состоянием
		//про новое состояние и как выебнуться и спрятать все, что не нужно видеть опоненту подробнее в maskMatchStateForUser
		middleware.WriteJSON(w, http.StatusOK, maskMatchStateForUser(newState, userID))
	}
}

/*
А это маппинг всех возможных ошибок, которые могут возникнуть в ходе проверок. Смысл в чем ? На уровне движка у нас
есть куча проблем, которые так или иначе будут заебывать игрока. Чтобы не тащить их в хендлеры и более менее держаться
правила -слои абстрагированы-, тем не менее нужно отдавать эти ошибки наружу. Как это сделать ? С помощью карт, которые
будут заниматься интерпритацией движковых ошибок, переводя их в понятные статус код. Собственно вот все что пока что* есть
*/
func MapEngineErr(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	//если челик борщанул с кликами
	case errors.Is(err, game.ErrStaleAction):
		return http.StatusConflict
		//право на действие
	case errors.Is(err, game.ErrNotYourTurn):
		return http.StatusForbidden
		//ошибка фазы
	case errors.Is(err, game.ErrMatchFinished):
		return http.StatusConflict
		//ресурсы
	case errors.Is(err, game.ErrNotEnoughMana):
		return http.StatusBadRequest
		//ошибки стола
	case errors.Is(err, game.ErrTablesFull):
		return http.StatusBadRequest
	case errors.Is(err, game.ErrSlotOccupied):
		return http.StatusBadRequest
		//правила боя
	case errors.Is(err, game.ErrAttackerOnCooldown),
		errors.Is(err, game.ErrAttackerSummoneddThisTurn),
		errors.Is(err, game.ErrMustAttackTank),
		errors.Is(err, game.ErrCannotAttackHeroWithTanks),
		errors.Is(err, game.ErrCannotHitHeroWhileTanks),
		errors.Is(err, game.ErrHeroOnCooldown),
		errors.Is(err, game.ErrHeroAttackIsZero),
		errors.Is(err, game.ErrHealerCannotAttack):
		return http.StatusBadRequest
	//дека
	case errors.Is(err, game.ErrDeckSizeNot20),
		errors.Is(err, game.ErrDeckCountInvalid),
		errors.Is(err, game.ErrDeckTooManyCopies),
		errors.Is(err, game.ErrDeckNotOwnedEnough),
		errors.Is(err, game.ErrDeckUnknownCard),
		errors.Is(err, game.ErrDeckUnknownKind):
		return http.StatusBadRequest
	//способности героя
	case errors.Is(err, game.ErrHeroAbilityOnCooldown),
		errors.Is(err, game.ErrHeroAbilityBadTarget),
		errors.Is(err, game.ErrHeroAbilityUnknown):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

/*ФИКСИРУЕМ. Я понимаю что отдавать действительные err.Error() это не самая хорошая идея, когда дело
касается реального прода. Извините. Но при текущем положении кода, я тупо заебусь на каждый косяк писать
-something went wrong. Да, я понимаю что я могу плюс минус из контекста понять, что происходит, НО! Я
не хочу сейчас этим заниматься. Я понимаю что при отдаче дейсвтительных ошибок я палю внутренню архитектуру
проекта (в частности БД), но пока имеем что имеем. Все равно игра бесплатная...*/

/*
Данная функция занимается исключетльно тем, что ставит в БД к конкретному игроку (который отправил запрос)
персонажа, с которым тот пойдем в бой. В целом ничего специфичного. Отдлаем хендлер под мукс. Теперь к реализации.
*/
func NewSelectedHeroHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		au, ok := middleware.FromContext(r.Context()) //<-проверяем на авторизацию
		if !ok {
			middleware.WriteErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		userID := au.UserID           //<-записываем юзера в память
		var req dto.SelectHeroRequest //<-здесь мы заняты декодированием намерений клиента в понятную движку DTO структуру
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteErr(w, http.StatusBadRequest, err.Error())
			return
		}
		var tpl heroes.CharacterTemplate //<-а здесь мы занимаемся поиском героя в наших стандартных шаблонах (проще говоря :
		//есть ли вообще такой персонаж. Ато пришлют хер знает кого). Это кстати кривой хендлер да... Поскольку здесь Use-Case
		//(внатуре логика, типа что то сделай, а не обратись к...), но настоящим пацанам похуй!!!(почти)
		if err := db.Where("character_code = ?", req.HeroCode).First(&tpl).Error; err != nil { //<-ну и собственно ищем
			middleware.WriteErr(w, http.StatusNotFound, err.Error())
			return
		}
		var owned repository.GamerCharacter //<-НО! После того как мы убедились что игрок не пиздит, нам нужно отдать ему
		//именно его персонажа, поскольку у него есть уровень прокачки и все такое. Так и вот, здесь мы проверяем есть ли
		//такой перс у игрока, и, что важнее, достаем именно его характеристики
		if err := db.Where("gamer_id = ? AND character_template_id = ?", userID, tpl.ID).First(&owned).Error; err != nil {
			middleware.WriteErr(w, http.StatusForbidden, err.Error())
			return
		}
		//а здесь, после всех проверок просто ставим его перса в активных персонажей, с которыми он собственно пойдет в матч. Круто
		if err := db.Model(&player.TelegramUser{}).Where("id = ?", userID).Update("selected_hero_template_id", tpl.ID).Error; err != nil {
			middleware.WriteErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		//пишем что все заебись
		middleware.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "hero selected",
		})
	}
}

/*Таким образом я реализую генеральную логику хендлера (пускай в случае с NewSelectedHeroHandler кривоватую) следующим образом:
Мы принимаем некую сущность, которая собственно отвечает за зависимости, к примеру ApplyActionHandlerDeps, где лежит все необходимое
для того, чтобы действие собственно совершить. Круто. Наружу же я отдаю хендлер, который и занят валидацией всего того, что может
пойти не так. Внутри же самого хендлера я (в идеальном раскладе) вызываю различные функции, которые отвечают тому, что хочет клиент.
При необходимости я паршу его PATCH, где конкретно проверяю че он хочет. Сериализую тело его запроса, передавая его намерения в
функции. Иными словами хендлер это -Принимаем аргументы (все то, что необходимо для операции)->десериализируем тело запроса->
валидируем значения->вызываем нужные функции(в идаеле, а не как я в уже понятно где)->проверяем ошибки->и далее мутим все то, что
нужно для исполнения задачи (в моем случае либо пушим обновления по SSE, либо просто отдаем ответ о том, что все кул. Так то можно
было развернуть телегу дальше и хз, сериализовать че нибудь клиенту, но пока этого не надо. Чилим)*/
