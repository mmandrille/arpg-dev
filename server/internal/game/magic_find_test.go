package game

import "testing"

func TestMagicFindDerivedStatFromEquipment(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_magic_find_stats", "mf-stat", rules)
	ring := addRolledInventoryItem(t, sim, 9701, "ring", map[string]int{"magic_find_percent": 25})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_mf", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "equip_mf")

	view := sim.CharacterProgressionView()
	if got := view.DerivedStats.MagicFindPercent; got != 25 {
		t.Fatalf("derived magic_find_percent = %v, want 25", got)
	}
	breakdown := findStatBreakdown(view.StatBreakdowns, "magic_find_percent")
	if breakdown == nil || breakdown.Value != 25 || !statBreakdownHasSourceKind(*breakdown, "equipment_roll") {
		t.Fatalf("magic find breakdown = %+v all=%+v", breakdown, view.StatBreakdowns)
	}
}

func TestMagicFindBiasesRarityWeights(t *testing.T) {
	rules := loadRules(t)

	if got, want := rules.magicFindAdjustedRarityWeight("common", 500), rules.Rarities["common"].Weight; got != want {
		t.Fatalf("common weight with magic find = %d, want %d", got, want)
	}
	for _, rarityID := range []string{"magic", "rare", "unique"} {
		base := rules.Rarities[rarityID].Weight
		if got, want := rules.magicFindAdjustedRarityWeight(rarityID, 100), base*2; got != want {
			t.Fatalf("%s weight with 100 magic find = %d, want %d", rarityID, got, want)
		}
	}
}

func TestMagicFindDoesNotChangeBaselineShopRoll(t *testing.T) {
	rules := loadRules(t)
	rng := NewRNG(SeedToUint64("shop-baseline-magic-find"))
	got, ok := rules.rollItemTemplateWithRNG("ring", rng, 1)
	if !ok {
		t.Fatal("baseline shop-style roll failed")
	}
	rng = NewRNG(SeedToUint64("shop-baseline-magic-find"))
	want, ok := rules.rollItemTemplateWithMagicFind("ring", rng, 1, 0)
	if !ok {
		t.Fatal("zero magic find roll failed")
	}
	if got.Rarity != want.Rarity || got.DisplayName != want.DisplayName {
		t.Fatalf("zero magic find changed baseline roll: got %+v want %+v", got, want)
	}
}
