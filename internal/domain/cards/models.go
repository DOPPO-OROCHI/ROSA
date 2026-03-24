package cards

import "gorm.io/gorm"

/*Файл целиком и полностью посвящен описанию чертежей карт. Можно сказать что это именно что шаблон, по которому
в предыдущем файле я собирал карты. Так вот... Коль что добавится в геймплей в игре, если речь идет о картах, все
это должно быть отражено здесь, в чертеже. Тут полное описание всего, что может быть. Простейший файл, ничего сложного*/

type BattleCardTemplate struct {
	gorm.Model
	Name         string `gorm:"not null"`             //<-имя карты
	CodeString   string `gorm:"not null;uniqueIndex"` //<-уникальный код карты
	HealthPoints int    `gorm:"not null;default:0"`   //<-количество хп карты
	Attack       int    `gorm:"not null;default:0"`   //<-атакующая сила карты
	SplashRadius int    `gorm:"not null;default:0"`   //<-радиус сплеша
	IsTank       bool   `gorm:"not null"`             //<-является ли карта танком
	CardType     string `gorm:"not null"`             //<-тип карты Mech, Organic, Demonic, Healer
	CoolDown     int    `gorm:"not null;default:1"`   //<-кд карты
	ManaCost     int    `gorm:"not null;default:1"`   //<-сколько стоит карта
	BuffSlot     bool   `gorm:"not null"`             //<-можно ли улучшить карту
	MaxCopies    int    `gorm:"not null;default:1"`   //<-максимальное число копий у игрока

	Description     string `gorm:"not null"` //<-описание карты
	ImageKey        string //<-картинка карты
	AssetBaseKey    string //<-в будущем я сделаю анимации
	SkillImageKey   string `gorm:"column:skill_image_key"` //<-картинка скилла карты
	SkillName       string //<-имя скилла
	SkillCode       string //<-код скилла
	SkillTrigger    string //<-когда скилл активируется ?
	SkillTarget     string //<-таргет скила
	SkillValue      int    //<-значение скилла
	SkillDuration   int    //<-продолжительность скилла
	SkillCooldown   int    //<-кд скилла
	SkillParamsJSON string //<-для спец приколов типа -обойти танка
	
	PassiveImageKey       string //<-картинка пассивки
	PassiveName           string //<-имя
	PassiveCode           string //<-код
	PassiveTrigger        string //<-когда действует пассивка ?
	PassiveTarget         string //<-цель пассивки
	PassiveEffect         string //<-еффект пассивки
	PassiveCondition      string //<-при каком условии пассивка активна ?
	PassiveValue          int    //<-значение пассивки
	PassiveDuration       int    //<-0-постоянный
	PassiveScale          string //<-от чего скейлится пассивка (если на столе 2 демона)
	PassiveCountOwner     string //<-где считаем условия ? (наш стол, противника, оба)
	PassiveConditionCount int    //<-порог для условия
	PassiveCountType      string //<-кого считаем для условия
	PassiveCountCode      string //<-ЗАПОМНИТЬ!!! Карты могут считать и по просто классам (Demonical, Organical).
	//А могут считаться по кодам других карт
}

type BuffCardsTemplate struct {
	gorm.Model
	Name         string `gorm:"not null"`             //<-имя баф карты
	CodeString   string `gorm:"not null;uniqueIndex"` //<-уникальный код карты
	ManaCost     int    `gorm:"not null;default:1"`   //<-стоимость в мане
	BuffType     string `gorm:"not null"`             //<-тип бафа
	BuffValue    int    `gorm:"not null;default:1"`   //<-значение бафа
	OnlyFor      string `gorm:"not null"`             //<-для какого типа карт этот баф
	MaxCopies    int    `gorm:"not null;default:1"`   //<-максимальное число копий у игрока
	Duration     int    `gorm:"not null;default:0"`   //<-длительность бафа
	Description  string `gorm:"not null"`
	ImageKey     string
	AssetBaseKey string
}

/*Короче, потом надо добавить шкурки на карты*/
