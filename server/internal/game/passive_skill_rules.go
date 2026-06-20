package game

import "fmt"

func validatePassiveStatSkillPayload(skillID string, skill SkillDef) error {
	if len(skill.PassiveStats.Stats) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.passive_stats.stats: required", skillID)
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
	if len(skill.Effects) > 0 || skill.Execute.ThresholdPercentBase > 0 || skill.Projectile.Range > 0 || skill.Cone.Range > 0 || skill.Dash.RangeBase > 0 || skill.Mobility.RangeBase > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: passive_stat_bonus does not support active payloads", skillID)
	}
	return nil
}

func isSupportedPassiveSkillStat(stat string) bool {
	switch stat {
	case "all_skills", "hotbar_slots", "inventory_rows":
		return false
	case "damage_percent", "armor_percent", "max_hp_percent", "max_mana_percent", "health_regen_percent", "mana_regen_percent", "light_radius_percent":
		return true
	default:
		return isSupportedItemStat(stat)
	}
}
