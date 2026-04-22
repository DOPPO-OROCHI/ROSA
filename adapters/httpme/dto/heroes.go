package dto

/*Файл посвящен DTO для работы с героями. Здесь происходит ровно та же история что и с картами,
и деками, только здесь уже для героев во владении игрока. Здесь вся та же операционка, поэтому
не вижу смысла подробно описывать каждую мелочь*/

type OwnedHeroDTO struct {
	HeroID         uint           `json:"hero_id"`
	HeroCode       string         `json:"hero_code"`
	Name           string         `json:"name"`
	Level          int            `json:"level"`
	HealthPoints   int            `json:"health_points"`
	Ability        HeroAbilityDTO `json:"ability"`
	AttackPower    int            `json:"attack_power"`
	AttackCooldown int            `json:"attack_cooldown"`
	SplashRadius   int            `json:"splash_radius"`
	Description    string         `json:"description"`
	ImageKey       string         `json:"image_key"`
	AssetBaseKey   string         `json:"asset_base_key"`
	SkillImageKey  string         `json:"skill_image_key"`
	AttackImageKey string         `json:"attack_image_key"`
}
type HeroAbilityDTO struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	Description  string `json:"description"`
	Kind         string `json:"kind"`
	Target       string `json:"target"`
	CoolDown     int    `json:"cool_down"`
	ManaCost     int    `json:"mana_cost"`
	Power        int    `json:"power"`
	Duration     int    `json:"duration"`
	ExtraValue   int    `json:"extra_value"`
	ApplyCount   int    `json:"apply_count"`
	BuffEffect   string `json:"buff_effect"`
	DebuffEffect string `json:"debuff_effect"`
	IgnoreTank   bool   `json:"ignore_tank"`
}

/*Но кое что объяснить все таки важно, как в целях обучения, так и в целях объяснения механик. Зачем
нам столько структур ? А затем, что они все отвечают на разные запросы и ответы. К примеру эта структура
отвечает за запрос на выбор героя игрока. То есть чувак выбирает персонажа, это понятно, но как он это
делает ? Вот вот... Отправляет JSON, который я потом в коде сериализую в эту структуру, а потом эта инфа
отправляется на сервер, где уже и проходит конкретная логика выбора персонажа*/
type SelectHeroRequest struct {
	HeroCode string `json:"hero_code"`
}

/*А эта структура нужна для того, чтобы отдавать список героев игрока, которые он вообще может выбрать*/
type HeroesListResponce struct {
	Heroes []OwnedHeroDTO `json:"heroes"`
}
