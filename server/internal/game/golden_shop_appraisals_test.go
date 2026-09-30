package game

import (
	"encoding/json"
	"testing"
)

type shopAppraisalExpected struct {
	DisplayName  string   `json:"display_name"`
	Slot         string   `json:"slot"`
	Category     string   `json:"category"`
	BuyPrice     int      `json:"buy_price"`
	SellPrice    int      `json:"sell_price"`
	SummaryLines []string `json:"summary_lines"`
	Comparison   []struct {
		Stat     string `json:"stat"`
		Offered  int    `json:"offered"`
		Equipped int    `json:"equipped"`
		Delta    int    `json:"delta"`
	} `json:"comparison"`
}

func assertShopComparison(t *testing.T, label string, got *ShopComparisonView, want shopAppraisalExpected) {
	t.Helper()
	if len(want.Comparison) == 0 {
		if got != nil && len(got.Deltas) != 0 {
			t.Fatalf("%s comparison = %+v, want none", label, got)
		}
		return
	}
	if got == nil {
		t.Fatalf("%s comparison missing, want %+v", label, want.Comparison)
	}
	if len(got.Deltas) != len(want.Comparison) {
		t.Fatalf("%s comparison deltas = %+v, want %+v", label, got.Deltas, want.Comparison)
	}
	for i, w := range want.Comparison {
		d := got.Deltas[i]
		if d.Stat != w.Stat || d.Offered != w.Offered || d.Equipped != w.Equipped || d.Delta != w.Delta {
			t.Fatalf("%s comparison[%d] = %+v, want %+v", label, i, d, w)
		}
	}
}

// assertSummaryLinesContain requires every golden summary line to appear in got,
// in order. Extra server-authored lines (e.g. weapon range) are tolerated: the
// golden pins the stable lines, not the full tooltip.
func assertSummaryLinesContain(t *testing.T, label string, got, want []string) {
	t.Helper()
	next := 0
	for _, line := range got {
		if next < len(want) && line == want[next] {
			next++
		}
	}
	if next != len(want) {
		t.Fatalf("%s summary_lines = %q, want ordered superset of %q (missing %q)", label, got, want, want[next])
	}
}

// TestShopAppraisalsGolden asserts shared/golden/shop_appraisals.json against the
// Go shop: fixed offer view, generated offer view with equipped comparison,
// sell appraisal, and equipped-item exclusion from sell appraisals.
func TestShopAppraisalsGolden(t *testing.T) {
	var golden struct {
		ShopID     string `json:"shop_id"`
		FixedOffer struct {
			ItemDefID string                `json:"item_def_id"`
			Expected  shopAppraisalExpected `json:"expected"`
		} `json:"fixed_offer"`
		GeneratedOffer struct {
			ItemTemplateID string                `json:"item_template_id"`
			Rarity         string                `json:"rarity"`
			RolledStats    map[string]int        `json:"rolled_stats"`
			EquippedStats  map[string]int        `json:"equipped_stats"`
			Expected       shopAppraisalExpected `json:"expected"`
		} `json:"generated_offer"`
		SellAppraisal struct {
			ItemInstanceID string                `json:"item_instance_id"`
			ItemTemplateID string                `json:"item_template_id"`
			Rarity         string                `json:"rarity"`
			RolledStats    map[string]int        `json:"rolled_stats"`
			Expected       shopAppraisalExpected `json:"expected"`
		} `json:"sell_appraisal"`
		EquippedItemExclusion struct {
			ItemInstanceID  string `json:"item_instance_id"`
			ExpectedExclude bool   `json:"expected_excluded"`
		} `json:"equipped_item_exclusion"`
	}
	loadGolden(t, "shop_appraisals.json", &golden)

	// deepest depth 1 == potion tier level 1, matching the golden's level-1 potion.
	sim := newTownVendorSim(t, 5000, 1)
	shop, ok := sim.rules.Shops[golden.ShopID]
	if !ok {
		t.Fatalf("unknown golden shop %q", golden.ShopID)
	}

	// --- inventory fixture: equipped main-hand weapon, unequipped sellable item, equipped item to exclude.
	genTemplate := sim.rules.ItemTemplates[golden.GeneratedOffer.ItemTemplateID]
	equippedWeapon := &invItem{
		instanceID: 7001, itemDefID: golden.GeneratedOffer.ItemTemplateID, slot: mainHandSlot, equipped: true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: golden.GeneratedOffer.ItemTemplateID,
			DisplayName:    genTemplate.Name,
			Rarity:         "common",
			Stats:          golden.GeneratedOffer.EquippedStats,
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	sellTemplate := sim.rules.ItemTemplates[golden.SellAppraisal.ItemTemplateID]
	sellName := sim.rules.affixDisplayName(sellTemplate, golden.SellAppraisal.Rarity, golden.SellAppraisal.RolledStats)
	sellItem := &invItem{
		instanceID: 2001, itemDefID: golden.SellAppraisal.ItemTemplateID, slot: offHandSlot,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: golden.SellAppraisal.ItemTemplateID,
			DisplayName:    sellName,
			Rarity:         golden.SellAppraisal.Rarity,
			Stats:          golden.SellAppraisal.RolledStats,
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	excludedItem := &invItem{
		instanceID: 2002, itemDefID: golden.SellAppraisal.ItemTemplateID,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: golden.SellAppraisal.ItemTemplateID,
			DisplayName:    sellName,
			Rarity:         golden.SellAppraisal.Rarity,
			Stats:          golden.SellAppraisal.RolledStats,
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	sim.inventory = append(sim.inventory, equippedWeapon, sellItem, excludedItem)
	sim.equipped[mainHandSlot] = equippedWeapon.instanceID
	// The golden's excluded instance is "equipped": put it in a slot that is not the
	// weapon slot, marking it equipped without disturbing the comparison baseline.
	excludedItem.equipped = true
	excludedItem.slot = "ring_left"
	sim.equipped[ringLeftSlot] = excludedItem.instanceID

	// --- generated offer: name, price, view, comparison.
	t.Run("generated_offer", func(t *testing.T) {
		g := golden.GeneratedOffer
		exp := g.Expected
		gotName := sim.rules.affixDisplayName(genTemplate, g.Rarity, g.RolledStats)
		if gotName != exp.DisplayName {
			t.Fatalf("generated display name = %q, want %q", gotName, exp.DisplayName)
		}
		buyPrice, ok := shop.generatedBuyPrice(g.ItemTemplateID, g.Rarity, g.RolledStats, sim.rules)
		if !ok || buyPrice != exp.BuyPrice {
			t.Fatalf("generated buy price = %d (ok=%v), want %d", buyPrice, ok, exp.BuyPrice)
		}
		if genTemplate.Slot != exp.Slot || genTemplate.Category != exp.Category {
			t.Fatalf("generated slot/category = %q/%q, want %q/%q", genTemplate.Slot, genTemplate.Category, exp.Slot, exp.Category)
		}
		payload, err := json.Marshal(ItemRollPayload{
			ItemTemplateID: g.ItemTemplateID,
			DisplayName:    gotName,
			Rarity:         g.Rarity,
			Stats:          g.RolledStats,
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		})
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		sim.LoadShopStock([]PersistedShopStockItem{{
			ShopID: golden.ShopID, RefreshKey: sim.shopRefreshKey(), OfferIndex: 0,
			OfferID: "generated:golden:000", SourceDepth: 2, ItemTemplateID: g.ItemTemplateID,
			RolledPayload: payload, BuyPrice: buyPrice, Available: true,
		}})
		offers, ok := sim.shopCatalog(golden.ShopID)
		if !ok {
			t.Fatalf("shop catalog %q unavailable", golden.ShopID)
		}
		offer := findOffer(offers, "generated:golden:000")
		if offer == nil {
			t.Fatalf("generated golden offer missing from catalog: %+v", offers)
		}
		if offer.DisplayName != exp.DisplayName || offer.Slot != exp.Slot || offer.Category != exp.Category || offer.BuyPrice != exp.BuyPrice {
			t.Fatalf("generated offer view = %+v, want %+v", offer, exp)
		}
		assertSummaryLinesContain(t, "generated offer", offer.SummaryLines, exp.SummaryLines)
		assertShopComparison(t, "generated offer", offer.Comparison, exp)
	})

	// --- sell appraisal.
	t.Run("sell_appraisal", func(t *testing.T) {
		exp := golden.SellAppraisal.Expected
		buyPrice, ok := shop.generatedBuyPrice(golden.SellAppraisal.ItemTemplateID, golden.SellAppraisal.Rarity, golden.SellAppraisal.RolledStats, sim.rules)
		if !ok {
			t.Fatal("sell item has no generated buy price")
		}
		if got := shop.sellPrice(buyPrice); got != exp.SellPrice {
			t.Fatalf("sell price from buy %d = %d, want %d", buyPrice, got, exp.SellPrice)
		}
		if got, ok := sim.inventorySellPrice(golden.ShopID, sellItem); !ok || got != exp.SellPrice {
			t.Fatalf("inventorySellPrice = %d (ok=%v), want %d", got, ok, exp.SellPrice)
		}
		var appraisal *ShopSellAppraisalView
		appraisals := sim.shopSellAppraisals(golden.ShopID)
		for i := range appraisals {
			if appraisals[i].ItemInstanceID == golden.SellAppraisal.ItemInstanceID {
				appraisal = &appraisals[i]
			}
		}
		if appraisal == nil {
			t.Fatalf("sell appraisal for %s missing: %+v", golden.SellAppraisal.ItemInstanceID, appraisals)
		}
		if appraisal.DisplayName != exp.DisplayName || appraisal.Slot != exp.Slot || appraisal.Category != exp.Category || appraisal.SellPrice != exp.SellPrice {
			t.Fatalf("sell appraisal = %+v, want %+v", appraisal, exp)
		}
		assertSummaryLinesContain(t, "sell appraisal", appraisal.SummaryLines, exp.SummaryLines)
		assertShopComparison(t, "sell appraisal", appraisal.Comparison, exp)
	})

	// --- equipped item exclusion.
	t.Run("equipped_item_exclusion", func(t *testing.T) {
		if !golden.EquippedItemExclusion.ExpectedExclude {
			t.Fatal("golden equipped_item_exclusion must expect exclusion")
		}
		for _, appraisal := range sim.shopSellAppraisals(golden.ShopID) {
			if appraisal.ItemInstanceID == golden.EquippedItemExclusion.ItemInstanceID {
				t.Fatalf("equipped item %s appears in sell appraisals: %+v", appraisal.ItemInstanceID, appraisal)
			}
		}
	})

	// --- fixed offer.
	t.Run("fixed_offer", func(t *testing.T) {
		exp := golden.FixedOffer.Expected
		offers, ok := sim.shopCatalog(golden.ShopID)
		if !ok {
			t.Fatalf("shop catalog %q unavailable", golden.ShopID)
		}
		offer := findOffer(offers, "fixed:"+golden.FixedOffer.ItemDefID)
		if offer == nil {
			t.Fatalf("fixed offer for %s missing", golden.FixedOffer.ItemDefID)
		}
		if offer.DisplayName != exp.DisplayName || offer.Slot != exp.Slot || offer.Category != exp.Category || offer.BuyPrice != exp.BuyPrice {
			t.Fatalf("fixed offer view = %+v, want %+v", offer, exp)
		}
		// Load-bearing heal line (the contract tools/validate_shared.py also pins).
		if !containsShopString(offer.SummaryLines, "Restores 3 HP") {
			t.Fatalf("fixed offer summary_lines = %q, want to contain %q", offer.SummaryLines, "Restores 3 HP")
		}
		// Full golden summary_lines, in order. Leveled potions render "Level N" instead
		// of "Kind: consumable", so this is a golden-vs-Go disagreement (see report).
		assertSummaryLinesContain(t, "fixed offer", offer.SummaryLines, exp.SummaryLines)
	})
}
