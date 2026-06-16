package game

import "fmt"

type BossEnrageDef struct {
	HealthRatioThreshold float64 `json:"health_ratio_threshold"`
	CooldownMultiplier   float64 `json:"cooldown_multiplier"`
}

func validateBossTemplateEnrage(templateID string, enrage *BossEnrageDef) error {
	if enrage == nil {
		return nil
	}
	if enrage.HealthRatioThreshold <= 0 || enrage.HealthRatioThreshold > 1 {
		return fmt.Errorf("game: invalid rules boss_templates.%s.enrage.health_ratio_threshold: must be > 0 and <= 1", templateID)
	}
	if enrage.CooldownMultiplier <= 0 {
		return fmt.Errorf("game: invalid rules boss_templates.%s.enrage.cooldown_multiplier: must be positive", templateID)
	}
	return nil
}
