package cards

/*Описание дефолтных карт. В будущем будут и донатные, но сейчас речь не об этом. Здесь в общем то сердце
всех карт, которые были, есть и будут. Как уже говорилось в описании констант, у каждой карты есть свое
уникальное имя, код (для операционки), хп, атака и все то, что увидишьт внутри описания. Так вот, важно
здесь не то, что и ежу понятно (в роли атаки, хп), а такие штуки как BuffSlot->который отвечает за то, можно
ли улучшить карту, MaxCopies->который отвечает за то, сколько карт конкретного типа можно взять в матч.
Сделано чисто на тот вариант, чтобы игрок не мог с собой взять все 20 карт одной и той же масти, если можно
так сказать конечно. К слову это и сидится вместе с миграциями в моем замечательном пакете bootstrap.*/

var DefaultBattleCards = []BattleCardTemplate{
	{
		Name:         "Имперский пехотинец", //<-имя карты
		CodeString:   "imperial_guardian",   //<-код карты для простого обращения к нему. Используется в операционке
		HealthPoints: 15,                    //<-хп карты
		Attack:       5,                     //<-сила атаки карты
		IsTank:       false,                 //<-является ли карта танком (это вообще отдельный разговор)
		CardType:     OrganicalCard,         //<-тип карты, нужен для бафов и прочей ерунды
		CoolDown:     1,                     //<-кд атаки карты. Если значение 1-значит карта может бить только через ход
		ManaCost:     1,                     //<-стоимость карты в мане
		BuffSlot:     true,                  //<-есть ли возможность  улучшить карту с помощью бафа
		MaxCopies:    5,                     //<-максимально допустимое количество копий, которое игрок может взять в матч
		Description:  "",                    //<-описание. Используется для UI, чтобы игрок смог прочесть че вообще за карта
		ImageKey:     "",                    //<-ключ картинки, которая отражает то, как выглядит карта
		AssetBaseKey: "",                    //<-а здесь будут храниться ключи анимации (вообще отдельная тема)
	},
	{
		Name:         "Механический рыцарь",
		CodeString:   "mechanical_knight",
		HealthPoints: 60,
		Attack:       3,
		IsTank:       true,
		CardType:     MechanicalCard,
		CoolDown:     1,
		ManaCost:     3,
		BuffSlot:     true,
		MaxCopies:    3,
		Description:  "",
		ImageKey:     "",
		AssetBaseKey: "",
	},
	{
		Name:         "Беспилотные дроны",
		CodeString:   "drones",
		HealthPoints: 20,
		Attack:       8,
		IsTank:       false,
		CardType:     MechanicalCard,
		CoolDown:     2,
		ManaCost:     5,
		BuffSlot:     true,
		MaxCopies:    4,
		Description:  "",
		ImageKey:     "",
		AssetBaseKey: "",
	},
}

/*По аналогии с картами, это стандартное описание всех возможных бафф карт, которые есть в игре (так же будут платные)
Как ты можешь увидеть здесь все работает по схожему принципу, за исключением того, что у баф карт очевидно не может
быть силы атаки, кд и всего того, что может бить, стрелять и производить действия на поле боя. Так вот...*/
var DefaultBuffCards = []BuffCardsTemplate{
	{
		Name:         "Адреналин",   //<-имя
		CodeString:   "adrenalin",   //<-уникальный код карты
		ManaCost:     1,             //<-стоимость в мане
		BuffType:     DamageUpdate,  //<-тип улучшения
		BuffValue:    5,             //<-число, на которое баф увеличивает то, или иное значение карты
		OnlyFor:      OrganicalCard, //<-то, к какому типу карт можно применить баф (в случае адреналина только к органическим картам)
		MaxCopies:    4,             //<-максимально допустимое количество копий одной карты в матче
		Duration:     1,             //<-длительность бафа. Если значение 0-баф постоянный. Если значение 1-баф действует 1 ход и далее
		Description:  "",            //<-описание для UI
		ImageKey:     "",            //<-картинка. Тоже для UI
		AssetBaseKey: "",            //<-и ключи анимаций соответственно
	},
	{
		Name:         "Линейный привод",
		CodeString:   "linear_actuator",
		ManaCost:     2,
		BuffType:     DamageUpdate,
		BuffValue:    3,
		OnlyFor:      MechanicalCard,
		MaxCopies:    3,
		Duration:     1,
		Description:  "",
		ImageKey:     "",
		AssetBaseKey: "",
	},
	{
		Name:         "Обновление процессоров",
		CodeString:   "processor_update",
		ManaCost:     2,
		BuffType:     CoolDownUpdate,
		BuffValue:    1,
		OnlyFor:      MechanicalCard,
		MaxCopies:    2,
		Duration:     0,
		Description:  "",
		ImageKey:     "",
		AssetBaseKey: "",
	},
}
