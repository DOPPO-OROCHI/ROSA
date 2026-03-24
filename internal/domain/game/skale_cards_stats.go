package game

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
