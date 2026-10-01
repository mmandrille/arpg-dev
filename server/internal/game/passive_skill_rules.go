package game

import "fmt"

func validatePassiveStatSkillPayload(skillID string, skill SkillDef) error {
	if len(skill.PassiveStats.Stats) == 0 && len(skill.PassiveStats.AffinityScaling.Stats) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.passive_stats: at least one stat bonus is required", skillID)
	}
	for stat, value := range skill.PassiveStats.Stats {
		if !isSupportedPassiveSkillStat(stat) {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.stats.%s: unsupported stat", skillID, stat)
		}
		if value.Base < 0 || value.PerRank < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.stats.%s: values must be non-negative", skillID, stat)
		}
		if value.Base == 0 && value.PerRank == 0 {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.stats.%s: must grant a bonus", skillID, stat)
		}
	}
	if len(skill.PassiveStats.AffinityScaling.Stats) > 0 {
		if skill.PassiveStats.AffinityScaling.MaxActiveAffinities <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.affinity_scaling.max_active_affinities: must be positive", skillID)
		}
		for stat, value := range skill.PassiveStats.AffinityScaling.Stats {
			if !isSupportedPassiveSkillStat(stat) {
				return fmt.Errorf("game: invalid rules skills.%s.passive_stats.affinity_scaling.stats.%s: unsupported stat", skillID, stat)
			}
			if value.Base < 0 || value.PerRank < 0 {
				return fmt.Errorf("game: invalid rules skills.%s.passive_stats.affinity_scaling.stats.%s: values must be non-negative", skillID, stat)
			}
			if value.Base == 0 && value.PerRank == 0 {
				return fmt.Errorf("game: invalid rules skills.%s.passive_stats.affinity_scaling.stats.%s: must grant a bonus", skillID, stat)
			}
		}
	}
	if cfg := skill.PassiveStats.RangerAffinityDamage; cfg != nil {
		if skillID != "deadeye" || skill.Class != "ranger" {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.ranger_affinity_damage: only Ranger Deadeye may own this rule", skillID)
		}
		if cfg.AffinityStat != "str" && cfg.AffinityStat != "dex" && cfg.AffinityStat != "vit" && cfg.AffinityStat != "magic" {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.ranger_affinity_damage.affinity_stat", skillID)
		}
		if cfg.BasePercentPerAffinity < 0 || cfg.BasePercentPerAffinity > 10 || cfg.StatPointsPerExtraPercent < 1 || cfg.StatPointsPerExtraPercent > 100 || cfg.MaxActiveAffinities < 1 || cfg.MaxActiveAffinities > 3 || cfg.MaxBonusPercent < 1 || cfg.MaxBonusPercent > 25 {
			return fmt.Errorf("game: invalid rules skills.%s.passive_stats.ranger_affinity_damage: coefficients outside supported bounds", skillID)
		}
	}
	if len(skill.Effects) > 0 || skill.Execute.ThresholdPercentBase > 0 || skill.Projectile.Range > 0 || skill.Cone.Range > 0 || skill.Dash.RangeBase > 0 || skill.Mobility.RangeBase > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: passive_stat_bonus does not support active payloads", skillID)
	}
	return nil
}

func isSupportedPassiveSkillStat(stat string) bool {
	switch stat {
	case "all_skills", "hotbar_slots", "inventory_rows":
		return false
	case "damage_percent", "armor_percent", "max_hp_percent", "max_mana_percent", "health_regen_percent", "mana_regen_percent", "light_radius_percent", "movement_speed_percent":
		return true
	default:
		return isSupportedItemStat(stat)
	}
}
