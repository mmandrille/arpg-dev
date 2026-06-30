package game

import "testing"

func TestMaxItemLevelForDepthUsesTierBands(t *testing.T) {
	tiers := ItemLevelTierRules{LevelsPerTier: 10}
	cases := map[int]int{
		0:  1,
		1:  1,
		9:  1,
		10: 1,
		17: 1,
		25: 2,
		36: 3,
	}
	for depth, want := range cases {
		if got := MaxItemLevelForDepth(depth, tiers); got != want {
			t.Fatalf("depth %d max item level = %d, want %d", depth, got, want)
		}
	}
}

func TestRollItemLevelIsDeterministicAndWithinBand(t *testing.T) {
	rules := loadRules(t)
	tiers := rules.DungeonGeneration.ItemLevelTiers
	first, ok := rules.rollItemTemplateWithRNG("cave_blade", NewRNG(11), 36)
	if !ok {
		t.Fatal("roll cave_blade")
	}
	second, ok := rules.rollItemTemplateWithRNG("cave_blade", NewRNG(11), 36)
	if !ok {
		t.Fatal("repeat roll cave_blade")
	}
	if first.ItemLevel != second.ItemLevel {
		t.Fatalf("repeat item level = %d vs %d", first.ItemLevel, second.ItemLevel)
	}
	if first.ItemLevel < 1 || first.ItemLevel > MaxItemLevelForDepth(36, tiers) {
		t.Fatalf("item level %d outside 1..%d", first.ItemLevel, MaxItemLevelForDepth(36, tiers))
	}
}

func TestRolledItemLevelAtShallowDepthIsOne(t *testing.T) {
	rules := loadRules(t)
	low, ok := rules.rollItemTemplateWithRNG("cave_blade", NewRNG(11), 0)
	if !ok {
		t.Fatal("roll low-depth cave_blade")
	}
	if low.ItemLevel != 1 {
		t.Fatalf("low item level = %d, want 1", low.ItemLevel)
	}
}

func TestItemLevelScalingMatchesMonsterDamageBand(t *testing.T) {
	rules := loadRules(t)
	def, ok := rules.Monsters["dungeon_mob"]
	if !ok {
		t.Fatal("dungeon_mob missing")
	}
	rarity := rules.DungeonGeneration.MonsterRarities[0]
	sim := &Sim{rules: rules}
	monster := sim.generatedMonsterStats(def, -11, rarity)
	itemLevel := 2
	tiers := rules.DungeonGeneration.ItemLevelTiers
	template := rules.ItemTemplates["cave_blade"]
	payload := FinalizeItemRollPayload(ItemRollPayload{
		ItemTemplateID: "cave_blade",
		DisplayName:    template.Name,
		Rarity:         "common",
		ItemLevel:      1,
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
	}, itemLevel, rules.DungeonGeneration.MonsterDepthScaling, tiers)
	if payload.Stats["damage_max"] <= template.BaseStats["damage_max"] {
		t.Fatalf("tier 2 damage_max = %d, want above base %d", payload.Stats["damage_max"], template.BaseStats["damage_max"])
	}
	if monster.attackDamage == nil {
		t.Fatal("expected monster attack damage")
	}
	wantFactor := DepthFactor(rules.DungeonGeneration.MonsterDepthScaling.DamagePerDepth, DepthIndexForItemLevel(itemLevel, tiers))
	wantDamageMax := roundPositive(float64(template.BaseStats["damage_max"]) * wantFactor)
	if payload.Stats["damage_max"] != wantDamageMax {
		t.Fatalf("tier 2 damage_max = %d, want %d", payload.Stats["damage_max"], wantDamageMax)
	}
	if monster.attackDamage.Max != roundPositive(float64(def.AttackDamage.Max)*wantFactor) {
		t.Fatalf("monster damage_max = %d, want %d", monster.attackDamage.Max, roundPositive(float64(def.AttackDamage.Max)*wantFactor))
	}
}

func TestUpgradeItemLevelPayloadRescalesProportionally(t *testing.T) {
	rules := loadRules(t)
	scaling := rules.DungeonGeneration.MonsterDepthScaling
	tiers := rules.DungeonGeneration.ItemLevelTiers
	stats := map[string]int{"damage_min": 2, "damage_max": 4, "item_level": 1}
	nextStats, _, nextLevel, err := UpgradeItemLevelPayload(stats, map[string]int{"level": 1}, 1, scaling, tiers)
	if err != nil {
		t.Fatal(err)
	}
	if nextLevel != 2 {
		t.Fatalf("next level = %d, want 2", nextLevel)
	}
	if nextStats["damage_max"] <= stats["damage_max"] {
		t.Fatalf("upgraded damage_max = %d, want > %d", nextStats["damage_max"], stats["damage_max"])
	}
}
