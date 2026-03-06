package game

import (
	"errors"
	"math/rand/v2"
	"time"
)

// В этом файле описаны все утилиты, необходимые для анализа матча на предмет ничьи.

// проверяем можно ли объявить ничью, принимая как аргумент состояние матча
func CheckDraw(m *MatchState) bool {
	//проверяем на завершение матча
	if m.Finished {
		return false
	}
	//вводим переменные наших игроков, чтобы
	p1 := m.Players[0]
	p2 := m.Players[1]
	if p1 == nil || p2 == nil {
		return false
	}
	//оценить состояние игроков, чтобы понять, можно ли ебануть ничью или еще нет
	//принимаем состояния игроков и если отдаем тру фолс
	empty := func(p *PlayerState) bool {
		return len(p.Hand) == 0 && len(p.Deck) == 0 && TableEmpty(p)

	}
	/*А идея ничьи вот в чем. Если у игрока нет карт ни в руке (len(p.Hand)), ни в деке (len(p.Deck)), а
	еще и стол пустой TableEmpty, да еще и такое состояние встречается у обоих игроков -ничья. Иначе -играем*/
	return empty(p1) && empty(p2)
}

/*
Функция постановки ничьи в матче. Если CheckDraw занимался валидацией условий перед ничьей, то это
уже все, поставить ничью и пиздец. Здесь мы принимаем состояние матча и если CheckDraw истина-ничья
*/
func ApplyDrawNeed(m *MatchState) bool {
	if CheckDraw(m) {
		//переводим статус матча в завершенный
		m.Finished = true
		//записываем ничью
		m.Result = MatchDraw
		return true
	}
	return false
}

/*Функция запуска матча. Принимает состояние матча, для того, чтобы влиять на него*/
func EnsureStartTurn(m *MatchState) error {
	//если матч завершен -возвращаем ошибку, которая об этом говорит
	if m.Finished {
		return ErrMatchFinished
	}
	//если фаза в матче старт (системная подготовка к ходу)
	if m.Phase == PhaseStart {
		//запускаем функцию, которая помимо прочего ограничивает время на ход. Подробнее о ней в следующих файлах
		StartTurn(m, time.Now().Unix())
	}
	return nil
}

/*Данная функция по большей части служит для принудительного старта хода игроком, который на ход претендует.
Делает он это с помощью принятия состояния матча, откуда и берет информацию о том, чей сейчас ход. Так же проверяет
обязательные условия перед стартом хода, по результату которой и инициализирует ход*/

/*
Функция, которая является главным адапретор действия внутри матча. Здесь обрабатываются пользовательские решения,
которые он вообще может отправить в рамках матча. Во входящих аргументах принимается состояние матча, структура
действия (где описано, что такое действие в принципе) и резолверы, о роли которых много позже, поскольку вещь фундаментальная.
В случае чего отдаем ошибку. Приступим
*/
func ApplyAction(m *MatchState, a Action, r Resolvers) error {
	if m.Finished { //<-проверяем на завершение матча
		return ErrMatchFinished
	}
	if a.ExpectedVersion != 0 && a.ExpectedVersion != m.Version { //<-защита от действий, которые уже были записаны
		return ErrStaleAction
	}
	m.Events = m.Events[:0] //<-очищаем массив с ивентами (UI тема), для того чтобы не отдавать клиенту кашу из спецэффектов
	//а здесь начинается уже менеджмент принимаемых игроком решений
	if a.Type == ActionLeaveMatch { //<-если чел захотел ливнуть из матча
		if err := LeaveMatch(m, a.PlayerIndex); err != nil { //<-то мы вызываем соответствующую функцию
			return err
		}
		m.Version++ //<-обязательно версионируем, поскольку лив из матча, это по сути тоже версия
		return nil
	}
	if err := EnsureStartTurn(m); err != nil { //<-стартуем ход
		return err
	}
	now := time.Now().Unix()                                                    //<-объявляем переменную времени для отсчета дедлайна на ход
	if m.Phase == PhaseMain && m.TurnDeadLineAt > 0 && now > m.TurnDeadLineAt { //<-и если время вышло, возвращаем ошибку
		return ErrTurnTimeOut
	}
	if a.PlayerIndex != m.ActivePlayer { //<-если игрок, который в момент времени НЕ является активным игроком (не может ходить)
		return ErrNotYourTurn //<-возвращаем ошибку
	}
	var err error   //<-объявляем переменную ошибки, чтобы я смог обработать ее по результатам работы той, или иной функции
	switch a.Type { //<-свичим тип действия
	case ActionPlayBattle: //<-если действие атаки картой
		if r.Battle == nil { //<-проверяем подключенный резолвер, без него домен не сможет получить шаблон карты
			return errors.New("battle resolver is nil")
		}
		err = PlayBattleCard(m, a.PlayerIndex, a.CardInstanceID, a.TargetSlot, r.Battle) //<-вызываем функцию атаки
		//и далее по списку возможных действий
	case ActionPlayBuff:
		if r.Buff == nil {
			return errors.New("buff resolver is nil")
		}
		err = PlayBuffCard(m, a.PlayerIndex, a.CardInstanceID, a.TargetInstanceID, r.Buff)
	case ActionCardAttack:
		if r.Battle == nil {
			return errors.New("battle resolver is nil")
		}
		err = CardAttack(m, a.PlayerIndex, a.CardInstanceID, a.TargetInstanceID, a.AttackHero, r.Battle)
	case ActionHeroAttack:
		err = HeroAttack(m, a.PlayerIndex, a.TargetInstanceID, a.AttackHero)
	case ActionPlayHeroSpell:
		err = PlayHeroSpell(m, a, r)
	case ActionEndTurn:
		EndTurn(m)
	default:
		return errors.New("unknown action type: " + string(a.Type))
	}
	if err != nil {
		return err
	}
	m.Version++
	ApplyDrawNeed(m)
	return nil
}

/*Таким образом прменяются действия в рамках конкретного матча, проходя строго определенный путь проверок и вызовов.
Любая операция, которая будет добавляться в будущем обязана пройти следующий путь :
-Доменная часть : Добавляем новый ActionType (game.Const),
-(только если нужно) Добавляем новый ивент тип (в том числе Source, Kind)
-Добавляем новую доменную функцию, по аналогии с функциями PlayBattle/BuffCard и так далее, внутри которой валидировать
значения, добавлять ивенты и заниматься всей херней, которая нужна для изменения состояния матча
-Добавляем новый case в ApplyAction
-Если для действия нужен Resolver, проверяем и на него тоже
-Вызываем соответствующую функцию нового действия в новом кейсе
-Добавляем новую DTO, сохраняя при этом оригинальную структуру типа -instanceID- и прочего, для взаимодействия
с другими полями других DTO
-Добавляем все это добро в ApplyActionToMatchTX, который должен уметь собирать новый Action, при этом, все должно
происходить транзакционно (например анмаршаллинг дейсвтий, аппли экшн и тд)
-добавляем новый эндпоинт на новое действие, где обрабатываем все ошибки
-после успешного применения действия все публикуется в SSE, где клиент получает обновленный maskedState.
-ну и наконец проверяем инваиранты, где действие должно обязательно проходить по Version, turn и всего того, что напрямую
влияет на защиту матча от случайностей

По сути, новое действие считается добавленным тогда, когда в коде реализована цепочка :
DTO (где мы составляем транспортировочный файл)->application tx(где действие регистрируется в БД)->domain apply (где происходит
доменная логика)->persistence save ()-> masked HTTP response -> SSE publish.*/

func NewMatchState(matchID uint, p1 *PlayerState, p2 *PlayerState) *MatchState {
	firstPlayer := rand.IntN(2)
	m := &MatchState{
		MatchID:      matchID,
		Version:      1,
		Players:      [2]*PlayerState{p1, p2},
		ActivePlayer: firstPlayer,
		Phase:        PhaseStart,
		Finished:     false,
		Result:       MatchOnGoing,
	}
	for _, p := range m.Players {
		if p == nil {
			continue
		}
		p.Turns = 0
		p.Mana = 1
		p.Discard = nil
	}
	DrawCards(p1, 2)
	DrawCards(p2, 2)
	return m
}

func DrawCards(p *PlayerState, n int) {
	if p == nil || n <= 0 {
		return
	}
	if len(p.Deck) < n {
		n = len(p.Deck)
	}
	for i := 0; i < n; i++ {
		c := p.Deck[0]
		p.Deck = p.Deck[1:]
		p.Hand = append(p.Hand, c)
	}
}

func TableEmpty(p *PlayerState) bool {
	for i := 0; i < TableSize; i++ {
		if p.Table[i] != nil {
			return false
		}
	}
	return true
}
