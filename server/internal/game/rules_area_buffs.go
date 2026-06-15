package game

import "fmt"

func validateAreaStatPercentBuffEffect(skillID string, idx int, effect SkillEffectDef) error {
	if len(effect.Stats) == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats: at least one stat is required", skillID, idx)
	}
	seen := map[string]bool{}
	for _, stat := range effect.Stats {
		if stat != "armor" && stat != "block_percent" && !isSupportedRequirementStat(stat) {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats.%s: unsupported stat", skillID, idx, stat)
		}
		if seen[stat] {
			return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats.%s: duplicate stat", skillID, idx, stat)
		}
		seen[stat] = true
	}
	if effect.PercentBase < 0 || effect.PercentPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values must be non-negative", skillID, idx)
	}
	if effect.PercentBase == 0 && effect.PercentPerRank == 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: percent values cannot both be zero", skillID, idx)
	}
	if err := validateAreaBuffCommon(skillID, idx, effect); err != nil {
		return err
	}
	return validateSkillMagicScaling(fmt.Sprintf("skills.%s.effects[%d].magic_scaling", skillID, idx), effect.MagicScaling)
}

func validateAreaImmunityBuffEffect(skillID string, idx int, effect SkillEffectDef) error {
	if len(effect.Stats) != 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].stats: immunity buffs do not support stats", skillID, idx)
	}
	if effect.PercentBase != 0 || effect.PercentPerRank != 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d]: immunity buffs do not support percent values", skillID, idx)
	}
	if err := validateAreaBuffCommon(skillID, idx, effect); err != nil {
		return err
	}
	return validateSkillMagicScaling(fmt.Sprintf("skills.%s.effects[%d].magic_scaling", skillID, idx), effect.MagicScaling)
}

func validateAreaBuffCommon(skillID string, idx int, effect SkillEffectDef) error {
	if effect.DurationTicks <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].duration_ticks: must be positive", skillID, idx)
	}
	if effect.Target != "allies" {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].target: unsupported %s", skillID, idx, effect.Target)
	}
	if effect.Range <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].range: must be positive", skillID, idx)
	}
	if effect.Radius <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].radius: must be positive", skillID, idx)
	}
	if effect.EffectID == "" {
		return fmt.Errorf("game: invalid rules skills.%s.effects[%d].effect_id: required", skillID, idx)
	}
	return nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}
