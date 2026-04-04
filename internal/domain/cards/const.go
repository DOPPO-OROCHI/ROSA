package cards

/*Константы, служащие для описания типов карт и эффектов от бафа. К примеру, есть карта
MachanicalGuardian. Ее тип MechanicalCard соответстввенно, а значит, что для такого типа
карты сработают только определенные бафы, поскольку те тоже не для всех (хотя я сейчас так
думаю, что нахер это и не надо, но пока давай сделаем так, чтобы моя игра была типа сложной
и продуманной...). Так вот, здесь да, чисто константы чтобы использовать их при описании
карт. Ато мало ли ошибусь.*/

const (
	//типы карт
	Mechanical = "mechanical"
	Human      = "organical"
	Demonical  = "demonical"
	Healer     = "healer"
)

const (
	CoolDownUpdate        = "cooldown_update"
	HealthPointsUpdate    = "health_points_update"
	MaxHealthPointsUpdate = "max_health_points_update"
	DamageUpdate          = "damage_update"
	MakeTankUpdate        = "make_tank"
	SkillDamageUpdate     = "skill_damage_update"
	SkillCooldownUpdate   = "skill_cooldown_update"
)
