package game

import "fmt"

// SkillCompanionDef defines a server-owned summoned companion.
type SkillCompanionDef struct {
	MonsterDefID string  `json:"monster_def_id"`
	VisualModel  string  `json:"visual_model"`
	VisualTint   string  `json:"visual_tint"`
	VisualScale  float64 `json:"visual_scale"`
	Limit        int     `json:"limit"`
}

// SkillReviveDef defines rank-scaled revived-monster companion power.
type SkillReviveDef struct {
	PowerPercentBase    int `json:"power_percent_base"`
	PowerPercentPerRank int `json:"power_percent_per_rank"`
	Limit               int `json:"limit"`
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
	if skill.Companion.Limit != 1 {
		return fmt.Errorf("game: invalid rules skills.%s.companion.limit: must be 1 for this slice", skillID)
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
	if skill.Revive.Limit != 1 {
		return fmt.Errorf("game: invalid rules skills.%s.revive.limit: must be 1 for this slice", skillID)
	}
	if skill.Damage.Type != "" || skill.Projectile.Range > 0 || len(skill.Effects) > 0 {
		return fmt.Errorf("game: invalid rules skills.%s: revive_companion does not support damage, projectile, or effects", skillID)
	}
	return nil
}

func revivePowerPercent(def SkillDef, rank int) int {
	if rank < 1 {
		rank = 1
	}
	return def.Revive.PowerPercentBase + def.Revive.PowerPercentPerRank*(rank-1)
}
