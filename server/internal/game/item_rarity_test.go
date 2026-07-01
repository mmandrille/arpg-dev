package game

import "testing"

func TestGeneratedMagicAndRareAffixPools(t *testing.T) {
	rules := loadRules(t)
	base := []RollableStatDef{
		{Stat: "max_hp", Min: 1, Max: 1, Weight: 1},
		{Stat: "max_mana", MinRarity: "magic", Min: 1, Max: 1, Weight: 1},
		{Stat: "attack_speed_percent", MinRarity: "rare", Min: 1, Max: 1, Weight: 1},
	}
	common := rules.rollableStatsForRarity(base, "common", 1)
	if _, ok := findRollableStat(common, "max_mana"); ok {
		t.Fatal("common pool should not include magic-gated max_mana")
	}
	magic := rules.rollableStatsForRarity(base, "magic", 1)
	if _, ok := findRollableStat(magic, "max_mana"); !ok {
		t.Fatal("magic pool missing magic-gated max_mana")
	}
	if _, ok := findRollableStat(magic, "attack_speed_percent"); ok {
		t.Fatal("magic pool should not include rare-gated attack_speed_percent")
	}
	for _, stat := range []string{"str", "dex", "vit", "magic"} {
		roll, ok := findRollableStat(magic, stat)
		if !ok {
			t.Fatalf("magic pool missing %s", stat)
		}
		if roll.Min != 1 || roll.Max != 3 {
			t.Fatalf("magic %s range = %d-%d, want 1-3", stat, roll.Min, roll.Max)
		}
	}
	if _, ok := findRollableStat(magic, "all_skills"); ok {
		t.Fatal("magic pool should not include all_skills")
	}
	rare := rules.rollableStatsForRarity(base, "rare", 20)
	if _, ok := findRollableStat(rare, "attack_speed_percent"); !ok {
		t.Fatal("rare pool missing rare-gated attack_speed_percent")
	}
	roll, ok := findRollableStat(rare, "all_skills")
	if !ok {
		t.Fatal("rare depth 20 pool missing all_skills")
	}
	if roll.Min != 1 || roll.Max != 2 {
		t.Fatalf("rare depth 20 all_skills range = %d-%d, want 1-2", roll.Min, roll.Max)
	}
}

func TestRarityRollCountRanges(t *testing.T) {
	rules := loadRules(t)
	cases := map[string][2]int{
		"common": {1, 1},
		"magic":  {1, 2},
		"rare":   {2, 4},
		"unique": {3, 5},
		"set":    {3, 5},
	}
	for rarityID, want := range cases {
		rarity, ok := rules.Rarities[rarityID]
		if !ok {
			t.Fatalf("missing rarity %s", rarityID)
		}
		if rarity.StatRollsMin != want[0] || rarity.StatRollsMax != want[1] {
			t.Fatalf("%s roll count = %d-%d, want %d-%d", rarityID, rarity.StatRollsMin, rarity.StatRollsMax, want[0], want[1])
		}
	}
	if rules.rarityRandomRollable("set") {
		t.Fatal("set rarity should not enter random item rolls")
	}
}

func TestRolledItemLevelFollowsSourceDepth(t *testing.T) {
	rules := loadRules(t)
	tiers := rules.DungeonGeneration.ItemLevelTiers
	deep, ok := rules.rollItemTemplateWithRNG("long_sword", NewRNG(11), 36)
	if !ok {
		t.Fatal("roll deep long_sword")
	}
	maxLevel := MaxItemLevelForDepth(36, tiers)
	if deep.ItemLevel < 1 || deep.ItemLevel > maxLevel {
		t.Fatalf("deep item level = %d, want 1..%d", deep.ItemLevel, maxLevel)
	}
}

func TestJewelryTemplatesCanRollInventoryRows(t *testing.T) {
	rules := loadRules(t)
	for _, templateID := range []string{"ring", "amulet"} {
		template := rules.ItemTemplates[templateID]
		roll, ok := findRollableStat(template.RollableStats, "inventory_rows")
		if !ok {
			t.Fatalf("%s rollable stats missing inventory_rows", templateID)
		}
		if roll.Min < 1 || roll.Max < roll.Min {
			t.Fatalf("%s inventory_rows roll range = %d-%d, want positive valid range", templateID, roll.Min, roll.Max)
		}
	}
}

func TestAffixDisplayNameUsesSkillAffixFamily(t *testing.T) {
	rules := loadRules(t)
	template := rules.ItemTemplates["starter_sorcerer_staff"]
	stats := cloneIntMap(template.BaseStats)
	stats["skill_cooldown_reduction_percent"] = 12
	stats["skill_mana_cost_reduction"] = 1

	got := rules.affixDisplayName(template, "rare", stats)
	if got != "Focused Sorcerer Staff" {
		t.Fatalf("affix display name = %q, want Focused Sorcerer Staff", got)
	}
}

func TestAffixDisplayNamePreservesCommonRarityName(t *testing.T) {
	rules := loadRules(t)
	template := rules.ItemTemplates["long_sword"]
	stats := cloneIntMap(template.BaseStats)
	stats["damage_max"] += 2

	got := rules.affixDisplayName(template, "common", stats)
	if got != "Long Sword" {
		t.Fatalf("common affix display name = %q, want Long Sword", got)
	}
}

func findRollableStat(stats []RollableStatDef, stat string) (RollableStatDef, bool) {
	for _, roll := range stats {
		if roll.Stat == stat {
			return roll, true
		}
	}
	return RollableStatDef{}, false
}
