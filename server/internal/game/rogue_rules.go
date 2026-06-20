package game

import "fmt"

type SkillPoisonDef struct {
	DamagePercentBase          int    `json:"damage_percent_base"`
	DamagePercentPerRank       int    `json:"damage_percent_per_rank"`
	DurationTicks              int    `json:"duration_ticks"`
	MagicDurationTicksPerPoint int    `json:"magic_duration_ticks_per_point"`
	MarkDamageBonusPercent     int    `json:"mark_damage_bonus_percent"`
	MarkDurationTicks          int    `json:"mark_duration_ticks"`
	MarkEffectID               string `json:"mark_effect_id"`
}

type SkillDashDef struct {
	RangeBase             float64 `json:"range_base"`
	RangePerRank          float64 `json:"range_per_rank"`
	DamagePercentBase     int     `json:"damage_percent_base"`
	DamagePercentPerMagic int     `json:"damage_percent_per_magic"`
	MaxDamageBonusPercent int     `json:"max_damage_bonus_percent"`
	StunEffectID          string  `json:"stun_effect_id"`
	StunDurationTicks     int     `json:"stun_duration_ticks"`
}

type SkillMobilityDef struct {
	RangeBase            float64  `json:"range_base"`
	RangePerRank         float64  `json:"range_per_rank"`
	Mode                 string   `json:"mode"`
	Visual               string   `json:"visual"`
	IgnoreObstacleKinds  []string `json:"ignore_obstacle_kinds,omitempty"`
	SpeedTilesPerSecond  float64  `json:"speed_tiles_per_second"`
	SpeedMultiplier      float64  `json:"speed_multiplier"`
	ChannelManaPer10Sec  int      `json:"channel_mana_per_10_seconds"`
	DamagePercentBase    int      `json:"damage_percent_base"`
	DamagePercentPerRank int      `json:"damage_percent_per_rank"`
	ImpactRadius         float64  `json:"impact_radius"`
	PushMin              float64  `json:"push_min"`
	PushMax              float64  `json:"push_max"`
	StunEffectID         string   `json:"stun_effect_id"`
	StunDurationTicks    int      `json:"stun_duration_ticks"`
	RootEffectID         string   `json:"root_effect_id"`
	RootDurationTicks    int      `json:"root_duration_ticks"`
}

func validateRogueConeSkillPayload(skillID string, skill SkillDef) error {
	if skill.Poison.DurationTicks > 0 &&
		(skill.Poison.DamagePercentBase <= 0 || skill.Poison.DamagePercentPerRank < 0 || skill.Poison.MagicDurationTicksPerPoint < 0 ||
			skill.Poison.MarkDamageBonusPercent < 0 || skill.Poison.MarkDurationTicks < 0) {
		return fmt.Errorf("game: invalid rules skills.%s.poison: values must be valid", skillID)
	}
	if skill.Poison.MarkDurationTicks > 0 && (skill.Poison.MarkDamageBonusPercent <= 0 || skill.Poison.MarkEffectID == "") {
		return fmt.Errorf("game: invalid rules skills.%s.poison.mark: bonus and effect id are required", skillID)
	}
	if (skill.Dash.RangeBase > 0 || skill.Dash.DamagePercentBase > 0) &&
		(skill.Dash.RangeBase <= 0 || skill.Dash.RangePerRank < 0 || skill.Dash.DamagePercentBase <= 0 ||
			skill.Dash.DamagePercentPerMagic < 0 || skill.Dash.MaxDamageBonusPercent < 0 || skill.Dash.StunDurationTicks < 0) {
		return fmt.Errorf("game: invalid rules skills.%s.dash: values must be valid", skillID)
	}
	if skill.Dash.StunDurationTicks > 0 && skill.Dash.StunEffectID == "" {
		return fmt.Errorf("game: invalid rules skills.%s.dash.stun_effect_id: required", skillID)
	}
	if skill.Kind == "mobility" {
		if skill.Mobility.RangeBase <= 0 || skill.Mobility.RangePerRank < 0 || skill.Mobility.Visual == "" {
			return fmt.Errorf("game: invalid rules skills.%s.mobility: range and visual must be valid", skillID)
		}
		switch skill.Mobility.Mode {
		case "teleport", "leap", "charge", "disengage":
		default:
			return fmt.Errorf("game: invalid rules skills.%s.mobility.mode: unsupported %s", skillID, skill.Mobility.Mode)
		}
		if skill.Mobility.SpeedTilesPerSecond < 0 || skill.Mobility.SpeedMultiplier < 0 || skill.Mobility.ChannelManaPer10Sec < 0 || skill.Mobility.DamagePercentBase < 0 || skill.Mobility.DamagePercentPerRank < 0 || skill.Mobility.ImpactRadius < 0 ||
			skill.Mobility.PushMin < 0 || skill.Mobility.PushMax < skill.Mobility.PushMin ||
			skill.Mobility.StunDurationTicks < 0 || skill.Mobility.RootDurationTicks < 0 {
			return fmt.Errorf("game: invalid rules skills.%s.mobility: effect values must be non-negative", skillID)
		}
		for _, kind := range skill.Mobility.IgnoreObstacleKinds {
			switch kind {
			case obstacleKindWater, obstacleKindHole:
			default:
				return fmt.Errorf("game: invalid rules skills.%s.mobility.ignore_obstacle_kinds: unsupported %s", skillID, kind)
			}
		}
		if skill.Mobility.Mode == "charge" && skill.Mobility.SpeedMultiplier <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.mobility.speed_multiplier: required for charge", skillID)
		}
		if skill.Mobility.Mode == "charge" && skill.Mobility.ChannelManaPer10Sec <= 0 {
			return fmt.Errorf("game: invalid rules skills.%s.mobility.channel_mana_per_10_seconds: required for charge", skillID)
		}
		if skill.Mobility.StunDurationTicks > 0 && skill.Mobility.StunEffectID == "" {
			return fmt.Errorf("game: invalid rules skills.%s.mobility.stun_effect_id: required", skillID)
		}
		if skill.Mobility.RootDurationTicks > 0 && skill.Mobility.RootEffectID == "" {
			return fmt.Errorf("game: invalid rules skills.%s.mobility.root_effect_id: required", skillID)
		}
	}
	return nil
}
