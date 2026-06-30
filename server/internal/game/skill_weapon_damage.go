package game

import "math"

func absoluteSkillDamageRange(def SkillDef, rank int) DamageRange {
	if rank < 1 {
		rank = 1
	}
	minDamage := def.Damage.MinBase + def.Damage.MinPerRank*(rank-1)
	maxDamage := def.Damage.MaxBase + def.Damage.MaxPerRank*(rank-1)
	if minDamage < 0 {
		minDamage = 0
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func skillWeaponMultiplierPercent(def SkillDef, rank int, min bool) int {
	if rank < 1 {
		rank = 1
	}
	if min {
		return def.Damage.MinBase + def.Damage.MinPerRank*(rank-1)
	}
	return def.Damage.MaxBase + def.Damage.MaxPerRank*(rank-1)
}

func weaponPercentDamageRange(base DamageRange, minPercent, maxPercent int) DamageRange {
	minDamage := 0
	maxDamage := 0
	if minPercent > 0 {
		minDamage = int(math.Round(float64(base.Min) * float64(minPercent) / 100.0))
		if minDamage < 1 {
			minDamage = 1
		}
	}
	if maxPercent > 0 {
		maxDamage = int(math.Round(float64(base.Max) * float64(maxPercent) / 100.0))
		if maxDamage < 1 {
			maxDamage = 1
		}
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func (s *Sim) skillDamageRange(def SkillDef, rank int) DamageRange {
	switch def.Damage.Type {
	case "", "rank_linear_range":
		return absoluteSkillDamageRange(def, rank)
	case "weapon_multiplier_range":
		base := s.resolvePlayerAttackDamage()
		return weaponPercentDamageRange(
			base,
			skillWeaponMultiplierPercent(def, rank, true),
			skillWeaponMultiplierPercent(def, rank, false),
		)
	default:
		return DamageRange{}
	}
}
