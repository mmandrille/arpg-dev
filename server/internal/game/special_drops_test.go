package game

import "testing"

func TestBossSpecialTreasureClassDropsAuthoredPackages(t *testing.T) {
	rules := loadRules(t)
	drops := rules.LootDrops("boss_drop_tier_1", NewRNG(1))
	if len(drops) != 3 {
		t.Fatalf("boss special drops = %+v, want 3 drops", drops)
	}
	if drops[0].UniqueItemID != "conduit_staff" {
		t.Fatalf("boss unique drop = %+v, want conduit_staff", drops[0])
	}
	if drops[1].SetItemID != "stormrunner_covenant_bow" {
		t.Fatalf("boss set drop = %+v, want stormrunner_covenant_bow", drops[1])
	}
	if drops[2].ItemTemplateID != "cave_amulet" {
		t.Fatalf("boss equipment drop = %+v, want cave_amulet", drops[2])
	}
}

func TestEliteObjectiveSpecialTreasureClassUsesWeightedSetPool(t *testing.T) {
	rules := loadRules(t)
	if rules.DungeonGeneration.EliteObjective.LootTable != "elite_objective_special_drop" {
		t.Fatalf("elite objective loot table = %s, want elite_objective_special_drop", rules.DungeonGeneration.EliteObjective.LootTable)
	}
	tc := rules.TreasureClasses["elite_objective_special_tc_1"]
	if len(tc.Attempts) < 2 {
		t.Fatalf("elite special treasure class = %+v, want set and equipment attempts", tc.Attempts)
	}
	setAttempt := tc.Attempts[0]
	if setAttempt.AttemptID != "set_piece" || setAttempt.SuccessWeight != 20 || setAttempt.NoDropWeight != 80 {
		t.Fatalf("elite set attempt = %+v, want 20/80 weighted chance", setAttempt)
	}
	setIDs := map[string]bool{}
	for _, entry := range setAttempt.Entries {
		if entry.SetItemID == "" {
			t.Fatalf("elite set attempt entry = %+v, want set_item_id", entry)
		}
		setIDs[entry.SetItemID] = true
	}
	for setItemID := range rules.SetItems {
		if !setIDs[setItemID] {
			t.Fatalf("elite set attempt missing %s in %+v", setItemID, setAttempt.Entries)
		}
	}
	if len(setIDs) != len(rules.SetItems) {
		t.Fatalf("elite set attempt entries = %d, want %d", len(setIDs), len(rules.SetItems))
	}
}

func TestAuthoredSpecialDropsSpawnFixedPayloads(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_special_drops", "special_drops_seed", rules)
	res := &TickResult{}
	sim.spawnLootDrops(rules.LootDrops("boss_drop_tier_1", NewRNG(3)), sim.entities[sim.playerID].pos, playerRadius, "corr_special", res, goldRollContext{levelNum: -5})

	conduit := findLootByDisplayName(sim, "Conduit Staff")
	if conduit == nil || conduit.rollPayload == nil || conduit.rollPayload.Rarity != "unique" || !sameStringSlice(conduit.rollPayload.EffectIDs, []string{"arcane_conduit"}) {
		t.Fatalf("Conduit Staff loot payload = %+v", conduit)
	}
	stormrunner := findLootByDisplayName(sim, "Stormrunner Covenant Bow")
	if stormrunner == nil || stormrunner.rollPayload == nil || stormrunner.rollPayload.Rarity != "set" || stormrunner.rollPayload.ItemTemplateID != "cave_bow" {
		t.Fatalf("Stormrunner loot payload = %+v", stormrunner)
	}
	amulet := findLootByTemplateID(sim, "cave_amulet")
	if amulet == nil || amulet.rollPayload == nil || amulet.rollPayload.Rarity == "" || amulet.rollPayload.DisplayName == "" {
		t.Fatalf("rolled amulet payload = %+v", amulet)
	}
	if got := countSpecialDropEvents(res.Events, "loot_dropped"); got != 3 {
		t.Fatalf("loot_dropped events = %d, want 3: %+v", got, res.Events)
	}
}

func findLootByDisplayName(sim *Sim, displayName string) *entity {
	for _, entity := range sim.activeLevel().entities {
		if entity.kind == lootEntity && entity.rollPayload != nil && entity.rollPayload.DisplayName == displayName {
			return entity
		}
	}
	return nil
}

func findLootByTemplateID(sim *Sim, templateID string) *entity {
	for _, entity := range sim.activeLevel().entities {
		if entity.kind == lootEntity && entity.rollPayload != nil && entity.rollPayload.ItemTemplateID == templateID {
			return entity
		}
	}
	return nil
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
