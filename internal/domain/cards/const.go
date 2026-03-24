package cards

/*Константы, служащие для описания типов карт и эффектов от бафа. К примеру, есть карта
MachanicalGuardian. Ее тип MechanicalCard соответстввенно, а значит, что для такого типа
карты сработают только определенные бафы, поскольку те тоже не для всех (хотя я сейчас так
думаю, что нахер это и не надо, но пока давай сделаем так, чтобы моя игра была типа сложной
и продуманной...). Так вот, здесь да, чисто константы чтобы использовать их при описании
карт. Ато мало ли ошибусь.*/

const (
	//типы карт
	MechanicalCard = "mechanical"
	OrganicalCard  = "organical"
	DemonicalCard  = "demonical"
	HealerCard     = "healer"
	All            = "all"

	//улучшения для карт, которые можно засунуть в матче
	CoolDownUpdate     = "cooldown_update"
	HealthPointsUpdate = "health_points_update"
	DamageUpdate       = "damage_update"
	MakeTankUpdate     = "make_tank"
)

// ГЛОБАЛЬНЫЙ ПАТЧ
// активные умения карт, либо активный скилл тогда, когда что то произойдет
// триггеры карт. Типа когда срабатывают скилы ?
const (
	TriggerOnPlay    = "on_play"   //срабатывает скил при поставлении карты на стол
	TriggerOnAttack  = "on_attack" //срабатывает, когда карта совершает атаку
	TriggerOnDeath   = "on_death"  //когда карта умирает
	TriggerTurnStart = "on_start"  //в начале хода игрока
	TriggerActive    = "active"    //активный скилл карты
)

// на кого срабатывают карты ?
const (
	TargetNone            = "none"         //цели не требуется, например когда карта умирает она бьет всех
	TargetSelf            = "self"         //на себя
	TargetAllyUnit        = "ally_unit"    //на союзную карту
	TargetEnemyUnit       = "enemy_unit"   //на вражескую карту
	TargetAllyAll         = "ally_all"     //весь союзный стол
	TargetEnemyAll        = "enemy_all"    //весь вражеский стол
	TargetBothAll         = "both_all"     //вообще весь стол
	TargetEnemySplash     = "enemy_splash" //сплеш по противникам
	TargetAllySplash      = "ally_splash"  //сплеш по своим
	TargetAllyGraveSingle = "ally_grave_single"
)

// коды скиллов карт
const (
	SkillDamageSingle             = "damage_single"               //бьем точечно в одну карту done
	SkillDamageSplash             = "damage_splash"               //бем по сплешу done
	SkillApplyDebuff              = "apply_debuff"                //накладываем дебаф при атаке, или просто по использованию
	SkillApplyBuff                = "apply_buff"                  //накладываем бафф
	SkillSummonSelfCopy           = "summon_self_copy"            //призывает копии себя же
	SkillBanishUnit               = "banish_unit"                 //убирает карту со стола
	SkillRevealEnemyHand          = "reveal_enemy_hand"           //смотрим в руку противника
	SkillIncEnemyCdAllOnDeath     = "inc_enemy_cd_all_on_death"   //после смерти карты увеличиваем кд всего вражеского стола
	SkillIncEnemyCdSingle         = "inc_enemy_cd_single"         //точечно увеличиваем кд вражеской карты
	SkillDecAllyCdSingle          = "dec_ally_cd_single"          //снимаем/уменьшаем кд союзной карты
	SkillResurrectTargetFromGrave = "resurrect_target_from_grave" //воскрешает карту
	SkillResurrectNextTurn        = "resurrect_next_turn"         //карта воскресает на следующий ход после смерти
	SkillDeathAoe                 = "death_aoe"                   //при уничтожении бьет по столу
)

// Для JSON скиллов карт
const (
	IgnoreTankTrue = "ignore_tank=true"
)

// Негативные эффекты на скилы
const (
	DotHPUpdate       = "dot_hp_update"
	DotAttackUpdate   = "dot_attack_update"
	DotCooldownUpdate = "dot_cooldown_update"
)

// Пассивные умения карт
// Триггеры. Когда срабатывают пассивки ?
const (
	PassiveTriggerContinious  = "continious"
	PassiveTriggerTurnStart   = "turn_start"
	PassiveTriggerTurnEnd     = "turn_end"
	PassiveTriggerOnPlay      = "on_play"
	PassiveTriggerOnAttack    = "on_attack"
	PassiveTriggerOnDeath     = "on_death"
	PassiveTriggerOnAllyDead  = "on_ally_dead"
	PassiveTriggerOnEnemyDead = "on_enemy_dead"
)

// Таргеты. Кто получает эффекты пассивки ?
const (
	PassiveTargetSelf          = "self"
	PassiveTargetAllyAll       = "ally_all"
	PassiveTargetEnemyAll      = "enemy_all"
	PassiveTargetBothAll       = "both_all"
	PassiveTargetAllyLeftRight = "ally_left_right"

	PassiveTargetAllyTypeDemonical  = "ally_type_demonical"
	PassiveTargetAllyTypeMechanical = "ally_type_mechanical"
	PassiveTargetAllyTypeOrganical  = "ally_type_organical"
	PassiveTargetAllyTypeHealer     = "ally_type_healer"

	PassiveTargetEnemyTypeDemonical  = "enemy_type_demonical"
	PassiveTargetEnemyTypeMechanical = "enemy_type_mechanical"
	PassiveTargetEnemyTypeOrganical  = "enemy_type_organical"
	PassiveTargetEnemyTypeHealer     = "enemy_type_healer"
)

// Эффект от пассивок. Что произойдет ?
const (
	PassiveEffectDamageUp          = "damage_up"
	PassiveEffectHPUp              = "hp_up"
	PassiveEffectCoolDownDown      = "coolwodn_down"
	PassiveEffectSkillDamageUp     = "skill_damage_up"
	PassiveEffectSkillCooldownDown = "skill_cooldown_down"
)

// Когда действует пассивка ?
const (
	PassiveConditionAlways            = "always"
	PassiveConditionDemonicalOnTable  = "demonical_on_table"
	PassiveConditionOrganicalOnTable  = "organical_on_table"
	PassiveConditionMechanicalOnTable = "mechanical_on_table"
	PassiveConditionHealerOnTable     = "healer_on_table"

	//если на столе определенное количество карт
	PassiveConditionCountAtLeats = "count_at_least" //<-если карт больше N
	PassiveConditionCountAtMost  = "count_at_most"  //<-если карт меньше или ровно N
	PassiveConditionExact        = "count_exact"    //<-если карт ровно N
)

// Где считать карты для активации пассивок ?
const (
	PassiveCountOwnerAlly  = "ally"
	PassiveCountOwnerEnemy = "enemy"
	PassiveCountOwnerBoth  = "both"
)

// Как считать бонусы ?
const (
	PassiveScaleFlat     = "flat"      //<- просто добавляем Х
	PassiveSaclePerCount = "per_count" //<- за каждого +Х
)
