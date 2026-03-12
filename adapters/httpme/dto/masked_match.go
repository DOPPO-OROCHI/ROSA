package dto

import "TheWar/internal/domain/game"

/*А вот это самый интересный файл. Тут происходит конкретная движуха в матче. Смотри в чем
прикол. Раньше я просто отдавал вообще все состояние, наивно, так делать не надо. Почему ?
Потому что противник не должен знать все данные о противнике. Соответственно нам надо как то
прятать эту инфу. Как ? Сами по себе эти структуры не маскируют данные, а просто отдают их
в формате, который удобен для клиента. Сама маскировка происходит в maskMatchStateForUser, из
файла handlers/mask.go. Вот там да, происходит уже конкретная маскировка, которая отдает игроку
только ту инфу, которую он должен видеть. Но здесь мы трем про DTO, так что вернемся к ним. Как
ты можешь видеть, здесь в некоторых полях мы засовываем доменные приколы. Почему? Потому что я
нуб. Так делать не надо, но тем не менее я так сделал. Сосите. */

type MaskedPlayerState struct {
	PlayerID               int    `json:"player_id"`                 //<-индекс игрока в матче, 0 или 1
	UserID                 uint   `json:"user_id"`                   //<-айди пользователя, нужен клиенту чтобы понять кто из players является текущим игроком
	HeroID                 uint   `json:"hero_id"`                   //<-айди шаблона героя
	HeroCode               string `json:"hero_code"`                 //<-код героя, который выбрал игрок
	HeroHP                 int    `json:"hero_hp"`                   //<-текущее здоровье героя игрока
	HeroLevel              int    `json:"hero_level"`                //<-уровень героя игрока
	HeroAttackPower        int    `json:"hero_attack_power"`         //<-текущая сила атаки героя игрока
	HeroAttackCooldown     int    `json:"hero_attack_cooldown"`      //<-текущее время перезарядки атаки героя игрока
	HeroAttackBaseCooldown int    `json:"hero_attack_base_cooldown"` //<-базовое кд атаки героя игрока
	HeroSplashRadius       int    `json:"hero_splash_radius"`        //<-радиус атаки героя игрока
	HeroAbilityCooldown    int    `json:"hero_ability_cooldown"`     //<-текущее кд способности героя игрока

	Mana  int `json:"mana"`  //<-понятно
	Turns int `json:"turns"` //<-кол-во ходов, которые игрок провел в матче
	//состояние стола игрока. Тут уже есть конкретные карты, которые стоят на столе, их характеристики и прочее
	Table [game.TableSize]*game.UnitState `json:"table"` //<-здесь проблема, поскольку я отдаю доменную часть, что хуево
	//А здесь уже инфа о руке, колоде, сбросе. НО! Она отдается только тому игроку, которому она принадлежит.
	//Типа, это отдается тому чуваку, который смотрит на это состояние. Можно сказать, он смотрит на себя
	Hand      []game.CardsInMatch `json:"hand,omitempty"`    //<-что в руке у игрока
	Deck      []game.CardsInMatch `json:"deck,omitempty"`    //<-что в колоде у игрока (сколько карт)
	Discard   []game.CardsInMatch `json:"discard,omitempty"` //<- сколько карт проебал
	HandCount int                 `json:"hand_count,omitempty"`
	DeckCount int                 `json:"deck_count,omitempty"`
	DiscCount int                 `json:"discard_count,omitempty"`
}

/*А это маскированное состояние матча. Эта структура описывает то, что игроку нужно знать о текущем состоянии матча,
к примеру, айдишник мачта, кто активен, фаза, и так далее*/
type MaskedMatchState struct {
	MatchID      uint                  `json:"match_id"`
	Version      int64                 `json:"version"`
	ActivePlayer int                   `json:"active_player"`
	Phase        game.TurnPhase        `json:"phase"`
	Finished     bool                  `json:"finished"`
	Result       game.MatchResult      `json:"result"`
	Players      [2]*MaskedPlayerState `json:"players"`
	Event        []game.Event          `json:"events,omitempty"`
}
