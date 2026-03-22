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

//ГЛОБАЛЬНЫЙ ПАТЧ
//триггеры карт. Типа когда срабатывают скилы ?
const (
	TriggerOnPlay    = "on_play"   //срабатывает скил при поставлении карты на стол
	TriggerOnAttack  = "on_attack" //срабатывает, когда карта совершает атаку
	TriggerOnDeath   = "on_death"  //когда карта умирает
	TriggerTurnStart = "on_start"  //в начале хода игрока
	TriggerActive    = "active"    //активный скилл карты
)

//на кого срабатывают карты ?
const (
	TargetNone        = "none"         //цели не требуется, например когда карта умирает она бьет всех
	TargetSelf        = "self"         //на себя
	TargetAllyUnit    = "ally_unit"    //на союзную карту
	TargetEnemyUnit   = "enemy_unit"   //на вражескую карту
	TargetAllyAll     = "ally_all"     //весь союзный стол
	TargetEnemyAll    = "enemy_all"    //весь вражеский стол
	TargetBothAll     = "both_all"     //вообще весь стол
	TargetEnemySplash = "enemy_splash" //сплеш по противникам
	TargetAllySplash  = "ally_splash"  //сплеш по своим
)

//коды скиллов карт
const (
	SkillDamageSingle         = "damage_single"             //бьем точечно в одну карту done
	SkillDamageSplash         = "damage_splash"             //бем по сплешу done
	SkillApplyDebuff          = "apply_debuff"              //накладываем дебаф при атаке, или просто по использованию
	SkillApplyBuff            = "apply_buff"                //накладываем бафф
	SkillSummonSelfCopy       = "summon_self_copy"          //призывает копии себя же
	SkillBanishUnit           = "banish_unit"               //убирает карту со стола
	SkillRevealEnemyHand      = "reveal_enemy_hand"         //смотрим в руку противника
	SkillIncEnemyCdAllOnDeath = "inc_enemy_cd_all_on_death" //после смерти карты увеличиваем кд всего вражеского стола
	SkillIncEnemyCdSingle     = "inc_enemy_cd_single"       //точечно увеличиваем кд вражеской карты
	SkillDecAllyCdSingle      = "dec_ally_cd_single"        //снимаем/уменьшаем кд союзной карты
	SkillResurrectFromGrave   = "resurrect_from_grave"      //воскрешает карту
	SkillResurrectNextTurn    = "resurrect_next_turn"       //карта воскресает на следующий ход после смерти
	SkillDeathAoe             = "death_aoe"                 //при уничтожении бьет по столу
)
