package game

import "testing"

func TestVendorLabMoveToMysterySeller(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_vendor_mystery_nav", "v51_mystery_0504", rules, "vendor_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	player := sim.activeLevel().entities[sim.playerID]
	goal := Vec2{X: 17, Y: 12}
	steps, ok := sim.planPath(sim.activeNav(), player.pos, goal, sim.buildBlockedFn())
	if !ok {
		t.Fatalf("no path from %+v to %+v within bounds %+v", player.pos, goal, sim.activeNav().GridBounds)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty path to mystery seller")
	}
}
