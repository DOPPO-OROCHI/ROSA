package dto

type CardKind string

const (
	CardKindBattle CardKind = "battle"
	CardKindBuff   CardKind = "buff"
)

type CardsListResponse struct {
	Battle []OwnedBattleCardsDTO `json:"battle"`
	Buff   []OwnedBuffCardsDTO   `json:"buff"`
}
type OwnedBattleCardsDTO struct {
	Kind         CardKind `json:"kind"`
	TemplateID   string   `json:"template_id"`
	Name         string   `json:"name"`
	CardType     string   `json:"card_type"`
	ManaCost     int      `json:"mana_cost"`
	HealthPoints int      `json:"health_points"`
	Attack       int      `json:"attack"`
	SplashRadius int      `json:"splash_radius"`
	Cooldown     int      `json:"cooldown"`
	IsTank       bool     `json:"is_tank"`
	BuffSlot     bool     `json:"buff_slot"`
	MaxCopies    int      `json:"max_copies"`

	OwnedCardID uint `json:"owned_card_id"`
	Copies      int  `json:"copies"`
	Level       int  `json:"level"`
	XP          int  `json:"xp"`

	ImageKey     string `json:"image_key"`
	AssetBaseKey string `json:"asset_base_key"`
}
type OwnedBuffCardsDTO struct {
	Kind       CardKind `json:"kind"`
	TemplateID string   `json:"template_id"`
	Name       string   `json:"name"`
	ManaCost   int      `json:"mana_cost"`
	BuffType   string   `json:"buff_type"`
	BuffValue  int      `json:"buff_value"`
	OnlyFor    string   `json:"only_for"`
	Duration   int      `json:"duration"`
	MaxCopies  int      `json:"max_copies"`

	OwnedCardID  uint   `json:"owned_card_id"`
	Copies       int    `json:"copies"`
	Level        int    `json:"level"`
	XP           int    `json:"xp"`
	ImageKey     string `json:"image_key"`
	AssetBaseKey string `json:"asset_base_key"`
}
