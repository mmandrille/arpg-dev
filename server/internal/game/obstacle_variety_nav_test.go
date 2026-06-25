package game

import "testing"

func TestObstacleVarietyLabMoveToLoot(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_obstacle_variety_nav", "v299_obstacle_variety_pack", rules, "obstacle_variety_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	player := sim.activeLevel().entities[sim.playerID]
	goal := Vec2{X: 13, Y: 5}
	stop := sim.activeNav().StopDistance

	r := sim.Tick([]Input{{MessageID: "go_loot", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: goal}}})
	assertAck(t, r, "go_loot")

	lastPos := player.pos
	stuckTicks := 0
	for tick := 0; tick < 400; tick++ {
		sim.Tick(nil)
		if sim.activeLevel().autoNav == nil {
			break
		}
		if distance(player.pos, goal) <= stop {
			break
		}
		if player.pos == lastPos {
			stuckTicks++
		} else {
			stuckTicks = 0
		}
		lastPos = player.pos
		if stuckTicks >= 20 {
			t.Fatalf("stuck at %+v after %d ticks (goal %+v, dist=%.3f, autoNav=%+v)",
				player.pos, tick, goal, distance(player.pos, goal), sim.activeLevel().autoNav)
		}
	}

	if got := distance(player.pos, goal); got > stop {
		t.Fatalf("player stopped %.3f from goal; pos=%+v goal=%+v autoNav=%+v",
			got, player.pos, goal, sim.activeLevel().autoNav)
	}
}

func TestObstacleVarietyLabPathPlansAroundBlockers(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_obstacle_variety_path", "v299", rules, "obstacle_variety_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	player := sim.activeLevel().entities[sim.playerID]
	goal := Vec2{X: 13, Y: 5}
	steps, ok := sim.planPath(sim.activeNav(), player.pos, goal, sim.buildBlockedFn())
	if !ok {
		t.Fatal("planPath returned ok=false")
	}
	if len(steps) < 8 {
		t.Fatalf("planPath steps = %d, want detour around blocker strip", len(steps))
	}
}
