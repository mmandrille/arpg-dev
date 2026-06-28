package game

import "testing"

func TestPlanPlayerPathRespectsPerSearchNodeCap(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_player_path_cap", "player_path_cap_seed", rules, "combat_control_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	nav := sim.activeNav()
	player := sim.activeLevel().entities[sim.playerID]
	blocked := func(gx, gy int) bool {
		return gx >= 10
	}
	goal := Vec2{X: 15, Y: player.pos.Y}
	_, ok := sim.planPlayerPath(nav, player.pos, goal, blocked)
	if ok {
		t.Fatal("planPlayerPath ok=true, want false for blocked goal")
	}
	dist := octile(worldToGrid(nav, player.pos), worldToGrid(nav, goal))
	nodeCap := nav.PlayerPathNodesPerSearch
	if scaled := dist*128 + 64; scaled > nodeCap {
		nodeCap = scaled
	}
	if sim.playerPathNodesThisTick > nodeCap+1 {
		t.Fatalf("nodes visited = %d, want <= %d", sim.playerPathNodesThisTick, nodeCap+1)
	}
}

func TestPlanPlayerPathRespectsPerTickNodeCap(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_player_path_tick_cap", "player_path_tick_cap_seed", rules, "combat_control_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	nav := sim.activeNav()
	player := sim.activeLevel().entities[sim.playerID]
	blocked := func(gx, gy int) bool {
		return gx >= 10
	}
	for i := 0; i < 3; i++ {
		sim.planPlayerPath(nav, player.pos, Vec2{X: 15, Y: player.pos.Y + float64(i)}, blocked)
	}
	if sim.playerPathNodesThisTick > rules.Navigation.PlayerPathNodesPerTick+1 {
		t.Fatalf("aggregate nodes = %d, want <= %d", sim.playerPathNodesThisTick, rules.Navigation.PlayerPathNodesPerTick+1)
	}
}

func TestFindMeleeApproachGoalStaysWithinPlayerPathTickBudget(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_player_approach_cap", "player_approach_cap_seed", rules, "player_path_budget_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	var target *entity
	for _, e := range sim.activeLevel().entities {
		if e != nil && e.kind == monsterEntity {
			target = e
			break
		}
	}
	if target == nil {
		t.Fatal("missing monster in player_path_budget_lab")
	}
	_, _, ok := sim.findMeleeApproachGoal(target)
	if ok {
		t.Fatal("findMeleeApproachGoal ok=true, want unreachable walled monster")
	}
	if sim.playerPathNodesThisTick > rules.Navigation.PlayerPathNodesPerTick+1 {
		t.Fatalf("approach scan nodes = %d, want <= %d", sim.playerPathNodesThisTick, rules.Navigation.PlayerPathNodesPerTick+1)
	}
}

func TestPlayerMoveIntentUsesPlayerMaxAutoSteps(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_player_max_steps", "player_max_steps_seed", rules, "path_maze")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	goal := Vec2{X: 12, Y: 5}
	res := sim.Tick([]Input{{
		MessageID: "move-far",
		Type:      "move_to_intent",
		MoveTo:    &MoveToIntent{Position: goal},
	}})
	if len(res.Rejects) > 0 {
		t.Fatalf("move rejected: %+v", res.Rejects)
	}
	if len(res.Acks) == 0 {
		t.Fatalf("move not acked: %+v", res)
	}
	if sim.activeLevel().autoNav == nil {
		t.Fatal("autoNav nil after move_to_intent")
	}
	if got := len(sim.activeLevel().autoNav.steps); got > rules.Navigation.PlayerMaxAutoSteps {
		t.Fatalf("planned steps = %d, want <= player_max_auto_steps %d", got, rules.Navigation.PlayerMaxAutoSteps)
	}
}
