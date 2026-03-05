package dto

import "TheWar/internal/domain/game"

/*Файл целиком и полностью посвящен DTO. Здесь я написал функцию ApplyActionReplace, которая отвечает
за применение действия. Как пример, клиент может отправить что-то типа : "type": "play_buff_card" и
далее по списку. В общем эта структура отвечает за прием намерений пользователя о том, что он вообще
хочет, над чем, что конкретно, версию и куда. */

type ApplyActionRequest struct {
	Type             game.ActionType `json:"type"`               //<-тип действия (Диспетчер действий...)
	CardInstanceID   string          `json:"card_instance_id"`   //<-индекс карты, которая будет что то делать
	TargetInstanceID string          `json:"target_instance_id"` //<-индекс цели
	AttackHero       bool            `json:"attack_hero"`        //<-инфа о том, будем пиздить персонажа или нет
	ExpectedVersion  int64           `json:"expected_version"`   //<-версия хода (Optimistic Lock)
	TargetSlot       int             `json:"target_slot"`        //<-слот, в который игрок будет класть карту
}
