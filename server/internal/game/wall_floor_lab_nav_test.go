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
	sim, err := NewSimWithWorld("sess_wall_floor_offset", "wall_seed_00", rules, "generated_wall_lab")
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
	goal := offsetInteractableMoveGoal(player.pos, stairs.pos, 1.275)
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
		t.Fatalf("did not reach stairs from %+v (goal=%+v stairs=%+v)", player.pos, goal, stairs.pos)
	}
	descend := sim.Tick([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	assertAck(t, descend, "descend")
	if sim.currentLevel != -2 {
		t.Fatalf("level = %d, want -2", sim.currentLevel)
	}
}

func offsetInteractableMoveGoal(playerPos, targetPos Vec2, stopDistance float64) Vec2 {
	delta := Vec2{X: targetPos.X - playerPos.X, Y: targetPos.Y - playerPos.Y}
	dist := distance(playerPos, targetPos)
	if dist <= stopDistance || dist <= 1e-9 {
		return targetPos
	}
	dir := Vec2{X: delta.X / dist, Y: delta.Y / dist}
	return Vec2{X: targetPos.X - dir.X*stopDistance, Y: targetPos.Y - dir.Y*stopDistance}
}
