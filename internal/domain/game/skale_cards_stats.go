package game

// просто скейлим статы внутрь рантайма в зависимости от уровня боевой карты игрока из БД (что пришло из резолвера)
func ScaleBattleStats(baseHP, baseAttack, level int) (int, int) {
	if level < 1 {
		level = 1
	}
	if level > MaxCardLevel {
		level = MaxCardLevel
	}
	bonus := level - 1
	return baseHP + bonus, baseAttack + bonus
}

// та же тема
func ScaleBuffStats(baseValue, level int) int {
	if level < 1 {
		level = 1
	}
	if level > MaxCardLevel {
		level = MaxCardLevel
	}
	bonus := level - 1
	return baseValue + bonus
}
