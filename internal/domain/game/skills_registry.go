package game

type SkillHandler func(m *MatchState, a Action, caster *UnitState) error

var SkillHandlers = map[string]SkillHandler{
	"fragmentation_grenades":   CastSplashDamageSkill,
	"expansive_projectiles":    CastBuffSkill,
	"burning":                  CastDebuffSkill,
	"urgent_reload":            CastBuffSkill,
	"suppression":              CastBuffSkill,
	"infection_with_pathogens": CastDebuffSkill,
	"second_wind":              CastSingleDamageSkill,
	"mathematical_accuracy":    CastDebuffSkill,
	"iron_will":                CastBuffSkill,
	"chain_induction":          CastBuffSkill,
	"energy_repeater":          CastBuffSkill,
	"materialized_energy":      CastAllEnemiesDamageSkill,
	"pressure":                 CastHighestHPDamageSkill,
	"liquidation":              CastKillTargetSkill,
	"shadow_strike":            CastDebuffSkill,
	"deep_wounds":              CastDebuffSkill,
	"fury":                     CastRandomMultiEnemyDamageSkill,
	"fearlessness":             CastHighestAttackDamageSkill,
	"continuous_battle":        CastBuffSkill,
	"headship":                 CastBuffSkill,
	"phosphorus_shells":        CastDebuffSkill,
	"group_work":               CastSummonSelfCopySkill,
	"missile_guidance":         CastRandomSingleEnemyDamageSkill,
	"replenishment":            CastHealSingleSkill,
}
