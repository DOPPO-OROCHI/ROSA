package dto

type MeResponse struct {
	UserID              uint   `json:"user_id"`
	TGID                int    `json:"tg_id"`
	Username            string `json:"username"`
	FirstName           string `json:"first_name"`
	Rating              int    `json:"rating"`
	XP                  int    `json:"xp"`
	SelectedHeroRequest *uint  `json:"selected_hero_template_id"`
	SelectedHeroCode    string `json:"selected_hero_code,omitempty"`
	SelectedHeroName    string `json:"selected_hero_name,omitempty"`
}
