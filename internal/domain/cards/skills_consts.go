package cards

/*Новая система скиллов карт.
Здесь все будет по чистому.*/

type UnitSkillState struct {
	Name         string //имя скилла
	Code         string //уникальный код скилла
	Kind         string //тип скилла, типа damage,heal,buff и так далее
	Target       string //цель скилла
	Power        int    //значение скилла
	BaseCooldown int    //базовый КД скилла
	CooldownLeft int    //фактический кд скилла, как если бы мы только отыграли его
	Duration     int    //длительность скилла
	ExtraValue   int    //спец значение, типа хил + атака и так далее
	BuffEffect   string //эффект от баффа (типа че баффаем ?)
	DebuffEffect string //эффект от дебафа (типа че дебафаем ?)
	CleanseMode  string //что снимает скилл ? Ну типа, есть положительные и отрицательные эффекты, круто, их можно снять
	IgnoreTank   bool   //игронит ли скилл танка ?
	ApplyCount   int    //сколько раз, или на сколько целей применяется скилл ?
}

const (
	SkillKindDamage = "damage"
	SkillKindHeal   = "heal"
	SkillKindBuff   = "buff"
	SkillKindHybrid = "hybrid"
	SkillKindKill   = "kill"
	SkillKindVision = "vision"
	SkillKindDebuff = "debuff"
)

const (
	//свои
	SkillTargetSelf              = "self"                //на себя
	SkillTargetAllySingle        = "ally_single"         //соло союзник
	SkillTargetAllyAdjacent      = "ally_adjacent"       //соседний союзник (вокруг карты источника)
	SkillTargetAllyAll           = "ally_all"            //все союзники
	SkillTargetAllyLowestHP      = "ally_lowest_hp"      //союзник с наименьшим ХП
	SkillTargetAllyHighestAttack = "ally_highest_attack" //союзник с наибольшей атакой
	//враги
	SkillTargetEnemySingle        = "enemy_single"         //соло противник
	SkillTargetEnemySplash        = "enemy_splash"         //несколько противник (цель+-1)
	SkillTargetEnemyAll           = "enemy_all"            //все противники
	SkillTargetEnemyRandom        = "enemy_random"         //случайный противник
	SkillTargetEnemyRandomMulti   = "enemy_random_multi"   //несколько случайных противников
	SkillTargetEnemyLowestHP      = "enemy_lowest_hp"      //противник с наименьшим ХП
	SkillTargetEnemyHighestHP     = "enemy_highest_hp"     //противник с наибольшим ХП
	SkillTargetEnemyHighestAttack = "enemy_highest_attack" //противник с наибольшей атакой
	SkillTargetEnemyLowestAttack  = "enemy_lowest_attack"  //противник с наименьшей атакой
)

//обычные бафы, которые висят на цели
const (
	BuffEffectNone            = ""                 //никакого эффекта. Существует чтобы описывать скилы, смысл которых в том, что это не бафы
	BuffEffectHP              = "hp"               //апаем ХП (именно максимальное значение)
	BuffEffectAttack          = "attack"           //апаем основную атаку
	BuffEffectAttackCooldown  = "attack_cooldown"  //режем кд основной атаки (не базовое значение, а актуальное)
	BuffEffectSkillCooldown   = "skill_cooldown"   //режем кд скилла (опять же, не базовое)
	BuffEffectHealPerTurn     = "heal_per_turn"    //накидываем хил на длину ходов
	BuffEffectSplash          = "splash"           //расширяем сплеш атаки героя
	BuffEffectOverdrive       = "overdrive"        //даруем возможность бить дважды за ход
	BuffEffectMulticast       = "multicast"        //даруем возможность использовать скилл дважды (нельзя навесить на себя)
	BuffEffectMakeTank        = "make_tank"        //делаем из карты танка (если карта уже танк, нельзя навесить)
	BuffEffectShield          = "shield"           //накидываем щит на карту
	BuffEffectReflectShield   = "reflect_shield"   //накидываем щит отражающий урон на карту
	BuffEffectRedirectDamage  = "redirect_damage"  //перенаправляем урон на целевую карту в сторону кастера (нельзя навесить на себя)
	BuffEffectVampiricStrike  = "vampiric_strike"  //даруем вамиризм целевой карте
	BuffEffectChainAttack     = "chain_attack"     //атака целевой карты перескакивает на случайных противников
	BuffEffectDamageReduction = "damage_reduction" //снижаем урон входящий по целевой карте на Х
)

//бафы, которые активируются после тригера
const (
	BuffEffectDeathExplosion   = "death_explosion"    //после смерти карты, на которую повешан бафф -карта взрывается, нанося урон
	BuffEffectDeathMassHeal    = "death_mass_heal"    //после смерти карты, отхиливаются союзники
	BuffEffectCounterattack    = "counterattack"      //накидываем баф, после которого жертва контратакует обидчика
	BuffEffectLifeOnHit        = "life_on_hit"        //фиксированный хил тогда, когда цель атакуют
	BuffEffectBonusAfterAttack = "bonus_after_attack" //бонус посе того, как карта атаковала (+к ХП, +к атаке, -к КД)
)

//эффекты, которые накидываются на союзную цель сразу
const (
	EffectSetFixedHP     = "set_fixed_hp"     //ставим фиксированное кол-во ХП
	EffectEqualizeAllyHP = "equalize_ally_hp" //уравниваем ХП и атаку целевой карты
)

//дебафы
const (
	DebuffEffectNone            = ""                  //никакого эффекта, нужно для описания карты
	DebuffEffectAttackDown      = "attack_down"       //снижаем силу атаки
	DebuffEffectCooldownUp      = "cooldown_up"       //повышаем кд основной атаки
	DebuffEffectSkillCooldownUp = "skill_cooldown_up" //повышаем кд скилла
	DebuffEffectDamageOverTime  = "damage_over_time"  //накидываем доту на дамаг
	DebuffEffectSilence         = "silence"           //запрещаем цели использовать способности
	DebuffEffectNoHeal          = "no_heal"           //запрещаем хилить цель
	DebuffEffectVulnerable      = "vulnerable"        //делаем цель более уязвимой. Она принимает на Х больше дамага
	DebuffEffectDisarm          = "disarm"            //станим карту (она не может атаковать в течении Х ходов, но может колдовать)
	DebuffEffectStun            = "stun"              //цель вообще не может ничего делать в течении Х ходов
)

//снятие эффектов с карты
const (
	CleanseModeNone             = ""                   //никакого эффекта, нужно для описания карты
	CleanseModeRemoveDebuff     = "remove_debuff"      //снимаем один дебаф с цели
	CleanseModeRemoveAllDebuffs = "remove_all_debuffs" //снимаем все негативные эффекты с карты
	CleanseModeRemoveBuff       = "remove_buff"        //снимаем один положительный эффект
	CleanseModeRemoveAllBuffs   = "remove_all_buffs"   //снимаем все положительные эффекты с карты
	CleanseModeRemoveAllEffects = "remove_all_effects" //снимаем вообще все эффекты с цели
)
