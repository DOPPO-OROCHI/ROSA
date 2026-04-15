package game

import "TheWar/internal/domain/cards"

/*Файл посвященный доменной части зависимостей, от которых зависит состояние матча. Используется исключительно внутриигровой
логикой (функциями). Ладно не все. Тут я тупанул и ебанул некоторые DTO, но изменить это я уже не смогу, во избежание взаимного
импорта пакетов. Поэтому имеем что имеем. Грязный код да...*/

/*
Структура данных о том, что такое матч в логике домена. Все то, что будет добавляться в матч,
в том числе таймер (который я так и не ввел) будет добавляться сюда, а в последствии и в DTO.
*/
type MatchState struct {
	MatchID       uint            //<-чтобы внутри JSON было понятно, к какой строке (внутри БД) матча относится состояние
	Version       int64           //<-чтобы избежать повторный ход и тд
	Players       [2]*PlayerState //<-строго два игрока
	ActivePlayer  int             //<-кто сейчас ходит ?
	Phase         TurnPhase       //<-фазы, типа старт или мэйн. Нужно для того, чтобы отделить системные изменения от игрока
	Finished      bool            //<-завершен ли матч, или еще нет
	Result        MatchResult     //<-ну и результат, кто победил
	Events        []Event         //<-чисто для анимаций.
	TurnStartedAt int64           //<-время начала хода, для таймера
	TurnDeadline  int64           //<-время, когда ход закончится, для таймера
	TurnTimeSec   int             //<-сколько секунд на ход
}

/*
По аналогии с состоянием матча, существует так же состояние игрока, которые так же имеют
свои айди, героев, и так далее и тому подобное.
*/
type PlayerState struct {
	PlayerID                int
	UserID                  uint   //<-это внутренний айди пользователя из ТГ
	HeroID                  uint   //<-айдищник героя, с которым пришел юзер (БД)
	HeroCode                string //<-код этого же героя. Так надо, потому что так стабильнее
	HeroHP                  int    //<-ресурсная часть
	HeroLevel               int    //<-здесь будет отражаться левел персонажа
	HeroAttackPower         int    //<-ресурсная часть
	HeroAttackCooldown      int    //<-ресурсная часть
	HeroAttackBaseCooldown  int    //<-ресурсная часть
	HeroSplashRadius        int    //<-новая механика, типа слпешь по остальным картам
	HeroAbilityCooldown     int    //<-кулдаун способности
	HeroAbilityBaseCooldown int    //<-базовый кулдаун способности
	HeroAbilityManaCost     int
	Mana                    int                   //<-ресурсная часть
	Turns                   int                   //<-то, сколько ходов сделал чувак
	Deck                    []CardsInMatch        //<-дека игрока, которую он должен собрать до матча
	Hand                    []CardsInMatch        //<-что у персонажа в руке
	Discard                 []CardsInMatch        //<-сколько карт игрок уже проебал
	Table                   [TableSize]*UnitState //<-что у игрока не столе
	GraveYard               []GraveEntry
	PendingRes              []PendingResurrected
}

/*Карты внутри матча. Или с какими картами пришел игрок в матч.*/
type CardsInMatch struct {
	InstanceID    string `json:"instance_id"`   //<-уникальный айди карты (каждый матч разный айди)
	Kind          string `json:"kind"`          //<-тип картыю "battle" или "buff"
	TemplateID    string `json:"template_id"`   //<-CodeString из шаблона карты
	GamerCardID   uint   `json:"gamer_card_id"` //<-айдишник владеня карты (потенциально нужна для начисления опыта и тд)
	CardLevel     int    `json:"card_level"`    //<-уровень карты, с которым пришел игрок в момент создания матча
	Name          string `json:"name"`
	Description   string `json:"description"`
	ManaCost      int    `json:"mana_cost"`
	Attack        int    `json:"attack"`
	HealthPoints  int    `json:"health_points"`
	CardType      string `json:"card_type"`
	ImageKey      string `json:"image_key"`
	AssetBaseKey  string `json:"asset_base_key"`
	SplashRadius  int    `json:"splash_radius"`
	BaseCooldown  int    `json:"base_cooldown"`
	HasSkill      bool   `json:"has_skill"`
	SkillImageKey string `json:"skill_image_key"`
}

/*Состояние отдельно взятой карты внутри матча*/
type UnitState struct {
	InstanceID      string               `json:"instance_id"`   //<-рандомносгенерированный айди
	TemplateID      string               `json:"template_id"`   //<-чертеж карты
	GamerCardID     uint                 `json:"gamer_card_id"` //<-айдишник карты с владения
	CardLevel       int                  `json:"card_level"`    //<-уровень карты владения
	HP              int                  `json:"hp"`            //<-это копия статки карты, на момент выхода на стол
	MaxHP           int                  `json:"max_hp"`
	Attack          int                  `json:"attack"`
	SplashRadius    int                  `json:"splash_radius"`
	IsTank          bool                 `json:"is_tank"`
	CardType        string               `json:"card_type"`
	BaseCooldown    int                  `json:"base_cooldown"`
	Cooldown        int                  `json:"cooldown"`
	SummonedInTurn  int                  `json:"summoned_in_turn"`
	ImageKey        string               `json:"image_key"`
	AssetBaseKey    string               `json:"asset_base_key"`
	HasSkill        bool                 `json:"has_skill"`
	SkillImageKey   string               `json:"skill_image_key"`
	Skill           cards.UnitSkillState `json:"skill"`
	Effects         []UnitEffect         `json:"effects"`
	ResurrectedUsed bool                 `json:"resurrected_used"`
}

// эффект, или баф, который можно наложить на карту
type UnitEffect struct {
	EffectType       string `json:"effect_type"`        //<-тип эффекта, бафа
	TurnsLeft        int    `json:"turns_left"`         //<-то, сколько ходов длится баф (важно, если 0-то баф перманентный)
	Value            int    `json:"value"`              //<-величина бафа
	ExtraValue       int    `json:"extra_value"`        //<-специальное число под отражения (щиты)
	SourceType       string `json:"source_type"`        //<-от кого пришел эффект
	Polarity         string `json:"polarity"`           // <-типа бафф, дебаф
	SourceInstanceID string `json:"source_instance_id"` //<-от кого пришел эффект ?
	Dispellable      bool   `json:"dispellable"`        //<-можно ли снять эффект ?
	Targeting        string `json:"targeting"`          //<-это короче для эффектов после смерти (к примеру, взрыв после)
}

//Ну вот и пошли косяки по разграничению доменов и DTO. Как уже говорил, я знаю об этом косяке, не надо тут это...

// так и вот. Это структура ивента, служащая для отдачи игроку. Как уже понятно это DTO. Здесь описывается все то, что необходимо для анимации
type Event struct {
	Type                  string        `json:"type"`                              //<-тип действия
	PlayerIndex           int           `json:"player_index,omitempty"`            //<-сторона, которая совершила действие
	SourceKind            string        `json:"source_kind,omitempty"`             //<-источник, который совершил действие
	SourceInstanceID      string        `json:"source_instance_id,omitempty"`      //<-айди карты
	SourceTemplateID      string        `json:"source_template_id,omitempty"`      //<-темплейт карты
	SourceHeroCode        string        `json:"source_hero_code,omitempty"`        //<-это на случай, если источник-перс
	SourceCardTemplateID  string        `json:"source_card_template_id,omitempty"` //<-каст,баф из руки (не со стола)
	VFXKey                string        `json:"vfx_key,omitempty"`                 //<-видеоэффект
	SFXKey                string        `json:"sfx_key,omitempty"`                 //<-саундэффект
	TargetSlot            int           `json:"target_slot,omitempty"`             //<-для суммона
	Targets               []EventTarget `json:"targets,omitempty"`                 //<-для атак, или бафов, или хила
	VisibleForPlayerIndex *int          `json:"visible_for_player_index,omitempty"`
}

// а это структура, которая обращается уже к тому, на кого было воздействие. Типа, есть анимации у того кто бьет и у того, кого бьют
type EventTarget struct {
	InstanceID string `json:"instance_id"`
	TemplateID string `json:"template_id,omitempty"` //<-чтобы UI смог выбрать hit/death анимацию
	Amount     int    `json:"amount,omitempty"`      //<-значение действия (сколько кто отхилил, нанес урона)
	Died       bool   `json:"died,omitempty"`        //<-умер ли чувак после применения
	NewHP      int    `json:"new_hp,omitempty"`      //<-новые хп
}

/*
а это структура действия. Здесь выражаются все намерения игрока о том, что он сейчас будет вообще мутить.
Эта структура нужна дла того, чтобы игрок смог взаимодействовать путем передачи данных с нашим основным
диспетчером вызова функций (ApplyAction). Дивгло в свою очередь читает эту структуру и делает свои приколы
*/
type Action struct {
	PlayerIndex      int        `json:"player_index"`                 //<-тут прикол. В моей игре есть два индекса. 0 и 1. Это поле отражает одного из игроков
	Type             ActionType `json:"type"`                         //<-какое действие делает игрок, исходя из наших констант
	CardInstanceID   string     `json:"card_instance_id,omitempty"`   //<-какую карту использует игрок (play_battle/buff_card карта из руки, card_attack -атакующая карта на столе)
	TargetInstanceID string     `json:"target_instance_id,omitempty"` //<-цель. Для атаки-карта на столе противника, либо герой. Для бафа -карта на столе, либо для атаки героя
	AttackHero       bool       `json:"attack_hero,omitempty"`        //<-отражает истинность намерений типа -мы ебем персонажа или нет ?
	ExpectedVersion  int64      `json:"expected_version"`             //<-защита от повторов или устаревших действий. Игрок не сможет дважды атаковать одной и той же картой за ход
	TargetSlot       int        `json:"target_slot,omitempty"`        //<-новая механика, которая позволяет в добавок ко всему пиздить соседние карты сплешем
	KillerInstanceID string     `json:"killer_instance_id"`
	KillerOwnerIdx   int        `json:"killer_owner_index"`
}

// ГЛОБАЛЬНЫЙ ПАТЧ. Внедрение механики возраждения после смерти
type GraveEntry struct {
	Unit       UnitState
	DiedAtTurn int
}

type PendingResurrected struct {
	InstanceID string
	DueTurn    int
}

/*
Функция поиска карты на столе по его айди. Тут мы и принимаем этот айдишни, при этом возвращая
инт(это слот где находится карта) и указатель на карту, которую мы нашли. Эта функция нужна для того,
чтобы дать корректную цель для всех действий которые только есть, будь то атака или баф
*/
func (p *PlayerState) FindSlot(instanceID string) (int, *UnitState) {
	//идем циклом по всем ячейкам на столе
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		//если нашли -заебись, отдаем слот и карту
		if u != nil && u.InstanceID == instanceID {
			return i, u
		}
	}
	//если нет, то такой карты не сущесвтует
	return -1, nil
}

/*
Функция удаления карты со стола. Вызывается в тех моментах, когда по нескольким причинам карта
больше не может существовать на столе (к примеру -у карты нет хп). Вызывается чаще всего при смерти
(и только, потому что пока что разраб других механик не завез)
*/
func (p *PlayerState) RemoveAt(slot int) {
	//валидируем входящий слот. Если ничего не нашли -тупо выходим
	if slot < 0 || slot >= TableSize || p.Table[slot] == nil {
		return
	}
	//удаляем карту со стола
	p.Table[slot] = nil
}

/*Как то так описываются структуры для доменного взаимодействия. Понимаю, так то тут еще есть
дтошки, что в целом ломает чистый код и все такое но кого ебет, кроме пытливого взгляда опытного
чела. Так вот. Все это нужно для того, чтобы круто, удобно и приятно организовывать собственно
взаимодействие эти структур, дтошек и функций, о которых мы поговорим в файле turn.go*/
