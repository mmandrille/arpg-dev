package game

import "testing"

func TestDungeonEliteObjectiveChestRequiresEliteLeader(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	level, err := GenerateDungeonLevel("v158_forced_elite_objective", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate forced elite objective: %v", err)
	}
	objectives := 0
	for _, chest := range level.chests {
		if chest.eliteObjective {
			objectives++
			if chest.defID != treasureChestDefID || chest.lootTable != rules.DungeonGeneration.EliteObjective.LootTable {
				t.Fatalf("elite objective chest = %+v", chest)
			}
		}
	}
	if objectives != 1 {
		t.Fatalf("elite objective chests = %d in %+v, want 1", objectives, level.chests)
	}

	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 0
	level, err = GenerateDungeonLevel("v158_forced_elite_objective", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate no elite objective: %v", err)
	}
	for _, chest := range level.chests {
		if chest.eliteObjective {
			t.Fatalf("unexpected elite objective without elite leader: %+v", level.chests)
		}
	}
}
