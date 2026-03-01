package dto

import "TheWar/internal/domain/game"

type MaskedPlayerState struct {
	PlayerID               int    `json:"player_id"`
	UserID                 uint   `json:"user_id"`
	HeroID                 uint   `json:"hero_id"`
	HeroCode               string `json:"hero_code"`
	HeroHP                 int    `json:"hero_hp"`
	HeroLevel              int    `json:"hero_level"`
	HeroAttackPower        int    `json:"hero_attack_power"`
	HeroAttackCooldown     int    `json:"hero_attack_cooldown"`
	HeroAttackBaseCooldown int    `json:"hero_attack_base_cooldown"`
	HeroSplashRadius       int    `json:"hero_splash_radius"`
	HeroAbilityCooldown    int    `json:"hero_ability_cooldown"`

	Mana  int `json:"mana"`
	Turns int `json:"turns"`

	Table [game.TableSize]*game.UnitState `json:"table"`

	Hand      []game.CardsInMatch `json:"hand,omitempty"`
	Deck      []game.CardsInMatch `json:"deck,omitempty"`
	Discard   []game.CardsInMatch `json:"discard,omitempty"`
	HandCount int                 `json:"hand_count,omitempty"`
	DeckCount int                 `json:"deck_count,omitempty"`
	DiscCount int                 `json:"discard_count,omitempty"`
}
type MaskedMatchState struct {
	MatchID      uint                  `json:"match_id"`
	Version      int64                 `json:"version"`
	ActivePlayer int                   `json:"active_player"`
	Phase        game.TurnPhase        `json:"phase"`
	Finished     bool                  `json:"finished"`
	Result       game.MatchResult      `json:"result"`
	Players      [2]*MaskedPlayerState `json:"players"`
	Event        []game.Event          `json:"events,omitempty"`
}
