package game

/*Файл с содержанием карточных чертежей. Первый для боевых карт (Battle), второй для баф карт (Buff).
Особенностью данных структур является то, что они нужны непосредственно домену. Почему ? Смотрите в чем
прикол дорогие мои... У нас уже есть модели карт, да (cards/models). Даааа, но как бы то слой БД шаришь ?
А вот эти две структуры нужны для того, чтобы мапить в них те структуры, которые живут в БД. Живите с этим.
Подробнее об этом будем тереть в резолверах, а пока закрепим то, что данные структуры онли рантайм
компонент из DTO, в который мы мапим нужные для двигла поля, чтобы доменная часть оставалась чистой.*/

type BattleTemplate struct {
	TemplateID            string //<-айдишник чертежа
	HealthPoints          int    //<-хп карты
	Attack                int    //<-сила атаки карты
	SplashRadius          int    //<-радиус сплеша (если 0-сплеша нет, если 1-бьет по одной цели справа и слева)
	Cooldown              int    //<-кд карты
	Manacost              int    //<-стоимость в мане
	IsTank                bool   //<-является ли танком
	CardType              string //<-тип карты
	CanBeUpgraded         bool   //<-может ли быть улучшенной
	ImageKey              string //<-картинка
	AssetBaseKey          string //<-набор ассетов
	SkillImageKey         string
	SkillName             string
	SkillCode             string
	SkillTrigger          string
	SkillTarget           string
	SkillValue            int
	SkillDuration         int
	SkillCooldown         int
	SkillParamsJSON           string
	PassiveImageKey       string
	PassiveName           string
	PassiveCode           string
	PassiveTrigger        string
	PassiveTarget         string
	PassiveEffect         string
	PassiveCondition      string
	PassiveValue          int
	PassiveDuration       int
	PassiveScale          string
	PassiveCountOwner     string
	PassiveConditionCount int
	PassiveCountType      string
	PassiveCountCode      string
}

// та же тема только с бафами
type BuffTemplate struct {
	TemplateID   string //<-понятно
	ManaCost     int    //<-тоже понятно
	BuffType     string //<-тип бафа (хп,атака,танк)
	BuffValue    int    //<-значение бафа
	OnlyFor      string //<-только для...
	Duration     int    //<-длительность бафа
	ImageKey     string //<-понятно
	AssetBaseKey string //<-заебись понятно
}
