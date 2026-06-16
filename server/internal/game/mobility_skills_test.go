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
