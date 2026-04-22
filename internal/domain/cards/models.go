package cards

import "gorm.io/gorm"

/*Файл целиком и полностью посвящен описанию чертежей карт. Можно сказать что это именно что шаблон, по которому
в предыдущем файле я собирал карты. Так вот... Коль что добавится в геймплей в игре, если речь идет о картах, все
это должно быть отражено здесь, в чертеже. Тут полное описание всего, что может быть. Простейший файл, ничего сложного*/

type BattleCardTemplate struct {
	gorm.Model
	Name              string `gorm:"not null"`             //<-имя карты
	CodeString        string `gorm:"not null;uniqueIndex"` //<-уникальный код карты
	Description       string `gorm:"not null"`             //<-описание карты
	HealthPoints      int    `gorm:"not null;default:0"`   //<-ХП карты
	Attack            int    `gorm:"not null;default:0"`   //<-атака карты
	SplashRadius      int    `gorm:"not null;default:0"`   //<-сплеш основной атаки карты
	IsTank            bool   `gorm:"not null"`             //<-является ли карта танком
	CardType          string `gorm:"not null"`             //<-типа карты (типа человек, демон, или еще кто)
	BaseCooldown      int    `gorm:"not null"`             //<-базовый КД основной атаки
	ManaCost          int    `gorm:"not null;default:1"`   //<-стоимость карты в мане
	MaxCopies         int    `gorm:"not null;default:1"`   //<-максимальное количество копий карты
	ImageKey          string //<-ключ картинки карты
	AssetBaseKey      string //<-ключ анимаций карты
	SkillImageKey     string //<-картинка скилла карты (если есть)
	HasSkill          bool   `gorm:"not null;default:false"` //<-есть ли у карты скилл ?
	SkillName         string //<-имя скилла
	SkillCode         string //<-уникальный код скилла
	SkillKind         string //<-типа скилла (подробнее вы skills.go)
	SkillTargeting    string //<-таргет скилла
	SkillPower        int    //<-сила\значение скилла
	SkillBaseCooldown int    //<-базовый КД скилла
	SkillDuration     int    //<-длительность скилла (актуально для скиллов -дебафов)
	SkillExtraValue   int    //<-экстра значение скилла (когда карта одновременно и хилит и пиздится)
	SkillBuffEffect   string //<-эффект от бафа (если пусто, значит это не баф эффект)
	SkillDebuffEffect string //<-та же тема только для дебафа
	SkillCleanseMode  string //<-может ли карта снимать положительные-отрицательные эффекты (skills.go)
	SkillIgnoreTank   bool   `gorm:"not null;default:false"` //<-игнорит ли скилл танк ?
	SkillApplyCount   int    `gorm:"not null;default:0"`     //<-сколько раз или на сколько целей применяется скилл ?
	HasPassive        bool
	Passive           PassiveSpec `gorm:"embedded;embeddedPrefix:passive_"`
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
