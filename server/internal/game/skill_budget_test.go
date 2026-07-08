package game

import "testing"

func TestSkillBudgetDefersOverflowCastsDeterministically(t *testing.T) {
	rules := loadRules(t)
	rules.Combat.CombatProcessing.SkillResolutionsPerTick = 1
	bolt := rules.Skills["magic_bolt"]
	bolt.Cooldown = SkillCooldownDef{Type: "none"}
	rules.Skills["magic_bolt"] = bolt
	sim := MustNewSim("sess_skill_budget", "skill_budget", rules)
	ps := sim.players[sim.playerID]
	ps.Progression.CharacterClass = "sorcerer"
	ps.Progression.SkillRanks["magic_bolt"] = 1
	sim.usePlayer(ps)
	player := sim.activeLevel().entities[sim.playerID]
	player.mana = player.maxMana
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 4, Y: player.pos.Y},
		hp:           30,
		maxHP:        30,
		monsterDefID: monsterDefID,
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster
	sim.activeLevel().entities[monster.id] = monster

	first := Input{
		MessageID:     "cast_a",
		Sequence:      1,
		Type:          "cast_skill_intent",
		ActorPlayerID: sim.playerID,
		CastSkill:     &CastSkillIntent{SkillID: "magic_bolt", Direction: &Vec2{X: 1}},
	}
	second := Input{
		MessageID:     "cast_b",
		Sequence:      2,
		Type:          "cast_skill_intent",
		ActorPlayerID: sim.playerID,
		CastSkill:     &CastSkillIntent{SkillID: "magic_bolt", Direction: &Vec2{X: 1}},
	}
	results := sim.TickResults([]Input{first, second})
	if !messageAcked(results, "cast_a") {
		t.Fatalf("first cast should ack immediately: %+v", results)
	}
	if messageAcked(results, "cast_b") {
		t.Fatalf("second cast should defer without ack on same tick: %+v", results)
	}
	next := sim.TickResults(nil)
	if !messageAcked(next, "cast_b") {
		t.Fatalf("deferred cast should ack on next tick: %+v", next)
	}
}

func messageAcked(results []TickResult, messageID string) bool {
	for _, res := range results {
		if hasAck(res, messageID) {
			return true
		}
	}
	return false
}
