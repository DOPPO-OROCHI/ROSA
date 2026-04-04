package game

type SkillHandler func(m *MatchState, a Action, caster *UnitState) error

var SkillHandlers = map[string]SkillHandler{
	"fragmentation_grenades": CastSplashDamageSkill,
}
