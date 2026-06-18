package game

import "testing"

func TestGeneratedDungeonDoorGeneration(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("v40_obstacles", -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	again, err := GenerateDungeonLevel("v40_obstacles", -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate again: %v", err)
	}
	if len(level.doors) == 0 {
		t.Fatal("generated dungeon missing generated door")
	}
	if len(level.doors) != len(again.doors) {
		t.Fatalf("repeat doors = %d, want %d", len(again.doors), len(level.doors))
	}
	for i, door := range level.doors {
		if door != again.doors[i] {
			t.Fatalf("door %d = %+v, repeat %+v", i, door, again.doors[i])
		}
		if door.defID != woodenDoorDefID || door.state != interactableClosed {
			t.Fatalf("generated door = %+v, want closed wooden door", door)
		}
		if !generatedTargetReachable(rules.DungeonGeneration.RulesForLevel(level.levelNum), level, door.pos) {
			t.Fatalf("generated door unreachable: %+v", door)
		}
	}
}

func TestBossFloorExcludesGeneratedDoors(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("boss_floor_gate", -5, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor: %v", err)
	}
	if len(level.doors) != 0 {
		t.Fatalf("boss floor generated doors = %+v, want none", level.doors)
	}
}

func TestGeneratedDungeonDoorsPopulateAsClosedInteractables(t *testing.T) {
	sim := MustNewSim("v261_generated_doors", "01", loadRules(t))
	level, err := sim.ensureDungeonLevel(-2)
	if err != nil {
		t.Fatalf("ensure dungeon: %v", err)
	}
	found := false
	for _, e := range level.entities {
		if e.kind != interactableEntity || e.interactableDefID != woodenDoorDefID {
			continue
		}
		found = true
		if e.state != interactableClosed {
			t.Fatalf("generated door state = %s, want closed", e.state)
		}
	}
	if !found {
		t.Fatal("missing generated door interactable")
	}
}
