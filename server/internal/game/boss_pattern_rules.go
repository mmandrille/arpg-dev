package game

import "fmt"

func validateBossPatterns(patterns map[string]BossPatternDef, minTelegraphTicks int, monsters ...map[string]MonsterDef) error {
	if len(patterns) == 0 {
		return fmt.Errorf("game: invalid rules boss_patterns.patterns: required")
	}
	for patternID, pattern := range patterns {
		if len(pattern.Phases) == 0 {
			return fmt.Errorf("game: invalid rules boss_patterns.%s.phases: required", patternID)
		}
		if pattern.CooldownTicks < 0 {
			return fmt.Errorf("game: invalid rules boss_patterns.%s.cooldown_ticks: must be non-negative", patternID)
		}
		var priorTelegraph *BossPatternPhase
		for idx, phase := range pattern.Phases {
			if phase.DurationTicks <= 0 {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].duration_ticks: must be positive", patternID, idx)
			}
			switch phase.Kind {
			case "telegraph":
				if phase.DurationTicks < minTelegraphTicks {
					return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].duration_ticks: below minimum telegraph", patternID, idx)
				}
				if phase.TelegraphType == "" || phase.HitShape == "" {
					return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: telegraph_type and hit_shape required", patternID, idx)
				}
				if phase.Radius <= 0 {
					return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].radius: must be positive", patternID, idx)
				}
				if (phase.HitShape == "line" || phase.HitShape == "cone") && phase.Width <= 0 {
					return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].width: must be positive for %s", patternID, idx, phase.HitShape)
				}
				copy := phase
				priorTelegraph = &copy
			case "active":
				if phase.SummonMonsterDefID != "" || phase.SummonCount != 0 || phase.SummonRadius != 0 {
					if phase.SummonMonsterDefID == "" {
						return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].summon_monster_def_id: required", patternID, idx)
					}
					if len(monsters) > 0 {
						if _, ok := monsters[0][phase.SummonMonsterDefID]; !ok {
							return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].summon_monster_def_id: unknown monster %s", patternID, idx, phase.SummonMonsterDefID)
						}
					}
					if phase.SummonCount <= 0 {
						return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].summon_count: must be positive", patternID, idx)
					}
					if phase.SummonRadius <= 0 {
						return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].summon_radius: must be positive", patternID, idx)
					}
				}
				if phase.Damage != nil {
					if priorTelegraph == nil {
						return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: damage requires prior telegraph", patternID, idx)
					}
					if err := validateDamageRange(fmt.Sprintf("boss_patterns.%s.phases[%d].damage", patternID, idx), *phase.Damage); err != nil {
						return err
					}
					if phase.Shape != priorTelegraph.HitShape || phase.Radius != priorTelegraph.Radius || phase.Width != priorTelegraph.Width {
						return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: active hit predicate must match telegraph", patternID, idx)
					}
				}
			case "recovery":
			default:
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d].kind: %s", patternID, idx, phase.Kind)
			}
		}
	}
	return nil
}
