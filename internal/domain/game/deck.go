package game

import (
	"errors"
	"math/rand/v2"

	"github.com/google/uuid"
)

type DeckEntry struct {
	Kind       string `json:"kind"`
	TemplateID string `json:"template_id"`
	Count      int    `json:"count"`
}

type OwnedCardInfo struct {
	GamerCardID uint
	Copies      int
	Level       int
}

func ValidateDeckList(
	entries []DeckEntry,
	battleMax map[string]int,
	buffMax map[string]int,
	ownedBattle map[string]int,
	ownedBuff map[string]int,
) error {
	total := 0
	for _, e := range entries {
		if e.Count <= 0 {
			return ErrDeckCountInvalid
		}
		total += e.Count
		switch e.Kind {
		case "battle":
			max, ok := battleMax[e.TemplateID]
			if !ok {
				return ErrDeckUnknownCard
			}
			if e.Count > max {
				return ErrDeckTooManyCopies
			}
			if ownedBattle[e.TemplateID] < e.Count {
				return ErrDeckNotOwnedEnough
			}
		case "buff":
			max, ok := buffMax[e.TemplateID]
			if !ok {
				return ErrDeckUnknownCard
			}
			if e.Count > max {
				return ErrDeckTooManyCopies
			}
			if ownedBuff[e.TemplateID] < e.Count {
				return ErrDeckNotOwnedEnough
			}
		default:
			return ErrDeckUnknownKind
		}
	}
	if total != 20 {
		return ErrDeckSizeNot20
	}
	return nil
}

func BuildDeck(
	entires []DeckEntry,
	battle map[string]OwnedCardInfo,
	buff map[string]OwnedCardInfo,
) ([]CardsInMatch, error) {
	deck := make([]CardsInMatch, 0, 20)
	for _, e := range entires {
		var inf OwnedCardInfo
		var ok bool
		switch e.Kind {
		case "battle":
			inf, ok = battle[e.TemplateID]
		case "buff":
			inf, ok = buff[e.TemplateID]
		default:
			return nil, ErrDeckUnknownKind
		}
		if !ok {
			return nil, ErrDeckUnknownCard
		}
		for i := 0; i < e.Count; i++ {
			deck = append(deck, CardsInMatch{
				InstanceID:  NewInstanceID(),
				Kind:        e.Kind,
				TemplateID:  e.TemplateID,
				GamerCardID: inf.GamerCardID,
				CardLevel:   inf.Level,
			})
		}
	}
	if len(deck) != 20 {
		return nil, errors.New("internal: buit deck is not 20")
	}
	shuffleDeck(deck)
	return deck, nil
}

func BuildDeckInMatch(entries []DeckEntry) []CardsInMatch {
	deck := make([]CardsInMatch, 0, 20)
	for _, e := range entries {
		for i := 0; i < e.Count; i++ {
			deck = append(deck, CardsInMatch{
				InstanceID:  NewInstanceID(),
				Kind:        e.Kind,
				TemplateID:  e.TemplateID,
				GamerCardID: 0,
			})
		}
	}
	shuffleDeck(deck)
	return deck
}

// перемешиваем карты
func shuffleDeck(deck []CardsInMatch) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func NewInstanceID() string {
	return uuid.NewString()
}
