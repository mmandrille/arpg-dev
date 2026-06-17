package game

import "testing"

func TestMoveToIntentCompletesExactTargetOffset(t *testing.T) {
	sim := MustNewSim("sess_move_to_exact", "01", loadRules(t))
	sim.activeLevel().walls = nil
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 0.9, Y: 0.9}
	target := Vec2{X: 3.1, Y: 0.1}

	r := sim.Tick([]Input{{MessageID: "go_exact", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: target}}})
	assertAck(t, r, "go_exact")
	for i := 0; i < 12 && sim.autoNav != nil; i++ {
		sim.Tick(nil)
	}
	if sim.autoNav != nil {
		t.Fatalf("autoNav still active after exact-target ticks: %+v", sim.autoNav)
	}
	if got := distance(player.pos, target); got > sim.activeNav().StopDistance {
		t.Fatalf("player stopped %.3f from exact target; pos=%+v target=%+v", got, player.pos, target)
	}
}
