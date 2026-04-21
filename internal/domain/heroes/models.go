package heroes

import "gorm.io/gorm"

//описание чертежей героев с их способностями. На основе этой херни создаются новые герои

type CharacterTemplate struct {
	gorm.Model
	Name           string      `gorm:"not null"`                         //<-имя персонажа для UI
	CharacterCode  string      `gorm:"not null;uniqueIndex"`             //<-код персонажа для операционки
	AttackPower    int         `gorm:"not null;default:0"`               //<-сила атаки героя
	HealthPoints   int         `gorm:"not null;default:0"`               //<-хп героя
	AttackCooldown int         `gorm:"not null;default:0"`               //<-кд атаки
	SplashRadius   int         `gorm:"not null;default:0"`               //<-радиус сплеша если вообще есть
	Ability        AbilitySpec `gorm:"embedded;embeddedPrefix:ability_"` //<-ну понятно
	Description    string      `gorm:"text;not null"`                    //<-описание для UI
	ImageKey       string      //<-картинка
	AssetBaseKey   string      //<-набор эффектов
	SkillImageKey  string      `gorm:"column:skill_image_key"`  //<-ключ картинки скилла
	AttackImageKey string      `gorm:"column:attack_image_key"` //<-ключ картинки атаки героя
}

//а это абилити персонажа, уникальны
type AbilitySpec struct {
	Name         string
	Code         string `gorm:"not null"` //<-код абилки (отражает целеполагание)
	Description  string
	Kind         string
	Target       string //<-цель абилки
	CoolDown     int    `gorm:"not null;default:1"` //<-кд абилити
	ManaCost     int    `gorm:"not null;default:0"` //<-стоимость в мане
	Power        int    `gorm:"not null;default:0"` //<-значение способности (сколько урона,хила,или еще какой хуйни наносит)
	Duration     int    `gorm:"not null;default:0"` //<-длительность
	ExtraValue   int    `gorm:"not null;default:0"`
	ApplyCount   int    `gorm:"not null;default:0"`
	BuffEffect   string
	DebuffEffect string
	IgnoreTank   bool `gorm:"not null;default:false"`
}
