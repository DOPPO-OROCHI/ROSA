package dto

import "TheWar/internal/domain/game"

type ApplyActionReplace struct {
	Type             game.ActionType `json:"type"`               //<-тип действия (Диспетчер действий...)
	CardInstanceID   string          `json:"card_instance_id"`   //<-индекс карты, которая будет что то делать
	TargetInstanceID string          `json:"target_instance_id"` //<-индекс цели
	AttackHero       bool            `json:"attack_hero"`        //<-инфа о том, будем пиздить персонажа или нет
	ExpectedVersion  int64           `json:"expected_version"`   //<-версия хода (Optimistic Lock)
	TargetSlot       int             `json:"target_slot"`        //<-слот, в который игрок будет класть карту
}
