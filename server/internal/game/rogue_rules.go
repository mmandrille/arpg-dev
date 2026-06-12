package game

import "fmt"

type SkillPoisonDef struct {
	DamagePercentBase          int `json:"damage_percent_base"`
	DamagePercentPerRank       int `json:"damage_percent_per_rank"`
	DurationTicks              int `json:"duration_ticks"`
	MagicDurationTicksPerPoint int `json:"magic_duration_ticks_per_point"`
}

type SkillDashDef struct {
	RangeBase             float64 `json:"range_base"`
	RangePerRank          float64 `json:"range_per_rank"`
	DamagePercentBase     int     `json:"damage_percent_base"`
	DamagePercentPerMagic int     `json:"damage_percent_per_magic"`
	MaxDamageBonusPercent int     `json:"max_damage_bonus_percent"`
}

func validateRogueConeSkillPayload(skillID string, skill SkillDef) error {
	if skill.Poison.DurationTicks > 0 &&
		(skill.Poison.DamagePercentBase <= 0 || skill.Poison.DamagePercentPerRank < 0 || skill.Poison.MagicDurationTicksPerPoint < 0) {
		return fmt.Errorf("game: invalid rules skills.%s.poison: values must be valid", skillID)
	}
	if (skill.Dash.RangeBase > 0 || skill.Dash.DamagePercentBase > 0) &&
		(skill.Dash.RangeBase <= 0 || skill.Dash.RangePerRank < 0 || skill.Dash.DamagePercentBase <= 0 ||
			skill.Dash.DamagePercentPerMagic < 0 || skill.Dash.MaxDamageBonusPercent < 0) {
		return fmt.Errorf("game: invalid rules skills.%s.dash: values must be valid", skillID)
	}
	return nil
}
