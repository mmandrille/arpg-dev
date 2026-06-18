package game

import "testing"

func TestCrowdedMovementLODDefersFarLowPriorityMonsters(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_movement_lod", "movement_lod_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	if !sim.monsterMovementLODActive() {
		t.Fatal("movement LOD inactive in crowded probe")
	}

	var highPrecision, lowAllowed, lowDeferred int
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		monster := sim.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		if sim.monsterMovementHighPrecision(monster) {
			highPrecision++
			if !sim.monsterMovementLODAllowsTick(monster) {
				t.Fatalf("high precision monster %d was LOD-deferred", monster.id)
			}
			continue
		}
		if sim.monsterMovementLODAllowsTick(monster) {
			lowAllowed++
		} else {
			lowDeferred++
		}
	}
	if highPrecision == 0 {
		t.Fatal("no high precision monsters found in crowded probe")
	}
	if lowAllowed == 0 || lowDeferred == 0 {
		t.Fatalf("low-priority LOD split = allowed %d deferred %d, want both", lowAllowed, lowDeferred)
	}
}

func TestMovementLODDoesNotAffectSmallFightsOrImportantMonsters(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_movement_lod_small", "movement_lod_small_seed", rules, "chase_lab")
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	if sim.monsterMovementLODActive() {
		t.Fatal("movement LOD active in small fight")
	}
	monster := firstEntityByKind(sim, monsterEntity)
	if !sim.monsterMovementLODAllowsTick(monster) {
		t.Fatal("small-fight monster was LOD-deferred")
	}

	crowded, err := NewSimWithWorld("sess_movement_lod_important", "movement_lod_important_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("crowded world: %v", err)
	}
	var far *entity
	for _, id := range sortedEntityIDs(crowded.activeLevel().entities) {
		monster := crowded.activeLevel().entities[id]
		if monster != nil && monster.kind == monsterEntity && !crowded.monsterMovementHighPrecision(monster) {
			far = monster
			break
		}
	}
	if far == nil {
		t.Fatal("no far low-priority monster found")
	}
	far.monsterRarityID = "test_elite"
	if !crowded.monsterMovementHighPrecision(far) || !crowded.monsterMovementLODAllowsTick(far) {
		t.Fatal("important far monster was not kept high precision")
	}
}
