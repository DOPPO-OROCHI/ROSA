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
	//для UI
	Description  string `gorm:"not null"` //<-описание карты
	ImageKey     string
	AssetBaseKey string
	BackPic      string `gorm:"column:back_pic"` //<-шкурка карты
}

type BuffCardsTemplate struct {
	gorm.Model
	Name       string `gorm:"not null"`             //<-имя баф карты
	CodeString string `gorm:"not null;uniqueIndex"` //<-уникальный код карты
	ManaCost   int    `gorm:"not null;default:1"`   //<-стоимость в мане
	BuffType   string `gorm:"not null"`             //<-тип бафа
	BuffValue  int    `gorm:"not null;default:1"`   //<-значение бафа
	OnlyFor    string `gorm:"not null"`             //<-для какого типа карт этот баф
	MaxCopies  int    `gorm:"not null;default:1"`   //<-максимальное число копий у игрока
	Duration   int    `gorm:"not null;default:0"`   //<-длительность бафа
	//для UI
	Description  string `gorm:"not null"`
	ImageKey     string
	AssetBaseKey string
	BackPic      string `gorm:"column:back_pic"` //<-шкурка карты
}

/*Короче, потом надо добавить шкурки на карты*/
