package cache

import (
	"TheWar/internal/domain/cards"
	"TheWar/internal/domain/game"
	"sync"

	"gorm.io/gorm"
)

/*И так. Здесь мы по полной будем теперть на тему резолверов. В общем. Сам по себе концепт резолвера отражает связь между доменной
логикой (в нашем случае матча) и источником данных (в нашем случае БД). Притом источник данных может быть любым. По сути, домен
запрашивает шаблон карты, бафа, перса по ключу, а резолвер отвечает. Это если коротко... Куда более занятно то, вообще почему его
надо добавлять. А причины тут вполне конкретные. Это нужно для того, чтобы не размазывать доменную логику с БД телодвижениями. Почему
это важно ? Потому что так наш домен независим от БД, тем самым становясь сугубо логическим сегментом. Если бы не было резолверов,
пришлось бы ходить в БД каждый раз, когда вызывается та или иная сущность. Представь обилие переменных, запросов, а вместе с тем,
возможных ошибок. Таким образом резолверы выполняют роль поставщика данных из БД в память, без участия напрямую домена. Далее по
текущему функционалу...*/

/*
А это прям концептуальная хрень. Смотри в чем прикол. Здесь есть три поля. res и init отвечают за определенную часть
считывания темплейтов. А мьютекс тут нужен затем, что он предотвращает конфликт чтения между этими двумя сущностями. Для
того, чтобы лучше объяснить смысл, перейдем к функциям
*/
type ResolverCache struct {
	mu   sync.RWMutex   //<-это мьютекс, прикольно. Служит для того, чтобы не было гонок чтения в методе Reload()
	res  game.Resolvers //<-набор резолверов, куда мы прогружаем все данные из метода reload()
	init bool           //<-флаг, который сигнаризирует о том, успешно ли завершился reload() и res валиден
}

/*
Метод reload служит для прогрузки всех сущностей из БД, которые указаны ниже. То есть это пишущий метод,
в котором мы принимаем в качестве аргумента нашу базу данных, после чего заполняем все необходимое в кэше
*/
func (c *ResolverCache) Reload(db *gorm.DB) error {
	//вводим переменную, куда будем писать все данные из БД
	var battles []cards.BattleCardTemplate
	//а здесь идем в БД в посиках шаблонов баттл карт, чтобы в последствии записать все данные в нее
	if err := db.Find(&battles).Error; err != nil {
		return err
	}
	//создаем мапу из всех найденных записей
	battleMap := make(map[string]game.BattleTemplate, len(battles))
	//и в цикле заполняем ее
	for _, t := range battles {
		battleMap[t.CodeString] = game.BattleTemplate{ //<-записываем данные по ключу
			TemplateID:            t.CodeString,
			HealthPoints:          t.HealthPoints,
			Attack:                t.Attack,
			SplashRadius:          t.SplashRadius,
			Cooldown:              t.CoolDown,
			Manacost:              t.ManaCost,
			IsTank:                t.IsTank,
			CardType:              t.CardType,
			CanBeUpgraded:         t.BuffSlot,
			ImageKey:              t.ImageKey,
			AssetBaseKey:          t.AssetBaseKey,
			SkillImageKey:         t.SkillImageKey,
			SkillName:             t.SkillName,
			SkillCode:             t.SkillCode,
			SkillTrigger:          t.SkillTrigger,
			SkillTarget:           t.SkillTarget,
			SkillValue:            t.SkillValue,
			SkillDuration:         t.SkillDuration,
			SkillCooldown:         t.SkillCooldown,
			SkillParamsJSON:       t.SkillParamsJSON,
			PassiveImageKey:       t.PassiveImageKey,
			PassiveName:           t.PassiveName,
			PassiveCode:           t.PassiveCode,
			PassiveTrigger:        t.PassiveTrigger,
			PassiveTarget:         t.PassiveTarget,
			PassiveEffect:         t.PassiveEffect,
			PassiveCondition:      t.PassiveCondition,
			PassiveValue:          t.PassiveValue,
			PassiveDuration:       t.PassiveDuration,
			PassiveScale:          t.PassiveScale,
			PassiveCountOwner:     t.PassiveCountOwner,
			PassiveConditionCount: t.PassiveConditionCount,
			PassiveCountType:      t.PassiveCountType,
			PassiveCountCode:      t.PassiveCountCode,
		}
		//таким образом я достал все данные из БД и положил их локально
	}
	//то же самое только для бафа
	var buffs []cards.BuffCardsTemplate
	if err := db.Find(&buffs).Error; err != nil {
		return err
	}
	buffMap := make(map[string]game.BuffTemplate, len(buffs))
	for _, t := range buffs {
		buffMap[t.CodeString] = game.BuffTemplate{
			TemplateID:   t.CodeString,
			ManaCost:     t.ManaCost,
			BuffType:     t.BuffType,
			BuffValue:    t.BuffValue,
			OnlyFor:      t.OnlyFor,
			Duration:     t.Duration,
			ImageKey:     t.ImageKey,
			AssetBaseKey: t.AssetBaseKey,
		}
	}
	//а здесь мы заполняем структуру резолверов, наполняя ее нужными данными о всех сущностях, необходимых движку
	res := game.Resolvers{
		//заполняем героев
		HeroAbility: func(herocode string) (game.HeroAbility, bool) {
			switch herocode {
			case "suprime_lider":
				return game.SupremeLiderAbilitySpec{}, true
			case "karn":
				return game.KarnAbilitySpec{}, true
			case "the_system":
				return game.TheSystemAbilitySpec{}, true
			case "imperial_commander":
				return game.ImperialCommanderAbilitySpec{}, true
			case "black_cell":
				return game.BlackCellAbilitySpec{}, true
			case "slavic_priest":
				return game.SlavicPriestAbilitySpec{}, true
			default:
				return nil, false
			}
		},
		//и передаем карты по ключу.
		Battle: game.BattleMapResolver{M: battleMap},
		Buff:   game.BuffMapResolver{M: buffMap},
	}
	/*И тут подробнее. Мы ставим лок здесь, чтобы никто не смог читать или писать с/в c.res и c.init
	пока операция не завершилась. НО! Ниже у нас есть еще один мьютекс в GET, но там прикол в другом.
	Гет допускает параллельное чтение этих полей. То есть картина блокировок такая -Reload() нахер все
	блокирует до тех пор, пока не завершит запись в резолвер, в то время как ГЕТ может читать все это
	гавно конкурентно безопасно, блокируясь в моменте, когда Reload() ведет запись*/
	c.mu.Lock()
	//гарантированно снимаем лок после выполнения программы
	defer c.mu.Unlock()
	//заполняем res (все резолверы)
	c.res = res
	//ставим трушку, чтобы дать сигнал о том, что прогрузка выполнена успешно
	c.init = true
	return nil
}

/*Таким образом лок в этой системе нужен для того, чтобы не было разночтений между двумя присваиваниями
(пока пишется c.res и обновляется c.init). Другие сущности не смогут увидеть полуобновления. Что бы было,
если бы не было лока ? Так как у нас есть метод Get(), который отвечает за считываение текущего кэша
резолверов, существует вероятность в чтении между пишущим (Reload) и читающим (Get). Я не могу гарантировать
отсутствие гонок без мьютексов, поскольку в таком случае Гет может прочитать полусостояние (когда Reload еще
не отработал, а init истина). Соответственно...*/

/*Метод который отвечает за чтение из резолвера актуального состояния данных.*/
func (c *ResolverCache) Get() (game.Resolvers, bool) {
	//здесь ставим блокировку на чтение
	c.mu.RLock()
	//а здесь снимаем ее после выхода
	defer c.mu.RUnlock()
	//отдаем резолверы и инфу о том, прогружены ли они или нет
	return c.res, c.init
}

/*И получается следующая картина. На старте сервера я обязательно создаю resolverCach, который служит для
доменной части игры. После этого я прогреваю его свежими данными при помощи метода Reload(), который читает
шаблоны боевых и баф карт, мапит это все в память по ключу карт, собирает резолверы (в том числе Battle, Buff,
HeroAbility) и под блокировкой пишет всю инфу в c.res, при этом по результату выполнения ставя сигнальный флаг
о том, что инициализация резолвера отработала успешно. НО. Помимо загрузки всех данных, нужно как то выгрузить
их. Этим занимается метод Get(), который точно так же под блокировкой этой же сущности мьютекса читае из c.rec
и c.init, и если все круто (c.init-истина)-отдает набор резорвера, а если все печально-паника. В свою очередь
вся информация из данных от Get() подключается к обработчику действий (ApplyActionHandlerDeps), который в свою
очередь вызывает ApplyActionToMatchTx, который вызывает ApplyAction, которому и нужен резолвер, чтобы собственно
совершить действие. В свою очередь ApplyAction вызывает различные функции, так или иначе действующие на матч.
Соответственно этим функциям нужна карта. Эта тема называется LookUp-поиск данных по ключу. PlayBattleCard к
примеру требует в себя templateID. Откуда нам его взять? Из резолвера. НО! Здесь в игру вступают хендлеры, которые
и принимают входящие данные, что как следствие вынуждает проверять и входящие карты, а значит, надо обработать
и эти данные. Если такая проверка дает ложь, то выскочет доменная ошибка (как в PlayBattleCard). Если рисовать все
это добро, то получится :
при старте приложение прогружаем все добро в кэш памяти -> Гет читает все это добро из мэйна -> передаем все данные в
сценарный слой (ApplyActionToMatchTx)-> который в свою очередь вызывает доменную логику -> после этого отрабатывает слой,
отвечающий за запись изменений в БД (apply_action_tx, который вызывает репозиторный слой, отвечающий за запись, куда уходят
JSON с новым состоянием и прочим) -> далее мы записываем все изменения в DTO -> отдавая все изменения по транспортному пути
(HTTP/SSE)

А если конкретно, то рисунок выглядит так :
(матч типа начался и игрок хочет поставить карту на стол)
-клиент шлет POST запрос со всеми данными
-MUX роутит запрос на ApplyAction хендлер
-AuthMiddleWare валидирует сессию и кладет AuthUser в контекст
-NewApplyActionHandler достает юзера из контекста, парсит айди матча, декодит JSON в DTO.ApplyActionRequest,
вызывает ApplyActionToMatchTx(айди матча, айди пользователя, запрос, резолверы)
-В свою очередь ApplyActionToMatchTx читает строку матча из репозитория (Match) из БД, определяет индексы игроков,
берет версию, анмаршалит состояния, ставит новую версию и собирает Action(намерения клиента по матчу) из запроса,
после этого вызывая ApplyAction
-ApplyAction проверяет таймауты, фазы, версии, активного игрока и всего того, что нужно проверить перед ходом,
вызывая в зависимости от типа действия нужную функцию (в нашем случае PlayBattleCard)
-PlayBattleCard проверяет все что касается геймплея (таргет слот, активную фазу, ману и все такое)
-После PlayBattleCard цикл возвращается, передавая управления опять в ApplyActionToMatchTx, который
сохраняет новое состояние в БД.
-Далее вызывается PublishMatchToSSE, чтобы запушить новое состояние двум игрокам
-После этого клиент получает HTTP ответ с новым состоянием или SSE апдейт матча.
Transport->Application->Domain->Repository->Transport
*/
