package game

import "fmt"

func (s *Sim) applyMonsterSlow(target *entity, sourceID uint64, skillID string, slow SkillSlowDef, correlationID string, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || slow.DurationTicks <= 0 {
		return
	}
	effectID := slow.EffectID
	if effectID == "" {
		effectID = skillID
	}
	stateKey := fmt.Sprintf("%s:%d", skillID, target.id)
	current := 0
	if existing, ok := s.skillEffects[stateKey]; ok && existing.EndsTick > s.tick {
		current = existing.Percent
	}
	percent := current + slow.Percent
	if percent > slow.MaxPercent {
		percent = slow.MaxPercent
	}
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:    skillID,
		TargetID:   target.id,
		Stats:      []string{"movement_speed"},
		Percent:    percent,
		EffectID:   effectID,
		EndsTick:   s.tick + uint64(slow.DurationTicks),
		TotalTicks: slow.DurationTicks,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Amount:         intPtr(percent),
		RemainingTicks: intPtr(slow.DurationTicks),
		TotalTicks:     intPtr(slow.DurationTicks),
	})
	s.replicateSkillEffectToNearbyMonsters(sourceID, target, stateKey, res)
}

func (s *Sim) uniqueDebuffReplicationTargets(playerID uint64, primary *entity) []replicatedDebuffTarget {
	targets := []replicatedDebuffTarget{}
	if primary == nil || primary.kind != monsterEntity {
		return targets
	}
	def, ok := s.liveUniqueEffect(replicatingBlightEffectID, "on_hero_damage_dealt")
	if !ok || !containsStringValue(s.equippedUniqueEffectIDs(playerID), replicatingBlightEffectID) {
		return targets
	}
	radius := uniqueEffectFloatParam(def, "replicate_radius_tiles", 0)
	if radius <= 0 {
		return targets
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.id == primary.id || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		if distance(primary.pos, target.pos) > radius+meleeRangeEpsilon {
			continue
		}
		targets = append(targets, replicatedDebuffTarget{entity: target})
	}
	return targets
}

type replicatedDebuffTarget struct {
	entity *entity
}

func (s *Sim) replicateUniqueBurnDot(playerID uint64, primary *entity, dot uniqueBurnDotState, res *TickResult) {
	if dot.EffectID == "" || dot.RemainingTicks <= 0 {
		return
	}
	for _, replicated := range s.uniqueDebuffReplicationTargets(playerID, primary) {
		target := replicated.entity
		clone := dot
		clone.TargetID = target.id
		s.uniqueBurnDots[uniqueBurnDotKey(clone.EffectID, target.id)] = clone
		target.effectIDs = sortedUniqueStrings(append(target.effectIDs, dot.EffectID))
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(target.id),
			SourceEntityID: idStr(playerID),
			TargetEntityID: idStr(target.id),
			CorrelationID:  dot.CorrelationID,
			SkillID:        dot.EffectID,
			Amount:         intPtr(dot.DamagePerTick),
			RemainingTicks: intPtr(dot.RemainingTicks),
			TotalTicks:     intPtr(dot.TotalTicks),
			DamageType:     dot.DamageType,
		})
	}
}

func (s *Sim) replicateSkillEffectToNearbyMonsters(playerID uint64, primary *entity, stateKey string, res *TickResult) {
	state, ok := s.skillEffects[stateKey]
	if !ok || state.EndsTick <= s.tick {
		return
	}
	effectID := state.EffectID
	if effectID == "" {
		effectID = state.SkillID
	}
	remainingTicks := int(state.EndsTick - s.tick)
	for _, replicated := range s.uniqueDebuffReplicationTargets(playerID, primary) {
		target := replicated.entity
		clone := state
		clone.TargetID = target.id
		s.skillEffects[fmt.Sprintf("replicated:%s:%d", stateKey, target.id)] = clone
		target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(target.id),
			SourceEntityID: idStr(playerID),
			TargetEntityID: idStr(target.id),
			SkillID:        state.SkillID,
			Amount:         intPtr(state.Percent),
			RemainingTicks: intPtr(remainingTicks),
			TotalTicks:     intPtr(state.TotalTicks),
		})
	}
}
