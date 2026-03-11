package dto

/*Ну тут и объяснять не нужно я думаю. Здесь все просто. Когда ты в доте нажимаешь
на свою аватарку, ты получаешь инфу о себе. Мысленно представь что это то-же самое.
Структура целиком и полностью отдает инфу о пользователе, которую мы берем к слову
из БД. Это пользователь, привет.*/

type MeResponse struct {
	UserID               uint   `json:"user_id"`
	TGID                 int64  `json:"tg_id"`
	Username             string `json:"username"`
	FirstName            string `json:"first_name"`
	Rating               int    `json:"rating"`
	XP                   int    `json:"xp"`
	SelectedHeroTemplate *uint  `json:"selected_hero_template_id"`
	SelectedHeroCode     string `json:"selected_hero_code,omitempty"`
	SelectedHeroName     string `json:"selected_hero_name,omitempty"`
}
