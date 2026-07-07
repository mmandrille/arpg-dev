package game

import "fmt"

type StatusEffectDef struct {
	ID         string   `json:"id"`
	Kind       string   `json:"kind"`
	DamageType string   `json:"damage_type,omitempty"`
	Stats      []string `json:"stats,omitempty"`
	Debuff     bool     `json:"debuff"`
}

type statusEffectState struct {
	StatusID       string
	EventSkillID   string
	SourcePlayerID uint64
	TargetID       uint64
	Rank           int
	Stats          []string
	Percent        int
	MaxPercent     int
	DamageType     string
	DamagePerTick  int
	NextTick       uint64
	IntervalTicks  int
	RemainingTicks int
	TotalTicks     int
	CorrelationID  string
}

func cloneStatusEffects(in map[string]statusEffectState) map[string]statusEffectState {
	if len(in) == 0 {
		return make(map[string]statusEffectState)
	}
	out := make(map[string]statusEffectState, len(in))
	for key, effect := range in { //nolint:determinism -- pure map clone, output is a map
		effect.Stats = cloneStringSlice(effect.Stats)
		out[key] = effect
	}
	return out
}

func statusEffectKey(statusID string, targetID uint64, eventSkillID string) string {
	return fmt.Sprintf("%s:%d:%s", statusID, targetID, eventSkillID)
}

func validateStatusEffects(statuses map[string]StatusEffectDef) error {
	for statusID, status := range statuses {
		if status.ID != statusID {
			return fmt.Errorf("game: invalid rules status_effects.%s.id: must match key", statusID)
		}
		switch status.Kind {
		case "dot":
			if canonicalDamageType(status.DamageType) == "" {
				return fmt.Errorf("game: invalid rules status_effects.%s.damage_type: required", statusID)
			}
		case "slow":
			if len(status.Stats) == 0 {
				return fmt.Errorf("game: invalid rules status_effects.%s.stats: required", statusID)
			}
			for i, stat := range status.Stats {
				if stat != "movement_speed" && stat != "attack_speed" {
					return fmt.Errorf("game: invalid rules status_effects.%s.stats[%d]: unsupported %s", statusID, i, stat)
				}
			}
		default:
			return fmt.Errorf("game: invalid rules status_effects.%s.kind: unsupported %s", statusID, status.Kind)
		}
	}
	return nil
}

func (s *Sim) statusDef(statusID string) (StatusEffectDef, bool) {
	if s == nil || s.rules == nil || statusID == "" {
		return StatusEffectDef{}, false
	}
	status, ok := s.rules.StatusEffects[statusID]
	return status, ok
}

func (s *Sim) startDotStatus(target *entity, statusID, eventSkillID string, sourcePlayerID uint64, rank int, damagePerTick, intervalTicks, durationTicks int, corr string, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || damagePerTick < 0 || intervalTicks <= 0 || durationTicks <= 0 {
		return
	}
	def, ok := s.statusDef(statusID)
	if !ok || def.Kind != "dot" {
		return
	}
	key := statusEffectKey(statusID, target.id, eventSkillID)
	s.statusEffects[key] = statusEffectState{
		StatusID:       statusID,
		EventSkillID:   eventSkillID,
		SourcePlayerID: sourcePlayerID,
		TargetID:       target.id,
		Rank:           rank,
		DamageType:     canonicalDamageType(def.DamageType),
		DamagePerTick:  damagePerTick,
		NextTick:       s.tick + uint64(intervalTicks),
		IntervalTicks:  intervalTicks,
		RemainingTicks: durationTicks,
		TotalTicks:     durationTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, statusID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(sourcePlayerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        eventSkillID,
		Rank:           intPtr(rank),
		Amount:         intPtr(damagePerTick),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
		DamageType:     canonicalDamageType(def.DamageType),
	})
}

func (s *Sim) applySlowStatus(target *entity, statusID, eventSkillID string, sourcePlayerID uint64, percent, maxPercent, durationTicks int, corr string, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || percent <= 0 || durationTicks <= 0 {
		return
	}
	def, ok := s.statusDef(statusID)
	if !ok || def.Kind != "slow" {
		return
	}
	key := statusEffectKey(statusID, target.id, eventSkillID)
	current := 0
	if existing, ok := s.statusEffects[key]; ok && existing.RemainingTicks > 0 {
		current = existing.Percent
	}
	total := current + percent
	if maxPercent > 0 && total > maxPercent {
		total = maxPercent
	}
	s.statusEffects[key] = statusEffectState{
		StatusID:       statusID,
		EventSkillID:   eventSkillID,
		SourcePlayerID: sourcePlayerID,
		TargetID:       target.id,
		Stats:          cloneStringSlice(def.Stats),
		Percent:        total,
		MaxPercent:     maxPercent,
		RemainingTicks: durationTicks,
		TotalTicks:     durationTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, statusID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(sourcePlayerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        eventSkillID,
		Amount:         intPtr(total),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}

func (s *Sim) replicateDotStatus(playerID uint64, primary *entity, statusID, eventSkillID string, rank int, damagePerTick, intervalTicks, durationTicks int, corr string, res *TickResult) {
	for _, replicated := range s.uniqueDebuffReplicationTargets(playerID, primary) {
		s.startDotStatus(replicated.entity, statusID, eventSkillID, playerID, rank, damagePerTick, intervalTicks, durationTicks, corr, res)
	}
}

func (s *Sim) replicateSlowStatus(playerID uint64, primary *entity, statusID, eventSkillID string, percent, maxPercent, durationTicks int, corr string, res *TickResult) {
	for _, replicated := range s.uniqueDebuffReplicationTargets(playerID, primary) {
		s.applySlowStatus(replicated.entity, statusID, eventSkillID, playerID, percent, maxPercent, durationTicks, corr, res)
	}
}

func (s *Sim) advanceStatusEffects(res *TickResult) {
	if len(s.statusEffects) == 0 {
		return
	}
	for _, key := range sortedStringKeys(s.statusEffects) {
		effect := s.statusEffects[key]
		target := s.activeLevel().entities[effect.TargetID]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || effect.RemainingTicks <= 0 {
			s.endStatusEffect(key, effect, target, res)
			continue
		}
		if effect.DamagePerTick > 0 && effect.IntervalTicks > 0 && s.tick >= effect.NextTick {
			rawDamage := effect.DamagePerTick
			if effect.DamageType == damageTypePoison {
				markedDamage := s.applyRogueMarkDamageBonus(target, DamageRange{Min: rawDamage, Max: rawDamage})
				rawDamage = markedDamage.Min
			}
			damage := s.applyResistanceToDamage(rawDamage, s.monsterResistance(target, effect.DamageType))
			if damage > target.hp {
				damage = target.hp
			}
			target.hp -= damage
			effect.NextTick += uint64(effect.IntervalTicks)
			res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
			res.Events = append(res.Events, Event{
				EventType:       "monster_damaged",
				EntityID:        idStr(target.id),
				SourceEntityID:  idStr(effect.SourcePlayerID),
				TargetEntityID:  idStr(target.id),
				CorrelationID:   effect.CorrelationID,
				SkillID:         effect.EventSkillID,
				Rank:            intPtr(effect.Rank),
				Damage:          intPtr(damage),
				DamageType:      effect.DamageType,
				Outcome:         "hit",
				RawDamage:       intPtr(rawDamage),
				MitigatedDamage: intPtr(rawDamage),
			})
			if damage > 0 {
				s.tryPassiveExecute(target, effect.SourcePlayerID, effect.CorrelationID, res)
			}
			if target.hp == 0 {
				s.finishMonsterKill(target, effect.SourcePlayerID, effect.CorrelationID, res)
				s.endStatusEffect(key, effect, target, res)
				continue
			}
		}
		effect.RemainingTicks--
		if effect.RemainingTicks <= 0 {
			s.endStatusEffect(key, effect, target, res)
			continue
		}
		s.statusEffects[key] = effect
	}
}

func (s *Sim) endStatusEffect(key string, effect statusEffectState, target *entity, res *TickResult) {
	delete(s.statusEffects, key)
	if target != nil && !s.targetHasStatus(target.id, effect.StatusID) {
		target.effectIDs = removeStringValue(target.effectIDs, effect.StatusID)
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	}
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_ended",
		EntityID:       idStr(effect.TargetID),
		SourceEntityID: idStr(effect.SourcePlayerID),
		TargetEntityID: idStr(effect.TargetID),
		CorrelationID:  effect.CorrelationID,
		SkillID:        effect.EventSkillID,
	})
}

func (s *Sim) targetHasStatus(targetID uint64, statusID string) bool {
	for _, key := range sortedStringKeys(s.statusEffects) {
		effect := s.statusEffects[key]
		if effect.TargetID == targetID && effect.StatusID == statusID && effect.RemainingTicks > 0 {
			return true
		}
	}
	return false
}

func (s *Sim) targetStatusPercent(targetID uint64, stat string) int {
	maxPercent := 0
	for _, key := range sortedStringKeys(s.statusEffects) {
		effect := s.statusEffects[key]
		if effect.TargetID != targetID || effect.RemainingTicks <= 0 || effect.Percent <= maxPercent {
			continue
		}
		if !containsStringValue(effect.Stats, stat) {
			continue
		}
		maxPercent = effect.Percent
	}
	return maxPercent
}
