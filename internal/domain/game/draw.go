package game

import (
	"errors"
	"math/rand/v2"
	"time"
)

func CheckDraw(m *MatchState) bool {
	if m.Finished {
		return false
	}
	p1 := m.Players[0]
	p2 := m.Players[1]
	if p1 == nil || p2 == nil {
		return false
	}
	empty := func(p *PlayerState) bool {
		return len(p.Hand) == 0 && len(p.Deck) == 0 && TableEmpty(p)

	}
	return empty(p1) && empty(p2)
}

func ApplyDrawNeed(m *MatchState) bool {
	if CheckDraw(m) {
		m.Finished = true
		m.Result = MatchDraw
		return true
	}
	return false
}

func EnsureStartTurn(m *MatchState) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if m.Phase == PhaseStart {
		StartTurn(m, time.Now().Unix())
	}
	return nil
}

func ApplyAction(m *MatchState, a Action, r Resolvers) error {
	if m.Finished {
		return ErrMatchFinished
	}
	if a.ExpectedVersion != 0 && a.ExpectedVersion != m.Version {
		return ErrStaleAction
	}
	m.Events = m.Events[:0]
	if a.Type == ActionLeaveMatch {
		if err := LeaveMatch(m, a.PlayerIndex); err != nil {
			return err
		}
		m.Version++
		return nil
	}
	if err := EnsureStartTurn(m); err != nil {
		return err
	}
	now := time.Now().Unix()
	if m.Phase == PhaseMain && m.TurnDeadLineAt > 0 && now > m.TurnDeadLineAt {
		return ErrTurnTimeOut
	}
	if a.PlayerIndex != m.ActivePlayer {
		return ErrNotYourTurn
	}
	var err error
	switch a.Type {
	case ActionPlayBattle:
		if r.Battle == nil {
			return errors.New("battle resolver is nil")
		}
		err = PlayBattleCard(m, a.PlayerIndex, a.CardInstanceID, a.TargetSlot, r.Battle)
	case ActionPlayBuff:
		if r.Buff == nil {
			return errors.New("buff resolver is nil")
		}
		err = PlayBuffCard(m, a.PlayerIndex, a.CardInstanceID, a.TargetInstanceID, r.Buff)
	case ActionCardAttack:
		if r.Battle == nil {
			return errors.New("battle resolver is nil")
		}
		err = CardAttack(m, a.PlayerIndex, a.CardInstanceID, a.TargetInstanceID, a.AttackHero, r.Battle)
	case ActionHeroAttack:
		err = HeroAttack(m, a.PlayerIndex, a.TargetInstanceID, a.AttackHero)
	case ActionPlayHeroSpell:
		err = PlayHeroSpell(m, a, r)
	case ActionEndTurn:
		EndTurn(m)
	default:
		return errors.New("unknown action type: " + string(a.Type))
	}
	if err != nil {
		return err
	}
	m.Version++
	ApplyDrawNeed(m)
	return nil
}

func NewMatchState(matchID uint, p1 *PlayerState, p2 *PlayerState) *MatchState {
	firstPlayer := rand.IntN(2)
	m := &MatchState{
		MatchID:      matchID,
		Version:      1,
		Players:      [2]*PlayerState{p1, p2},
		ActivePlayer: firstPlayer,
		Phase:        PhaseStart,
		Finished:     false,
		Result:       MatchOnGoing,
	}
	for _, p := range m.Players {
		if p == nil {
			continue
		}
		p.Turns = 0
		p.Mana = 1
		p.Discard = nil
	}
	DrawCards(p1, 2)
	DrawCards(p2, 2)
	return m
}

func DrawCards(p *PlayerState, n int) {
	if p == nil || n <= 0 {
		return
	}
	if len(p.Deck) < n {
		n = len(p.Deck)
	}
	for i := 0; i < n; i++ {
		c := p.Deck[0]
		p.Deck = p.Deck[1:]
		p.Hand = append(p.Hand, c)
	}
}

func TableEmpty(p *PlayerState) bool {
	for i := 0; i < TableSize; i++ {
		if p.Table[i] != nil {
			return false
		}
	}
	return true
}
