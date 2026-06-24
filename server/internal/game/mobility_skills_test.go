package game

import (
	"math"
	"testing"
)

func TestSorcererTeleportMovesWithoutDamage(t *testing.T) {
	sim := mobilitySkillSim(t, "sorcerer", "teleport")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	target := addRangerSkillMonster(sim, Vec2{X: 4, Y: 2}, 40)
	startX := player.pos.X

	cast := sim.Tick([]Input{{
		MessageID:     "teleport",
		CorrelationID: "corr_teleport",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "teleport", Direction: &Vec2{X: 1}},
	}})

	assertAck(t, cast, "teleport")
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("teleport events = %+v", cast.Events)
	}
	if player.pos.X <= startX {
		t.Fatalf("teleport player x = %.2f, want moved from %.2f", player.pos.X, startX)
	}
	if target.hp != target.maxHP {
		t.Fatalf("teleport target hp = %d/%d, want no damage", target.hp, target.maxHP)
	}
}

func TestBarbarianLeapMovesDamagesAndStunsLandingTargets(t *testing.T) {
	sim := mobilitySkillSim(t, "barbarian", "leap")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	nearLanding := addRangerSkillMonster(sim, Vec2{X: 10, Y: 2}, 40)
	far := addRangerSkillMonster(sim, Vec2{X: 13, Y: 2}, 40)

	cast := sim.Tick([]Input{{
		MessageID:     "leap",
		CorrelationID: "corr_leap",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "leap", Direction: &Vec2{X: 1}},
	}})

	assertAck(t, cast, "leap")
	if !hasEvent(cast, "skill_cast") || !hasSkillDamageEvent(cast, "leap") || !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("leap events = %+v", cast.Events)
	}
	if player.pos.X <= 2 {
		t.Fatalf("leap player x = %.2f, want moved forward", player.pos.X)
	}
	if nearLanding.hp >= nearLanding.maxHP || !containsStringValue(nearLanding.effectIDs, "stun") {
		t.Fatalf("near landing hp/effects = %d/%d %v, want damaged and stunned", nearLanding.hp, nearLanding.maxHP, nearLanding.effectIDs)
	}
	if far.hp != far.maxHP || containsStringValue(far.effectIDs, "stun") {
		t.Fatalf("far hp/effects = %d/%d %v, want unaffected", far.hp, far.maxHP, far.effectIDs)
	}
}

func TestPaladinChargeMovesDamagesStunsAndPushesLineTargets(t *testing.T) {
	sim := mobilitySkillSim(t, "paladin", "charge")
	def := sim.rules.Skills["charge"]
	def.Mobility.ChannelManaPer10Sec = 200
	sim.rules.Skills["charge"] = def
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	player.mana = player.maxMana
	lineTarget := addRangerSkillMonster(sim, Vec2{X: 5, Y: 2}, 40)
	turnTarget := addRangerSkillMonster(sim, Vec2{X: 4.5, Y: 5}, 40)
	far := addRangerSkillMonster(sim, Vec2{X: 11, Y: 8}, 40)
	lineStartX := lineTarget.pos.X
	startMana := player.mana

	start := sim.Tick([]Input{{
		MessageID:     "charge_start",
		CorrelationID: "corr_charge",
		Type:          "channel_skill_intent",
		ChannelSkill:  &ChannelSkillIntent{SkillID: "charge", Phase: "start", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, start, "charge_start")
	if !hasEvent(start, "skill_channel_started") {
		t.Fatalf("charge start events = %+v", start.Events)
	}

	events := append([]Event{}, start.Events...)
	update := sim.Tick([]Input{{
		MessageID:     "charge_update",
		CorrelationID: "corr_charge",
		Type:          "channel_skill_intent",
		ChannelSkill:  &ChannelSkillIntent{SkillID: "charge", Phase: "update", Direction: &Vec2{Y: 1}},
	}})
	assertAck(t, update, "charge_update")
	events = append(events, update.Events...)
	for i := 0; i < 2; i++ {
		tick := sim.Tick(nil)
		events = append(events, tick.Events...)
	}
	stop := sim.Tick([]Input{{
		MessageID:     "charge_stop",
		CorrelationID: "corr_charge",
		Type:          "channel_skill_intent",
		ChannelSkill:  &ChannelSkillIntent{SkillID: "charge", Phase: "stop"},
	}})
	assertAck(t, stop, "charge_stop")
	events = append(events, stop.Events...)
	if !hasEvent(stop, "skill_channel_ended") {
		t.Fatalf("charge stop events = %+v", stop.Events)
	}
	if hasEvent(start, "skill_cooldown_started") || hasEvent(update, "skill_cooldown_started") || hasEvent(stop, "skill_cooldown_started") || len(skillCooldownUpdate(start)) > 0 || len(skillCooldownUpdate(update)) > 0 || len(skillCooldownUpdate(stop)) > 0 {
		t.Fatalf("charge should not start cooldowns, start=%+v update=%+v stop=%+v", start.Events, update.Events, stop.Events)
	}
	if player.pos.X <= 2 {
		t.Fatalf("charge player x = %.2f, want moved forward", player.pos.X)
	}
	if player.pos.Y <= 2 {
		t.Fatalf("charge player y = %.2f, want turned upward", player.pos.Y)
	}
	if player.mana >= startMana {
		t.Fatalf("charge mana = %d, want spent below %d", player.mana, startMana)
	}
	if lineTarget.hp >= lineTarget.maxHP || lineTarget.pos.X <= lineStartX || !hasEventInSlice(events, "skill_effect_started") || !hasEventInSlice(events, "monster_pushed") {
		t.Fatalf("line target hp/effects/pos/events = %d/%d %v %.2f %+v, want damaged, stunned, and pushed", lineTarget.hp, lineTarget.maxHP, lineTarget.effectIDs, lineTarget.pos.X, events)
	}
	if turnTarget.hp >= turnTarget.maxHP {
		t.Fatalf("turn target hp/effects = %d/%d %v, want damaged", turnTarget.hp, turnTarget.maxHP, turnTarget.effectIDs)
	}
	if far.hp != far.maxHP || containsStringValue(far.effectIDs, "stun") {
		t.Fatalf("far hp/effects = %d/%d %v, want unaffected", far.hp, far.maxHP, far.effectIDs)
	}
}

func TestPaladinChargeChannelSpeedScalesFromPlayerMoveSpeed(t *testing.T) {
	sim := mobilitySkillSim(t, "paladin", "charge")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	def := sim.rules.Skills["charge"]
	if def.Mobility.SpeedMultiplier != 2.5 {
		t.Fatalf("charge speed multiplier = %.2f, want 2.5", def.Mobility.SpeedMultiplier)
	}
	start := player.pos
	result := sim.Tick([]Input{{
		MessageID:    "charge_start",
		Type:         "channel_skill_intent",
		ChannelSkill: &ChannelSkillIntent{SkillID: "charge", Phase: "start", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, result, "charge_start")
	want := start.X + sim.playerMoveSpeed()*def.Mobility.SpeedMultiplier
	if math.Abs(player.pos.X-want) > 0.001 {
		t.Fatalf("charge x = %.4f, want %.4f from player speed %.4f * multiplier %.2f", player.pos.X, want, sim.playerMoveSpeed(), def.Mobility.SpeedMultiplier)
	}
}

func hasEventInSlice(events []Event, eventType string) bool {
	for _, e := range events {
		if e.EventType == eventType {
			return true
		}
	}
	return false
}

func TestRangerDisengageMovesAndSnaresStartTargets(t *testing.T) {
	sim := mobilitySkillSim(t, "ranger", "disengage")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 8, Y: 2}
	pursuer := addRangerSkillMonster(sim, Vec2{X: 7, Y: 2}, 40)
	endpoint := addRangerSkillMonster(sim, Vec2{X: 3, Y: 2}, 40)

	cast := sim.Tick([]Input{{
		MessageID:     "disengage",
		CorrelationID: "corr_disengage",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "disengage", Direction: &Vec2{X: -1}},
	}})

	assertAck(t, cast, "disengage")
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("disengage events = %+v", cast.Events)
	}
	if hasSkillDamageEvent(cast, "disengage") {
		t.Fatalf("disengage should not damage, events = %+v", cast.Events)
	}
	if player.pos.X >= 8 {
		t.Fatalf("disengage player x = %.2f, want moved backward", player.pos.X)
	}
	if pursuer.hp != pursuer.maxHP || !containsStringValue(pursuer.effectIDs, "disengage_snare") {
		t.Fatalf("pursuer hp/effects = %d/%d %v, want snared without damage", pursuer.hp, pursuer.maxHP, pursuer.effectIDs)
	}
	if endpoint.hp != endpoint.maxHP || containsStringValue(endpoint.effectIDs, "disengage_snare") {
		t.Fatalf("endpoint hp/effects = %d/%d %v, want unaffected", endpoint.hp, endpoint.maxHP, endpoint.effectIDs)
	}
}

func mobilitySkillSim(t *testing.T, classID string, skillID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim := MustNewSim("sess_"+skillID, skillID+"_seed", rules)
	sim.progression.CharacterClass = classID
	sim.progression.BaseStats = rules.CharacterProgression.Classes[classID].BaseStats
	sim.progression.BaseStats.Magic = 20
	sim.progression.BaseStats.Str = 20
	sim.progression.BaseStats.Dex = 20
	sim.progression.SkillRanks[skillID] = 1
	ps := sim.defaultPlayer()
	ps.Progression = sim.progression
	player := sim.activeLevel().entities[sim.playerID]
	player.maxMana = 100
	player.mana = 100
	for id, e := range sim.activeLevel().entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	return sim
}
