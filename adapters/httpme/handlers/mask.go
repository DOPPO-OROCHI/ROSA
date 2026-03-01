package handlers

import (
	"TheWar/adapters/httpme/dto"
	"TheWar/internal/domain/game"
)

func maskMatchStateForUser(st *game.MatchState, viewerUserID uint) *dto.MaskedMatchState {
	if st == nil {
		return nil
	}
	viewerIndex := -1
	if st.Players[0] != nil && st.Players[0].UserID == viewerUserID {
		viewerIndex = 0
	}
	if st.Players[1] != nil && st.Players[1].UserID == viewerUserID {
		viewerIndex = 1
	}
	out := dto.MaskedMatchState{
		MatchID:      st.MatchID,
		Version:      st.Version,
		ActivePlayer: st.ActivePlayer,
		Phase:        st.Phase,
		Finished:     st.Finished,
		Result:       st.Result,
		Event:        append([]game.Event(nil), st.Events...),
	}
	common := func(p *game.PlayerState) *dto.MaskedPlayerState {
		if p == nil {
			return nil
		}
		mp := &dto.MaskedPlayerState{
			PlayerID:               p.PlayerID,
			UserID:                 p.UserID,
			HeroID:                 p.HeroID,
			HeroCode:               p.HeroCode,
			HeroHP:                 p.HeroHP,
			HeroLevel:              p.HeroLevel,
			HeroAttackPower:        p.HeroAttackPower,
			HeroAttackCooldown:     p.HeroAttackCooldown,
			HeroAttackBaseCooldown: p.HeroAttackBaseCooldown,
			HeroSplashRadius:       p.HeroSplashRadius,
			HeroAbilityCooldown:    p.HeroAbilityCooldown,
			Mana:                   p.Mana,
			Turns:                  p.Turns,
			Table:                  p.Table,
		}
		return mp
	}
	for i := 0; i < 2; i++ {
		p := st.Players[i]
		mp := common(p)
		if mp == nil {
			out.Players[i] = nil
			continue
		}
		if i == viewerIndex {
			mp.Hand = append(mp.Hand, p.Hand...)
			mp.Deck = append(mp.Deck, p.Deck...)
			mp.Discard = append(mp.Discard, p.Discard...)
		} else {
			mp.HandCount = len(p.Hand)
			mp.DeckCount = len(p.Deck)
			mp.DiscCount = len(p.Discard)
		}
		out.Players[i] = mp
	}
	return &out
}
