package game

import "testing"

func TestGeneratedWallLabStairsOffsetMoveGoalV40(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_wall_v40", "v40_obstacles", rules, "generated_wall_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	assertStairsDescendReachable(t, sim)
}

func TestGeneratedWallLabStairsOffsetMoveGoal(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_wall_floor_offset", "wall_seed_00", rules, "generated_wall_reachability_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	assertStairsDescendReachable(t, sim)
}

func assertStairsDescendReachable(t *testing.T, sim *Sim) {
	t.Helper()
	player := sim.activeLevel().entities[sim.playerID]
	var stairs *entity
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == stairsDownDefID {
			stairs = e
			break
		}
	}
	if stairs == nil {
		t.Fatal("stairs_down not found")
	}
	goal, _, ok := sim.findApproachGoal(stairs)
	if !ok {
		t.Fatal("stairs_down has no reachable approach goal")
	}
	stop := sim.activeNav().StopDistance
	assertAck(t, sim.Tick([]Input{{MessageID: "go", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: goal}}}), "go")
	for tick := 0; tick < 400; tick++ {
		sim.Tick(nil)
		if sim.activeLevel().autoNav == nil {
			break
		}
		if distance(player.pos, goal) <= stop {
			break
		}
	}
	if !sim.inMeleeRange(stairs) {
		remaining, reachable := sim.planPlayerPath(sim.activeNav(), player.pos, goal, sim.buildBlockedFn())
		t.Fatalf("did not reach stairs from %+v (goal=%+v stairs=%+v directPath=%t steps=%v autoNav=%+v)", player.pos, goal, stairs.pos, reachable, remaining, sim.activeLevel().autoNav)
	}
	descend := sim.Tick([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	assertAck(t, descend, "descend")
	if sim.currentLevel != -2 {
		t.Fatalf("level = %d, want -2", sim.currentLevel)
	}
}
