package game

import "testing"

func TestBossTreasureClassUsesRandomEquipmentPool(t *testing.T) {
	rules := loadRules(t)
	table, ok := rules.LootTables["boss_drop_tier_1"]
	if !ok || table.TreasureClassID != "boss_tc_tier_1" {
		t.Fatalf("boss loot table = %+v, want boss_tc_tier_1", table)
	}
	tc := rules.TreasureClasses["boss_tc_tier_1"]
	if len(tc.Attempts) < 3 {
		t.Fatalf("boss treasure class = %+v, want primary/bonus/extra attempts", tc.Attempts)
	}
	foundSunwake := false
	for _, attempt := range tc.Attempts {
		for _, entry := range attempt.Entries {
			if entry.SetItemID != "" {
				t.Fatalf("boss attempt %s has authored set drop = %+v, want none", attempt.AttemptID, entry)
			}
			if entry.UniqueItemID != "" {
				if entry.UniqueItemID != "sunwake_guard" {
					t.Fatalf("boss attempt %s has unexpected authored unique = %+v", attempt.AttemptID, entry)
				}
				foundSunwake = true
			}
		}
	}
	if !foundSunwake {
		t.Fatal("boss treasure class missing sunwake_guard unique pin")
	}
	if tc.Attempts[0].AttemptID != "primary" || tc.Attempts[0].SuccessWeight != 100 || tc.Attempts[0].NoDropWeight != 0 {
		t.Fatalf("boss primary attempt = %+v, want guaranteed primary roll", tc.Attempts[0])
	}
	if tc.Attempts[1].SuccessWeight != rules.MainConfig.Gameplay.BossBonusDropRatePercent {
		t.Fatalf("boss bonus success = %d, want %d", tc.Attempts[1].SuccessWeight, rules.MainConfig.Gameplay.BossBonusDropRatePercent)
	}
	if tc.Attempts[2].SuccessWeight != rules.MainConfig.Gameplay.BossExtraDropRatePercent {
		t.Fatalf("boss extra success = %d, want %d", tc.Attempts[2].SuccessWeight, rules.MainConfig.Gameplay.BossExtraDropRatePercent)
	}
}

func TestEliteObjectiveTreasureClassUsesGuardedChestPool(t *testing.T) {
	rules := loadRules(t)
	if rules.DungeonGeneration.EliteObjective.LootTable != "elite_objective_special_drop" {
		t.Fatalf("elite objective loot table = %s, want elite_objective_special_drop", rules.DungeonGeneration.EliteObjective.LootTable)
	}
	table := rules.LootTables["elite_objective_special_drop"]
	if table.TreasureClassID != "guarded_chest_tc_depth_3_plus" {
		t.Fatalf("elite objective treasure class = %s, want guarded_chest_tc_depth_3_plus", table.TreasureClassID)
	}
}

func TestBossLootRollsRandomEquipmentPayloads(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_boss_random_loot", "boss_random_loot_seed", rules)
	res := &TickResult{}
	sim.spawnLootDrops(
		rules.LootDrops("boss_drop_tier_1", NewRNG(3)),
		sim.entities[sim.playerID].pos,
		playerRadius,
		"corr_boss_loot",
		res,
		goldRollContext{levelNum: -5, magicFind: true, magicFindBonusPercent: rules.MainConfig.Gameplay.BossLootMagicFindBonusPercent},
	)

	rolledEquipment := 0
	for _, entity := range sim.activeLevel().entities {
		if entity.kind != lootEntity || entity.rollPayload == nil {
			continue
		}
		if entity.rollPayload.SetPieceID != "" {
			t.Fatalf("boss loot has authored set payload = %+v", entity.rollPayload)
		}
		if entity.rollPayload.NamedUniqueID != "" {
			continue
		}
		if entity.rollPayload.Rarity == "unique" || entity.rollPayload.Rarity == "set" {
			t.Fatalf("boss loot has authored rarity = %+v", entity.rollPayload)
		}
		if entity.rollPayload.ItemTemplateID != "" && entity.rollPayload.Rarity != "" {
			rolledEquipment++
		}
	}
	if rolledEquipment == 0 {
		t.Fatalf("boss loot spawned no rolled equipment; events=%+v", res.Events)
	}
	if got := countSpecialDropEvents(res.Events, "loot_dropped"); got == 0 {
		t.Fatalf("loot_dropped events = 0, want at least one: %+v", res.Events)
	}
}

func countSpecialDropEvents(events []Event, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}

	return count
}
