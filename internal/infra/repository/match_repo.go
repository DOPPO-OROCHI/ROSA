package repository

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/*В данном файле представлено все необходимое для создания матча внутри БД. В том числе сама структура матча и основная функция,
которая именно что создает матч. И суть вот в чем. Сам по себе матч живет только внутри базы данных, откуда мы и можем считывать
текущее состояние в любой момент времени. Из чего следует то, что по сути, данный файл является непосредственным входом в матч для
игрока. А теперь к сути*/

// каждый игрок может участвовать только в одном матче в момент времени
var ErrActiveMatchExists = errors.New("active match already exists")

/*
структура матча, где учитывается вообще все то, что влияет на матч. В будущем, когда будут добавляться новые приколы, которые
так или иначе будут участвовать в самом понятии матч как таковом, будут обрабатываться здесь (учитывая все детали, которые необходимы
в других местах)
*/
type Match struct {
	gorm.Model                    //<-уникальный айди матча
	PlayerID1      uint           `gorm:"not null;index"`                                   //<-айдишник первого пользователя
	PlayerID2      uint           `gorm:"not null;index"`                                   //<-второго
	State          datatypes.JSON `gorm:"type:jsonb;not null"`                              //<-состояние матча, которое складывается из JSON
	Version        int64          `gorm:"not null;default:1"`                               //<-версия матча. Добавлен для Optimistic Lock
	Finished       bool           `gorm:"not null;default:false"`                           //<-считается ли матч завершенным
	TurnDeadlineAt int64          `gorm:"column:turn_deadline_at;not null;default:0;index"` //<-время дедлайна текущего хода
}

/*
Функция хелпер, которая служит для того, чтобы внести апдейт в существующий матч, со всеми изменениями.
Принимаем БД, айди матча (состояние которого хотим изменить), актуальную версию (которую тоже будем менять),
новое состояние (вообще все то, что происходит на столе), новую версию (на которую будем менять старую),
информацию о завершении матча и то, сколько времени осталось на текущем ходе
*/
func SaveMatchState(tx *gorm.DB,
	matchID uint,
	expectedDBVersion int64,
	newStateJSON []byte,
	newVersion int64,
	finished bool,
	turnDeadlineAt int64) error {
	//читаем все из соответствующего матча с актуальной версией
	res := tx.Model(&Match{}).Where("id = ? AND version = ?", matchID, expectedDBVersion).
		//после чего апдейтим состояние матча на актуальное (как уже говорилось, функция эндер действия)
		Updates(map[string]any{
			"state":            datatypes.JSON(newStateJSON),
			"version":          newVersion,
			"finished":         finished,
			"turn_deadline_at": turnDeadlineAt,
		})
		//и занимаемся вполне рутинной обработкой ошибок
	if res.Error != nil {
		return res.Error
	}
	//за исключением этого. Это чисто версионная ошибка, которая говорит о том, что присланное
	//действие (версия) устарела, а значит нельзя обновлять состояние матча
	if res.RowsAffected == 0 {
		return game.ErrStaleAction
	}
	return nil
}

/*
И теперь самое вкусное и думаю, самое большое что можно встретить в моей игре с точки зрения бэк кода.
Это непосредственное создание матча в БД. Здесь собирается все, оба игрока, вместе с их деками и их персы.
Отдается же здесь доменная структура состояния матча, которая нужна вообще для всей логики в игре. Поскольку
матч завязан на БД, то и вся инфа будет браться отсюда
*/
func CreateMatchTX(db *gorm.DB,
	p1UserID, p2UserID uint,
	p1HeroCode, p2HeroCode string) (*game.MatchState, error) {
	var out *game.MatchState                        //<-формируем аут, который будем отдавать
	err := db.Transaction(func(tx *gorm.DB) error { //<-обязательно все делаем в транзакции
		if p1UserID == 0 || p2UserID == 0 { //<-проверяем на наличие пользователей (чтобы челы не смогли абузить)
			return errors.New("bad user id")
		}
		if p1UserID == p2UserID { //<-проверяем на абуз, чтобы игрок не смог сделать матч с собой
			return errors.New("cannot create match with yourself")
		}
		if p1HeroCode == "" {
			return errors.New("player 1 has no hero")
		}
		if p2HeroCode == "" {
			return errors.New("player 2 has no hero")
		}
		/*Фундаментальная вещь. Здесь мы проверяем обоих игроков на то, чтобы у них не было активных матчей.
		Честно сказать я уже тестил эту игру без такого инваиранта, получилось так себе. Поэтому да, вот так*/
		var activeCount int64
		//ищем активные матчи в списке матчей обоих игроков
		if err := tx.Model(&Match{}).
			Where("finished = false").
			//объяснить почему так
			Where("(player_id1 = ? OR player_id2 = ?) OR (player_id1 = ? OR player_id2 = ?)", p1UserID, p1UserID, p2UserID, p2UserID).
			Count(&activeCount).Error; err != nil {
			return err
		}
		//если у игрока уже есть активные матчи -отдаем ошибку чтобы шел доигрывал существующий
		if activeCount > 0 {
			return ErrActiveMatchExists
		}
		//проверяем все лимиты копий по шаблонам, чтобы передать все добро валидации. Берем инфу из БД* Откуда же еще**
		battleMax, buffMax, err := LoadTemplateLimits(tx)
		if err != nil {
			return err
		}
		//ниже читаем сохраненные деки для обоих игроков с проверкой, не пустые ли они
		p1Entires, err := LoadDeckTx(tx, p1UserID)
		if err != nil {
			return err
		}
		p2Entires, err := LoadDeckTx(tx, p2UserID)
		if err != nil {
			return err
		}
		if len(p1Entires) == 0 {
			return errors.New("p1 deck is empty")
		}
		if len(p2Entires) == 0 {
			return errors.New("p2 deck is empty")
		}
		//ниже грузим владение картами и их количество, чтобы проверить в натуре ли игрок владеет тем, что кладет в деку
		p1BattleInfo, p1BattleCopies, err := LoadOwnedBattleCards(tx, p1UserID)
		if err != nil {
			return err
		}
		p1BuffInfo, p1BuffCopies, err := LoadOwnedBuff(tx, p1UserID)
		if err != nil {
			return err
		}
		p2BattleInfo, p2BattleCopies, err := LoadOwnedBattleCards(tx, p2UserID)
		if err != nil {
			return err
		}
		p2BuffInfo, p2BuffCopies, err := LoadOwnedBuff(tx, p2UserID)
		if err != nil {
			return err
		}
		/*Может справделиво показаться, что слишком много кода вокруг карт. Грузим несколько раз одно и то же и все такое.
		Замечание справедливо, но разница в том, что все эти функции отвечают за разные слои логики работы с деками игроков.
		Разжую : LoadTemplateLimits отвечает за глобальные правила шаблонов. Тобишь она занимается исключительно тем, чтобы
		проверять количество копий в каждой отдельной деке. LoadDeckTx заявляет намерения игрока, иными словами то, что игрок
		хочет положить в деку (в том числе типы карт, чертежи, количество). LoadOwnedBattle/BuffCards отвечает на вопрос -
		реально ли игрок владеет тем, что кладет в деку. В связи с чем рисуется следующая связка работы :
		-игрок присылает деку на добавление с помощью SaveDeckRequest
		-Репозиторий читает все это и дека проходит проверку через LoadTemplateLimits (где считываются лимиты)
		-LoadDeckTX (что игрок заявляет)
		-LoadOwned(Battle/Buff)Cards где проверяются владения
		Но в этой схеме не хватает главного...

		Не хватает валидации присланной деки. Этот кусок отвечает как раз за это
		Он сверяет размеры деки, лимиты копий по шаблонам, вообще фактическое существование
		шаблонов и владение картами. Если все заебись -дека проходит вызывая SaveDeckTx, которая
		реплейсит деки внутри БД*/
		if err := game.ValidateDeckList(p1Entires,
			battleMax, buffMax, p1BattleCopies,
			p1BuffCopies); err != nil {
			return err
		}
		if err := game.ValidateDeckList(p2Entires,
			battleMax, buffMax, p2BattleCopies,
			p2BuffCopies); err != nil {
			return err
		}
		/*Местачковая функция, которая позволяет не только работать с кодом героя из БД, но и превратить
		это в полноценный рантайм. Че здесь происходит ? Мы ищем игрока вместе с героем, с которым он пришел
		в матч. Отдаем при этом геройский темплейт, целое число, которое говорит об уровне персонажа и, если
		что, ошибку*/
		loadHero := func(userID uint, heroCode string) (heroes.CharacterTemplate, int, error) {
			var tpl heroes.CharacterTemplate                                                   //<-сюда будем грузить найденный шаблон
			if err := tx.Where("character_code = ?", heroCode).First(&tpl).Error; err != nil { //<-ищем героя ИЗ ШАБЛОНА
				return heroes.CharacterTemplate{}, 0, err
			}
			var g GamerCharacter //<-а сюда будем класть именно владение
			//ищем персонажа по айди игрока и чертежу самого персонажа
			if err := tx.Where("gamer_id = ? AND character_template_id = ?", userID, tpl.ID).First(&g).Error; err != nil {
				return heroes.CharacterTemplate{}, 0, err
			}
			//возвращаем мурняк
			return tpl, g.CharacterLevel, nil
		}
		//а ниже прогружаем героев для обоих игроков
		p1HeroTpl, p1HeroLevel, err := loadHero(p1UserID, p1HeroCode)
		if err != nil {
			return err
		}
		p2HeroTpl, p2HeroLevel, err := loadHero(p2UserID, p2HeroCode)
		if err != nil {
			return err
		}
		/*Но на этом песня с деками не закончилась. Теперь нам нужно после всех проверок, сверок блять и так далее
		загрузить эту деку в рантайм. Делается это все здесь, в сигнатуре функции которой мы проверяем кол-во карт,
		проверяем их типы а так же перемешиваем колоду (массив с картами), так же присваивая каждой карте уникальный айди*/
		p1Deck, err := game.BuildDeck(p1Entires, p1BattleInfo, p1BuffInfo)
		if err != nil {
			return err
		}
		p2Deck, err := game.BuildDeck(p2Entires, p2BattleInfo, p2BuffInfo)
		if err != nil {
			return err
		}
		//а здесь создаем сам матч, чтобы записать его в БД. Откуда получаем айди матча
		row := Match{
			PlayerID1: p1UserID,
			PlayerID2: p2UserID,
			Version:   1,
			State:     datatypes.JSON([]byte(`{}`)),
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		//И наконец, собираем стартовое состояние обоих игроков для доменной логики. Здесь описано все
		p1 := game.PlayerState{
			PlayerID:                0,                          //<-место игрока в матче (матч, это массив из двух игроков в этом смысле)
			UserID:                  p1UserID,                   //<-фактический айди пользователя
			HeroID:                  p1HeroTpl.ID,               //<-геройский айди
			HeroCode:                p1HeroTpl.CharacterCode,    //<-код героя
			HeroLevel:               p1HeroLevel,                //<-уровень героя
			HeroHP:                  p1HeroTpl.HealthPoints,     //<-его хп
			HeroAttackPower:         p1HeroTpl.AttackPower,      //<-его силу атаки
			HeroAttackCooldown:      0,                          //<-существующее кд атаки
			HeroAttackBaseCooldown:  p1HeroTpl.AttackCooldown,   //<-глобальное кд атаки
			HeroSplashRadius:        p1HeroTpl.SplashRadius,     //<-есть ли у героя сплеш
			HeroAbilityCooldown:     0,                          //<-кд его способности
			HeroAbilityBaseCooldown: p1HeroTpl.Ability.CoolDown, //<-глобальное кд способности
			HeroAbilityManaCost:     p1HeroTpl.Ability.ManaCost, //<-стоимость способности в мане
			Deck:                    p1Deck,                     //<-дека
		}
		p2 := game.PlayerState{
			PlayerID:                1,
			UserID:                  p2UserID,
			HeroID:                  p2HeroTpl.ID,
			HeroCode:                p2HeroTpl.CharacterCode,
			HeroLevel:               p2HeroLevel,
			HeroHP:                  p2HeroTpl.HealthPoints,
			HeroAttackPower:         p2HeroTpl.AttackPower,
			HeroAttackCooldown:      0,
			HeroAttackBaseCooldown:  p2HeroTpl.AttackCooldown,
			HeroSplashRadius:        p2HeroTpl.SplashRadius,
			HeroAbilityCooldown:     0,
			HeroAbilityBaseCooldown: p2HeroTpl.Ability.CoolDown,
			HeroAbilityManaCost:     p2HeroTpl.Ability.ManaCost,
			Deck:                    p2Deck,
		}
		//после чего мы собираем матч
		st := game.NewMatchState(row.ID, &p1, &p2)
		b, err := json.Marshal(st)
		if err != nil {
			return err
		}
		//после чего записываем фактическое состояние и версию в только что созданную строку матча
		if err := tx.Model(&Match{}).Where("id = ?", row.ID).Updates(map[string]any{
			"state":            datatypes.JSON(b),
			"version":          st.Version,
			"turn_deadline_at": st.TurnDeadline,
		}).Error; err != nil {
			return err
		}
		//сохраняем результат транзакции наружу
		out = st
		return nil
	})
	//если что то идет не по плану, то мы даем заднюю...
	if err != nil {
		return nil, err
	}
	//а если нет, то мы возвращаем вполне рабочий рантайм компонент, который участвует в функиональных аспектах доменной логики
	return out, nil
}

/*И так. С вами была можно сказать единственная оркестровая функция на тему -я хочу создать валидный матч.
Архитектурно кейс выглядит следующим образом : во-первых все внутри транзакции, ибо либо создаем фул состояние,
либо не создаем ничего. Во-вторых -соблюдаем все инваиранты на создание матча, а-ля проверка user-id, запрет
матча с самим собой, запрет второго активного матча. В-третьих, собираем все компоненты для составления деки,
в том числе :LoadTemplateLimits (где мы собираем глобальные лимиты шаблонов), LoadDeckTX(где вообще читается,
что игрок заявил в деке), LoadOwnedBattle/BuffCards(чем реально владеет игрок). В-четвертых, валидируем все то,
что прислал клиент (мало ли наебал). В-пятых, но это не касается конкретно этой функции но все же, мы должны
прогрузить резолвы. Прелесть в том, что делается это не здесь а в генеральнй доменной функции ApplyAction.
В-шестых, конструируем рантайм компонент матча с помощью BuildDeck, где превращаем декларацию в непосредственно
матчевые карты (где даем каждой карте instanseID, учитываем уровень, перемешиваем колоду). В-седьмых, записываем
все в строку матча с полным состоянием, для получения айдишника матча. В-восьмых, ебашим всю муру в состояние
уже пользователей (PlayerState). В-девятых, фиксируем стартовое состояние на уровне JSON, чтобы вернуть это
состояние наружу, что по итогу дает пользователю полное состояние матча, которое он может буквально посмотреть*/
