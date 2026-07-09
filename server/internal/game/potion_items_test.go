package game

import (
	"strconv"
	"testing"
)

func TestPotionRestoreAmount(t *testing.T) {
	rules := loadRules(t)
	if got := rules.PotionRestoreAmount(10); got != 30 {
		t.Fatalf("restore amount = %d, want 30", got)
	}
}

func TestRejuvRestorePercent(t *testing.T) {
	rules := loadRules(t)
	if got := rules.RejuvRestorePercent(10); got != 33 {
		t.Fatalf("rejuv percent at 10 = %d, want 33", got)
	}
	if got := rules.RejuvRestorePercent(70); got != 70 {
		t.Fatalf("rejuv percent at 70 = %d, want 70", got)
	}
}

func TestResolvePotionDropKindDistribution(t *testing.T) {
	rules := loadRules(t)
	counts := map[string]int{}
	for seed := 0; seed < 1000; seed++ {
		id := rules.ResolvePotionDropKind(NewRNG(SeedToUint64("potion-drop-"+strconv.Itoa(seed))), "red_potion")
		counts[id]++
	}
	if counts[RejuvPotionItemDefID] < 150 || counts[RejuvPotionItemDefID] > 250 {
		t.Fatalf("rejuv drop count = %d, want ~200", counts[RejuvPotionItemDefID])
	}
}

func TestTreasureClassPotionDropVariety(t *testing.T) {
	rules := loadRules(t)
	counts := map[string]int{}
	for seed := uint64(0); seed < 5000; seed++ {
		for _, drop := range rules.RollTreasureClass("dungeon_mob_tc_1", NewRNG(seed)) {
			switch drop.ItemDefID {
			case "red_potion", "blue_potion", RejuvPotionItemDefID:
				counts[drop.ItemDefID]++
			}
		}
	}
	if counts["red_potion"] < 50 {
		t.Fatalf("red potion drops = %d, want healthy health potion share; counts=%+v", counts["red_potion"], counts)
	}
	if counts[RejuvPotionItemDefID] < 20 {
		t.Fatalf("rejuv potion drops = %d, want non-trivial rejuv share; counts=%+v", counts[RejuvPotionItemDefID], counts)
	}
	if counts["blue_potion"] < 50 {
		t.Fatalf("blue potion drops = %d, want healthy mana potion share; counts=%+v", counts["blue_potion"], counts)
	}
}

func TestPotionShopBuyPriceScalesWithLevel(t *testing.T) {
	rules := loadRules(t)
	if got := rules.PotionShopBuyPrice("red_potion", 10, 5); got != 400 {
		t.Fatalf("level-10 red buy price = %d, want 400", got)
	}
}

func TestLeveledPotionLootCarriesFloorLevel(t *testing.T) {
	sim := MustNewSim("sess_potion_depth", "01", loadRules(t))
	res := &TickResult{}
	sim.spawnLootDrops(
		[]LootDrop{{ItemDefID: "red_potion"}},
		Vec2{X: 1, Y: 1},
		1.0,
		"corr",
		res,
		goldRollContext{levelNum: -10},
	)
	loot := findLootByDef(sim, "red_potion")
	if loot == nil {
		loot = findLootByDef(sim, "blue_potion")
	}
	if loot == nil {
		loot = findLootByDef(sim, RejuvPotionItemDefID)
	}
	if loot == nil || loot.rollPayload == nil || loot.rollPayload.ItemLevel != 10 {
		t.Fatalf("loot payload = %+v, want level 10 potion", loot)
	}
}