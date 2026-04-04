package game

/*Файл с содержанием карточных чертежей. Первый для боевых карт (Battle), второй для баф карт (Buff).
Особенностью данных структур является то, что они нужны непосредственно домену. Почему ? Смотрите в чем
прикол дорогие мои... У нас уже есть модели карт, да (cards/models). Даааа, но как бы то слой БД шаришь ?
А вот эти две структуры нужны для того, чтобы мапить в них те структуры, которые живут в БД. Живите с этим.
Подробнее об этом будем тереть в резолверах, а пока закрепим то, что данные структуры онли рантайм
компонент из DTO, в который мы мапим нужные для двигла поля, чтобы доменная часть оставалась чистой.*/

type BattleTemplate struct {
	TemplateID    string
	Name          string
	Description   string
	HealthPoints  int
	Attack        int
	SplashRadius  int
	BaseCooldown  int
	ManaCost      int
	IsTank        bool
	CardType      string
	ImageKey      string
	AssetBaseKey  string
	SkillImageKey string
	HasSkill      bool
	Skill         BattleSkillTemplate
}

type BattleSkillTemplate struct {
	Code         string
	Kind         string
	Target       string
	Power        int
	BaseCooldown int
	Duration     int
	ExtraValue   int
	IgnoreTank   bool
	HitCount     int
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
