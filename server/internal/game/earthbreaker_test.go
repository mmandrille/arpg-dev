package game

import "testing"

func TestEarthbreakerDamagesRadialTargets(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_earthbreaker_radial", "01", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.BaseStats.Str = 12
	sim.progression.BaseStats.Vit = 10
	sim.progression.SkillRanks["cleave"] = 1
	sim.progression.SkillRanks["earthbreaker"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 20, Y: 20}
	player.mana = player.maxMana
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	front := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 2, Y: player.pos.Y}, hp: 20, maxHP: 20, monsterDefID: monsterDefID, lootTable: "no_drop"}
	back := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X - 2, Y: player.pos.Y}, hp: 20, maxHP: 20, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[front.id] = front
	sim.entities[back.id] = back

	cast := sim.Tick([]Input{{
		MessageID:     "cast_earthbreaker",
		CorrelationID: "corr_earthbreaker",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "earthbreaker", Direction: &Vec2{X: 1, Y: 0}},
	}})
	assertAck(t, cast, "cast_earthbreaker")
	if front.hp >= 20 || back.hp >= 20 {
		t.Fatalf("earthbreaker hp front=%d back=%d, want both radial targets damaged; events=%+v", front.hp, back.hp, cast.Events)
	}
	var castEvent *Event
	for i := range cast.Events {
		if cast.Events[i].EventType == "skill_cast" {
			castEvent = &cast.Events[i]
			break
		}
	}
	if castEvent == nil || castEvent.AngleDegrees == nil || *castEvent.AngleDegrees != 360 {
		t.Fatalf("earthbreaker cast event = %+v, want 360 degree radial geometry", castEvent)
	}
}
