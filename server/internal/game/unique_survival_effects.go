package game

import (
	"fmt"
	"math"
)

const (
	veilOfTheLastOathEffectID = "veil_of_the_last_oath"
	frostglassWardEffectID    = "frostglass_ward"
	mirrorsteelSkinEffectID   = "mirrorsteel_skin"
	ashenReprisalEffectID     = "ashen_reprisal"
)

type uniqueIncomingDamageSource struct {
	Projectile         bool
	MonsterAttackStyle string
}

type uniqueAshenReprisalState struct {
	EndsTick uint64
}

func (s *Sim) damagePlayerByMonster(monster *entity, player *entity, damageRange DamageRange, corr string, res *TickResult) combatResolution {
	return s.damagePlayerByMonsterWithSource(monster, player, damageRange, corr, res, uniqueIncomingDamageSource{MonsterAttackStyle: s.monsterCombatAttackStyle(monster)})
}

func (s *Sim) damagePlayerByMonsterWithSource(monster *entity, player *entity, damageRange DamageRange, corr string, res *TickResult, source uniqueIncomingDamageSource) combatResolution {
	if outcome, immune := s.playerDamageImmunityOutcome(player); immune {
		res.Events = append(res.Events, combatEventWithAttackStyle(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome, source.MonsterAttackStyle))
		return outcome
	}
	damageRange = s.applyEliteAuraToMonsterDamage(monster, damageRange)
	attackerStats := s.monsterEffectiveCombatStats(monster, damageRange)
	defenderStats, _ := s.playerEffectiveCombatStats()
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	if !outcome.Hit || outcome.Blocked {
		s.triggerUniqueEffectsAfterPlayerAvoidedHit(player, monster, corr, res)
		res.Events = append(res.Events, combatEventWithAttackStyle(s.combatEventType(playerEntity, outcome), monster.id, player.id, corr, outcome, source.MonsterAttackStyle))
		return outcome
	}
	outcome = s.applyUniqueEffectsBeforePlayerDamage(player, monster, corr, res, outcome, source)
	player.hp -= outcome.Damage
	if player.hp < 0 {
		player.hp = 0
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	eventType := "player_damaged"
	if player.hp == 0 {
		eventType = "player_killed"
	}
	res.Events = append(res.Events, combatEventWithAttackStyle(eventType, monster.id, player.id, corr, outcome, source.MonsterAttackStyle))
	s.triggerUniqueEffectsAfterPlayerDamage(player, monster, corr, res, outcome)
	return outcome
}

func (s *Sim) monsterCombatAttackStyle(monster *entity) string {
	if monster == nil || monster.kind != monsterEntity {
		return ""
	}
	def, ok := s.rules.Monsters[monster.monsterDefID]
	if !ok || def.effectiveAttackStyle() == monsterAttackStyleMelee {
		return ""
	}
	return def.effectiveAttackStyle()
}

func combatEventWithAttackStyle(eventType string, sourceID, targetID uint64, corr string, outcome combatResolution, attackStyle string) Event {
	event := combatEvent(eventType, sourceID, targetID, corr, outcome)
	event.AttackStyle = attackStyle
	return event
}

func cloneUniqueAshenReprisals(in map[uint64]uniqueAshenReprisalState) map[uint64]uniqueAshenReprisalState {
	if len(in) == 0 {
		return make(map[uint64]uniqueAshenReprisalState)
	}
	out := make(map[uint64]uniqueAshenReprisalState, len(in))
	for key, state := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = state
	}
	return out
}

func (s *Sim) triggerUniqueEffectsAfterPlayerAvoidedHit(player *entity, attacker *entity, corr string, res *TickResult) {
	if player == nil || player.kind != playerEntity || player.hp <= 0 {
		return
	}
	for _, effectID := range s.equippedUniqueEffectIDs(player.id) {
		if effectID != ashenReprisalEffectID {
			continue
		}
		def, ok := s.liveUniqueEffect(effectID, "on_block_or_evade")
		if !ok {
			continue
		}
		durationTicks := uniqueEffectIntParam(def, "primed_duration_seconds", 0) * 10
		if durationTicks <= 0 {
			continue
		}
		s.uniqueAshenReprisals[player.id] = uniqueAshenReprisalState{EndsTick: s.tick + uint64(durationTicks)}
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(player.id),
			SourceEntityID: idStr(entityID(attacker)),
			TargetEntityID: idStr(player.id),
			CorrelationID:  corr,
			SkillID:        def.ID,
			RemainingTicks: intPtr(durationTicks),
			TotalTicks:     intPtr(durationTicks),
		})
		return
	}
}

func (s *Sim) applyUniqueEffectsBeforePlayerDamage(player *entity, attacker *entity, corr string, res *TickResult, outcome combatResolution, source uniqueIncomingDamageSource) combatResolution {
	if player == nil || player.kind != playerEntity || player.hp <= 0 || outcome.Damage <= 0 {
		return outcome
	}
	for _, effectID := range s.equippedUniqueEffectIDs(player.id) {
		switch effectID {
		case mirrorsteelSkinEffectID:
			if source.Projectile {
				def, ok := s.liveUniqueEffect(effectID, "on_projectile_hit_taken")
				if ok {
					outcome = s.applyMirrorsteelSkin(player, attacker, def, corr, res, outcome)
				}
			}
		case veilOfTheLastOathEffectID:
			def, ok := s.liveUniqueEffect(effectID, "on_lethal_damage_taken")
			if ok {
				outcome = s.applyVeilOfTheLastOath(player, attacker, def, corr, res, outcome)
			}
		}
	}
	return outcome
}

func (s *Sim) triggerUniqueEffectsAfterPlayerDamage(player *entity, attacker *entity, corr string, res *TickResult, outcome combatResolution) {
	if player == nil || player.kind != playerEntity || outcome.Damage <= 0 {
		return
	}
	for _, effectID := range s.equippedUniqueEffectIDs(player.id) {
		if effectID != frostglassWardEffectID {
			continue
		}
		def, ok := s.liveUniqueEffect(effectID, "on_large_hit_taken")
		if ok {
			s.tryFrostglassWard(player, attacker, def, corr, res, outcome)
		}
	}
}

func (s *Sim) applyVeilOfTheLastOath(player *entity, attacker *entity, def UniqueEffectDef, corr string, res *TickResult, outcome combatResolution) combatResolution {
	if outcome.Damage < player.hp || !uniqueEffectBoolParam(def, "prevent_lethal_damage", false) {
		return outcome
	}
	if _, active := s.skillCooldownRemaining(def.ID); active {
		return outcome
	}
	durationTicks := uniqueEffectIntParam(def, "cloak_duration_seconds", 0) * 10
	cooldownTicks := uniqueEffectIntParam(def, "cooldown_seconds", 0) * 10
	statusID := uniqueEffectStringParam(def, "cloak_status_id", def.ID)
	if durationTicks <= 0 || cooldownTicks <= 0 || statusID == "" {
		return outcome
	}
	outcome.Damage = 0
	outcome.MitigatedDamage = outcome.RawDamage
	player.effectIDs = sortedUniqueStrings(append(player.effectIDs, statusID))
	s.skillEffects[fmt.Sprintf("%s:%d", def.ID, player.id)] = skillEffectState{
		SkillID:    def.ID,
		TargetID:   player.id,
		EffectID:   statusID,
		EndsTick:   s.tick + uint64(durationTicks),
		TotalTicks: durationTicks,
	}
	s.skillCooldowns[def.ID] = skillCooldownState{EndsTick: s.tick + uint64(cooldownTicks), TotalTicks: cooldownTicks}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendSkillCooldownUpdate(res)
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(player.id),
		SourceEntityID: idStr(entityID(attacker)),
		TargetEntityID: idStr(player.id),
		CorrelationID:  corr,
		SkillID:        def.ID,
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
	s.appendSkillCooldownStartedEvent(res, player, def.ID, corr, cooldownTicks)
	return outcome
}

func (s *Sim) applyMirrorsteelSkin(player *entity, attacker *entity, def UniqueEffectDef, corr string, res *TickResult, outcome combatResolution) combatResolution {
	if _, active := s.skillCooldownRemaining(def.ID); active {
		return outcome
	}
	reductionPercent := uniqueEffectIntParam(def, "damage_reduction_percent", 0)
	reflectPercent := uniqueEffectIntParam(def, "reflect_damage_percent", 0)
	cooldownTicks := uniqueEffectIntParam(def, "cooldown_seconds", 0) * 10
	if reductionPercent <= 0 || cooldownTicks <= 0 {
		return outcome
	}
	originalDamage := outcome.Damage
	reduced := originalDamage - percentOf(originalDamage, reductionPercent)
	if reduced < 0 {
		reduced = 0
	}
	outcome.Damage = reduced
	outcome.MitigatedDamage += originalDamage - reduced
	s.skillCooldowns[def.ID] = skillCooldownState{EndsTick: s.tick + uint64(cooldownTicks), TotalTicks: cooldownTicks}
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, def.ID, corr, cooldownTicks)
	if attacker != nil && attacker.kind == monsterEntity && attacker.hp > 0 && reflectPercent > 0 {
		reflect := percentOf(originalDamage, reflectPercent)
		if reflect < 1 {
			reflect = 1
		}
		s.applyUniqueDirectDamage(player.id, attacker, def.ID, reflect, damageTypeForce, corr, res)
	}
	return outcome
}

func (s *Sim) tryFrostglassWard(player *entity, attacker *entity, def UniqueEffectDef, corr string, res *TickResult, outcome combatResolution) {
	if _, active := s.skillCooldownRemaining(def.ID); active {
		return
	}
	threshold := uniqueEffectIntParam(def, "hit_damage_percent_max_hp_threshold", 0)
	if threshold <= 0 || percentOf(player.maxHP, threshold) > outcome.Damage {
		return
	}
	durationTicks := uniqueEffectIntParam(def, "duration_seconds", 0) * 10
	cooldownTicks := uniqueEffectIntParam(def, "cooldown_seconds", 0) * 10
	if durationTicks <= 0 || cooldownTicks <= 0 {
		return
	}
	armorPercent := uniqueEffectIntParam(def, "armor_bonus_percent", 0)
	if armorPercent > 0 {
		s.skillEffects[fmt.Sprintf("%s:%d", def.ID, player.id)] = skillEffectState{
			SkillID:    def.ID,
			TargetID:   player.id,
			Stats:      []string{"armor"},
			Percent:    armorPercent,
			EffectID:   def.ID,
			EndsTick:   s.tick + uint64(durationTicks),
			TotalTicks: durationTicks,
		}
		player.effectIDs = sortedUniqueStrings(append(player.effectIDs, def.ID))
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
		res.Events = append(res.Events, uniqueEffectStartedEvent(player, attacker, def.ID, corr, armorPercent, durationTicks))
	}
	s.applyFrostglassMonsterSlows(player, def, corr, res, durationTicks)
	s.skillCooldowns[def.ID] = skillCooldownState{EndsTick: s.tick + uint64(cooldownTicks), TotalTicks: cooldownTicks}
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, def.ID, corr, cooldownTicks)
}

func (s *Sim) applyFrostglassMonsterSlows(player *entity, def UniqueEffectDef, corr string, res *TickResult, durationTicks int) {
	statusID := uniqueEffectStringParam(def, "slow_status_id", "ice_slow")
	slowPercent := uniqueEffectIntParam(def, "slow_percent", 0)
	radius := uniqueEffectFloatParam(def, "slow_radius_tiles", 0)
	if statusID == "" || slowPercent <= 0 || radius <= 0 {
		return
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 || distance(player.pos, monster.pos) > radius+meleeRangeEpsilon {
			continue
		}
		stateKey := fmt.Sprintf("%s:%d", def.ID, monster.id)
		s.skillEffects[stateKey] = skillEffectState{
			SkillID:    def.ID,
			TargetID:   monster.id,
			Stats:      []string{"movement_speed"},
			Percent:    slowPercent,
			EffectID:   statusID,
			EndsTick:   s.tick + uint64(durationTicks),
			TotalTicks: durationTicks,
		}
		monster.effectIDs = sortedUniqueStrings(append(monster.effectIDs, statusID))
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(monster))})
		res.Events = append(res.Events, uniqueEffectStartedEvent(monster, player, def.ID, corr, slowPercent, durationTicks))
	}
}

func (s *Sim) applyAshenReprisalOnHeroHit(target *entity, playerID uint64, corr string, res *TickResult, sourceDamage int) {
	state, ok := s.uniqueAshenReprisals[playerID]
	if !ok || s.tick >= state.EndsTick || sourceDamage <= 0 {
		delete(s.uniqueAshenReprisals, playerID)
		return
	}
	def, ok := s.liveUniqueEffect(ashenReprisalEffectID, "on_block_or_evade")
	if !ok {
		delete(s.uniqueAshenReprisals, playerID)
		return
	}
	delete(s.uniqueAshenReprisals, playerID)
	bonusPercent := uniqueEffectIntParam(def, "bonus_fire_damage_percent", 0)
	if bonusPercent > 0 {
		bonus := percentOf(sourceDamage, bonusPercent)
		if bonus < 1 {
			bonus = 1
		}
		s.applyUniqueDirectDamage(playerID, target, def.ID, bonus, damageTypeFire, corr, res)
	}
	s.startAshenReprisalBurn(playerID, target, def, sourceDamage, corr, res)
}

func (s *Sim) startAshenReprisalBurn(playerID uint64, target *entity, def UniqueEffectDef, sourceDamage int, corr string, res *TickResult) {
	durationTicks := uniqueEffectIntParam(def, "burn_duration_seconds", 0) * 10
	intervalTicks := uniqueEffectIntParam(def, "burn_tick_interval_seconds", 0) * 10
	statusID := uniqueEffectStringParam(def, "burn_status_id", "burning")
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || sourceDamage <= 0 || durationTicks <= 0 || intervalTicks <= 0 || statusID == "" {
		return
	}
	damage := int(math.Round(float64(sourceDamage) * 0.10))
	if damage < 1 {
		damage = 1
	}
	s.uniqueBurnDots[uniqueBurnDotKey(def.ID, target.id)] = uniqueBurnDotState{
		SourcePlayerID: playerID,
		TargetID:       target.id,
		EffectID:       def.ID,
		DamageType:     damageTypeFire,
		DamagePerTick:  damage,
		NextTick:       s.tick + uint64(intervalTicks),
		IntervalTicks:  intervalTicks,
		RemainingTicks: durationTicks,
		TotalTicks:     durationTicks,
		CorrelationID:  corr,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, statusID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, uniqueEffectStartedEvent(target, nil, def.ID, corr, damage, durationTicks))
}

func uniqueEffectStartedEvent(target *entity, source *entity, skillID string, corr string, amount int, durationTicks int) Event {
	return Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(entityID(target)),
		SourceEntityID: idStr(entityID(source)),
		TargetEntityID: idStr(entityID(target)),
		CorrelationID:  corr,
		SkillID:        skillID,
		Amount:         intPtr(amount),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	}
}

func entityID(e *entity) uint64 {
	if e == nil {
		return 0
	}
	return e.id
}
