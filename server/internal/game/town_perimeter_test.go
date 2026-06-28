package game

import "testing"

func TestTownExitGateRejectsWhileClosed(t *testing.T) {
	sim, err := NewSimWithWorld("sess_town_exit_gate", "town_exit_gate_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon_levels: %v", err)
	}
	gate := findTownInteractable(t, sim, "town_exit_gate")
	sim.entities[sim.playerID].pos = Vec2{X: gate.pos.X, Y: gate.pos.Y - 1.5}
	result := sim.Tick([]Input{{MessageID: "gate", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(gate.id)}}})
	assertReject(t, result, "gate", "town_exit_locked")
	if gate.state != interactableClosed {
		t.Fatalf("gate state = %s, want closed", gate.state)
	}
}

func TestTownPerimeterBlocksMovement(t *testing.T) {
	sim, err := NewSimWithWorld("sess_town_perimeter_block", "town_perimeter_block_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon_levels: %v", err)
	}
	outside := Vec2{X: 11, Y: 29}
	sim.Tick([]Input{{MessageID: "outside", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: outside}}})
	for i := 0; i < 40; i++ {
		sim.Tick(nil)
	}
	got := sim.entities[sim.playerID].pos
	if got.Y > 26.5 {
		t.Fatalf("player escaped south past perimeter: pos=%+v", got)
	}
	if distance(got, outside) < 2.0 {
		t.Fatalf("player reached outside goal: pos=%+v want~%+v", got, outside)
	}
}

func TestTownServicesReachableFromSpawn(t *testing.T) {
	sim, err := NewSimWithWorld("sess_town_service_reach", "town_service_reach_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon_levels: %v", err)
	}
	player := sim.entities[sim.playerID]
	nav := sim.navigationForLevel(sim.activeLevel())
	blocked := sim.buildBlockedFn()
	vendor := findTownInteractable(t, sim, "town_vendor")
	if _, ok := sim.planPath(nav, player.pos, vendor.pos, blocked); !ok {
		t.Fatalf("no path from spawn %+v to vendor %+v", player.pos, vendor.pos)
	}
	stairs := findTownInteractable(t, sim, "stairs_down")
	if _, ok := sim.planPath(nav, player.pos, stairs.pos, blocked); !ok {
		t.Fatalf("no path from spawn %+v to stairs %+v", player.pos, stairs.pos)
	}
}

func findTownInteractable(t *testing.T, sim *Sim, defID string) *entity {
	t.Helper()
	level := sim.levels[townLevel]
	if level == nil {
		t.Fatalf("missing town level")
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == defID {
			return e
		}
	}
	t.Fatalf("missing interactable %s", defID)
	return nil
}
