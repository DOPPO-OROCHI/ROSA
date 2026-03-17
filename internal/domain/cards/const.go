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
