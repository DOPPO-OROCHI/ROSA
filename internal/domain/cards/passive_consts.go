package cards

const (
	PassiveKindAura     = "aura"     //<-постоянный эффект карты
	PassiveKindReactive = "reactive" //<-типа реакция на какое либо событие, чтобы пассивка активировалась
)

const (
	PassiveTriggerWhileAlive       = "while_alive"   //<-пока карта-источник жива
	PassiveTriggerOnEnter          = "on_enter"      //<-пасствка активируется в тот момент, когда карта зашла на стол
	PassiveTriggerOnLeave          = "on_leave"      //<-пассивка активируется тогда, когда карта умирает
	PassiveTriggerTurnStart        = "turn_start"    //<-пассивка активируется при старте хода
	PassiveTriggerTurnEnd          = "turn_end"      //<-пассивка активируется при конце хода
	PassiveTriggerOnAllyPlay       = "on_ally_play"  //<-пассивка активируется, когда союзная карта разыгрывается на стол
	PassiveTriggerOnEnemyPlay      = "on_enemy_play" //<-пассивка активируется тогда, когда противник разыгрывает карту
	PassiveTriggerOnAllySkill      = "on_ally_skill"
	PassiveTriggerOnEnemySkill     = "on_enemy_skill"
	PassiveTriggerOnEnemyDeath     = "on_enemy_death"
	PassiveTriggerOnAllyDeath      = "on_ally_death"
	PassiveTriggerOnEnemyHeroSkill = "on_enemy_hero_skill"
	PassiveTriggerOnHeroSkill      = "on_hero_skill"
	PassiveTriggerOnDamaged        = "on_damaged"
	PassiveTriggerOnAttack         = "on_attack"
)

const (
	PassiveConditionNone           = ""                 //<-не требует особых условий для активации
	PassiveConditionOwnerHasTank   = "owner_has_tank"   //<-активируется тогда, когда у игрока есть танк на столе
	PassiveConditionEnemyHasTank   = "enemy_has_tank"   //<-тогда, когда у противника есть танк
	PassiveConditionOwnerRaceCount = "owner_race_count" //<-тогда, когда у игрока Х определенной расы
	PassiveConditionEnemyRaceCount = "enemy_race_count" //<-у противника*
	PassiveConditionOwnerAllRace   = "owner_all_race"   //<-тогда, когда на столе только определенная раса
	PassiveConditionEnemyAllRace   = "enemy_all_race"   //<-тогда, когда на столе противника только определенная раса
)

const (
	PassiveLayerNone       = ""
	EffectLayerSkill       = "skill"
	EffectLayerHeroAbility = "hero_ability"
	EffectLayerPassive     = "passive"
)

const (
	PassiveEventFilterNone       = ""
	PassiveEventFilterCardPlayed = "card_played"
	PassiveEventFilterSkillUsed  = "skill_used"
)

const (
	PassiveScaleNone              = ""
	PassiveScalePerEnemyUnit      = "per_enemy_unit"
	PassiveScalePerAllyUnit       = "per_ally_unit"
	PassiveScalePerOwnerRaceCount = "per_owner_race_count"
	PassiveScalePerEnemyRaceCount = "per_enemy_race_count"
	PassiveScalePerEnemyTank      = "per_enemy_tank"
	PassiveScalePerOwnerTank      = "per_owner_tank"
)

const (
	PassiveEffectBuff          = "buff"
	PassiveEffectDebuff        = "debuff"
	PassiveEffectDamage        = "damage"
	PassiveEffectHeal          = "heal"
	PassiveEffectCounterattack = "counterattack"
)
