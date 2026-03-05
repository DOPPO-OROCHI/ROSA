package dto

type CardKind string

/*Файл посвящен DTO для работы с картами. Здесь мы по сути формируем ответы на то, какие карты есть
у пользователя и все такое. Константы ниже нужны для определения типа карт, чтобы не писать вручную
а конкретно зафиксировать типы (battle\buff).*/

const (
	CardKindBattle CardKind = "battle"
	CardKindBuff   CardKind = "buff"
)

/*
Это структура, которая нужна для ответа на запрос о картах. Она содержит в себе две части,
карты для боя и карты баффов. Каждая из чатей является массивом из тех карт, которые есть у
игрока с их полным описанием. По существу, когда пользователь отправляет запрос на получение
списка карт, мы формируем DTO с этими двумя массивами, после чего игрок получает приличный список
*/
type CardsListResponse struct {
	Battle []OwnedBattleCardsDTO `json:"battle"`
	Buff   []OwnedBuffCardsDTO   `json:"buff"`
}

/*
DTO с картами битвы которые привязаны к конкретному игроку. Здесь есть все необходимое для описания,
начиная от типа карты, заканчивая уроввнем прокачки конкретной карты и ключей эффектов. В общем все, что нужно
*/
type OwnedBattleCardsDTO struct {
	Kind         CardKind `json:"kind"`
	TemplateID   string   `json:"template_id"`
	Name         string   `json:"name"`
	CardType     string   `json:"card_type"`
	ManaCost     int      `json:"mana_cost"`
	HealthPoints int      `json:"health_points"`
	Attack       int      `json:"attack"`
	SplashRadius int      `json:"splash_radius"`
	Cooldown     int      `json:"cooldown"`
	IsTank       bool     `json:"is_tank"`
	BuffSlot     bool     `json:"buff_slot"`
	MaxCopies    int      `json:"max_copies"`

	OwnedCardID uint `json:"owned_card_id"`
	Copies      int  `json:"copies"`
	Level       int  `json:"level"`
	XP          int  `json:"xp"`

	ImageKey     string `json:"image_key"`
	AssetBaseKey string `json:"asset_base_key"`
}

/*
То же самое и для Бафф карт, за исключением тех мест, где играет роль вообще структура бафа.
Баф карты не могут бить, например, и прочие приколы
*/
type OwnedBuffCardsDTO struct {
	Kind       CardKind `json:"kind"`
	TemplateID string   `json:"template_id"`
	Name       string   `json:"name"`
	ManaCost   int      `json:"mana_cost"`
	BuffType   string   `json:"buff_type"`
	BuffValue  int      `json:"buff_value"`
	OnlyFor    string   `json:"only_for"`
	Duration   int      `json:"duration"`
	MaxCopies  int      `json:"max_copies"`

	OwnedCardID  uint   `json:"owned_card_id"`
	Copies       int    `json:"copies"`
	Level        int    `json:"level"`
	XP           int    `json:"xp"`
	ImageKey     string `json:"image_key"`
	AssetBaseKey string `json:"asset_base_key"`
}

/*Таким образом я создал структуры DTO для работы с картами игрока,
Это позволяет мне возвращать данные в формате, который удобен для
игрока, но что самое важное, это позволяет мне абстрагироваться от
доменной модели, не передавая игроку лишние данные.*/
