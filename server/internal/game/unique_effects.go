package game

import (
	"fmt"
	"math"
)

const everburningWoundEffectID = "everburning_wound"

type uniqueBurnDotState struct {
	SourcePlayerID uint64
	TargetID       uint64
	EffectID       string
	DamageType     string
	DamagePerTick  int
	NextTick       uint64
	IntervalTicks  int
	RemainingTicks int
	TotalTicks     int
	CorrelationID  string
}

func cloneUniqueBurnDots(in map[string]uniqueBurnDotState) map[string]uniqueBurnDotState {
	if len(in) == 0 {
		return make(map[string]uniqueBurnDotState)
	}
	out := make(map[string]uniqueBurnDotState, len(in))
	for key, dot := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = dot
	}
	return out
}

func uniqueBurnDotKey(effectID string, targetID uint64) string {
	return fmt.Sprintf("%s:%d", effectID, targetID)
}

func (s *Sim) triggerUniqueEffectsAfterHeroDamage(target *entity, playerID uint64, corr string, res *TickResult, outcome combatResolution) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || outcome.Damage <= 0 {
		return
	}
	for _, effectID := range s.equippedUniqueEffectIDs(playerID) {
		if effectID != everburningWoundEffectID {
			continue
		}
		def, ok := s.rules.UniqueEffects[effectID]
		if !ok || !def.Enabled || def.Status != "ready" || def.Hook != "on_hero_damage_dealt" {
			continue
		}
		s.startUniqueBurnDot(playerID, target, def, outcome.Damage, corr, res)
	}
}

func (s *Sim) equippedUniqueEffectIDs(playerID uint64) []string {
	ps := s.players[playerID]
	if ps == nil || len(ps.Equipped) == 0 || len(ps.Inventory) == 0 {
		return nil
	}
	effectIDs := []string{}
	for _, slot := range sortedStringKeys(ps.Equipped) {
		instanceID := ps.Equipped[slot]
		if instanceID == 0 {
			continue
		}
		item := inventoryItemByID(ps.Inventory, instanceID)
		if item == nil || item.rollPayload == nil {
			continue
		}
		effectIDs = append(effectIDs, item.rollPayload.EffectIDs...)
	}
	return sortedUniqueStrings(effectIDs)
}

func inventoryItemByID(items []*invItem, instanceID uint64) *invItem {
	for _, item := range items {
		if item != nil && item.instanceID == instanceID {
			return item
		}
	}
	return nil
}

func (s *Sim) startUniqueBurnDot(playerID uint64, target *entity, def UniqueEffectDef, sourceDamage int, corr string, res *TickResult) {
	if target == nil || sourceDamage < 0 {
		return
	}
	damageType := canonicalDamageType(uniqueEffectStringParam(def, "damage_type", damageTypeFire))
	percent := uniqueEffectIntParam(def, "tick_damage_percent_of_original_hit", 0)
	durationSeconds := uniqueEffectIntParam(def, "duration_seconds", 0)
	intervalSeconds := uniqueEffectIntParam(def, "tick_interval_seconds", 0)
	statusID := uniqueEffectStringParam(def, "status_id", "burning")
	if percent <= 0 || durationSeconds <= 0 || intervalSeconds <= 0 || statusID == "" {
		return
	}
	damage := int(math.Round(float64(sourceDamage) * float64(percent) / 100.0))
	if sourceDamage > 0 && damage < 1 {
		damage = 1
	}
	intervalTicks := intervalSeconds * 10
	totalTicks := durationSeconds * 10
	if intervalTicks < 1 || totalTicks < 1 {
		return
	}
	s.uniqueBurnDots[uniqueBurnDotKey(def.ID, target.id)] = uniqueBurnDotState{
		SourcePlayerID: playerID,
		TargetID:       target.id,
		EffectID:       def.ID,
		DamageType:     damageType,
		DamagePerTick:  damage,
		NextTick:       s.tick + uint64(intervalTicks),
		IntervalTicks:  intervalTicks,
		RemainingTicks: totalTicks,
		TotalTicks:     totalTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, def.ID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(playerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        def.ID,
		Amount:         intPtr(damage),
		RemainingTicks: intPtr(totalTicks),
		TotalTicks:     intPtr(totalTicks),
		DamageType:     damageType,
	})
}

func (s *Sim) advanceUniqueBurnDots(res *TickResult) {
	if len(s.uniqueBurnDots) == 0 {
		return
	}
	for _, key := range sortedStringKeys(s.uniqueBurnDots) {
		dot := s.uniqueBurnDots[key]
		target := s.activeLevel().entities[dot.TargetID]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || dot.RemainingTicks <= 0 {
			delete(s.uniqueBurnDots, key)
			continue
		}
		if s.tick < dot.NextTick {
			continue
		}
		rawDamage := dot.DamagePerTick
		damage := s.applyResistanceToDamage(rawDamage, s.monsterResistance(target, dot.DamageType))
		if damage > target.hp {
			damage = target.hp
		}
		target.hp -= damage
		if dot.IntervalTicks <= 0 {
			dot.IntervalTicks = 10
		}
		dot.RemainingTicks -= dot.IntervalTicks
		dot.NextTick += uint64(dot.IntervalTicks)
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:       "monster_damaged",
			EntityID:        idStr(target.id),
			SourceEntityID:  idStr(dot.SourcePlayerID),
			TargetEntityID:  idStr(target.id),
			CorrelationID:   dot.CorrelationID,
			SkillID:         dot.EffectID,
			Damage:          intPtr(damage),
			DamageType:      dot.DamageType,
			Outcome:         "hit",
			RawDamage:       intPtr(rawDamage),
			MitigatedDamage: intPtr(rawDamage),
		})
		if target.hp == 0 {
			s.finishMonsterKill(target, dot.SourcePlayerID, dot.CorrelationID, res)
			delete(s.uniqueBurnDots, key)
			continue
		}
		if dot.RemainingTicks <= 0 {
			s.endUniqueBurnDot(target, dot, key, res)
			continue
		}
		s.uniqueBurnDots[key] = dot
	}
}

func (s *Sim) endUniqueBurnDot(target *entity, dot uniqueBurnDotState, key string, res *TickResult) {
	delete(s.uniqueBurnDots, key)
	if target != nil {
		target.effectIDs = removeStringValue(target.effectIDs, dot.EffectID)
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	}
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_ended",
		EntityID:       idStr(dot.TargetID),
		SourceEntityID: idStr(dot.SourcePlayerID),
		TargetEntityID: idStr(dot.TargetID),
		CorrelationID:  dot.CorrelationID,
		SkillID:        dot.EffectID,
	})
}

func uniqueEffectIntParam(def UniqueEffectDef, key string, fallback int) int {
	value, ok := def.Params[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	default:
		return fallback
	}
}

func uniqueEffectStringParam(def UniqueEffectDef, key string, fallback string) string {
	value, ok := def.Params[key]
	if !ok {
		return fallback
	}
	typed, ok := value.(string)
	if !ok || typed == "" {
		return fallback
	}
	return typed
}
