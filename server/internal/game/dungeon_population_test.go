package game

import "testing"

func TestPopulateDungeonLevelTracksEliteObjectiveChestIDs(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	sim, err := NewSimWithWorld("sess_population_objective", "v160_population_objective", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	descendFromCurrentLevel(t, sim, "descend")
	level := sim.activeLevel()
	chest := findEliteObjectiveChestEntity(level)
	if chest == nil {
		t.Fatalf("missing runtime elite objective chest id in %+v", level.entities)
	}
	if chest.kind != interactableEntity || chest.interactableDefID != treasureChestDefID {
		t.Fatalf("objective chest entity = %+v", chest)
	}
	if chest.lootTable != rules.DungeonGeneration.EliteObjective.LootTable {
		t.Fatalf("objective chest loot table = %s, want %s", chest.lootTable, rules.DungeonGeneration.EliteObjective.LootTable)
	}
}

func TestPopulateDungeonLevelPreservesBossAndRarityRuntimeState(t *testing.T) {
	var golden struct {
		Seed     string `json:"seed"`
		Level    int    `json:"level"`
		Expected struct {
			Boss struct {
				TemplateID       string  `json:"template_id"`
				BaseMonsterDefID string  `json:"base_monster_def_id"`
				VisualModel      string  `json:"visual_model"`
				VisualColor      string  `json:"visual_color"`
				VisualScale      float64 `json:"visual_scale"`
			} `json:"boss"`
		} `json:"expected"`
	}
	loadGolden(t, "boss_floor_-5.json", &golden)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_population_boss", golden.Seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	for levelNum := 0; levelNum > golden.Level; levelNum-- {
		descendFromCurrentLevel(t, sim, "descend")
	}

	boss := findRuntimeBoss(sim.activeLevel())
	if boss == nil {
		t.Fatalf("missing runtime boss on level %d", golden.Level)
	}
	if !boss.isBoss || boss.bossTemplateID != golden.Expected.Boss.TemplateID || boss.monsterDefID != golden.Expected.Boss.BaseMonsterDefID {
		t.Fatalf("boss identity = boss:%v template:%s def:%s, want template %s def %s", boss.isBoss, boss.bossTemplateID, boss.monsterDefID, golden.Expected.Boss.TemplateID, golden.Expected.Boss.BaseMonsterDefID)
	}
	if boss.visualModel != golden.Expected.Boss.VisualModel || boss.visualTint != golden.Expected.Boss.VisualColor || boss.visualScale != golden.Expected.Boss.VisualScale {
		t.Fatalf("boss visual = model:%s tint:%s scale:%f, want %+v", boss.visualModel, boss.visualTint, boss.visualScale, golden.Expected.Boss)
	}
	template := rules.BossTemplates[golden.Expected.Boss.TemplateID]
	base := rules.Monsters[golden.Expected.Boss.BaseMonsterDefID]
	wantHP := roundPositive(float64(base.MaxHP) * template.HPMultiplier)
	if boss.maxHP != wantHP {
		t.Fatalf("boss max hp = %d, want base %d * multiplier %.2f = %d", boss.maxHP, base.MaxHP, template.HPMultiplier, wantHP)
	}
	if boss.maxHP <= 0 || boss.hp != boss.maxHP || boss.lootTable == "" || boss.bossPatternDeckIndex != 0 || boss.bossPatternID == "" {
		t.Fatalf("boss runtime state incomplete: %+v", boss)
	}
}

func findRuntimeBoss(level *LevelState) *entity {
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e.kind == monsterEntity && e.isBoss {
			return e
		}
	}
	return nil
}
