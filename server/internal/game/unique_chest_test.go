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
	wantAmount := enabledUniqueEffectCount(rules) + enabledNamedUniqueCount(rules) + enabledSetItemCount(rules)
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
		if _, named := rules.UniqueItems[payloadNamedUniqueID(rules, payload.DisplayName)]; !named {
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
		} else {
			gotNamed[payload.DisplayName]++
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
	for _, uniqueID := range sortedStringKeys(rules.UniqueItems) {
		unique := rules.UniqueItems[uniqueID]
		if unique.Enabled && unique.Status == "ready" && gotNamed[unique.DisplayName] != 1 {
			t.Fatalf("unique chest named %s count = %d, want 1; all=%+v", unique.DisplayName, gotNamed[unique.DisplayName], ev.StashItems)
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
			templateID:   "cave_ring",
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
	if len(itemsA) != enabledUniqueEffectCount(rules)+enabledNamedUniqueCount(rules)+enabledSetItemCount(rules) {
		t.Fatalf("item count = %d, want enabled effects + named uniques + set items", len(itemsA))
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
		if payloadNamedUniqueID(rules, a.DisplayName) != "" {
			namedCounts[a.DisplayName]++
		}
	}
	for _, uniqueID := range sortedStringKeys(rules.UniqueItems) {
		unique := rules.UniqueItems[uniqueID]
		if unique.Enabled && unique.Status == "ready" && namedCounts[unique.DisplayName] != 1 {
			t.Fatalf("named unique %s count = %d, want 1; all=%+v", unique.DisplayName, namedCounts[unique.DisplayName], namedCounts)
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
			if payloadNamedUniqueID(rules, payload.DisplayName) != "" {
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
	wantAmount := enabledUniqueEffectCount(rules) + enabledNamedUniqueCount(rules) + enabledSetItemCount(rules)
	if state == nil || len(state.items) != wantAmount {
		t.Fatalf("seeded chest count = %d, want %d; state=%+v", len(state.items), wantAmount, state)
	}
}

func payloadNamedUniqueID(rules *Rules, displayName string) string {
	for _, uniqueID := range sortedStringKeys(rules.UniqueItems) {
		unique := rules.UniqueItems[uniqueID]
		if unique.DisplayName == displayName {
			return uniqueID
		}
	}
	return ""
}

func TestSetItemPayloadsAndEquippedBonuses(t *testing.T) {
	rules := loadRules(t)
	setItemIDs := sortedStringKeys(rules.SetItems)
	wantSetItems := enabledSetItemCount(rules)
	if len(setItemIDs) != wantSetItems {
		t.Fatalf("set item count = %d, want %d", len(setItemIDs), wantSetItems)
	}
	payload, ok := rules.setItemPayload("verdant_vanguard_blade")
	if !ok {
		t.Fatal("setItemPayload returned false")
	}
	if payload.Rarity != "set" {
		t.Fatalf("set payload identity = %+v", payload)
	}
	if payload.DisplayName != "Verdant Vanguard Blade" || payload.Stats["damage_max"] != 7 || payload.Requirements["level"] != 5 {
		t.Fatalf("set payload fields = %+v", payload)
	}

	sim := MustNewSim("sess_set_items", "set_item_seed", rules)
	sim.progression.Level = 5
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["magic_bolt"] = 1
	addSetItemToInventory(t, sim, "verdant_vanguard_blade", 9101)
	addSetItemToInventory(t, sim, "verdant_vanguard_helm", 9102)
	addSetItemToInventory(t, sim, "verdant_vanguard_mail", 9103)
	addSetItemToInventory(t, sim, "verdant_vanguard_gloves", 9104)
	addSetItemToInventory(t, sim, "verdant_vanguard_boots", 9105)

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_blade", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9101", Slot: mainHandSlot}}}), "equip_blade")
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_helm", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9102", Slot: "head"}}}), "equip_helm")
	twoPiece := sim.equippedSetBonusStats()
	if twoPiece["armor"] != 3 || twoPiece["max_hp"] != 0 || sim.equippedItemStatTotal("skill_damage_percent") != 0 {
		t.Fatalf("two-piece set stats = %+v", twoPiece)
	}
	helmView := sim.itemView(sim.findItemByID(9102))
	if !containsShopString(helmView.SummaryLines, "Set: Verdant Vanguard (2/5 equipped)") ||
		!containsShopString(helmView.SummaryLines, "2-piece set bonus: Armor +3 (active)") ||
		!containsShopString(helmView.SummaryLines, "3-piece set bonus: Max HP +8 (inactive)") {
		t.Fatalf("two-piece summary lines = %+v", helmView.SummaryLines)
	}

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_mail", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9103", Slot: "chest"}}}), "equip_mail")
	threePiece := sim.equippedSetBonusStats()
	if threePiece["armor"] != 3 || threePiece["max_hp"] != 8 || threePiece["attack_speed_percent"] != 0 {
		t.Fatalf("three-piece set stats = %+v", threePiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_gloves", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9104", Slot: "gloves"}}}), "equip_gloves")
	fourPiece := sim.equippedSetBonusStats()
	if fourPiece["armor"] != 3 || fourPiece["max_hp"] != 8 || fourPiece["attack_speed_percent"] != 8 || fourPiece["all_skills"] != 0 {
		t.Fatalf("four-piece set stats = %+v", fourPiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_boots", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9105", Slot: "boots"}}}), "equip_boots")
	full := sim.equippedSetBonusStats()
	if full["armor"] != 3 || full["max_hp"] != 8 || full["attack_speed_percent"] != 8 || full["all_skills"] != 1 || full["skill_damage_percent"] != 20 {
		t.Fatalf("full set stats = %+v", full)
	}
	bladeView := sim.itemView(sim.findItemByID(9101))
	if !containsShopString(bladeView.SummaryLines, "Set: Verdant Vanguard (5/5 equipped)") ||
		!containsShopString(bladeView.SummaryLines, "5-piece set bonus: All skills +1, HP regen / 10s +5, Skill damage % +20% (active)") {
		t.Fatalf("full set summary lines = %+v", bladeView.SummaryLines)
	}
	if sim.effectiveSkillRank("magic_bolt") != 2 {
		t.Fatalf("effective magic_bolt rank = %d, want 2", sim.effectiveSkillRank("magic_bolt"))
	}
	if sim.equippedItemStatTotal("skill_damage_percent") != 20 {
		t.Fatalf("skill damage bonus = %d, want 20", sim.equippedItemStatTotal("skill_damage_percent"))
	}
	if regen := findStatBreakdown(sim.StatBreakdownViews(), "health_regen_per_second"); regen == nil || !statBreakdownHasSourceKind(*regen, "set_bonus") {
		t.Fatalf("full set health regen breakdown missing set bonus: %+v", sim.StatBreakdownViews())
	}
}

func TestSecondSetPackagePayloadsAndBonuses(t *testing.T) {
	rules := loadRules(t)
	if len(rules.SetCatalogs) < 2 {
		t.Fatalf("set catalog count = %d, want at least 2", len(rules.SetCatalogs))
	}
	payload, ok := rules.setItemPayload("stormrunner_covenant_bow")
	if !ok {
		t.Fatal("stormrunner setItemPayload returned false")
	}
	if payload.Rarity != "set" || payload.DisplayName != "Stormrunner Covenant Bow" || payload.ItemTemplateID != "cave_bow" {
		t.Fatalf("stormrunner payload identity = %+v", payload)
	}
	if payload.Stats["damage_min"] != 3 || payload.Stats["damage_max"] != 6 || payload.Stats["dex"] != 1 || payload.Requirements["level"] != 5 {
		t.Fatalf("stormrunner payload fields = %+v", payload)
	}

	sim := MustNewSim("sess_second_set_items", "second_set_seed", rules)
	sim.progression.Level = 5
	sim.progression.BaseStats.Dex = 8
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["magic_bolt"] = 1
	addSetItemToInventory(t, sim, "stormrunner_covenant_bow", 9201)
	addSetItemToInventory(t, sim, "stormrunner_covenant_mask", 9202)
	addSetItemToInventory(t, sim, "stormrunner_covenant_grips", 9203)
	addSetItemToInventory(t, sim, "stormrunner_covenant_treads", 9204)
	addSetItemToInventory(t, sim, "stormrunner_covenant_loop", 9205)

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9201", Slot: mainHandSlot}}}), "equip_bow")
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_mask", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9202", Slot: "head"}}}), "equip_mask")
	twoPiece := sim.equippedSetBonusStats()
	if twoPiece["dex"] != 2 || twoPiece["crit_chance"] != 0 || twoPiece["all_skills"] != 0 {
		t.Fatalf("stormrunner two-piece stats = %+v", twoPiece)
	}
	maskView := sim.itemView(sim.findItemByID(9202))
	if !containsShopString(maskView.SummaryLines, "Set: Stormrunner Covenant (2/5 equipped)") ||
		!containsShopString(maskView.SummaryLines, "2-piece set bonus: Dexterity +2 (active)") ||
		!containsShopString(maskView.SummaryLines, "5-piece set bonus: All skills +1, Skill damage % +15%, Magic Find +20% (inactive)") {
		t.Fatalf("stormrunner two-piece summary lines = %+v", maskView.SummaryLines)
	}

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_grips", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9203", Slot: "gloves"}}}), "equip_grips")
	threePiece := sim.equippedSetBonusStats()
	if threePiece["dex"] != 2 || threePiece["crit_chance"] != 6 || threePiece["attack_speed_percent"] != 0 {
		t.Fatalf("stormrunner three-piece stats = %+v", threePiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_treads", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9204", Slot: "boots"}}}), "equip_treads")
	fourPiece := sim.equippedSetBonusStats()
	if fourPiece["dex"] != 2 || fourPiece["crit_chance"] != 6 || fourPiece["attack_speed_percent"] != 6 || fourPiece["all_skills"] != 0 {
		t.Fatalf("stormrunner four-piece stats = %+v", fourPiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_loop", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9205", Slot: ringLeftSlot}}}), "equip_loop")
	full := sim.equippedSetBonusStats()
	if full["dex"] != 2 || full["crit_chance"] != 6 || full["attack_speed_percent"] != 6 || full["all_skills"] != 1 || full["skill_damage_percent"] != 15 || full["magic_find_percent"] != 20 {
		t.Fatalf("stormrunner full set stats = %+v", full)
	}
	bowView := sim.itemView(sim.findItemByID(9201))
	if !containsShopString(bowView.SummaryLines, "Set: Stormrunner Covenant (5/5 equipped)") ||
		!containsShopString(bowView.SummaryLines, "5-piece set bonus: All skills +1, Skill damage % +15%, Magic Find +20% (active)") {
		t.Fatalf("stormrunner full summary lines = %+v", bowView.SummaryLines)
	}
	if sim.effectiveSkillRank("magic_bolt") != 2 {
		t.Fatalf("effective magic_bolt rank = %d, want 2", sim.effectiveSkillRank("magic_bolt"))
	}
	if sim.equippedItemStatTotal("magic_find_percent") != 38 {
		t.Fatalf("magic find total = %d, want 38", sim.equippedItemStatTotal("magic_find_percent"))
	}
	if magicFind := findStatBreakdown(sim.StatBreakdownViews(), "magic_find_percent"); magicFind == nil || !statBreakdownHasSourceKind(*magicFind, "set_bonus") {
		t.Fatalf("full set magic find breakdown missing set bonus: %+v", sim.StatBreakdownViews())
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

func enabledSetItemCount(rules *Rules) int {
	count := 0
	for _, setItemID := range sortedStringKeys(rules.SetItems) {
		if rules.SetItems[setItemID].SetID != "" {
			count++
		}
	}
	return count
}

func findInventoryAddChange(changes []Change) *Change {
	for i := range changes {
		if changes[i].Op == OpInventoryAdd {
			return &changes[i]
		}
	}
	return nil
}

func addSetItemToInventory(t *testing.T, sim *Sim, setItemID string, instanceID uint64) {
	t.Helper()
	payload, ok := sim.rules.setItemPayload(setItemID)
	if !ok {
		t.Fatalf("missing set payload %s", setItemID)
	}
	addTestInventoryItem(sim, &invItem{
		instanceID:  instanceID,
		itemDefID:   payload.ItemTemplateID,
		slot:        sim.itemSlot(payload.ItemTemplateID, &payload),
		rollPayload: cloneRollPayload(&payload),
	})
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
