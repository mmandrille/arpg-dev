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
