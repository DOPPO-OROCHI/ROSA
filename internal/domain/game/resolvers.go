package game

type BattleTemplateResolver interface {
	GetBattleTemplate(templateID string) (BattleTemplate, bool)
}
type BattleMapResolver struct {
	M map[string]BattleTemplate
}

func (m BattleMapResolver) GetBattleTemplate(id string) (BattleTemplate, bool) {
	t, ok := m.M[id]
	return t, ok
}

type BuffTemplateResolver interface {
	GetBuffTemplate(templateID string) (BuffTemplate, bool)
}
type BuffMapResolver struct {
	M map[string]BuffTemplate
}

func (r BuffMapResolver) GetBuffTemplate(id string) (BuffTemplate, bool) {
	t, ok := r.M[id]
	return t, ok
}

type Resolvers struct {
	HeroAbility func(herocode string) (HeroAbility, bool)
	Battle      BattleTemplateResolver
	Buff        BuffTemplateResolver
}
