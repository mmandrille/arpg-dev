package game

import "testing"

func TestOverloadDegradationDefersOnlyLowPriorityMovement(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_overload_guardrails", "overload_guardrails_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("world: %v", err)
	}

	var lowPriority *entity
	var highPrecision *entity
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		monster := sim.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		if sim.monsterMovementHighPrecision(monster) && highPrecision == nil {
			highPrecision = monster
			continue
		}
		if !sim.monsterMovementHighPrecision(monster) && lowPriority == nil {
			lowPriority = monster
		}
	}
	if lowPriority == nil || highPrecision == nil {
		t.Fatalf("need low and high precision monsters, got low=%v high=%v", lowPriority != nil, highPrecision != nil)
	}
	if !sim.ApplyOverloadDegradation() || !sim.overloadDegraded() {
		t.Fatal("overload degradation did not activate")
	}
	if sim.monsterMovementLODAllowsTick(lowPriority) {
		t.Fatalf("low-priority monster %d was allowed during overload degradation", lowPriority.id)
	}
	if !sim.monsterMovementLODAllowsTick(highPrecision) {
		t.Fatalf("high-priority monster %d was deferred during overload degradation", highPrecision.id)
	}

	sim.tick += uint64(rules.Navigation.MonsterOverloadDegradeTicks)
	if sim.overloadDegraded() {
		t.Fatal("overload degradation did not expire after configured ticks")
	}
}
