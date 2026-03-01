package game

type MatchState struct {
	MatchID        uint            //<-чтобы внутри JSON было понятно, к какой строке (внутри БД) матча относится состояние
	Version        int64           //<-чтобы избежать повторный ход и тд
	Players        [2]*PlayerState //<-строго два игрока
	ActivePlayer   int             //<-кто сейчас ходит ?
	Phase          TurnPhase       //<-фазы, типа старт или мэйн. Нужно для того, чтобы отделить системные изменения от игрока
	Finished       bool            //<-завершен ли матч, или еще нет
	Result         MatchResult     //<-ну и результат, кто победил
	Events         []Event         //<-чисто для анимаций.
	TurnStartedAt  int64           //<-время начала хода, для таймера
	TurnDeadLineAt int64           //<-время, когда ход закончится, для таймера
	TurnTimeSec    int             //<-сколько секунд на ход
}

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
}

type CardsInMatch struct {
	InstanceID  string //<-разобрали уже
	Kind        string //<-тип картыю "battle" или "buff"
	TemplateID  string //<-CodeString из шаблона карты
	GamerCardID uint   //<-айдишник владеня карты (потенциально нужна для начисления опыта и тд)
	CardLevel   int    //<-уровень карты, с которым пришел игрок в момент создания матча
}

type UnitState struct {
	InstanceID     string
	TemplateID     string
	GamerCardID    uint
	CardLevel      int
	HP             int  //<-это копия статки карты, на момент выхода на стол
	Attack         int  //<-это копия статки карты, на момент выхода на стол
	SplashRadius   int  //<-радиус атаки
	CanBeUpgraded  bool //<-тонкий момент, этот параметр копируется из шаблона, а дальше используется для бафов
	Cooldown       int  //<-кд карты
	IsTank         bool //<-является ли карта танком
	SummonedInTurn int  //<-этот параметр равен ходу, в который была призвана карта, чтобы она не могла
	MaxHP          int  //<-максимальное кличество ХП
	// атаковать тогда, когда была только разыграна
	Effects  []UnitEffect //<-эффекты на карте. Используется в последствии. Ни одна карта в начале не имеет эффектов
	CardType string       //<-тип карты. Используется для апгрейдов
}
type UnitEffect struct {
	EffectType string //<-объяснил
	TurnsLeft  int    //<-то, сколько ходов длится баф (важно, если 0-то баф перманентный)
	Value      int    //<-величина бафа
}

type Event struct {
	Type                 string        `json:"type"`                              //<-тип действия
	PlayerIndex          int           `json:"player_index,omitempty"`            //<-сторона, которая совершила действие
	SourceKind           string        `json:"source_kind,omitempty"`             //<-источник, который совершил действие
	SourceInstanceID     string        `json:"source_instance_id,omitempty"`      //<-айди карты
	SourceTemplateID     string        `json:"source_template_id,omitempty"`      //<-темплейт карты
	SourceHeroCode       string        `json:"source_hero_code,omitempty"`        //<-это на случай, если источник-перс
	SourceCardTemplateID string        `json:"source_card_template_id,omitempty"` //<-каст,баф из руки (не со стола)
	VFXKey               string        `json:"vfx_key,omitempty"`                 //<-видеоэффект
	SFXKey               string        `json:"sfx_key,omitempty"`                 //<-саундэффект
	TargetSlot           int           `json:"target_slot,omitempty"`             //<-для суммона
	Targets              []EventTarget `json:"targets,omitempty"`                 //<-для атак, или бафов, или хила
}

type EventTarget struct {
	InstanceID string `json:"instance_id"`
	TemplateID string `json:"template_id,omitempty"` //<-чтобы UI смог выбрать hit/death анимацию
	Amount     int    `json:"amount,omitempty"`      //<-значение действия (сколько кто отхилил, нанес урона)
	Died       bool   `json:"died,omitempty"`        //<-умер ли чувак после применения
	NewHP      int    `json:"new_hp,omitempty"`      //<-новые хп
}

type Action struct {
	PlayerIndex      int        `json:"player_index"`                 //<-тут прикол. В моей игре есть два индекса. 0 и 1. Это поле отражает одного из игроков
	Type             ActionType `json:"type"`                         //<-какое действие делает игрок, исходя из наших констант
	CardInstanceID   string     `json:"card_instance_id,omitempty"`   //<-какую карту использует игрок (play_battle/buff_card карта из руки, card_attack -атакующая карта на столе)
	TargetInstanceID string     `json:"target_instance_id,omitempty"` //<-цель. Для атаки-карта на столе противника, либо герой. Для бафа -карта на столе, либо для атаки героя
	AttackHero       bool       `json:"attack_hero,omitempty"`        //<-отражает истинность намерений типа -мы ебем персонажа или нет ?
	ExpectedVersion  int64      `json:"expected_version"`             //<-защита от повторов или устаревших действий. Игрок не сможет дважды атаковать одной и той же картой за ход
	TargetSlot       int        `json:"target_slot,omitempty"`        //<-новая механика, которая позволяет в добавок ко всему пиздить соседние карты сплешем
}

func (p *PlayerState) HasFreeSlot() bool {
	for i := 0; i < TableSize; i++ {
		if p.Table[i] == nil {
			return true
		}
	}
	return false
}

func (p *PlayerState) FindSlot(instanceID string) (int, *UnitState) {
	for i := 0; i < TableSize; i++ {
		u := p.Table[i]
		if u != nil && u.InstanceID == instanceID {
			return i, u
		}
	}
	return -1, nil
}

func (p *PlayerState) RemoveAt(slot int) bool {
	if slot < 0 || slot >= TableSize || p.Table[slot] == nil {
		return false
	}
	p.Table[slot] = nil
	return true
}

func (p *PlayerState) AdjacentSlots(slot int) (int, int) {
	return slot - 1, slot + 1
}

func (p *PlayerState) AdjacentUnits(slot int) []*UnitState {
	out := make([]*UnitState, 0, 2)
	left, right := slot-1, slot+1
	if left >= 0 && p.Table[left] != nil {
		out = append(out, p.Table[left])
	}
	if right < TableSize && p.Table[right] != nil {
		out = append(out, p.Table[right])
	}
	return out
}
