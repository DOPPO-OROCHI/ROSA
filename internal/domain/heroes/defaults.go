package heroes

//По примеру с картами, это дефолтный темплейт всех персонажей, которые представлены в игре

//здесь идет перечисление всех персов, их характеристики, имена, и абилки
var DefaultHeroTemplate = []CharacterTemplate{
	{
		Name:           "Imperial Commander", //<-имя, которое отображается в UI
		CharacterCode:  "imperial_commander", //<-уникальный код, который нужен для операционки
		AttackPower:    5,                    //<-сила атаки
		HealthPoints:   60,                  //<-хп
		AttackCooldown: 1,                    //<-кд атаки (не спел)
		Ability: AbilitySpec{ //<-а это спел
			Code:     BUFF_ATK, //<-с его кодом, который отражает смысл абилки
			Target:   OWN_UNIT, //<-цель спела
			CoolDown: 3,        //<-кд
			ManaCost: 2,        //<-стоимость в мане
			Value:    10,       //<-значение абилки (в случае атаки -урон, в случае бафа -плюсы)
			Duration: 3,        //<-длительность (если 0-баф вечный, далее по значению)
		},
		Description: "Раз в 3 хода повышает урон на 10 любой карты. Баф длится 3 хода", //<-описание для UI. Пока криво, не нравится
		//потому что описание должно быть о персонаже, а не его скиле. Ну потом пофикшу
	},
	{
		//далее по образу и подобию
		Name:           "Karn",
		CharacterCode:  "karn",
		AttackPower:    6,
		HealthPoints:   70,
		AttackCooldown: 2,
		Ability: AbilitySpec{
			Code:     MAKE_TANK,
			Target:   OWN_UNIT,
			CoolDown: 5,
			ManaCost: 5,
			Value:    0,
			Duration: 3,
		},
		Description: "Раз в 5 ходов делает из обычной карты танка. Баф длится 3 хода",
	},
	{
		Name:           "The System",
		CharacterCode:  "the_system",
		AttackPower:    1,
		HealthPoints:   80,
		AttackCooldown: 1,
		Ability: AbilitySpec{
			Code:     ATTACK_SPLASH,
			Target:   ENEMY_UNIT,
			CoolDown: 3,
			ManaCost: 3,
			Value:    4,
			Duration: 0,
		},
		Description: "Раз в три хода запускает ракеты, поражая цель и соседей на 4 урона",
	},
	{
		Name:           "Black Cell",
		CharacterCode:  "black_cell",
		AttackPower:    5,
		HealthPoints:   50,
		AttackCooldown: 1,
		Ability: AbilitySpec{
			Code:     BUFF_ATK_PERM,
			Target:   OWN_UNIT,
			CoolDown: 2,
			ManaCost: 2,
			Value:    10,
			Duration: 0,
		},
		Description: "Каждые 2 хода повышает атаку карты на 10 едениц навсегда, но здороввье такой карты всегда равно 1",
	},
	{
		Name:           "Жрец",
		CharacterCode:  "slavic_priest",
		AttackPower:    5,
		HealthPoints:   40,
		AttackCooldown: 1,
		Ability: AbilitySpec{
			Code:     HEAL_UNIT,
			Target:   OWN_UNIT,
			CoolDown: 2,
			ManaCost: 3,
			Value:    10,
			Duration: 0,
		},
		Description: "Каждые 2 хода лечит одну карту на 10",
	},
	{
		Name:           "Citadele",
		CharacterCode:  "suprime_lider",
		AttackPower:    2,
		HealthPoints:   50,
		AttackCooldown: 1,
		Ability: AbilitySpec{
			Code:     ATTACK_ANY,
			Target:   ENEMY_ANY,
			CoolDown: 3,
			ManaCost: 5,
			Value:    30,
			Duration: 0,
		},
	},
}
