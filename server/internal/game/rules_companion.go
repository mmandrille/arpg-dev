package game

import "fmt"

// SkillCompanionDef defines a server-owned summoned companion.
type SkillCompanionDef struct {
	MonsterDefID string                 `json:"monster_def_id"`
	VisualModel  string                 `json:"visual_model"`
	VisualTint   string                 `json:"visual_tint"`
	VisualScale  float64                `json:"visual_scale"`
	Limit        SkillCompanionLimitDef `json:"limit"`
}

// SkillReviveDef defines rank-scaled revived-monster companion power.
type SkillReviveDef struct {
	PowerPercentBase       int                    `json:"power_percent_base"`
	PowerPercentPerRank    int                    `json:"power_percent_per_rank"`
	DurationSecondsBase    int                    `json:"duration_seconds_base"`
	DurationSecondsPerRank int                    `json:"duration_seconds_per_rank"`
	Limit                  SkillCompanionLimitDef `json:"limit"`
}

// SkillCompanionLimitDef defines active companion quantity scaling.
type SkillCompanionLimitDef struct {
	Base         int `json:"base"`
	PerRankStep  int `json:"per_rank_step"`
	RanksPerStep int `json:"ranks_per_step"`
}

func validateSummonCompanionSkillPayload(skillID string, skill SkillDef, monsters map[string]MonsterDef) error {
	if skill.Companion.MonsterDefID == "" {
		return fmt.Errorf("game: invalid rules skills.%s.companion.monster_def_id: required", skillID)
	}
	if monsters != nil {
		if _, ok := monsters[skill.Companion.MonsterDefID]; !ok {
			return fmt.Errorf("game: invalid rules skills.%s.companion.monster_def_id: unknown monster %s", skillID, skill.Companion.MonsterDefID)
		}
	}
	if err := validateCompanionLimit("skills."+skillID+".companion.limit", skill.Companion.Limit); err != nil {
		return err
	}
	if skill.Companion.VisualModel == "" {
		return fmt.Errorf("game: invalid rules skills.%s.companion.visual_model: required", skillID)
	}
	if skill.Companion.VisualScale < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.companion.visual_scale: must be non-negative", skillID)
	}
	if skill.Damage.Type != "" || skill.Projectile.Range > 0 || len(skill.Effects) > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: summon_companion does not support damage, projectile, or effects", skillID)
	}
	return nil
}

func validateSkillCompanionMonsterRefs(skills map[string]SkillDef, monsters map[string]MonsterDef) error {
	for skillID, skill := range skills {
		if skill.Kind != "summon_companion" {
			continue
		}
		if _, ok := monsters[skill.Companion.MonsterDefID]; !ok {
			return fmt.Errorf("game: invalid rules skills.%s.companion.monster_def_id: unknown monster %s", skillID, skill.Companion.MonsterDefID)
		}
	}
	return nil
}

func validateReviveCompanionSkillPayload(skillID string, skill SkillDef) error {
	if skill.Targeting != "direction_or_target" {
		return fmt.Errorf("game: invalid rules skills.%s.targeting: unsupported %s for revive_companion", skillID, skill.Targeting)
	}
	if skill.Revive.PowerPercentBase <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.revive.power_percent_base: must be positive", skillID)
	}
	if skill.Revive.PowerPercentPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.revive.power_percent_per_rank: must be non-negative", skillID)
	}
	if skill.Revive.DurationSecondsBase <= 0 {
		return fmt.Errorf("game: invalid rules skills.%s.revive.duration_seconds_base: must be positive", skillID)
	}
	if skill.Revive.DurationSecondsPerRank < 0 {
		return fmt.Errorf("game: invalid rules skills.%s.revive.duration_seconds_per_rank: must be non-negative", skillID)
	}
	if err := validateCompanionLimit("skills."+skillID+".revive.limit", skill.Revive.Limit); err != nil {
		return err
	}
	if skill.Damage.Type != "" || skill.Projectile.Range > 0 || len(skill.Effects) > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: revive_companion does not support damage, projectile, or effects", skillID)
	}
	return nil
}

func validateCompanionLimit(label string, limit SkillCompanionLimitDef) error {
	if limit.Base <= 0 {
		return fmt.Errorf("game: invalid rules %s.base: must be positive", label)
	}
	if limit.PerRankStep < 0 {
		return fmt.Errorf("game: invalid rules %s.per_rank_step: must be non-negative", label)
	}
	if limit.RanksPerStep <= 0 {
		return fmt.Errorf("game: invalid rules %s.ranks_per_step: must be positive", label)
	}
	return nil
}

func revivePowerPercent(def SkillDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	return def.Revive.PowerPercentBase + def.Revive.PowerPercentPerRank*(rank-1)
}

func reviveDurationTicks(def SkillDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	seconds := def.Revive.DurationSecondsBase + def.Revive.DurationSecondsPerRank*(rank-1)
	if seconds < 1 {
		seconds = 1
	}
	return seconds * 10
}

func companionLimitAtRank(limit SkillCompanionLimitDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	out := limit.Base + ((rank - 1) / limit.RanksPerStep * limit.PerRankStep)
	if out < 1 {
		return 1
	}
	return out
}
