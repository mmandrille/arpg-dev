package game

import "testing"

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
	nearLanding := addRangerSkillMonster(sim, Vec2{X: 7, Y: 2}, 40)
	far := addRangerSkillMonster(sim, Vec2{X: 11, Y: 2}, 40)

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
	if nearLanding.hp >= nearLanding.maxHP || !containsStringValue(nearLanding.effectIDs, "leap_stun") {
		t.Fatalf("near landing hp/effects = %d/%d %v, want damaged and stunned", nearLanding.hp, nearLanding.maxHP, nearLanding.effectIDs)
	}
	if far.hp != far.maxHP || containsStringValue(far.effectIDs, "leap_stun") {
		t.Fatalf("far hp/effects = %d/%d %v, want unaffected", far.hp, far.maxHP, far.effectIDs)
	}
}

func TestPaladinChargeMovesDamagesAndStunsEndpointTargets(t *testing.T) {
	sim := mobilitySkillSim(t, "paladin", "charge")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	endpoint := addRangerSkillMonster(sim, Vec2{X: 8, Y: 2}, 40)
	far := addRangerSkillMonster(sim, Vec2{X: 11, Y: 2}, 40)

	cast := sim.Tick([]Input{{
		MessageID:     "charge",
		CorrelationID: "corr_charge",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "charge", Direction: &Vec2{X: 1}},
	}})

	assertAck(t, cast, "charge")
	if !hasEvent(cast, "skill_cast") || !hasSkillDamageEvent(cast, "charge") || !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("charge events = %+v", cast.Events)
	}
	if player.pos.X <= 2 {
		t.Fatalf("charge player x = %.2f, want moved forward", player.pos.X)
	}
	if endpoint.hp >= endpoint.maxHP || !containsStringValue(endpoint.effectIDs, "charge_stun") {
		t.Fatalf("endpoint hp/effects = %d/%d %v, want damaged and stunned", endpoint.hp, endpoint.maxHP, endpoint.effectIDs)
	}
	if far.hp != far.maxHP || containsStringValue(far.effectIDs, "charge_stun") {
		t.Fatalf("far hp/effects = %d/%d %v, want unaffected", far.hp, far.maxHP, far.effectIDs)
	}
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
