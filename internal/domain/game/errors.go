package game

import "errors"

var (
	ErrMatchFinished    = errors.New("match finiched")
	ErrNotYourTurn      = errors.New("not your turn")
	ErrWrongPhase       = errors.New("wrong phase")
	ErrHandCardNotFound = errors.New("card not found in hand")
	ErrNotEnoughMana    = errors.New("not enough mana")
	ErrTablesFull       = errors.New("table is full")
	ErrSlotOccupied     = errors.New("slot occupied")

	ErrBuffCardNotFound = errors.New("buff card not found")
	ErrTargetNotFound   = errors.New("target unit not found on table")
	ErrWrongTargetType  = errors.New("buff cannot be applied to this card type")

	ErrAttackerNotFound          = errors.New("attacker not found on table")
	ErrDefenderNotFound          = errors.New("defender not found")
	ErrAttackerOnCooldown        = errors.New("attacker on cooldown")
	ErrAttackerSummoneddThisTurn = errors.New("attacker cannot atack on summon turn")
	ErrMustAttackTank            = errors.New("must attack tank while tank exist")
	ErrCannotAttackHeroWithTanks = errors.New("cannot attack hero while tank exist")
	ErrHealerCannotAttack        = errors.New("healer cannot attack")
	ErrHealerCannotAttackHero    = errors.New("healer cannot attack hero")
	ErrHealerCannotHealEnemy     = errors.New("healer cannot heal enemy units")

	ErrHeroOnCooldown          = errors.New("hero is on cooldawn")
	ErrUnknownHeroTemplate     = errors.New("unknown hero template")
	ErrHeroAttackIsZero        = errors.New("hero attack is zero")
	ErrCannotHitHeroWhileTanks = errors.New("cannot hit hero while tanks")

	ErrStaleAction = errors.New("stale action : state version mismatch")

	ErrDeckSizeNot20      = errors.New("deck size must be 20")
	ErrDeckCountInvalid   = errors.New("deck entry count invalid")
	ErrDeckTooManyCopies  = errors.New("too many copies of a card in deck")
	ErrDeckNotOwnedEnough = errors.New("not enough copies owned by gamer")
	ErrDeckUnknownCard    = errors.New("unknown card template")
	ErrDeckUnknownKind    = errors.New("deck unknown kind")

	ErrHeroAbilityOnCooldown       = errors.New("hero ability is on cooldown")
	ErrHeroAbilityUnknown          = errors.New("unknown hero ability")
	ErrHeroAbilityBadTarget        = errors.New("bad hero ability target")
	ErrHeroAbilityCannotAttackHero = errors.New("hero ability cannot attack hero")

	ErrTurnTimeOut = errors.New("turn time out")
)
