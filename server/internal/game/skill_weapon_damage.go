package game

import "math"

func absoluteSkillDamageRange(r *Rules, def SkillDef, rank int) DamageRange {
	if rank < 1 {
		rank = 1
	}
	minDamage := r.rankScaledInt(def.Damage.MinBase, def.Damage.MinPerRank, rank)
	maxDamage := r.rankScaledInt(def.Damage.MaxBase, def.Damage.MaxPerRank, rank)
	if maxDamage < minDamage {
		maxDamage = minDamage
	}

	return DamageRange{Min: minDamage, Max: maxDamage}
}

func skillWeaponMultiplierPercent(r *Rules, def SkillDef, rank int, min bool) int {
	if rank < 1 {
		rank = 1
	}
	if min {
		return r.rankScaledInt(def.Damage.MinBase, def.Damage.MinPerRank, rank)
	}

	return r.rankScaledInt(def.Damage.MaxBase, def.Damage.MaxPerRank, rank)
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
		return absoluteSkillDamageRange(s.rules, def, rank)
	case "weapon_multiplier_range":
		base := s.resolvePlayerAttackDamage()
		return weaponPercentDamageRange(
			base,
			skillWeaponMultiplierPercent(s.rules, def, rank, true),
			skillWeaponMultiplierPercent(s.rules, def, rank, false),
		)
	default:
		return DamageRange{}
	}
}
