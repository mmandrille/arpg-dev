package game

import "math"

func (s *Sim) applySkillDamageBonus(damageRange DamageRange) DamageRange {
	percent := s.equippedItemStatTotal("skill_damage_percent")
	if percent <= 0 {
		return damageRange
	}
	scale := 1.0 + float64(percent)/100.0
	damageRange.Min = int(math.Round(float64(damageRange.Min) * scale))
	damageRange.Max = int(math.Round(float64(damageRange.Max) * scale))
	if damageRange.Min < 0 {
		damageRange.Min = 0
	}
	if damageRange.Max < damageRange.Min {
		damageRange.Max = damageRange.Min
	}
	return damageRange
}

func (s *Sim) equippedItemStatTotal(stat string) int {
	total := 0
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil {
			continue
		}
		baseStats, rolledStats := s.itemBaseAndRollStats(item)
		total += baseStats[stat] + rolledStats[stat]
	}
	return total
}
