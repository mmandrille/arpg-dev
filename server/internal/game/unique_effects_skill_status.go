package game

func (s *Sim) applySkillHitStatusUnique(primary *entity, playerID uint64, skillID string, def UniqueEffectDef, sourceDamage int, corr string, res *TickResult) {
	if primary == nil || skillID == "" || uniqueEffectStringParam(def, "skill_id", "") != skillID || sourceDamage <= 0 {
		return
	}
	targets := []*entity{primary}
	radius := uniqueEffectFloatParam(def, "splash_radius_tiles", 0)
	if radius > 0 {
		targets = append(targets, s.uniqueNearbyMonsters(primary, radius)...)
	}
	if statusID := uniqueEffectStringParam(def, "dot_status_id", ""); statusID != "" {
		damage := percentOf(sourceDamage, uniqueEffectIntParam(def, "dot_tick_damage_percent_of_hit", 0))
		intervalTicks := uniqueEffectIntParam(def, "dot_tick_interval_seconds", 0) * 10
		durationTicks := uniqueEffectIntParam(def, "dot_duration_seconds", 0) * 10
		if damage > 0 && intervalTicks > 0 && durationTicks > 0 {
			for _, target := range targets {
				s.startDotStatus(target, statusID, def.ID, playerID, 1, damage, intervalTicks, durationTicks, corr, res)
			}
		}
	}
	if statusID := uniqueEffectStringParam(def, "slow_status_id", ""); statusID != "" {
		percent := uniqueEffectIntParam(def, "slow_percent", 0)
		maxPercent := uniqueEffectIntParam(def, "slow_max_percent", percent)
		durationTicks := uniqueEffectIntParam(def, "slow_duration_seconds", 0) * 10
		if percent > 0 && durationTicks > 0 {
			for _, target := range targets {
				s.applySlowStatus(target, statusID, def.ID, playerID, percent, maxPercent, durationTicks, corr, res)
			}
		}
	}
}

func (s *Sim) uniqueNearbyMonsters(primary *entity, radius float64) []*entity {
	if primary == nil || radius <= 0 {
		return nil
	}
	targets := []*entity{}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.id == primary.id || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		if distance(primary.pos, target.pos) > radius+meleeRangeEpsilon {
			continue
		}
		targets = append(targets, target)
	}
	return targets
}
