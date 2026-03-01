package dto

import "TheWar/internal/domain/game"

type SaveDeckRequest struct {
	Entries []game.DeckEntry `json:"entries"`
}
type DeckResponce struct {
	Entries []game.DeckEntry `json:"entries"`
}
type DeckEntryDTO struct {
	Kind       string `json:"kind"`
	TemplateID string `json:"template_id"`
	Count      int    `json:"count"`
}
