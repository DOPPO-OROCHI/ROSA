package adapters

import (
	"TheWar/adapters/httpme/middleware"
	"net/http"
)

/*В этом файле я собираю вообще всю суть приложения (что касается HTTP приколов). Здесь используется точно такая
же зависимость в виде структуры (потому что я в рот ебал все это запускать в аргументы), где описывается все то,
что игрок в принципе может делать в рамках моей игры. Все остальные действия типа -изменить аватарку, задонатить...
будут держаться здесь и обрабатываться в NewMux. К слову об этом...*/

/*
Данная структура отвечает за распределение раннее написанных хендлеров для того, чтобы передать их в муксы. Заполняется
она соответствующим образом в мэйне. И дело тут вот в чем... Каждый хендлер отвечает за определенные вещи внутри кода.
Здесь представлены все из них по именам полей, которые имеют тип хендлера. Далее мукс, в зависимости от того, какой эндпоинт
в него зашел будет вызывать конкретные поля структуры App, которые составлены соответствющим задаче образом в мэйне.
*/
type App struct {
	CreateMatch http.HandlerFunc
	GetMatch    http.HandlerFunc
	ApplyAction http.HandlerFunc

	GetMe        http.HandlerFunc
	UpdateAvatar http.HandlerFunc

	SaveDeck http.HandlerFunc
	GetDeck  http.HandlerFunc

	CardsList   http.HandlerFunc
	HeroesList  http.HandlerFunc
	MatchesList http.HandlerFunc

	SelectHero   http.HandlerFunc
	StreamMatch  http.HandlerFunc
	AuthTelegram http.HandlerFunc
	AuthDev      http.HandlerFunc

	JoinQueue    http.HandlerFunc
	LeaveQueue   http.HandlerFunc
	QueueStatus  http.HandlerFunc
	AcceptQueue  http.HandlerFunc
	DeclineQueue http.HandlerFunc
}

/*
Создаем мукс. По сути своей мукс-это контролер входящих патчей (данная формулировка может показаться странной в контексте
описания мукса, но помним что это в том числе мой учебный проект) (сборщик HTTP API). Здесь собственно и создается мукс, в
котором мы создаем его подмножества под каждый конкретный сценарий эндпоинта.
*/
func NewMux(app App) *http.ServeMux {
	mux := http.NewServeMux()
	//придуманный не мной мукс, который служит для проверки того, жив ли вообще сервер
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	/*мукс для операционки над матчами. Здесь производится проверка методов, в зависимости от которых вызываются
	разные методы App (пока что -> POST (создаем матч), GET (получаем список мачтей)). В свою очередь вызываются
	ране написанные хендлеры*/
	mux.HandleFunc("/matches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			//учебное пособие* -эти аргументы нужны для того, чтобы передать их в хендлеры
			app.CreateMatch(w, r)
		case http.MethodGet:
			app.MatchesList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	/*мукс для операционки внутри матча. Здесь, в зависимости от того, какой PATH пришел производится необходимая
	проверка для того, чтобы дернуть определенный хендлер, будь то получение текущего состояния матча (после реконекта
	например), или применение действия. Так же тут есть и SSE моменты, которые я (обещаю) выучу от и до. Но в целом,
	здесь как раз таки и происходит валидация PATH, исходя из данных которых будут вызываться конкретные функции.*/
	mux.HandleFunc("/matches/", func(w http.ResponseWriter, r *http.Request) {
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = matchID
		switch {
		case r.Method == http.MethodGet && tail == "":
			app.GetMatch(w, r)
			return
		case r.Method == http.MethodPost && tail == "actions":
			app.ApplyAction(w, r)
			return
		case r.Method == http.MethodGet && tail == "stream":
			app.StreamMatch(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
	/*Мукс для вызоваа хендлера обо всей инфе о пользователе. Ничего слонжого*/
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetMe(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	//очереди
	mux.HandleFunc("/queue/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		app.JoinQueue(w, r)
	})
	mux.HandleFunc("/queue/leave", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		app.LeaveQueue(w, r)
	})
	mux.HandleFunc("/queue/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		app.QueueStatus(w, r)
	})
	mux.HandleFunc("/queue/accept", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		app.AcceptQueue(w, r)
	})
	mux.HandleFunc("/queue/decline", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		app.DeclineQueue(w, r)
	})
	/*Поинтереснее, но тоже изи. В зависимости от присланного метода определеяем то, что клиент хочет
	сделать с декой. Либо посмотреть на свою деку, либо обновить существующую (в этом же блоке -создать деку)*/
	mux.HandleFunc("/deck", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetDeck(w, r)
		case http.MethodPost:
			app.SaveDeck(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	/*Тоже все легко. Тут по факту играет лишь один метод. Никаких дейсвтий в рамках этого кейса над картами
	не предусмотрено (движок не предусматривает операционку над картами вообще, ни улучшения (они статичны и для
	всех характеристик), ни какие то кастомные штуки типа поменять картинку), поэтому здесь чисто посмотреть
	список карт*/
	mux.HandleFunc("/cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.CardsList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	//аналогичная с картами тема. В будущем добавлю возможность операционки над обеими сущностями. Но пока не до этого
	mux.HandleFunc("/heroes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.HeroesList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	//мукс чисто для выбора персонажа. Ничего интересного
	mux.HandleFunc("/heroes/select", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.SelectHero(w, r)
	})
	//аутентификационный мукс
	mux.HandleFunc("/auth/telegram", func(w http.ResponseWriter, r *http.Request) {
		app.AuthTelegram(w, r)
	})
	mux.HandleFunc("/auth/dev", func(w http.ResponseWriter, r *http.Request) {
		app.AuthDev(w, r)
	})
	return mux
}

/*Таким образом я собрал мой мукс да. Схема такая -> структура APP собирается в мэйне, где я присваиваю каждому
полю свой хендлер->Все это добро передается в мукс, для того, чтобы в зависимости от патча и\или метода вызывать
строго определенные поля->мукс тоже инициализируется в мэйне
Аргументы нужны для того, чтобы передать их в хендлеры, чтобы те в свою очередь имели средства записи и чтения под
каждый конкретный запрос.*/
