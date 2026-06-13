package game

import (
	"reflect"
	"testing"
)

func TestUniqueTestChestOpensContainerAndTakesSelectedItem(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	sim.activeLevel().entities[sim.playerID].pos = chest.pos

	open := sim.Tick([]Input{{
		MessageID:     "open_unique_chest",
		CorrelationID: "corr_unique_chest",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(chest.id)},
	}})
	assertAck(t, open, "open_unique_chest")
	if chest.state != interactableOpen {
		t.Fatalf("chest state = %s, want open", chest.state)
	}
	ev := findEvent(open.Events, "unique_chest_opened")
	wantAmount := enabledUniqueEffectCount(rules) + enabledNamedUniqueCount(rules)
	if ev == nil || ev.Service != uniqueTestChestService || ev.Amount == nil || *ev.Amount != wantAmount || len(ev.StashItems) != wantAmount {
		t.Fatalf("unique chest event = %+v", ev)
	}
	if len(sim.inventory) != 0 {
		t.Fatalf("open added %d inventory items, want 0", len(sim.inventory))
	}

	gotEffects := map[string]int{}
	gotNamed := false
	for _, item := range ev.StashItems {
		payload := item.RollPayload()
		if payload == nil || payload.Rarity != "unique" {
			t.Fatalf("granted item missing unique payload: %+v", item)
		}
		if len(payload.EffectIDs) != 1 {
			t.Fatalf("granted item effects = %+v, want exactly one", payload.EffectIDs)
		}
		if payload.DisplayName != "Embercall Blade" && payload.DisplayName != "Stormstring Bow" {
			effectID := payload.EffectIDs[0]
			gotEffects[effectID]++
			effect := rules.UniqueEffects[effectID]
			template := rules.ItemTemplates[payload.ItemTemplateID]
			if payload.DisplayName != uniqueItemDisplayName(template, effect) {
				t.Fatalf("generated unique display name = %q, want %q", payload.DisplayName, uniqueItemDisplayName(template, effect))
			}
			if !uniqueChestEffectCompatible(effect, template.ItemType) {
				t.Fatalf("effect %s is not compatible with template %s type %s", effectID, payload.ItemTemplateID, template.ItemType)
			}
		}
		if payload.DisplayName == "Embercall Blade" {
			gotNamed = true
		}
	}
	for _, effectID := range sortedStringKeys(rules.UniqueEffects) {
		effect := rules.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" {
			continue
		}
		if gotEffects[effectID] != 1 {
			t.Fatalf("effect %s count = %d, want 1; inventory=%+v", effectID, gotEffects[effectID], sim.inventory)
		}
	}
	if !gotNamed {
		t.Fatalf("unique chest did not offer Embercall Blade: %+v", ev.StashItems)
	}

	take := sim.Tick([]Input{{
		MessageID:     "take_unique_item",
		CorrelationID: "corr_unique_take",
		Type:          "unique_chest_take_item_intent",
		UniqueChestTakeItem: &UniqueChestTakeItemIntent{
			ChestEntityID: idStr(chest.id),
			ChestItemID:   ev.StashItems[0].StashItemID,
		},
	}})
	assertAck(t, take, "take_unique_item")
	takeEv := findEvent(take.Events, "unique_chest_item_taken")
	if takeEv == nil || takeEv.StashItemID != ev.StashItems[0].StashItemID || takeEv.ItemInstanceID == "" || len(takeEv.StashItems) != wantAmount-1 {
		t.Fatalf("unique_chest_item_taken event = %+v", takeEv)
	}
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory count after take = %d, want 1", len(sim.inventory))
	}
}

func TestNamedUniquePayloadBuildsFixedPackages(t *testing.T) {
	rules := loadRules(t)

	tests := []struct {
		uniqueID     string
		templateID   string
		displayName  string
		wantStats    map[string]int
		requirements map[string]int
		effectIDs    []string
	}{
		{
			uniqueID:     "embercall_blade",
			templateID:   "cave_blade",
			displayName:  "Embercall Blade",
			wantStats:    map[string]int{"damage_min": 3, "damage_max": 9, "max_hp": 4},
			requirements: map[string]int{"level": 5, "str": 5},
			effectIDs:    []string{"everburning_wound"},
		},
		{
			uniqueID:     "stormstring_bow",
			templateID:   "cave_bow",
			displayName:  "Stormstring Bow",
			wantStats:    map[string]int{"damage_min": 2, "damage_max": 6, "attack_speed_percent": 6},
			requirements: map[string]int{"level": 5, "dex": 5},
			effectIDs:    []string{"stormbound_echo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.uniqueID, func(t *testing.T) {
			payload, ok := rules.namedUniquePayload(tc.uniqueID)
			if !ok {
				t.Fatal("namedUniquePayload returned false")
			}
			if payload.ItemTemplateID != tc.templateID || payload.DisplayName != tc.displayName || payload.Rarity != "unique" {
				t.Fatalf("named unique identity = %+v", payload)
			}
			for stat, want := range tc.wantStats {
				if payload.Stats[stat] != want {
					t.Fatalf("stat %s = %d, want %d in %+v", stat, payload.Stats[stat], want, payload.Stats)
				}
			}
			for stat, want := range tc.requirements {
				if payload.Requirements[stat] != want {
					t.Fatalf("requirement %s = %d, want %d in %+v", stat, payload.Requirements[stat], want, payload.Requirements)
				}
			}
			if !reflect.DeepEqual(payload.EffectIDs, tc.effectIDs) {
				t.Fatalf("effect ids = %+v, want %+v", payload.EffectIDs, tc.effectIDs)
			}
		})
	}
}

func TestUniqueTestChestDeterministicPayloadOrder(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	simA, err := NewSimWithWorld("sess_unique_test_chest_a", "seed_a", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim a: %v", err)
	}
	simB, err := NewSimWithWorld("sess_unique_test_chest_b", "seed_b", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim b: %v", err)
	}
	itemsA, ok := simA.uniqueTestChestItems()
	if !ok {
		t.Fatal("unique chest items a failed")
	}
	itemsB, ok := simB.uniqueTestChestItems()
	if !ok {
		t.Fatal("unique chest items b failed")
	}
	if len(itemsA) != len(itemsB) {
		t.Fatalf("item counts differ %d != %d", len(itemsA), len(itemsB))
	}
	if len(itemsA) != enabledUniqueEffectCount(rules)+enabledNamedUniqueCount(rules) {
		t.Fatalf("item count = %d, want enabled effects + named uniques", len(itemsA))
	}
	namedCounts := map[string]int{}
	for i := range itemsA {
		a := itemsA[i].rollPayload
		b := itemsB[i].rollPayload
		if a.ItemTemplateID != b.ItemTemplateID || a.Rarity != b.Rarity || len(a.EffectIDs) != 1 || len(b.EffectIDs) != 1 || a.EffectIDs[0] != b.EffectIDs[0] {
			t.Fatalf("payload %d differs: %+v != %+v", i, a, b)
		}
		if a.DisplayName == "Embercall Blade" || a.DisplayName == "Stormstring Bow" {
			namedCounts[a.DisplayName]++
		}
	}
	if namedCounts["Embercall Blade"] != 1 || namedCounts["Stormstring Bow"] != 1 {
		t.Fatalf("named unique counts = %+v", namedCounts)
	}
}

func TestUniqueTestChestRepeatActivationReopensRemainingItems(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest_repeat", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	first := sim.Tick([]Input{{MessageID: "open_unique_chest", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, first, "open_unique_chest")
	firstEv := findEvent(first.Events, "unique_chest_opened")
	if firstEv == nil || len(firstEv.StashItems) == 0 {
		t.Fatalf("first unique_chest_opened = %+v", firstEv)
	}

	again := sim.Tick([]Input{{MessageID: "open_unique_chest_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, again, "open_unique_chest_again")
	againEv := findEvent(again.Events, "unique_chest_opened")
	if againEv == nil || len(againEv.StashItems) != len(firstEv.StashItems) {
		t.Fatalf("repeat unique_chest_opened = %+v, first=%+v", againEv, firstEv)
	}
	if len(sim.inventory) != 0 {
		t.Fatalf("repeat activation inventory count = %d, want 0", len(sim.inventory))
	}
}

func TestUniqueTestChestHiddenWhenGameplayDebugDisabled(t *testing.T) {
	sim, err := NewSimWithWorld("sess_unique_test_chest_hidden", "unique_test_chest_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == "town_unique_chest" {
			t.Fatalf("unique chest spawned with gameplay debug disabled: %+v", e)
		}
	}
}

func findUniqueTestChest(t *testing.T, sim *Sim) *entity {
	t.Helper()
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == "town_unique_chest" {
			return e
		}
	}
	t.Fatalf("missing town_unique_chest: %+v", sim.activeLevel().entities)
	return nil
}

func enabledUniqueEffectCount(rules *Rules) int {
	count := 0
	for _, effect := range rules.UniqueEffects {
		if effect.Enabled && effect.Status == "ready" {
			count++
		}
	}
	return count
}

func enabledNamedUniqueCount(rules *Rules) int {
	count := 0
	for _, unique := range rules.UniqueItems {
		if unique.Enabled && unique.Status == "ready" {
			count++
		}
	}
	return count
}
