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

/*
Коды скиллов карт. Здесь описывает то, что скилл умеет делать вообще. В дальнейшем сюда будут добавться
еще константы, которые так или иначе влияют на скилы. Далее все это будет обсулживаться в скилловых хендлерах,
где карта будем мапить ключ (которыми выступают нижеописанные константы) и хендлер, который я напишу
*/
const (
	SkillDamageSingle             = "damage_single"               //бьем точечно в одну карту done
	SkillDamageSplash             = "damage_splash"               //бем по сплешу done
	SkillHealSingle               = "heal_single"                 //лечим соло цель
	SkillApplyDebuff              = "apply_debuff"                //накладываем дебаф при атаке, или просто по использованию
	SkillApplyDamageBuff          = "apply_buff"                  //накладываем бафф
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

/*
Триггеры скиллов. Не путать с триггерами пассивок*. Короче скиллы в игре существуют не только нажимаемыми,
но и действующими по ситуации. К примеру когда карта совершает атаку. В этом случае врубается скилл карты,
который потом так или иначе уходит на кд. Здесь так же описываются все возможные варианты. В будущем будут
добавляться еще сценарии, но пока что этого хватает
*/
const (
	TriggerOnPlay    = "on_play"   //срабатывает скил при поставлении карты на стол
	TriggerOnAttack  = "on_attack" //срабатывает, когда карта совершает атаку
	TriggerOnDeath   = "on_death"  //когда карта умирает
	TriggerTurnStart = "on_start"  //в начале хода игрока
	TriggerActive    = "active"    //активный скилл карты
)

/*
Цели скиллов. НА кого они рассчитаны ? Ну типа, существует сценарий, когда карту бьют, она лечит весь свой
стол. Или когда карта атакует, она увеличивает кд противнику. Так и вот, эти константы отвечают на вопрос,
в кого дейсвтует скилл ?
*/
const (
	TargetNone            = "none"              //цели не требуется, например когда карта умирает она бьет всех
	TargetSelf            = "self"              //на себя
	TargetAllyUnit        = "ally_unit"         //на союзную карту
	TargetEnemyUnit       = "enemy_unit"        //на вражескую карту
	TargetAllyAll         = "ally_all"          //весь союзный стол
	TargetEnemyAll        = "enemy_all"         //весь вражеский стол
	TargetBothAll         = "both_all"          //вообще весь стол
	TargetEnemySplash     = "enemy_splash"      //сплеш по противникам
	TargetAllySplash      = "ally_splash"       //сплеш по своим
	TargetAllyGraveSingle = "ally_grave_single" //поднять мертвую карту из могилы
)

// Для скиллов могут быть описаны спец условия. Например игнор танков. Все это описывается здесь
const IgnoreTankTrue = "ignore_tank=true"

/*
Переодические или временные эффекты в зависимости от цели. В чем прикол ? Здесь именно что константы
дот. Доты тоже бывают разными, здесь мы можем накинуть доту на хп, на кд и на атаку, уменьшая ту или
иную характеристику. Так же можно будет действовать и с положительными эффектами, но как правило этого
не нужно делать.
*/
const (
	DotHPUpdate       = "dot_hp_update"       //<-обновляем ХП
	DotAttackUpdate   = "dot_attack_update"   //<-обновляем кд основной атаки
	DotCooldownUpdate = "dot_cooldown_update" //<-обновляем кд скилла
)

/*
--------ПАССИВНЫЕ УМЕНИЯ КАРТ--------
Короче это пиздец. Пассивки устроены намного сложнее скиллов, поскольку их эффекты
постоянны и требуют ну ебать какого менеджмента (в чем ты можешь убедиться, зайдя в
файл passive_skills.go). Соответственно и всяк разных условий у них больше, к примеру:
если на столе противника есть 3 механические карты, то карта-владелец пассивки должна
апнуть всему своему столу урон на 1 ход, триггер в начале хода. Или вот еще прикол, при
смерти союзной демонической карты, карта-владелец пассивки-сосед умершей карты становится
танком на 3 хода. Круто да ?) Соответственно и условий больше. Перейдем к описанию...
*/

/*
Триггеры пассивок. Тут тоже все замудренно пиздец. В чем прикол ? Как я уже рофлил выше
условий для тригера пассивки бывает множество, от постоянных, до специфических в духе -умер
вражеский герой с таким то кодом (НЕ ТИПОМ, А ИМЕННО КОДОМ). Иными словами, когда можно вообще
проверять, можно ли включить пассивку ? Здесь все это описывается
*/
const (
	PassiveTriggerContinuous  = "continuous"    //<-постоянный эффект, не требует проверок
	PassiveTriggerTurnStart   = "turn_start"    //<-на старт хода (тут уже используется тогда, когда нужна проверка)
	PassiveTriggerTurnEnd     = "turn_end"      //<-при завершении хода (та же песня только от обратного)
	PassiveTriggerOnPlay      = "on_play"       //<-при разыгрывании карты (когда ставишь карту на стол)
	PassiveTriggerOnAttack    = "on_attack"     //<-когда карта атакует
	PassiveTriggerOnDeath     = "on_death"      //<-при смерти карты (к примеру, если карта умерла тогда, когда на столе есть Х то...)
	PassiveTriggerOnAllyDead  = "on_ally_dead"  //<-при смерти союзника
	PassiveTriggerOnEnemyDead = "on_enemy_dead" //<-при смерти противника
	PassiveTriggerHitMe       = "on_hit_me"     //<-при ударе карты-владельца
)

/*
Таргеты пассивок. Или кто получает эффекты пассивки ?
С условиями определились, но теперь нужно внести константы, которые будут отражать то, кто
цель пассивного умения ? Вот все они. Здесь так же описывается все, что может понадобиться.
*/
const (
	PassiveTargetSelf          = "self"              //<-эффект на себя
	PassiveTargetAllyAll       = "ally_all"          //<-на всех союзников
	PassiveTargetAttacker      = "attacker"          //<-на того, кто атакует карту-владельца пассивка
	PassiveTargetEnemyAll      = "enemy_all"         //<-на всех противников вообще
	PassiveTargetBothAll       = "both_all"          //<-ВООБЩЕ НА ВСЕХ НА СТОЛЕ
	PassiveTargetAllyLeftRight = "ally_left_right"   //<-на рядом стоящие от карты-владельца карты
	PassiveTargetRandomEnemy   = "random_enemy_unit" //<-на случайного противника
	PassiveTargetRandomAlly    = "random_ally_unit"  //<-на случайного союзника

	PassiveTargetAllyTypeDemonical  = "ally_type_demonical"  //<-на все демонические союзные карты
	PassiveTargetAllyTypeMechanical = "ally_type_mechanical" //<-на все механические союзные карты
	PassiveTargetAllyTypeOrganical  = "ally_type_organical"  //<-на все органические союзные карты
	PassiveTargetAllyTypeHealer     = "ally_type_healer"     //<-на всех хилеров
	//далее то же самое только уже на противников
	PassiveTargetEnemyTypeDemonical  = "enemy_type_demonical"
	PassiveTargetEnemyTypeMechanical = "enemy_type_mechanical"
	PassiveTargetEnemyTypeOrganical  = "enemy_type_organical"
	PassiveTargetEnemyTypeHealer     = "enemy_type_healer"
)

/*
Эффекты от пассивок. Что произойдет ? Ну вот мы описали выше когда действует пассивка и на кого.
И че ? Вот эти константы и отвечаеют на этот вопрос. Что будет ?
*/
const (
	PassiveEffectDamageUp          = "damage_up"           //<-увеличиваем урон
	PassiveEffectHPUp              = "hp_up"               //<-увеличиваем ХП (общие+хил)
	PassiveEffectHeal              = "heal"                //<-просто хилим
	PassiveEffectCooldownDown      = "cooldown_down"       //<-снижаем кд атаки
	PassiveEffectSkillDamageUp     = "skill_damage_up"     //<-увеличиваем дамаг скилла
	PassiveEffectSkillCooldownDown = "skill_cooldown_down" //<-снижаем кд скилла
	PassiveEffectDamage            = "damage"              //<-просто ебашим кого нибудь на столе (не путать с дотами)
)

/*
Когда действует пассивка ?
Пассивки не живут перманентно. В случае триггеров был дан ответ на вопрос -Когда проверять условия
включения пассивки ? То здесь я отвечаю на вопрос, можно ли сработать пассивке в момент времени ? То
есть триггер=момент, кондиции = дополнительные условия. То есть схема такая :
-в начале хода игрока проверяем триггер (например TriggetTurnStart)
-проверяем кондиции (например OrganicalOnTable)
-если все сработало - круто, врубаем пассивку с ее эффектами. Триггер = окно входа, кондиции = проверка,
можно ли включить пассивку.
*/
const (
	PassiveConditionAlways            = "always"              //<-всегда можно врубить при триггере
	PassiveConditionDemonicalOnTable  = "demonical_on_table"  //<-когда демон на столе
	PassiveConditionOrganicalOnTable  = "organical_on_table"  //<-когда человек на столе
	PassiveConditionMechanicalOnTable = "mechanical_on_table" //<-когда на столе машина
	PassiveConditionHealerOnTable     = "healer_on_table"     //<-ну и хилер
	//А тут вступает в силу прикол. Для активации некоторых пассивок нужно условие, к примеру
	//-Органических карт > X. Здесь это описывается.
	PassiveConditionCountAtLeast = "count_at_least" //<-если карт больше Х
	PassiveConditionCountAtMost  = "count_at_most"  //<-если карт меньше или ровно Х
	PassiveConditionExact        = "count_exact"    //<-если карт ровно Х
)

/*
А здесь отвечаем на вопрос, где считать карты, подходящие под критерии ? Ну типа, зацени пример
-увеличить всем хп, если на собзном столе Х людей. Или поднимай всем атаку, если на столе противника
Х демоном. Или вообще бля. Увеличь всем че нибудь если на общем столе хотябы одна карта с таким кодом(не
классом, а кодом). Так вот...
*/
const (
	PassiveCountOwnerAlly  = "ally"  //<-союзный стол
	PassiveCountOwnerEnemy = "enemy" //<-стол противника
	PassiveCountOwnerBoth  = "both"  //<-оба стола
)

// Как считать бонусы ? Тут все просто. Мы либо добавляем Х, либо считает за каждого подходящего под условие
const (
	PassiveScaleFlat     = "flat"      //<- просто добавляем Х
	PassiveScalePerCount = "per_count" //<- за каждого +Х
)

/*Таким образом реализован такой ебанутый коктейль из возможных вариаций скиллов+пассивок что это пиздец
В дальнейшем буду еще добавлять. Цепочка добавления такова -пишем константу-пиздуем в passive_skills.go,
добавляем или переиспользуем хендлер, если нужно добавлять новый хелпер или кейс -добавляем в passiveTargets,
passiveConditionOK (если появятся новые типы карт), регаем код пассивки (который пишем в карте) в мапу с пассивками,
*/
