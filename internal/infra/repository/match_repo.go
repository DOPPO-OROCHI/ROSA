package repository

import (
	"TheWar/internal/domain/game"
	"TheWar/internal/domain/heroes"
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrActiveMatchExists = errors.New("active match already exists")

type Match struct {
	gorm.Model
	PlayerID1      uint           `gorm:"not null;index"`
	PlayerID2      uint           `gorm:"not null;index"`
	State          datatypes.JSON `gorm:"type:jsonb;not null"`
	Version        int64          `gorm:"not null;default:1"`
	Finished       bool           `gorm:"not null;default:false"`
	TurnDeadLineAt int64          `gorm:"not null;default:0;index"`
}

func SaveMatchState(tx *gorm.DB,
	matchID uint,
	expectedDBVersion int64,
	newStateJSON []byte,
	newVersion int64,
	turnDeadlineAt int64) error {
	res := tx.Model(&Match{}).Where("id = ? AND version = ?", matchID, expectedDBVersion).
		Updates(map[string]any{
			"state":            datatypes.JSON(newStateJSON),
			"version":          newVersion,
			"turn_deadline_at": turnDeadlineAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return game.ErrStaleAction
	}
	return nil
}

func CreateMatchTX(db *gorm.DB,
	p1UserID, p2UserID uint,
	p1HeroCode, p2HeroCode string) (*game.MatchState, error) {
	var out *game.MatchState
	err := db.Transaction(func(tx *gorm.DB) error {
		if p1UserID == 0 || p2UserID == 0 {
			return errors.New("bad user id")
		}
		if p1UserID == p2UserID {
			return errors.New("cannot create match with yourself")
		}
		var activeCount int64
		if err := tx.Model(&Match{}).
			Where("finished = false").
			Where("(player_id1 = ? OR player_id2 = ?) OR (player_id1 = ? OR player_id2 = ?)", p1UserID, p1UserID, p2UserID, p2UserID).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrActiveMatchExists
		}
		battleMax, buffMax, err := LoadTemplateLimits(tx)
		if err != nil {
			return err
		}
		p1Entires, err := LoadDeckTx(tx, p1UserID)
		if err != nil {
			return err
		}
		p2Entires, err := LoadDeckTx(tx, p2UserID)
		if err != nil {
			return err
		}
		if len(p1Entires) == 0 {
			return errors.New("p1 deck is empty")
		}
		if len(p2Entires) == 0 {
			return errors.New("p2 deck is empty")
		}
		p1BattleInfo, p1BattleCopies, err := LoadOwnedBattleCards(tx, p1UserID)
		if err != nil {
			return err
		}
		p1BuffInfo, p1BuffCopies, err := LoadOwnedBuff(tx, p1UserID)
		if err != nil {
			return err
		}
		p2BattleInfo, p2BattleCopies, err := LoadOwnedBattleCards(tx, p2UserID)
		if err != nil {
			return err
		}
		p2BuffInfo, p2BuffCopies, err := LoadOwnedBuff(tx, p2UserID)
		if err != nil {
			return err
		}
		if err := game.ValidateDeckList(p1Entires,
			battleMax, buffMax, p1BattleCopies,
			p1BuffCopies); err != nil {
			return err
		}
		if err := game.ValidateDeckList(p2Entires,
			battleMax, buffMax, p2BattleCopies,
			p2BuffCopies); err != nil {
			return err
		}
		if p1HeroCode == "" {
			code, err := LoadSelectedHeroCodeTx(tx, p1UserID)
			if err != nil {
				return err
			}
			p1HeroCode = code
		}
		if p2HeroCode == "" {
			code, err := LoadSelectedHeroCodeTx(tx, p2UserID)
			if err != nil {
				return err
			}
			p2HeroCode = code
		}
		loadHero := func(userID uint, heroCode string) (heroes.CharacterTemplate, int, error) {
			var tpl heroes.CharacterTemplate
			if err := tx.Where("character_code = ?", heroCode).First(&tpl).Error; err != nil {
				return heroes.CharacterTemplate{}, 0, err
			}
			var g GamerCharacter
			if err := tx.Where("gamer_id = ? AND character_template_id = ?", userID, tpl.ID).First(&g).Error; err != nil {
				return heroes.CharacterTemplate{}, 0, err
			}
			return tpl, g.CharacterLevel, nil
		}
		p1HeroTpl, p1HeroLevel, err := loadHero(p1UserID, p1HeroCode)
		if err != nil {
			return err
		}
		p2HeroTpl, p2HeroLevel, err := loadHero(p2UserID, p2HeroCode)
		if err != nil {
			return err
		}
		p1Deck, err := game.BuildDeck(p1Entires, p1BattleInfo, p1BuffInfo)
		if err != nil {
			return err
		}
		p2Deck, err := game.BuildDeck(p2Entires, p2BattleInfo, p2BuffInfo)
		if err != nil {
			return err
		}
		row := Match{
			PlayerID1: p1UserID,
			PlayerID2: p2UserID,
			Version:   1,
			State:     datatypes.JSON([]byte(`{}`)),
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		p1 := game.PlayerState{
			PlayerID:                0,
			UserID:                  p1UserID,
			HeroID:                  p1HeroTpl.ID,
			HeroCode:                p1HeroTpl.CharacterCode,
			HeroLevel:               p1HeroLevel,
			HeroHP:                  p1HeroTpl.HealthPoints,
			HeroAttackPower:         p1HeroTpl.AttackPower,
			HeroAttackCooldown:      0,
			HeroAttackBaseCooldown:  p1HeroTpl.AttackCooldown,
			HeroSplashRadius:        p1HeroTpl.SplashRadius,
			HeroAbilityCooldown:     0,
			HeroAbilityBaseCooldown: p1HeroTpl.Ability.CoolDown,
			HeroAbilityManaCost:     p1HeroTpl.Ability.ManaCost,
			Deck:                    p1Deck,
		}
		p2 := game.PlayerState{
			PlayerID:                1,
			UserID:                  p2UserID,
			HeroID:                  p2HeroTpl.ID,
			HeroCode:                p2HeroTpl.CharacterCode,
			HeroLevel:               p2HeroLevel,
			HeroHP:                  p2HeroTpl.HealthPoints,
			HeroAttackPower:         p2HeroTpl.AttackPower,
			HeroAttackCooldown:      0,
			HeroAttackBaseCooldown:  p2HeroTpl.AttackCooldown,
			HeroSplashRadius:        p2HeroTpl.SplashRadius,
			HeroAbilityCooldown:     0,
			HeroAbilityBaseCooldown: p2HeroTpl.Ability.CoolDown,
			HeroAbilityManaCost:     p2HeroTpl.Ability.ManaCost,
			Deck:                    p2Deck,
		}
		st := game.NewMatchState(row.ID, &p1, &p2)
		b, err := json.Marshal(st)
		if err != nil {
			return err
		}
		if err := tx.Model(&Match{}).Where("id = ?", row.ID).Updates(map[string]any{
			"state":   datatypes.JSON(b),
			"version": st.Version,
		}).Error; err != nil {
			return err
		}
		out = st
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
