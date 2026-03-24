package dto

/*Файл посвящен DTO для работы с героями. Здесь происходит ровно та же история что и с картами,
и деками, только здесь уже для героев во владении игрока. Здесь вся та же операционка, поэтому
не вижу смысла подробно описывать каждую мелочь*/

type OwnedHeroDTO struct {
	HeroID         uint   `json:"hero_id"`
	HeroCode       string `json:"hero_code"`
	Name           string `json:"name"`
	Level          int    `json:"level"`
	HealthPoints   int    `json:"health_points"`
	AttackPower    int    `json:"attack_power"`
	AttackCooldown int    `json:"attack_cooldown"`
	SplashRadius   int    `json:"splash_radius"`
	Description    string `json:"description"`
	ImageKey       string `json:"image_key"`
	AssetBaseKey   string `json:"asset_base_key"`
	SkillImageKey  string `json:"skill_image_key"`
	AttackImageKey string `json:"attack_image_key"`
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
