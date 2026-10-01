package game

// RangerAffinityDamageDef is the bounded shared rule for Deadeye's ranged bonus.
type RangerAffinityDamageDef struct {
	AffinityStat              string `json:"affinity_stat"`
	BasePercentPerAffinity    int    `json:"base_percent_per_affinity"`
	StatPointsPerExtraPercent int    `json:"stat_points_per_extra_percent"`
	MaxActiveAffinities       int    `json:"max_active_affinities"`
	MaxBonusPercent           int    `json:"max_bonus_percent"`
}

// rangerAffinityPercent is a pure evaluator of Deadeye's shared rule. The
// caller supplies server-owned effective stats and equipped affinity count.
func rangerAffinityPercent(rule RangerAffinityDamageDef, effectiveStat, count, allocatedRank int) int {
	if allocatedRank <= 0 || count <= 0 || rule.StatPointsPerExtraPercent <= 0 || rule.MaxActiveAffinities <= 0 {
		return 0
	}
	if effectiveStat < 0 {
		effectiveStat = 0
	}
	if count > rule.MaxActiveAffinities {
		count = rule.MaxActiveAffinities
	}
	bonus := count * (rule.BasePercentPerAffinity + effectiveStat/rule.StatPointsPerExtraPercent)
	if bonus > rule.MaxBonusPercent {
		bonus = rule.MaxBonusPercent
	}
	if bonus < 0 {
		return 0
	}
	return bonus
}

func rangerAffinityStatValue(stats BaseStatsView, stat string) int {
	switch stat {
	case "str":
		return stats.Str
	case "dex":
		return stats.Dex
	case "vit":
		return stats.Vit
	case "magic":
		return stats.Magic
	default:
		return 0
	}
}

func (s *Sim) beneficialRangerAffinityCount() int {
	count := 0
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil || item.rollPayload == nil {
			continue
		}
		for _, affinity := range item.rollPayload.ClassAffinities {
			if affinity.Class == "ranger" && affinity.Value > 0 && classAffinityActive("ranger", affinity) {
				count++
			}
		}
	}
	return count
}

func (s *Sim) rangerAffinityBonusPercent() int {
	if s == nil || s.rules == nil || s.progression.CharacterClass != "ranger" || s.progression.SkillRanks["deadeye"] < 1 {
		return 0
	}
	deadeye, ok := s.rules.Skills["deadeye"]
	if !ok || deadeye.PassiveStats.RangerAffinityDamage == nil {
		return 0
	}
	rule := *deadeye.PassiveStats.RangerAffinityDamage
	stat := rangerAffinityStatValue(s.effectiveBaseStatsView(), rule.AffinityStat)
	return rangerAffinityPercent(rule, stat, s.beneficialRangerAffinityCount(), s.progression.SkillRanks["deadeye"])
}

func scaleRangerAffinityDamage(in DamageRange, bonusPercent int) DamageRange {
	if bonusPercent <= 0 {
		return in
	}
	return DamageRange{
		Min: in.Min * (100 + bonusPercent) / 100,
		Max: in.Max * (100 + bonusPercent) / 100,
	}
}

func (s *Sim) rangerAffinityBasicDamage(in DamageRange) DamageRange {
	if s.playerEquippedWeaponItemType() != "bow" {
		return in
	}
	return scaleRangerAffinityDamage(in, s.rangerAffinityBonusPercent())
}

func (s *Sim) rangerAffinitySkillDamage(def SkillDef, in DamageRange) DamageRange {
	if def.RangerAffinityEligible == nil || !*def.RangerAffinityEligible || def.Class != "ranger" || def.Kind != "projectile_attack" {
		return in
	}
	return scaleRangerAffinityDamage(in, s.rangerAffinityBonusPercent())
}
