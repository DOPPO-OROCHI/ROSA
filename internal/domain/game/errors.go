package game

import "errors"

/*А здесь ничего такого, за исключением констант ошибок, которые так или иначе могут возникнуть в ходе
выполнения функций. Так же эти ошибки обрабатываются в handlers (где я круто написал функцию, которая
мапит доменные ошибки в HTTP). Ничего сложного*/

var (
	ErrMatchFinished = errors.New("match finiched") //<-матч закончен
	ErrNotYourTurn   = errors.New("not your turn")  //<-чужой ход
	ErrWrongPhase    = errors.New("wrong phase")    //<-неверная фаза хода (если чел ну прям слишком
	// спешит походить, хотя сценарий близок к нереальному)
	ErrHandCardNotFound = errors.New("card not found in hand") //<-карты нет в матче
	ErrNotEnoughMana    = errors.New("not enough mana")        //<-не хватает маны
	ErrTablesFull       = errors.New("table is full")          //<-стол полон (не вариант поставить карту)
	ErrSlotOccupied     = errors.New("slot occupied")          //<-слот на столе занят

	ErrBuffCardNotFound = errors.New("buff card not found")                      //<-баф карты не существует в матче
	ErrTargetNotFound   = errors.New("target unit not found on table")           //<-цель не найдена
	ErrWrongTargetType  = errors.New("buff cannot be applied to this card type") //<-неверный таргет (как если чел пытается ударить
	//кого угодно кроме танка, пока тот есть на столе)

	ErrAttackerNotFound = errors.New("attacker not found on table") //<-атакующая карта не найдена (по большей части
	//защита от читаков, как)
	ErrDefenderNotFound          = errors.New("defender not found")                   //<-принимающей урон карты нет (ударить некого)
	ErrAttackerOnCooldown        = errors.New("attacker on cooldown")                 //<-атакующая единица на кд
	ErrAttackerSummoneddThisTurn = errors.New("attacker cannot atack on summon turn") //<-нельзя атаковать на момент размещения на столе
	ErrMustAttackTank            = errors.New("must attack tank while tank exist")    //<-необходимо атаковать именно танка (если таковой на столе)
	ErrCannotAttackHeroWithTanks = errors.New("cannot attack hero while tank exist")  //<-нельзя атаковать героя, пока на столе танк
	ErrHealerCannotAttack        = errors.New("healer cannot attack")                 //<-хил юнит не может атаковать
	ErrHealerCannotAttackHero    = errors.New("healer cannot attack hero")            //<-хил юнит не может атаковать персонажа
	ErrHealerCannotHealEnemy     = errors.New("healer cannot heal enemy units")       //<-хил карта не может хилить противника

	ErrHeroOnCooldown          = errors.New("hero is on cooldawn")         //<-герой на кд (как если бы мы захотели ударить рукой)
	ErrUnknownHeroTemplate     = errors.New("unknown hero template")       //<-непонятный перс в матче (защита от читаков)
	ErrHeroAttackIsZero        = errors.New("hero attack is zero")         //<-герой не может атаковать рукой
	ErrCannotHitHeroWhileTanks = errors.New("cannot hit hero while tanks") //<-нельзя атаковать пока на столе танк

	ErrStaleAction = errors.New("stale action : state version mismatch") //<-ошибка версионирования

	ErrDeckSizeNot20      = errors.New("deck size must be 20")              //<-дека не может быть не равна 20
	ErrDeckCountInvalid   = errors.New("deck entry count invalid")          //<-размер деки неправильный
	ErrDeckTooManyCopies  = errors.New("too many copies of a card in deck") //<-слишком много копий одной карты в деке
	ErrDeckNotOwnedEnough = errors.New("not enough copies owned by gamer")  //<-карта из деки не принадлежит игроку
	ErrDeckUnknownCard    = errors.New("unknown card template")             //<-неизвестная карта в деке
	ErrDeckUnknownKind    = errors.New("deck unknown kind")                 //<-неверный тип карты в деке

	ErrHeroAbilityOnCooldown       = errors.New("hero ability is on cooldown")     //<-способность героя на кд
	ErrHeroAbilityUnknown          = errors.New("unknown hero ability")            //<-неизвестная способность героя
	ErrHeroAbilityBadTarget        = errors.New("bad hero ability target")         //<-неверный таргет геройской способности
	ErrHeroAbilityCannotAttackHero = errors.New("hero ability cannot attack hero") //<-геройская способность не может атаковать героя

	ErrTurnTimeOut = errors.New("turn time out") //<-время на ход закончилось

	ErrCardSkillNotFound           = errors.New("card skill not found")
	ErrCardSkillNotActive          = errors.New("card skill is not active")
	ErrCardSkillOnCooldown         = errors.New("card skill is on cooldown")
	ErrCardSkillBadTarget          = errors.New("bad card skill target")
	ErrCardSkillUnsupported        = errors.New("unsupported card skill")
	ErrCardPassiveSkillUnsupported = errors.New("card passove skill unsupported")
	ErrCardSkillTargetTankBlocked  = errors.New("cannot use this card skill while tank exists")
)
