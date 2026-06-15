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

func (s *Sim) scaleSkillDamageForMagic(def SkillDef, rank int, damageRange DamageRange) DamageRange {
	scale := s.skillMagicScale(def, rank, def.Damage.MagicScaling)
	if scale <= 1 {
		return damageRange
	}
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

func (s *Sim) scaleSkillPercentForMagic(def SkillDef, rank int, effect SkillEffectDef, percent int) int {
	scale := s.skillMagicScale(def, rank, effect.MagicScaling)
	if scale <= 1 || percent <= 0 {
		return percent
	}
	scaled := int(math.Round(float64(percent) * scale))
	if scaled < percent {
		return percent
	}
	return scaled
}

func (s *Sim) scaleSkillRadiusForMagic(def SkillDef, rank int, effect SkillEffectDef) float64 {
	scale := s.skillMagicScale(def, rank, effect.MagicScaling)
	if scale <= 1 || effect.Radius <= 0 {
		return effect.Radius
	}
	return effect.Radius * scale
}

func (s *Sim) skillMagicScale(def SkillDef, rank int, scaling SkillScalingDef) float64 {
	if scaling.Stat == "" {
		return 1
	}
	if scaling.Stat != "magic" || scaling.PercentPerPoint <= 0 || scaling.MaxBonusPercent <= 0 {
		return 1
	}
	magic := s.effectiveBaseStatsView().Magic
	baseline := 0
	if scaling.UseRequirementBaseline {
		baseline = skillStatRequirementForRank(def, "magic", rank)
	}
	excess := magic - baseline
	if excess <= 0 {
		return 1
	}
	bonus := float64(excess) * scaling.PercentPerPoint
	if bonus > scaling.MaxBonusPercent {
		bonus = scaling.MaxBonusPercent
	}
	if bonus <= 0 {
		return 1
	}
	return 1 + bonus/100.0
}

func skillStatRequirementForRank(def SkillDef, stat string, rank int) int {
	if rank < 1 {
		rank = 1
	}
	base := def.Requirements.Stats[stat]
	perRank := def.Requirements.StatsPerRank[stat]
	return base + perRank*(rank-1)
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
	total += s.equippedSetBonusStats()[stat]
	return total
}
