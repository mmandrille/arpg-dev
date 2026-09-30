package game

import "testing"

// Two monsters equally far along the aim direction must tie-break by entity
// ID, never by Go map iteration order (replay determinism).
func TestMonsterAlongDirectionTieBreaksByEntityID(t *testing.T) {
	rules := loadRules(t)
	for attempt := range 32 {
		sim := MustNewSim("sess_skill_aim_tie", "01", rules)
		player := sim.entities[sim.playerID]
		player.pos = Vec2{X: 2, Y: 5}
		for id, e := range sim.entities {
			if e.kind == monsterEntity {
				delete(sim.entities, id)
			}
		}
		// Symmetric perpendicular offsets inside the aim tolerance: identical "along".
		var ids []uint64
		for _, dy := range []float64{-0.1, 0.1} {
			m := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: 12, Y: 5 + dy}, hp: 20, maxHP: 20, monsterDefID: monsterDefID}
			sim.entities[m.id] = m
			ids = append(ids, m.id)
		}

		got := sim.monsterAlongDirectionBeyondRange(player, Vec2{X: 1}, 4)
		if got == nil {
			t.Fatal("expected an aim target beyond cast range")
		}
		if got.id != ids[0] {
			t.Fatalf("attempt %d: tie resolved to entity %d, want lowest id %d", attempt, got.id, ids[0])
		}
	}
}
