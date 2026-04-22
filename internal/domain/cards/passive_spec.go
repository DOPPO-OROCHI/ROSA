package cards

type PassiveSpec struct {
	Name           string //<-понятно
	Code           string `gorm:"not null"` //<-уникальный код пассивки
	Description    string
	Kind           string `gorm:"not null"` //<-тип пассивки, либо аура, либо реакция
	Trigger        string `gorm:"not null"` //<-когда пассивка срабатывает ?
	Target         string `gorm:"not null"` //<-кого затрагивает пассивка ?
	TargetRace     string
	Power          int    `gorm:"not null;default:0"` //<-основное значение эффекта
	Duration       int    `gorm:"not null;default:0"` //<-если пассивка вешает временный эффект
	ExtraValue     int    `gorm:"not null;default:0"` //<-доп параметр под механики
	ApplyCount     int    `gorm:"not null;default:0"` //<-сколько раз применить, сколько целей задеть
	BuffEffect     string //<-понятно
	DebuffEffect   string //<-тоже понятно
	Condition      string `gorm:"not null"` //<-дополнительное условие активации
	ConditionRace  string //<-раса
	ConditionValue int    `gorm:"not null;default:0"` //<-число, сколько карт определенной расы должно быть
	EventFilter    string //<-какое событие интересует ? Ну типа, игрок разыграл человека, игрок кастанул скилл
	EventRace      string //<-если событие как раз связано с расой
	ScaleMode      string
	EventIsTank    bool `gorm:"not null;default:false"`
	IgnoreTank     bool `gorm:"not null;default:false"` //<-игрорим танка если бьем ?
}
