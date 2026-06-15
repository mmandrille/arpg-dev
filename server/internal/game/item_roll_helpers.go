package game

func weightedRollableStat(stats []RollableStatDef, rng *RNG) (RollableStatDef, bool) {
	total := 0
	for _, stat := range stats {
		total += stat.Weight
	}
	if total <= 0 {
		return RollableStatDef{}, false
	}
	roll := rng.IntN(total)
	for _, stat := range stats {
		roll -= stat.Weight
		if roll < 0 {
			return stat, true
		}
	}
	return stats[len(stats)-1], true
}
