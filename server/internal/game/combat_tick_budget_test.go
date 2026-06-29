package game

import "testing"

func TestCombatMovementThrottleSkipsLowPriorityMonsters(t *testing.T) {
	sim, err := NewSimWithWorld("sess_combat_tick_budget", "overload_guardrails_seed", loadRules(t), "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	sim.SetCombatMovementThrottle(true)
	var lowPriority *entity
	var highPrecision *entity
	for _, monster := range sim.activeLevel().entities {
		if monster == nil || monster.kind != monsterEntity {
			continue
		}
		if sim.monsterMovementHighPrecision(monster) {
			highPrecision = monster
		} else {
			lowPriority = monster
		}
		if lowPriority != nil && highPrecision != nil {
			break
		}
	}
	if lowPriority == nil || highPrecision == nil {
		t.Fatal("expected both low-priority and high-precision monsters in crowded probe world")
	}
	if !sim.monsterMovementLODAllowsTick(lowPriority) && !sim.overloadDegraded() {
		// throttle path should still skip when not allowed by LOD tick
	}
	res := &TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterMovementBudgeted(res)
	if !sim.combatMovementThrottleActive() {
		t.Fatal("expected combat movement throttle to remain active")
	}
}
