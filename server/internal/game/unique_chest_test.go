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
	wantAmount := rules.uniqueTestChestCatalogSize()
	if ev == nil || ev.Service != uniqueTestChestService || ev.Amount == nil || *ev.Amount != wantAmount || len(ev.StashItems) != wantAmount {
		t.Fatalf("unique chest event = %+v", ev)
	}
	if len(sim.inventory) != 0 {
		t.Fatalf("open added %d inventory items, want 0", len(sim.inventory))
	}

	gotEffects := map[string]int{}
	gotNamed := map[string]int{}
	gotSet := false
	for _, item := range ev.StashItems {
		payload := item.RollPayload()
		if payload == nil || (payload.Rarity != "unique" && payload.Rarity != "set") {
			t.Fatalf("granted item missing unique/set payload: %+v", item)
		}
		if payload.Rarity == "set" {
			gotSet = true
			if item.Rarity != "set" || item.DisplayName == "" {
				t.Fatalf("set item view missing presentation fields: %+v", item)
			}
			continue
		}
		if len(payload.EffectIDs) != 1 {
			t.Fatalf("granted item effects = %+v, want exactly one", payload.EffectIDs)
		}
		if namedID := payloadNamedUniqueID(rules, payload); namedID != "" {
			gotNamed[namedID]++
			continue
		}
		effectID := payload.EffectIDs[0]
		gotEffects[effectID]++
		effect := rules.UniqueEffects[effectID]
		template := rules.ItemTemplates[payload.ItemTemplateID]
		if payload.DisplayName != rules.uniqueItemDisplayName(template, payload.Stats, effect) {
			t.Fatalf("generated unique display name = %q, want %q", payload.DisplayName, rules.uniqueItemDisplayName(template, payload.Stats, effect))
		}
		if !uniqueChestEffectCompatible(effect, template.ItemType) {
			t.Fatalf("effect %s is not compatible with template %s type %s", effectID, payload.ItemTemplateID, template.ItemType)
		}
	}
	covered := rules.namedUniqueCoveredEffectIDs()
	for _, effectID := range sortedStringKeys(rules.UniqueEffects) {
		effect := rules.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" || covered[effectID] {
			continue
		}
		if gotEffects[effectID] != 1 {
			t.Fatalf("effect %s count = %d, want 1; inventory=%+v", effectID, gotEffects[effectID], sim.inventory)
		}
	}
	for _, uniqueID := range sortedStringKeys(rules.UniqueItems) {
		unique := rules.UniqueItems[uniqueID]
		if unique.Enabled && unique.Status == "ready" {
			if gotNamed[uniqueID] != 1 {
				t.Fatalf("unique chest named %s count = %d, want 1; all=%+v", uniqueID, gotNamed[uniqueID], ev.StashItems)
			}
		}
	}
	if !gotSet {
		t.Fatalf("unique chest did not offer a set item: %+v", ev.StashItems)
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
	add := findInventoryAddChange(take.Changes)
	if add == nil || add.Item == nil || add.Item.ItemInstanceID != takeEv.ItemInstanceID {
		t.Fatalf("unique chest take inventory add = %+v, event=%+v", add, takeEv)
	}
	if add.StashTransferID != "" {
		t.Fatalf("unique chest inventory add has stash transfer id %q; this would skip character-item persistence", add.StashTransferID)
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
			uniqueID:     "bloodbound_sigil",
			templateID:   "ring",
			displayName:  "Bloodbound Sigil",
			wantStats:    map[string]int{"max_hp": 6, "max_mana": 6},
			requirements: map[string]int{"level": 5, "magic": 5},
			effectIDs:    []string{"blood_price"},
		},
		{
			uniqueID:     "conduit_staff",
			templateID:   "starter_sorcerer_staff",
			displayName:  "Conduit Staff",
			wantStats:    map[string]int{"damage_min": 1, "damage_max": 3, "max_mana": 8},
			requirements: map[string]int{"level": 5, "magic": 5},
			effectIDs:    []string{"arcane_conduit"},
		},
		{
			uniqueID:     "embercall_blade",
			templateID:   "long_sword",
			displayName:  "Embercall Blade",
			wantStats:    map[string]int{"damage_min": 3, "damage_max": 9, "max_hp": 4},
			requirements: map[string]int{"level": 5, "str": 5},
			effectIDs:    []string{"everburning_wound"},
		},
		{
			uniqueID:     "stormstring_bow",
			templateID:   "bow",
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

func TestUniqueTestChestSameSeedProducesIdenticalCatalog(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	simA, err := NewSimWithWorld("sess_unique_test_chest_a", "seed_a", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim a: %v", err)
	}
	simB, err := NewSimWithWorld("sess_unique_test_chest_b", "seed_a", rules, "dungeon_levels")
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
	if len(itemsA) != rules.uniqueTestChestCatalogSize() {
		t.Fatalf("item count = %d, want %d", len(itemsA), rules.uniqueTestChestCatalogSize())
	}
	namedCounts := map[string]int{}
	for i := range itemsA {
		a := itemsA[i].rollPayload
		b := itemsB[i].rollPayload
		if a.ItemTemplateID != b.ItemTemplateID || a.Rarity != b.Rarity || a.DisplayName != b.DisplayName {
			t.Fatalf("payload %d differs: %+v != %+v", i, a, b)
		}
		if a.Rarity == "unique" && (len(a.EffectIDs) != 1 || len(b.EffectIDs) != 1 || a.EffectIDs[0] != b.EffectIDs[0]) {
			t.Fatalf("unique payload %d differs effects: %+v != %+v", i, a, b)
		}
		if payloadNamedUniqueID(rules, a) != "" {
			namedCounts[payloadNamedUniqueID(rules, a)]++
		}
	}
	for _, uniqueID := range sortedStringKeys(rules.UniqueItems) {
		unique := rules.UniqueItems[uniqueID]
		if unique.Enabled && unique.Status == "ready" && namedCounts[uniqueID] != 1 {
			t.Fatalf("named unique %s count = %d, want 1; all=%+v", uniqueID, namedCounts[uniqueID], namedCounts)
		}
	}
}

func TestUniqueTestChestEffectRollsVaryBySessionSeed(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	collectEffectTemplates := func(seed string) []string {
		sim, err := NewSimWithWorld("sess_unique_test_chest_roll_"+seed, seed, rules, "dungeon_levels")
		if err != nil {
			t.Fatalf("new sim %s: %v", seed, err)
		}
		items, ok := sim.uniqueTestChestItems()
		if !ok {
			t.Fatalf("unique chest items failed for seed %s", seed)
		}
		out := []string{}
		for _, item := range items {
			payload := item.rollPayload
			if payload == nil || payload.Rarity != "unique" || len(payload.EffectIDs) != 1 {
				continue
			}
			if payloadNamedUniqueID(rules, payload) != "" {
				continue
			}
			out = append(out, payload.ItemTemplateID)
		}
		return out
	}

	a := collectEffectTemplates("unique_chest_roll_seed_a")
	b := collectEffectTemplates("unique_chest_roll_seed_b")
	if len(a) == 0 || len(a) != len(b) {
		t.Fatalf("effect template rolls = %d and %d, want matching non-zero counts", len(a), len(b))
	}
	if reflect.DeepEqual(a, b) {
		t.Fatalf("expected different compatible template rolls for different session seeds")
	}
}

func TestUniqueTestChestSeededAtSessionStart(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest_seeded", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	state := sim.uniqueChests[chest.id]
	wantAmount := rules.uniqueTestChestCatalogSize()
	if state == nil || len(state.items) != wantAmount {
		t.Fatalf("seeded chest count = %d, want %d; state=%+v", len(state.items), wantAmount, state)
	}
}

func payloadNamedUniqueID(rules *Rules, payload *ItemRollPayload) string {
	if payload == nil {
		return ""
	}
	if payload.NamedUniqueID != "" {
		return payload.NamedUniqueID
	}

	return ""
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

func TestUniqueTestChestBackfillsPersistedCatalogGaps(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest_backfill", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	catalog, ok := sim.uniqueTestChestItems()
	if !ok || len(catalog) < 3 {
		t.Fatalf("unique test chest catalog = %d ok=%v", len(catalog), ok)
	}
	sim.uniqueChests[chest.id] = &uniqueChestState{items: []*stashItem{{
		stashItemID: sim.alloc(),
		itemDefID:   catalog[0].itemDefID,
		rollPayload: cloneRollPayload(catalog[0].rollPayload),
	}}}

	open := sim.Tick([]Input{{MessageID: "open_unique_chest_backfill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_unique_chest_backfill")
	ev := findEvent(open.Events, "unique_chest_opened")
	if ev == nil || len(ev.StashItems) != len(catalog) {
		t.Fatalf("backfilled unique chest event count = %d want %d event=%+v", len(ev.StashItems), len(catalog), ev)
	}
	if len(sim.uniqueChests[chest.id].items) != len(catalog) {
		t.Fatalf("persisted unique chest count = %d want %d", len(sim.uniqueChests[chest.id].items), len(catalog))
	}
	seen := map[string]bool{}
	for _, item := range sim.uniqueChests[chest.id].items {
		key := uniqueChestCatalogKey(item.itemDefID, item.rollPayload)
		if seen[key] {
			t.Fatalf("duplicate unique chest catalog key %s in %+v", key, sim.uniqueChests[chest.id].items)
		}
		seen[key] = true
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

func findInventoryAddChange(changes []Change) *Change {
	for i := range changes {
		if changes[i].Op == OpInventoryAdd {
			return &changes[i]
		}
	}
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
