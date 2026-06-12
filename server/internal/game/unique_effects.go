package game

import (
	"fmt"
	"math"
)

const everburningWoundEffectID = "everburning_wound"

const (
	stormboundEchoEffectID   = "stormbound_echo"
	executionersMarkEffectID = "executioners_mark"
	hungerOfTheDeepEffectID  = "hunger_of_the_deep"
)

type uniqueHeroDamageSource struct {
	BasicAttack bool
}

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

type uniqueExecutionMarkState struct {
	SourcePlayerID uint64
	TargetID       uint64
	EffectID       string
	DamageType     string
	Damage         int
	Radius         float64
	EndsTick       uint64
	CorrelationID  string
}

type uniqueHungerStackState struct {
	TargetID uint64
	Stacks   int
	EndsTick uint64
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

func cloneUniqueExecutionMarks(in map[uint64]uniqueExecutionMarkState) map[uint64]uniqueExecutionMarkState {
	if len(in) == 0 {
		return make(map[uint64]uniqueExecutionMarkState)
	}
	out := make(map[uint64]uniqueExecutionMarkState, len(in))
	for key, mark := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = mark
	}
	return out
}

func cloneUniqueHungerStacks(in map[uint64]uniqueHungerStackState) map[uint64]uniqueHungerStackState {
	if len(in) == 0 {
		return make(map[uint64]uniqueHungerStackState)
	}
	out := make(map[uint64]uniqueHungerStackState, len(in))
	for key, stack := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = stack
	}
	return out
}

func uniqueBurnDotKey(effectID string, targetID uint64) string {
	return fmt.Sprintf("%s:%d", effectID, targetID)
}

func (s *Sim) applyUniqueDamageBeforeHeroHit(target *entity, playerID uint64, damageRange DamageRange) DamageRange {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 {
		return damageRange
	}
	damageRange = s.applyPilgrimMomentumBeforeHeroHit(target, playerID, damageRange)
	for _, effectID := range s.equippedUniqueEffectIDs(playerID) {
		if effectID != hungerOfTheDeepEffectID {
			continue
		}
		def, ok := s.liveUniqueEffect(effectID, "on_repeated_same_target_hit")
		if !ok {
			continue
		}
		stack := s.uniqueHungerStacks[playerID]
		if stack.TargetID != target.id || stack.Stacks <= 0 || s.tick >= stack.EndsTick {
			continue
		}
		bonusPercent := uniqueEffectIntParam(def, "damage_bonus_percent_per_stack", 0) * stack.Stacks
		if bonusPercent <= 0 {
			continue
		}
		damageRange.Min = applyPercentBonus(damageRange.Min, bonusPercent)
		damageRange.Max = applyPercentBonus(damageRange.Max, bonusPercent)
	}
	return damageRange
}

func (s *Sim) triggerUniqueEffectsAfterHeroDamage(target *entity, playerID uint64, corr string, res *TickResult, outcome combatResolution, source uniqueHeroDamageSource) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || outcome.Damage <= 0 {
		return
	}
	for _, effectID := range s.equippedUniqueEffectIDs(playerID) {
		switch effectID {
		case everburningWoundEffectID:
			def, ok := s.liveUniqueEffect(effectID, "on_hero_damage_dealt")
			if ok {
				s.startUniqueBurnDot(playerID, target, def, outcome.Damage, corr, res)
			}
		case stormboundEchoEffectID:
			if source.BasicAttack {
				def, ok := s.liveUniqueEffect(effectID, "on_basic_attack_hit")
				if ok {
					s.tryStormboundEcho(playerID, target, def, outcome.Damage, corr, res)
				}
			}
		case executionersMarkEffectID:
			def, ok := s.liveUniqueEffect(effectID, "on_hero_damage_dealt")
			if ok {
				s.tryStartExecutionersMark(playerID, target, def, outcome.Damage, corr, res)
			}
		case hungerOfTheDeepEffectID:
			def, ok := s.liveUniqueEffect(effectID, "on_repeated_same_target_hit")
			if ok {
				s.updateHungerOfTheDeep(playerID, target, def)
			}
		case ashenReprisalEffectID:
			s.applyAshenReprisalOnHeroHit(target, playerID, corr, res, outcome.Damage)
		case pilgrimsMomentumEffectID:
			s.triggerPilgrimMomentumAfterHeroHit(target, playerID, corr, res, outcome)
		}
	}
}

func (s *Sim) triggerUniqueEffectsOnMonsterKilled(monster *entity, sourceID uint64, corr string, res *TickResult) {
	if monster == nil || monster.kind != monsterEntity {
		return
	}
	s.triggerResourceUniqueEffectsOnMonsterKilled(monster, sourceID, corr, res)
	mark, ok := s.uniqueExecutionMarks[monster.id]
	if !ok || mark.EffectID != executionersMarkEffectID || s.tick >= mark.EndsTick {
		delete(s.uniqueExecutionMarks, monster.id)
		removeEffectIDAndUpdate(monster, mark.EffectID, s, res)
		return
	}
	s.pulseExecutionersMark(monster, mark, corr, res)
	delete(s.uniqueExecutionMarks, monster.id)
	removeEffectIDAndUpdate(monster, mark.EffectID, s, res)
}

func (s *Sim) liveUniqueEffect(effectID string, hook string) (UniqueEffectDef, bool) {
	def, ok := s.rules.UniqueEffects[effectID]
	if !ok || !def.Enabled || def.Status != "ready" || def.Hook != hook {
		return UniqueEffectDef{}, false
	}
	return def, true
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

func (s *Sim) tryStormboundEcho(playerID uint64, primary *entity, def UniqueEffectDef, sourceDamage int, corr string, res *TickResult) {
	chancePercent := uniqueEffectIntParam(def, "trigger_chance_percent", 0)
	if chancePercent <= 0 || !s.rollChance(float64(chancePercent)/100.0) {
		return
	}
	target := s.nearestUniqueEffectTarget(primary, float64(uniqueEffectIntParam(def, "search_radius_tiles", 0)))
	if target == nil {
		return
	}
	damage := percentOf(sourceDamage, uniqueEffectIntParam(def, "chain_damage_percent", 0))
	if damage <= 0 {
		return
	}
	damageType := canonicalDamageType(uniqueEffectStringParam(def, "damage_type", damageTypeLightning))
	s.applyUniqueDirectDamage(playerID, target, def.ID, damage, damageType, corr, res)
}

func (s *Sim) nearestUniqueEffectTarget(primary *entity, radius float64) *entity {
	if primary == nil || radius <= 0 {
		return nil
	}
	var best *entity
	bestDist := 0.0
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		candidate := s.activeLevel().entities[id]
		if candidate == nil || candidate.id == primary.id || candidate.kind != monsterEntity || candidate.hp <= 0 {
			continue
		}
		dist := distance(primary.pos, candidate.pos)
		if dist > radius+meleeRangeEpsilon {
			continue
		}
		if best == nil || dist < bestDist-1e-9 || (math.Abs(dist-bestDist) <= 1e-9 && candidate.id < best.id) {
			best = candidate
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) tryStartExecutionersMark(playerID uint64, target *entity, def UniqueEffectDef, sourceDamage int, corr string, res *TickResult) {
	if target == nil || target.maxHP <= 0 || sourceDamage <= 0 {
		return
	}
	threshold := uniqueEffectIntParam(def, "target_hp_percent_threshold", 0)
	if threshold <= 0 || target.hp*100 > target.maxHP*threshold {
		return
	}
	durationTicks := uniqueEffectIntParam(def, "mark_duration_seconds", 0) * 10
	if durationTicks <= 0 {
		return
	}
	pulseDamage := percentOf(sourceDamage, uniqueEffectIntParam(def, "pulse_damage_percent_of_marking_hit", 0))
	if pulseDamage <= 0 {
		return
	}
	s.uniqueExecutionMarks[target.id] = uniqueExecutionMarkState{
		SourcePlayerID: playerID,
		TargetID:       target.id,
		EffectID:       def.ID,
		DamageType:     canonicalDamageType(uniqueEffectStringParam(def, "damage_type", damageTypeForce)),
		Damage:         pulseDamage,
		Radius:         float64(uniqueEffectIntParam(def, "pulse_radius_tiles", 0)),
		EndsTick:       s.tick + uint64(durationTicks),
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
		Amount:         intPtr(pulseDamage),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
		DamageType:     canonicalDamageType(uniqueEffectStringParam(def, "damage_type", damageTypeForce)),
	})
}

func (s *Sim) pulseExecutionersMark(dead *entity, mark uniqueExecutionMarkState, corr string, res *TickResult) {
	if mark.Radius <= 0 || mark.Damage <= 0 {
		return
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.id == dead.id || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		if distance(dead.pos, target.pos) > mark.Radius+meleeRangeEpsilon {
			continue
		}
		s.applyUniqueDirectDamage(mark.SourcePlayerID, target, mark.EffectID, mark.Damage, mark.DamageType, corr, res)
	}
}

func (s *Sim) applyUniqueDirectDamage(playerID uint64, target *entity, effectID string, rawDamage int, damageType string, corr string, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || rawDamage <= 0 {
		return
	}
	damageType = canonicalDamageType(damageType)
	damage := s.applyResistanceToDamage(rawDamage, s.monsterResistance(target, damageType))
	if damage > target.hp {
		damage = target.hp
	}
	target.hp -= damage
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:       "monster_damaged",
		EntityID:        idStr(target.id),
		SourceEntityID:  idStr(playerID),
		TargetEntityID:  idStr(target.id),
		CorrelationID:   corr,
		SkillID:         effectID,
		Damage:          intPtr(damage),
		DamageType:      damageType,
		Outcome:         "hit",
		RawDamage:       intPtr(rawDamage),
		MitigatedDamage: intPtr(rawDamage),
	})
	if target.hp == 0 {
		s.finishMonsterKill(target, playerID, corr, res)
	}
}

func (s *Sim) updateHungerOfTheDeep(playerID uint64, target *entity, def UniqueEffectDef) {
	if target == nil {
		return
	}
	maxStacks := uniqueEffectIntParam(def, "max_stacks", 0)
	expireTicks := uniqueEffectIntParam(def, "stack_expire_seconds", 0) * 10
	if maxStacks <= 0 || expireTicks <= 0 {
		return
	}
	stack := s.uniqueHungerStacks[playerID]
	if stack.TargetID != target.id || s.tick >= stack.EndsTick {
		stack = uniqueHungerStackState{TargetID: target.id}
	}
	stack.Stacks++
	if stack.Stacks > maxStacks {
		stack.Stacks = maxStacks
	}
	stack.EndsTick = s.tick + uint64(expireTicks)
	s.uniqueHungerStacks[playerID] = stack
}

func (s *Sim) advanceOffensiveUniqueEffectStates(res *TickResult) {
	for _, targetID := range sortedUint64Keys(s.uniqueExecutionMarks) {
		mark := s.uniqueExecutionMarks[targetID]
		if s.tick < mark.EndsTick {
			continue
		}
		delete(s.uniqueExecutionMarks, targetID)
		target := s.activeLevel().entities[targetID]
		removeEffectIDAndUpdate(target, mark.EffectID, s, res)
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_ended",
			EntityID:       idStr(targetID),
			SourceEntityID: idStr(mark.SourcePlayerID),
			TargetEntityID: idStr(targetID),
			CorrelationID:  mark.CorrelationID,
			SkillID:        mark.EffectID,
		})
	}
	for _, playerID := range sortedUint64Keys(s.uniqueHungerStacks) {
		stack := s.uniqueHungerStacks[playerID]
		if s.tick >= stack.EndsTick {
			delete(s.uniqueHungerStacks, playerID)
		}
	}
}

func removeEffectIDAndUpdate(target *entity, effectID string, s *Sim, res *TickResult) {
	if target == nil || effectID == "" {
		return
	}
	before := len(target.effectIDs)
	target.effectIDs = removeStringValue(target.effectIDs, effectID)
	if len(target.effectIDs) != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	}
}

func percentOf(value int, percent int) int {
	if value <= 0 || percent <= 0 {
		return 0
	}
	result := int(math.Round(float64(value) * float64(percent) / 100.0))
	if result < 1 {
		return 1
	}
	return result
}

func applyPercentBonus(value int, percent int) int {
	if value <= 0 || percent <= 0 {
		return value
	}
	result := int(math.Round(float64(value) * (1.0 + float64(percent)/100.0)))
	if result < value+1 {
		return value + 1
	}
	return result
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

func uniqueEffectFloatParam(def UniqueEffectDef, key string, fallback float64) float64 {
	value, ok := def.Params[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	default:
		return fallback
	}
}

func uniqueEffectBoolParam(def UniqueEffectDef, key string, fallback bool) bool {
	value, ok := def.Params[key]
	if !ok {
		return fallback
	}
	typed, ok := value.(bool)
	if !ok {
		return fallback
	}
	return typed
}
