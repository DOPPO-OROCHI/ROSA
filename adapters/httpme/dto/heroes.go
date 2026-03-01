package dto

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
}
type SelectedHeroRequest struct {
	HeroCode string `json:"hero_code"`
}
type HeroesListResponce struct {
	Heroes []OwnedHeroDTO `json:"heroes"`
}
