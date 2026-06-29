package game

import "testing"

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

func TestWayfarersAccordSetPayloadsAndBonuses(t *testing.T) {
	rules := loadRules(t)
	if len(rules.SetCatalogs) < 3 {
		t.Fatalf("set catalog count = %d, want at least 3", len(rules.SetCatalogs))
	}
	payload, ok := rules.setItemPayload("wayfarers_accord_pendant")
	if !ok {
		t.Fatal("wayfarers setItemPayload returned false")
	}
	if payload.Rarity != "set" || payload.DisplayName != "Wayfarer's Accord Pendant" || payload.ItemTemplateID != "cave_amulet" {
		t.Fatalf("wayfarers payload identity = %+v", payload)
	}
	if payload.Stats["max_mana"] != 5 || payload.Stats["magic"] != 1 || payload.Stats["max_hp"] != 3 || payload.Requirements["level"] != 5 {
		t.Fatalf("wayfarers payload fields = %+v", payload)
	}

	sim := MustNewSim("sess_wayfarers_accord", "wayfarers_accord_seed", rules)
	sim.progression.Level = 5
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["magic_bolt"] = 1
	addSetItemToInventory(t, sim, "wayfarers_accord_hood", 9301)
	addSetItemToInventory(t, sim, "wayfarers_accord_robes", 9302)
	addSetItemToInventory(t, sim, "wayfarers_accord_bindings", 9303)
	addSetItemToInventory(t, sim, "wayfarers_accord_treads", 9304)
	addSetItemToInventory(t, sim, "wayfarers_accord_pendant", 9305)

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_hood", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9301", Slot: "head"}}}), "equip_hood")
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_robes", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9302", Slot: "chest"}}}), "equip_robes")
	twoPiece := sim.equippedSetBonusStats()
	if twoPiece["str"] != 1 || twoPiece["dex"] != 1 || twoPiece["vit"] != 1 || twoPiece["magic"] != 1 || twoPiece["max_hp"] != 0 {
		t.Fatalf("wayfarers two-piece stats = %+v", twoPiece)
	}
	hoodView := sim.itemView(sim.findItemByID(9301))
	if !containsShopString(hoodView.SummaryLines, "Set: Wayfarer's Accord (2/5 equipped)") ||
		!containsShopString(hoodView.SummaryLines, "2-piece set bonus: Strength +1, Dexterity +1, Vitality +1, Magic +1 (active)") ||
		!containsShopString(hoodView.SummaryLines, "5-piece set bonus: All skills +1, HP regen / 10s +5, Skill cooldown reduction +5% (inactive)") {
		t.Fatalf("wayfarers two-piece summary lines = %+v", hoodView.SummaryLines)
	}

	assertAck(t, sim.Tick([]Input{{MessageID: "equip_bindings", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9303", Slot: "gloves"}}}), "equip_bindings")
	threePiece := sim.equippedSetBonusStats()
	if threePiece["max_hp"] != 10 || threePiece["max_mana"] != 8 || threePiece["skill_damage_percent"] != 0 {
		t.Fatalf("wayfarers three-piece stats = %+v", threePiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_treads", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9304", Slot: "boots"}}}), "equip_treads")
	fourPiece := sim.equippedSetBonusStats()
	if fourPiece["skill_damage_percent"] != 10 || fourPiece["all_skills"] != 0 {
		t.Fatalf("wayfarers four-piece stats = %+v", fourPiece)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_pendant", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "9305", Slot: "amulet"}}}), "equip_pendant")
	full := sim.equippedSetBonusStats()
	if full["all_skills"] != 1 || full["skill_cooldown_reduction_percent"] != 5 || full["health_regen_per_10_seconds"] != 5 || full["skill_damage_percent"] != 10 {
		t.Fatalf("wayfarers full set stats = %+v", full)
	}
	pendantView := sim.itemView(sim.findItemByID(9305))
	if !containsShopString(pendantView.SummaryLines, "Set: Wayfarer's Accord (5/5 equipped)") ||
		!containsShopString(pendantView.SummaryLines, "5-piece set bonus: All skills +1, HP regen / 10s +5, Skill cooldown reduction +5% (active)") {
		t.Fatalf("wayfarers full summary lines = %+v", pendantView.SummaryLines)
	}
	if sim.effectiveSkillRank("magic_bolt") != 2 {
		t.Fatalf("effective magic_bolt rank = %d, want 2", sim.effectiveSkillRank("magic_bolt"))
	}
	if sim.equippedItemStatTotal("skill_damage_percent") != 10 {
		t.Fatalf("skill damage bonus = %d, want 10", sim.equippedItemStatTotal("skill_damage_percent"))
	}
	if regen := findStatBreakdown(sim.StatBreakdownViews(), "health_regen_per_second"); regen == nil || !statBreakdownHasSourceKind(*regen, "set_bonus") {
		t.Fatalf("wayfarers health regen breakdown missing set bonus: %+v", sim.StatBreakdownViews())
	}
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
