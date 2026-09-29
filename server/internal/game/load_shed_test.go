package game

import "testing"

func TestSplitLoadShedInputsDropsDirectivesWithActor(t *testing.T) {
	d := &LoadShedDirective{OverloadDegrade: true}
	inputs := []Input{
		{Type: "move_intent", ActorPlayerID: 1001},
		{Type: SystemLoadShedInputType, LoadShed: d},
		{Type: SystemLoadShedInputType, ActorPlayerID: 1001, LoadShed: d},
	}
	players, directives := splitLoadShedInputs(inputs)
	if len(players) != 1 || players[0].Type != "move_intent" {
		t.Fatalf("player inputs = %+v, want only move_intent", players)
	}
	if len(directives) != 1 {
		t.Fatalf("directives = %d, want 1 (actor-bearing directive dropped)", len(directives))
	}
}

func TestRecordedLoadShedAppliesAfterTheTick(t *testing.T) {
	sim, err := NewSimWithWorld("sess_load_shed", "overload_guardrails_seed", loadRules(t), "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	if sim.activeNav().MonsterOverloadDegradeTicks <= 0 {
		t.Skip("overload degradation disabled by rules")
	}
	tick := sim.tick
	sim.TickResults([]Input{{Type: SystemLoadShedInputType, LoadShed: &LoadShedDirective{OverloadDegrade: true}}})
	want := tick + 1 + uint64(sim.activeNav().MonsterOverloadDegradeTicks)
	if sim.overloadDegradeUntilTick != want {
		t.Fatalf("degrade until = %d, want %d (window starts after the recorded tick)", sim.overloadDegradeUntilTick, want)
	}
	if !sim.combatMovementThrottleActive() || sim.LoadShedChanges(LoadShedDirective{CombatMovementThrottle: true}) {
		t.Fatal("applied directive should leave the throttle active")
	}
}
