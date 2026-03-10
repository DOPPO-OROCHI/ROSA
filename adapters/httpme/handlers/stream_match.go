package handlers

import (
	"TheWar/adapters/httpme/middleware"
	"TheWar/internal/transport"
	"fmt"
	"io"
	"net/http"
	"time"
)

/*Файл посвящен поднятию SSE стрим матча (GET /matches/{id}/stream) и держит открытый канал, в то время как сервак
пушит обноления состояний. Это чисто транспортный слой, логики матча тут нет и быть не может. Тут в целом такой же
набор заивисомстей (я имею в виду структуры) как и в случае с обычными хендлерами, за исключением того, что данный
файл строго к привязан к непознанной теме SSE, конспект на тему которого лежит в transport/hub. Перейдем к коду*/

/*
Как уже описывалось много где еще, это стандартная структура зависимостей для хендлера, где мы и указываем зависимости...
Здесь в полях наш SSE хаб, а так же валидатор авторизации, на случай, если в контексте нет корректного юзера. Грубо говоря
хаб здесь нужен только для того, чтобы подписать, отписать, запушить состояние на текущий поток. Круто
*/
type StreamMatchDeps struct {
	Hub   *transport.Hub
	Store *middleware.TokenStore
}

// Фукнция отвечает за создание соединения SSE. Здесь принимаем нужные зависимости, отдавая при этом хендлер
func NewStreamMatchHandler(deps StreamMatchDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //<-возвращаем хендлер
		matchID, tail, err := middleware.ParceMatchPath(r.URL.Path) //<-парсим патч запроса
		if err != nil || tail != "stream" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		au, ok := middleware.FromContext(r.Context()) //<-авторизуем пользователя через контекст
		if !ok || au.UserID == 0 {                    //<-если не получилось, то берем токен из Query и валадируем с помощью Store
			tok := r.URL.Query().Get("token")
			if tok == "" || deps.Store == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized) //<-если не получилось-пользователь не авторизован
				return
			}
			sess, valid := deps.Store.Validate(tok) //<-а если что то получилось, то надо валидирировать пользователя
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			au = middleware.AuthUser{UserID: sess.UserID, TGID: sess.TGID} //<-записываем пользователя в память
		}
		flusher, ok := w.(http.Flusher) //<-опишшу ниже
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		/*А это заголовки, каждый из которых не несет в себе конкретной функциональности для кода, но несет весомые инструкции
		для браузера. Делается это для того, чтобы клиент и промежуточные узлы обработали соединение как SSE. Здесь мы буквально
		говорим: -"Content-Type":"text/event-stream", типа не обычный JSON
		"Cache-Control":"no-cache", типа нельзя кэшировать стрим ответ
		"Connection":"keep-alive", типа просим держать соединение открытым для потока
		По сути все это инструкции для браузера, чтобы он сделал так, как мы хотим. Круууууто*/
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		key := transport.StreamKey{MatchID: matchID, ViewerUserID: au.UserID} //<-собираем ключ для подписки на SSE
		ch := deps.Hub.Subscribe(key)                                         //<-и подписываем игроков к этому ключу
		defer deps.Hub.Unsubscribe(key, ch)                                   //<-деферя отписку, закрывая соединение
		keepAlive := time.NewTicker(15 * time.Second)                         //<-ниже опишу
		defer keepAlive.Stop()
		ctx := r.Context()
		for { //<-а здесь в цикле читаем соеднинение с пользователем
			select {
			case <-ctx.Done(): //<-если клиент отключился-выходим
				return
			case msg, ok := <-ch: //<-пришел апдейт из hub.Publish
				if !ok {
					return
				}
				writeSSE(w, "state", msg) //<-соответственно мы пишем его
				flusher.Flush()
			case <-keepAlive.C: //<-пишем ping-комментарий
				io.WriteString(w, ": ping\n\n")
				flusher.Flush()
			}
		}
	}
}

// хелпер форматирования событий. Здесь принимаем средство записи (w), имя событий (state,error...) и всю инфу в байтах(JSON)
func writeSSE(w io.Writer, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\n", event) //<-че, просто пишем всю хуйню клиенту
	fmt.Fprintf(w, "data: %s\n\n", data)
}

/*Таким образом получается следующая схема... В пакете transport в файле hub.go мы видим описание структур, а так же методов
для операционки над SSE соединением. В том числе подписка на это соединение, отписка, публикация ивента от сервера к клиенту.
Здесь же я собрал рабочую лошкадку (хендлер) под подпись на SSE. Используется это тогда, когда игрок приходит в матч. Делается
это не клиентом напрямую, а в момент, когда он пришел в матч фронтом. Собственно так и реализован SSE прикол внутри моей игры.
По существу скажу, что сама по себе тема SSE будет раскрыта мной в отдельном файле, где я уже опишу каким образом работает эта
штука и как.*/
