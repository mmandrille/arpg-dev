package game

import (
	"fmt"
	"reflect"
	"testing"
)

func TestShopPricingGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		ShopID string `json:"shop_id"`
		Cases  []struct {
			Name  string `json:"name"`
			Input struct {
				ItemDefID      string         `json:"item_def_id"`
				ItemTemplateID string         `json:"item_template_id"`
				Rarity         string         `json:"rarity"`
				RolledStats    map[string]int `json:"rolled_stats"`
			} `json:"input"`
			Expected struct {
				BuyPrice  int `json:"buy_price"`
				SellPrice int `json:"sell_price"`
			} `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "shop_pricing.json", &golden)

	shop, ok := rules.Shops[golden.ShopID]
	if !ok {
		t.Fatalf("missing shop %s", golden.ShopID)
	}
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var buy int
			var ok bool
			if c.Input.ItemTemplateID != "" {
				buy, ok = shop.generatedBuyPrice(c.Input.ItemTemplateID, c.Input.Rarity, c.Input.RolledStats, rules)
			} else {
				buy, ok = shop.fixedBuyPrice(c.Input.ItemDefID)
			}
			if !ok {
				t.Fatalf("price lookup failed for %+v", c.Input)
			}
			if buy != c.Expected.BuyPrice {
				t.Fatalf("buy price = %d, want %d", buy, c.Expected.BuyPrice)
			}
			if sell := shop.sellPrice(buy); sell != c.Expected.SellPrice {
				t.Fatalf("sell price = %d, want %d", sell, c.Expected.SellPrice)
			}
		})
	}
}

func TestShopGeneratedOfferGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		ShopID      string `json:"shop_id"`
		Seed        string `json:"seed"`
		CharacterID string `json:"character_id"`
		Cases       []struct {
			Name                string            `json:"name"`
			DeepestDungeonDepth int               `json:"deepest_dungeon_depth"`
			ExpectedOfferCount  int               `json:"expected_offer_count"`
			Expected            []shopOfferGolden `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "shop_offers.json", &golden)

	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sim := MustNewSim(golden.ShopID+"_"+c.Name, golden.Seed, rules)
			offers := sim.generatedShopOffers(golden.ShopID, rules.Shops[golden.ShopID], golden.CharacterID, c.DeepestDungeonDepth)
			if len(offers) != c.ExpectedOfferCount {
				t.Fatalf("generated offers = %d, want %d: %+v", len(offers), c.ExpectedOfferCount, offers)
			}
			got := make([]shopOfferGolden, 0, len(offers))
			for _, offer := range offers {
				got = append(got, shopOfferGoldenFromView(offer))
			}
			if !reflect.DeepEqual(got, c.Expected) {
				t.Fatalf("generated offers drift:\ngot  %+v\nwant %+v", got, c.Expected)
			}
		})
	}
}

func TestShopRulesValidationRejectsBadReferencesAndPricing(t *testing.T) {
	t.Run("missing fixed item", func(t *testing.T) {
		rules := cloneRules(loadRules(t))
		shop := rules.Shops["town_vendor"]
		shop.FixedOffers[0].ItemDefID = "missing_item"
		rules.Shops["town_vendor"] = shop
		if err := validateShopRules(rules); err == nil {
			t.Fatal("validateShopRules accepted missing fixed item")
		}
	})

	t.Run("bad multiplier", func(t *testing.T) {
		rules := cloneRules(loadRules(t))
		shop := rules.Shops["town_vendor"]
		shop.Pricing.RarityMultipliers["common"] = 0
		rules.Shops["town_vendor"] = shop
		if err := validateShopRules(rules); err == nil {
			t.Fatal("validateShopRules accepted non-positive rarity multiplier")
		}
	})
}

func TestShopOpenBuyAndSell(t *testing.T) {
	sim := newTownVendorSim(t, 250, 3)
	vendor := townVendorEntity(t, sim)

	open := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_open_shop",
		Action:    &ActionIntent{TargetID: idStr(vendor.id)},
	}})
	if !hasAck(open, "msg_open_shop") {
		t.Fatalf("open shop was not acked: %+v", open)
	}
	opened := findEvent(open.Events, "shop_opened")
	if opened == nil || opened.ShopID != "town_vendor" || len(opened.Offers) != 7 {
		t.Fatalf("shop_opened event = %+v", opened)
	}

	buyPotion := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_potion",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: "fixed:red_potion"},
	}})
	if !hasAck(buyPotion, "msg_buy_potion") || sim.gold != 230 || len(sim.inventory) != 1 || sim.inventory[0].itemDefID != "red_potion" {
		t.Fatalf("buy potion result gold=%d inv=%+v res=%+v", sim.gold, sim.inventory, buyPotion)
	}
	if ev := findEvent(buyPotion.Events, "shop_purchase"); ev == nil || ev.Price == nil || *ev.Price != 20 || ev.TotalGold == nil || *ev.TotalGold != 230 {
		t.Fatalf("shop_purchase event = %+v", ev)
	}

	generated := firstGeneratedOffer(t, sim)
	beforeGold := sim.gold
	buyGenerated := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_generated",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: generated.OfferID},
	}})
	if !hasAck(buyGenerated, "msg_buy_generated") {
		t.Fatalf("buy generated was not acked: %+v", buyGenerated)
	}
	if sim.gold != beforeGold-generated.BuyPrice {
		t.Fatalf("gold after generated buy = %d, want %d", sim.gold, beforeGold-generated.BuyPrice)
	}
	bought := sim.inventory[len(sim.inventory)-1]
	if bought.rollPayload == nil || bought.rollPayload.ItemTemplateID != generated.ItemTemplateID {
		t.Fatalf("generated purchase item = %+v, offer %+v", bought, generated)
	}

	price, ok := sim.inventorySellPrice("town_vendor", bought)
	if !ok {
		t.Fatalf("sell price lookup failed for %+v", bought)
	}
	sell := sim.Tick([]Input{{
		Type:      "shop_sell_intent",
		MessageID: "msg_sell_generated",
		ShopSell:  &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: idStr(bought.instanceID)},
	}})
	if !hasAck(sell, "msg_sell_generated") || sim.findItemByID(bought.instanceID) != nil {
		t.Fatalf("sell generated result inv=%+v res=%+v", sim.inventory, sell)
	}
	if ev := findEvent(sell.Events, "shop_sale"); ev == nil || ev.Price == nil || *ev.Price != price {
		t.Fatalf("shop_sale event = %+v, want price %d", ev, price)
	}
}

func TestMysterySellerOpenShowsConcealedOffers(t *testing.T) {
	sim := newMysterySellerSim(t, 1000, 3)
	seller := townMysterySellerEntity(t, sim)

	open := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_open_mystery",
		Action:    &ActionIntent{TargetID: idStr(seller.id)},
	}})
	if !hasAck(open, "msg_open_mystery") {
		t.Fatalf("open mystery seller was not acked: %+v", open)
	}
	opened := findEvent(open.Events, "shop_opened")
	if opened == nil || opened.ShopID != "town_mystery_seller" {
		t.Fatalf("shop_opened event = %+v", opened)
	}
	if got, want := countShopOffersByKind(opened.Offers, shopOfferKindMystery), len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots); got != want {
		t.Fatalf("mystery offer count = %d, want %d: %+v", got, want, opened.Offers)
	}
	seenSlots := map[string]bool{}
	for _, offer := range opened.Offers {
		if offer.Kind != shopOfferKindMystery {
			t.Fatalf("unexpected mystery seller offer kind: %+v", offer)
		}
		if !offer.Concealed || offer.MysteryLabel == "" || offer.BuyPrice <= 0 {
			t.Fatalf("mystery offer missing concealed fields: %+v", offer)
		}
		if offer.ItemDefID != "" || offer.ItemTemplateID != "" || offer.DisplayName != "" || offer.Rarity != "" || len(offer.RolledStats) != 0 || offer.Comparison != nil || offer.EquipPreview != nil {
			t.Fatalf("mystery offer leaked item identity/detail: %+v", offer)
		}
		if offer.SourceDepthMin < 1 || offer.SourceDepthMax < offer.SourceDepthMin {
			t.Fatalf("mystery offer source window invalid: %+v", offer)
		}
		seenSlots[offer.Slot] = true
	}
	for _, slot := range sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots {
		if !seenSlots[slot] {
			t.Fatalf("missing mystery offer for slot %s in %+v", slot, opened.Offers)
		}
	}
}

func TestMysterySellerPurchaseRevealsAndConsumesOffer(t *testing.T) {
	sim := newMysterySellerSim(t, 2000, 3)
	seller := townMysterySellerEntity(t, sim)
	offer := firstMysteryOffer(t, sim)
	beforeGold := sim.gold

	buy := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_mystery",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(seller.id), OfferID: offer.OfferID},
	}})
	if !hasAck(buy, "msg_buy_mystery") {
		t.Fatalf("buy mystery was not acked: %+v", buy)
	}
	if !hasShopStockAvailability(buy, offer.OfferID, false) {
		t.Fatalf("mystery purchase did not consume stock: %+v", buy.Changes)
	}
	if sim.gold != beforeGold-offer.BuyPrice {
		t.Fatalf("gold after mystery buy = %d, want %d", sim.gold, beforeGold-offer.BuyPrice)
	}
	bought := sim.inventory[len(sim.inventory)-1]
	if bought.rollPayload == nil || !mysteryRarityAllowed(bought.rollPayload.Rarity, "magic", "rare") {
		t.Fatalf("mystery purchase item = %+v", bought)
	}
	ev := findEvent(buy.Events, "shop_purchase")
	if ev == nil || ev.Item == nil || ev.Item.ItemTemplateID == "" || ev.Item.Rarity == "" {
		t.Fatalf("shop_purchase reveal event = %+v", ev)
	}
	if ev.Item.ItemTemplateID != bought.rollPayload.ItemTemplateID || ev.Item.ItemInstanceID != idStr(bought.instanceID) {
		t.Fatalf("reveal item %+v does not match inventory item %+v", ev.Item, bought)
	}
	if findOffer(ev.Offers, offer.OfferID) != nil {
		t.Fatalf("consumed mystery offer still visible: %+v", ev.Offers)
	}
}

func TestMysterySellerPaidRerollReplacesStockAndSpendsGold(t *testing.T) {
	sim := newMysterySellerSim(t, 2000, 3)
	seller := townMysterySellerEntity(t, sim)
	before := mysteryOffers(t, sim)
	beforeGold := sim.gold
	cost := sim.rules.Shops["town_mystery_seller"].MysteryOffers.RerollCost

	reroll := sim.Tick([]Input{{
		Type:       "shop_reroll_intent",
		MessageID:  "msg_reroll_mystery",
		ShopReroll: &ShopRerollIntent{ShopEntityID: idStr(seller.id)},
	}})
	if !hasAck(reroll, "msg_reroll_mystery") {
		t.Fatalf("reroll mystery was not acked: %+v", reroll)
	}
	if sim.gold != beforeGold-cost {
		t.Fatalf("gold after mystery reroll = %d, want %d", sim.gold, beforeGold-cost)
	}
	if !hasShopStockReplace(reroll, "town_mystery_seller") {
		t.Fatalf("mystery reroll did not replace stock: %+v", reroll.Changes)
	}
	ev := findEvent(reroll.Events, "shop_reroll")
	if ev == nil || ev.ShopID != "town_mystery_seller" || ev.Price == nil || *ev.Price != cost || ev.TotalGold == nil || *ev.TotalGold != sim.gold || ev.RefreshKey == "" {
		t.Fatalf("shop_reroll event = %+v", ev)
	}
	after := ev.Offers
	if countShopOffersByKind(after, shopOfferKindMystery) != len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots) {
		t.Fatalf("rerolled mystery offer count mismatch: %+v", after)
	}
	if len(before) == 0 || len(after) == 0 || before[0].OfferID == after[0].OfferID {
		t.Fatalf("reroll did not produce a new refresh-keyed stock: before=%+v after=%+v", before, after)
	}
}

func TestMysterySellerPaidRerollRejectsWithoutMutation(t *testing.T) {
	t.Run("insufficient gold", func(t *testing.T) {
		sim := newMysterySellerSim(t, 0, 3)
		seller := townMysterySellerEntity(t, sim)
		before := mysteryOffers(t, sim)
		res := sim.Tick([]Input{{
			Type:       "shop_reroll_intent",
			MessageID:  "msg_reroll_no_gold",
			ShopReroll: &ShopRerollIntent{ShopEntityID: idStr(seller.id)},
		}})
		if !hasReject(res, "msg_reroll_no_gold", "insufficient_gold") {
			t.Fatalf("reroll no gold result = %+v", res)
		}
		after := mysteryOffers(t, sim)
		if sim.gold != 0 || len(after) != len(before) || after[0].OfferID != before[0].OfferID {
			t.Fatalf("insufficient gold reroll mutated state: gold=%d before=%+v after=%+v", sim.gold, before, after)
		}
	})

	t.Run("normal vendor", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 3)
		vendor := townVendorEntity(t, sim)
		beforeGold := sim.gold
		res := sim.Tick([]Input{{
			Type:       "shop_reroll_intent",
			MessageID:  "msg_reroll_vendor",
			ShopReroll: &ShopRerollIntent{ShopEntityID: idStr(vendor.id)},
		}})
		if !hasReject(res, "msg_reroll_vendor", "reroll_unavailable") {
			t.Fatalf("vendor reroll result = %+v", res)
		}
		if sim.gold != beforeGold {
			t.Fatalf("vendor reroll mutated gold=%d want %d", sim.gold, beforeGold)
		}
	})
}

func TestMysterySellerPurchaseRejectsWithoutMutation(t *testing.T) {
	t.Run("insufficient gold", func(t *testing.T) {
		sim := newMysterySellerSim(t, 0, 3)
		seller := townMysterySellerEntity(t, sim)
		offer := firstMysteryOffer(t, sim)
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_mystery_no_gold",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(seller.id), OfferID: offer.OfferID},
		}})
		if !hasReject(res, "msg_buy_mystery_no_gold", "insufficient_gold") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
		if sim.gold != 0 || len(sim.inventory) != 0 || findOffer(mysteryOffers(t, sim), offer.OfferID) == nil {
			t.Fatalf("insufficient gold mutated state: gold=%d inv=%+v offers=%+v", sim.gold, sim.inventory, mysteryOffers(t, sim))
		}
	})

	t.Run("full inventory", func(t *testing.T) {
		sim := newMysterySellerSim(t, 2000, 3)
		seller := townMysterySellerEntity(t, sim)
		offer := firstMysteryOffer(t, sim)
		for i := 0; i < sim.inventoryCapacity(); i++ {
			sim.inventory = append(sim.inventory, &invItem{instanceID: uint64(9100 + i), itemDefID: "red_potion"})
		}
		sim.savePlayer(sim.defaultPlayer())
		beforeGold := sim.gold
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_mystery_full",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(seller.id), OfferID: offer.OfferID},
		}})
		if !hasReject(res, "msg_buy_mystery_full", "inventory_full") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
		if sim.gold != beforeGold || len(sim.inventory) != sim.inventoryCapacity() || findOffer(mysteryOffers(t, sim), offer.OfferID) == nil {
			t.Fatalf("full inventory mutated state: gold=%d inv=%d offers=%+v", sim.gold, len(sim.inventory), mysteryOffers(t, sim))
		}
	})
}

func TestShopStockSourceDepthPolicyAndRarityCap(t *testing.T) {
	cases := []struct {
		name       string
		level      int
		deepest    int
		wantMin    int
		wantMax    int
		wantExact  int
		exactDepth bool
	}{
		{name: "level_24_depth_50", level: 24, deepest: 50, wantMin: 25, wantMax: 50},
		{name: "level_60_depth_50", level: 60, deepest: 50, wantMin: 1, wantMax: 50},
		{name: "level_1_depth_0", level: 1, deepest: 0, wantMin: 1, wantMax: 1, wantExact: 1, exactDepth: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim := newTownVendorSimWithLevel(t, 5000, c.deepest, c.level)
			vendor := townVendorEntity(t, sim)
			open := sim.Tick([]Input{{
				Type:      "action_intent",
				MessageID: "msg_open_" + c.name,
				Action:    &ActionIntent{TargetID: idStr(vendor.id)},
			}})
			opened := findEvent(open.Events, "shop_opened")
			if opened == nil {
				t.Fatalf("missing shop_opened: %+v", open)
			}
			generatedCount := 0
			for _, offer := range opened.Offers {
				if offer.Kind != shopOfferKindGenerated {
					continue
				}
				generatedCount++
				if offer.SourceDepth < c.wantMin || offer.SourceDepth > c.wantMax {
					t.Fatalf("%s source depth = %d, want %d..%d offer=%+v", c.name, offer.SourceDepth, c.wantMin, c.wantMax, offer)
				}
				if c.exactDepth && offer.SourceDepth != c.wantExact {
					t.Fatalf("%s source depth = %d, want %d", c.name, offer.SourceDepth, c.wantExact)
				}
				if !shopRarityAllowedByCap(offer.Rarity, "rare") {
					t.Fatalf("%s rarity %s exceeds rare", c.name, offer.Rarity)
				}
			}
			if generatedCount != sim.rules.Shops["town_vendor"].GeneratedOffers.OfferCount {
				t.Fatalf("%s generated count = %d", c.name, generatedCount)
			}
			if !hasShopStockReplace(open, "town_vendor") {
				t.Fatalf("%s missing stock replace change: %+v", c.name, open.Changes)
			}
		})
	}
}

func TestShopStockLifecycleGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		ShopID         string `json:"shop_id"`
		GeneratedStock struct {
			OfferCount        int    `json:"offer_count"`
			SourceDepthPolicy string `json:"source_depth_policy"`
			RefreshOn         string `json:"refresh_on"`
			MaxRarity         string `json:"max_rarity"`
			Cases             []struct {
				Name                   string `json:"name"`
				CharacterLevel         int    `json:"character_level"`
				DeepestDungeonDepth    int    `json:"deepest_dungeon_depth"`
				ExpectedMinSourceDepth int    `json:"expected_min_source_depth"`
				ExpectedMaxSourceDepth int    `json:"expected_max_source_depth"`
			} `json:"cases"`
		} `json:"generated_stock"`
		FiniteStock struct {
			AfterGeneratedPurchaseCount int `json:"after_generated_purchase_count"`
			FixedOfferCount             int `json:"fixed_offer_count"`
		} `json:"finite_stock"`
		Buyback struct {
			SellPrice int `json:"sell_price"`
			BuyPrice  int `json:"buy_price"`
		} `json:"buyback"`
	}
	loadGolden(t, "shop_stock_lifecycle.json", &golden)
	shop := rules.Shops[golden.ShopID]
	if shop.GeneratedOffers.SourceDepthPolicy != golden.GeneratedStock.SourceDepthPolicy ||
		shop.GeneratedOffers.RefreshOn != golden.GeneratedStock.RefreshOn ||
		shop.GeneratedOffers.MaxRarity != golden.GeneratedStock.MaxRarity {
		t.Fatalf("shop lifecycle rules = %+v, want golden %+v", shop.GeneratedOffers, golden.GeneratedStock)
	}
	if shop.GeneratedOffers.OfferCount != golden.GeneratedStock.OfferCount {
		t.Fatalf("offer count = %d, want %d", shop.GeneratedOffers.OfferCount, golden.GeneratedStock.OfferCount)
	}

	for _, c := range golden.GeneratedStock.Cases {
		t.Run(c.Name, func(t *testing.T) {
			sim := newTownVendorSimWithLevel(t, 5000, c.DeepestDungeonDepth, c.CharacterLevel)
			vendor := townVendorEntity(t, sim)
			open := sim.Tick([]Input{{
				Type:      "action_intent",
				MessageID: "msg_open_golden_" + c.Name,
				Action:    &ActionIntent{TargetID: idStr(vendor.id)},
			}})
			opened := findEvent(open.Events, "shop_opened")
			if opened == nil {
				t.Fatalf("missing shop open event: %+v", open)
			}
			got := generatedOfferSignatures(opened.Offers)
			if len(got) != golden.GeneratedStock.OfferCount {
				t.Fatalf("generated count = %d, want %d: %+v", len(got), golden.GeneratedStock.OfferCount, opened.Offers)
			}
			for i, offer := range got {
				wantOfferID := fmt.Sprintf("generated:wp:none:%03d", i)
				if offer.OfferID != wantOfferID {
					t.Fatalf("offer %d id = %s, want %s", i, offer.OfferID, wantOfferID)
				}
				if offer.SourceDepth < c.ExpectedMinSourceDepth || offer.SourceDepth > c.ExpectedMaxSourceDepth {
					t.Fatalf("source depth = %d, want %d..%d offer=%+v", offer.SourceDepth, c.ExpectedMinSourceDepth, c.ExpectedMaxSourceDepth, offer)
				}
				if !shopRarityAllowedByCap(offer.Rarity, golden.GeneratedStock.MaxRarity) {
					t.Fatalf("rarity %s exceeds %s", offer.Rarity, golden.GeneratedStock.MaxRarity)
				}
			}

			repeat := newTownVendorSimWithLevel(t, 5000, c.DeepestDungeonDepth, c.CharacterLevel)
			repeatOpen := repeat.Tick([]Input{{
				Type:      "action_intent",
				MessageID: "msg_open_golden_repeat_" + c.Name,
				Action:    &ActionIntent{TargetID: idStr(townVendorEntity(t, repeat).id)},
			}})
			repeatOpened := findEvent(repeatOpen.Events, "shop_opened")
			if repeatOpened == nil || !reflect.DeepEqual(got, generatedOfferSignatures(repeatOpened.Offers)) {
				t.Fatalf("generated stock not deterministic\n got=%+v\nrepeat=%+v", got, repeatOpened)
			}
		})
	}

	sim := newTownVendorSim(t, 1000, 3)
	vendor := townVendorEntity(t, sim)
	generated := firstGeneratedOffer(t, sim)
	buy := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_golden_generated",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: generated.OfferID},
	}})
	buyEvent := findEvent(buy.Events, "shop_purchase")
	if buyEvent == nil {
		t.Fatalf("missing buy event: %+v", buy)
	}
	if countShopOffersByKind(buyEvent.Offers, shopOfferKindGenerated) != golden.FiniteStock.AfterGeneratedPurchaseCount ||
		countShopOffersByKind(buyEvent.Offers, shopOfferKindFixed) != golden.FiniteStock.FixedOfferCount {
		t.Fatalf("post-buy offer counts generated=%d fixed=%d event=%+v",
			countShopOffersByKind(buyEvent.Offers, shopOfferKindGenerated),
			countShopOffersByKind(buyEvent.Offers, shopOfferKindFixed),
			buyEvent)
	}
	if got := shop.buybackPrice(golden.Buyback.SellPrice); got != golden.Buyback.BuyPrice {
		t.Fatalf("buyback price = %d, want %d", got, golden.Buyback.BuyPrice)
	}
}

func TestShopStockFiniteGeneratedAndBuybackLifecycle(t *testing.T) {
	sim := newTownVendorSim(t, 1000, 3)
	vendor := townVendorEntity(t, sim)
	open := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_open_stock_lifecycle",
		Action:    &ActionIntent{TargetID: idStr(vendor.id)},
	}})
	opened := findEvent(open.Events, "shop_opened")
	if opened == nil {
		t.Fatalf("missing open event: %+v", open)
	}
	generated := firstGeneratedOfferFrom(opened.Offers)
	if generated == nil {
		t.Fatalf("missing generated offer: %+v", opened.Offers)
	}

	buyGenerated := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_stock_generated",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: generated.OfferID},
	}})
	if !hasAck(buyGenerated, "msg_buy_stock_generated") || !hasShopStockAvailability(buyGenerated, generated.OfferID, false) {
		t.Fatalf("generated buy did not ack/consume stock: %+v", buyGenerated)
	}
	buyEvent := findEvent(buyGenerated.Events, "shop_purchase")
	if buyEvent == nil || findOffer(buyEvent.Offers, generated.OfferID) != nil || findOffer(buyEvent.Offers, "fixed:red_potion") == nil {
		t.Fatalf("purchase refreshed offers = %+v", buyEvent)
	}
	if offers, _ := sim.shopCatalog("town_vendor"); findOffer(offers, generated.OfferID) != nil {
		t.Fatalf("generated offer still available after buy: %+v", offers)
	}
	bought := sim.inventory[len(sim.inventory)-1]
	sellPrice, ok := sim.inventorySellPrice("town_vendor", bought)
	if !ok {
		t.Fatalf("sell price missing for bought item %+v", bought)
	}

	sell := sim.Tick([]Input{{
		Type:      "shop_sell_intent",
		MessageID: "msg_sell_to_buyback",
		ShopSell:  &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: idStr(bought.instanceID)},
	}})
	if !hasAck(sell, "msg_sell_to_buyback") {
		t.Fatalf("sell did not ack: %+v", sell)
	}
	buybackID := "buyback:" + idStr(bought.instanceID)
	saleEvent := findEvent(sell.Events, "shop_sale")
	if saleEvent == nil {
		t.Fatalf("missing sale event: %+v", sell)
	}
	buyback := findOffer(saleEvent.Offers, buybackID)
	if buyback == nil || buyback.BuyPrice != sim.rules.Shops["town_vendor"].buybackPrice(sellPrice) {
		t.Fatalf("buyback offer mismatch: event=%+v buyback=%+v sell_price=%d", saleEvent, buyback, sellPrice)
	}

	buybackPurchase := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buyback_purchase",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: buybackID},
	}})
	if !hasAck(buybackPurchase, "msg_buyback_purchase") || sim.findItemByID(bought.instanceID) == nil {
		t.Fatalf("buyback purchase failed: inv=%+v res=%+v", sim.inventory, buybackPurchase)
	}
	buybackEvent := findEvent(buybackPurchase.Events, "shop_purchase")
	if buybackEvent == nil || buybackEvent.ItemInstanceID != idStr(bought.instanceID) || findOffer(buybackEvent.Offers, buybackID) != nil {
		t.Fatalf("buyback purchase event mismatch: %+v", buybackEvent)
	}
}

func TestShopStockAndBuybackArePerCharacterInCoop(t *testing.T) {
	sim := newTownVendorSim(t, 1000, 3)
	hostID := sim.DefaultPlayerID()
	guestProgress := sim.rules.DefaultCharacterProgressionState()
	guestProgress.Gold = 1000
	guestProgress.DeepestDungeonDepth = 3
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", guestProgress)
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	vendor := townVendorEntity(t, sim)
	nearVendor := Vec2{X: vendor.pos.X, Y: vendor.pos.Y - 1}
	sim.levels[townLevel].entities[hostID].pos = nearVendor
	sim.levels[townLevel].entities[guestID].pos = nearVendor

	hostOpen := tickOne(t, sim, Input{
		ActorPlayerID: hostID,
		Type:          "action_intent",
		MessageID:     "msg_host_open_shop",
		Action:        &ActionIntent{TargetID: idStr(vendor.id)},
	})
	hostOpened := findEvent(hostOpen.Events, "shop_opened")
	if hostOpened == nil || countShopOffersByKind(hostOpened.Offers, shopOfferKindGenerated) != 5 {
		t.Fatalf("host open = %+v", hostOpen)
	}
	guestOpen := tickOne(t, sim, Input{
		ActorPlayerID: guestID,
		Type:          "action_intent",
		MessageID:     "msg_guest_open_shop",
		Action:        &ActionIntent{TargetID: idStr(vendor.id)},
	})
	guestOpened := findEvent(guestOpen.Events, "shop_opened")
	if guestOpened == nil || countShopOffersByKind(guestOpened.Offers, shopOfferKindGenerated) != 5 {
		t.Fatalf("guest open = %+v", guestOpen)
	}
	if sim.players[hostID].ShopStock["town_vendor"] == sim.players[guestID].ShopStock["town_vendor"] {
		t.Fatalf("host and guest share shop stock pointer")
	}

	hostGenerated := firstGeneratedOfferFrom(hostOpened.Offers)
	if hostGenerated == nil {
		t.Fatalf("host missing generated offer: %+v", hostOpened.Offers)
	}
	hostBuy := tickOne(t, sim, Input{
		ActorPlayerID: hostID,
		Type:          "shop_buy_intent",
		MessageID:     "msg_host_buy_generated",
		ShopBuy:       &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: hostGenerated.OfferID},
	})
	hostPurchase := findEvent(hostBuy.Events, "shop_purchase")
	if hostPurchase == nil || countShopOffersByKind(hostPurchase.Offers, shopOfferKindGenerated) != 4 {
		t.Fatalf("host purchase = %+v", hostBuy)
	}
	hostSell := tickOne(t, sim, Input{
		ActorPlayerID: hostID,
		Type:          "shop_sell_intent",
		MessageID:     "msg_host_sell_buyback",
		ShopSell:      &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: hostPurchase.ItemInstanceID},
	})
	hostSale := findEvent(hostSell.Events, "shop_sale")
	if hostSale == nil || countShopOffersByKind(hostSale.Offers, shopOfferKindBuyback) != 1 {
		t.Fatalf("host sale = %+v", hostSell)
	}

	guestOpenAfterHostMutation := tickOne(t, sim, Input{
		ActorPlayerID: guestID,
		Type:          "action_intent",
		MessageID:     "msg_guest_reopen_shop",
		Action:        &ActionIntent{TargetID: idStr(vendor.id)},
	})
	guestReopened := findEvent(guestOpenAfterHostMutation.Events, "shop_opened")
	if guestReopened == nil ||
		countShopOffersByKind(guestReopened.Offers, shopOfferKindGenerated) != 5 ||
		countShopOffersByKind(guestReopened.Offers, shopOfferKindBuyback) != 0 {
		t.Fatalf("guest stock leaked host mutation: %+v", guestReopened)
	}
}

func TestMysterySellerStockIsPerCharacterInCoop(t *testing.T) {
	sim := newMysterySellerSim(t, 2000, 3)
	hostID := sim.DefaultPlayerID()
	guestProgress := sim.rules.DefaultCharacterProgressionState()
	guestProgress.Gold = 2000
	guestProgress.DeepestDungeonDepth = 3
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", guestProgress)
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	seller := townMysterySellerEntity(t, sim)
	nearSeller := Vec2{X: seller.pos.X, Y: seller.pos.Y - 1}
	sim.levels[townLevel].entities[hostID].pos = nearSeller
	sim.levels[townLevel].entities[guestID].pos = nearSeller

	hostOpen := tickOne(t, sim, Input{
		ActorPlayerID: hostID,
		Type:          "action_intent",
		MessageID:     "msg_host_open_mystery",
		Action:        &ActionIntent{TargetID: idStr(seller.id)},
	})
	hostOpened := findEvent(hostOpen.Events, "shop_opened")
	if hostOpened == nil || countShopOffersByKind(hostOpened.Offers, shopOfferKindMystery) != len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots) {
		t.Fatalf("host mystery open = %+v", hostOpen)
	}
	guestOpen := tickOne(t, sim, Input{
		ActorPlayerID: guestID,
		Type:          "action_intent",
		MessageID:     "msg_guest_open_mystery",
		Action:        &ActionIntent{TargetID: idStr(seller.id)},
	})
	guestOpened := findEvent(guestOpen.Events, "shop_opened")
	if guestOpened == nil || countShopOffersByKind(guestOpened.Offers, shopOfferKindMystery) != len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots) {
		t.Fatalf("guest mystery open = %+v", guestOpen)
	}
	if sim.players[hostID].ShopStock["town_mystery_seller"] == sim.players[guestID].ShopStock["town_mystery_seller"] {
		t.Fatalf("host and guest share mystery stock pointer")
	}

	hostMystery := firstMysteryOfferFrom(hostOpened.Offers)
	if hostMystery == nil {
		t.Fatalf("host missing mystery offer: %+v", hostOpened.Offers)
	}
	hostBuy := tickOne(t, sim, Input{
		ActorPlayerID: hostID,
		Type:          "shop_buy_intent",
		MessageID:     "msg_host_buy_mystery",
		ShopBuy:       &ShopBuyIntent{ShopEntityID: idStr(seller.id), OfferID: hostMystery.OfferID},
	})
	hostPurchase := findEvent(hostBuy.Events, "shop_purchase")
	if hostPurchase == nil || countShopOffersByKind(hostPurchase.Offers, shopOfferKindMystery) != len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots)-1 {
		t.Fatalf("host mystery purchase = %+v", hostBuy)
	}

	guestOpenAfterHostMutation := tickOne(t, sim, Input{
		ActorPlayerID: guestID,
		Type:          "action_intent",
		MessageID:     "msg_guest_reopen_mystery",
		Action:        &ActionIntent{TargetID: idStr(seller.id)},
	})
	guestReopened := findEvent(guestOpenAfterHostMutation.Events, "shop_opened")
	if guestReopened == nil || countShopOffersByKind(guestReopened.Offers, shopOfferKindMystery) != len(sim.rules.Shops["town_mystery_seller"].MysteryOffers.EligibleSlots) {
		t.Fatalf("guest mystery stock leaked host mutation: %+v", guestReopened)
	}
}

func TestShopBuybackClearsOnLeavingTownAndGeneratedStockRefreshesOnNewWaypoint(t *testing.T) {
	sim := newTownVendorSim(t, 1000, 3)
	vendor := townVendorEntity(t, sim)
	generated := firstGeneratedOffer(t, sim)
	buy := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_before_clear",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: generated.OfferID},
	}})
	if !hasAck(buy, "msg_buy_before_clear") {
		t.Fatalf("buy before clear failed: %+v", buy)
	}
	item := sim.inventory[len(sim.inventory)-1]
	sell := sim.Tick([]Input{{
		Type:      "shop_sell_intent",
		MessageID: "msg_sell_before_clear",
		ShopSell:  &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: idStr(item.instanceID)},
	}})
	buybackID := "buyback:" + idStr(item.instanceID)
	saleEvent := findEvent(sell.Events, "shop_sale")
	if !hasAck(sell, "msg_sell_before_clear") || saleEvent == nil || findOffer(saleEvent.Offers, buybackID) == nil {
		t.Fatalf("sell before clear failed: %+v", sell)
	}

	initialKey := sim.shopStock["town_vendor"].RefreshKey
	stairs := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if stairs == nil {
		t.Fatal("missing town stairs down")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	descend1 := sim.TickResults([]Input{{Type: "descend_intent", MessageID: "msg_leave_town", Descend: &DescendIntent{}}})
	if len(descend1) == 0 || !hasAck(descend1[0], "msg_leave_town") {
		t.Fatalf("leave town failed: %+v", descend1)
	}
	if state := sim.shopStock["town_vendor"]; state != nil && len(state.Buyback) != 0 {
		t.Fatalf("buyback survived leaving town: %+v", state.Buyback)
	}

	for depth := 2; depth <= 3; depth++ {
		stairs = sim.findStair(sim.activeLevel(), stairsDownDefID)
		if stairs == nil {
			t.Fatalf("missing stairs down before depth %d", depth)
		}
		moveDefaultPlayerTo(sim, stairs.pos)
		res := sim.TickResults([]Input{{Type: "descend_intent", MessageID: "msg_descend_for_wp", Descend: &DescendIntent{}}})
		if len(res) == 0 {
			t.Fatalf("descend to depth %d produced no results", depth)
		}
	}
	teleporter := sim.findTeleporter(sim.activeLevel())
	if teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	moveDefaultPlayerTo(sim, teleporter.pos)
	discover := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_discover_shop_refresh",
		Action:    &ActionIntent{TargetID: idStr(teleporter.id)},
	}})
	if !hasAck(discover, "msg_discover_shop_refresh") || !hasShopStockReplace(discover, "town_vendor") {
		t.Fatalf("discover did not refresh stock: %+v", discover)
	}
	if got := sim.shopStock["town_vendor"].RefreshKey; got == initialKey {
		t.Fatalf("refresh key did not change: %s", got)
	}
	again := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_discover_shop_refresh_again",
		Action:    &ActionIntent{TargetID: idStr(teleporter.id)},
	}})
	if hasShopStockReplace(again, "town_vendor") {
		t.Fatalf("duplicate discovery refreshed stock: %+v", again)
	}
}

func TestShopOpenIncludesAppraisalsAndComparisons(t *testing.T) {
	sim := newTownVendorSim(t, 500, 2)
	vendor := townVendorEntity(t, sim)
	equipped := &invItem{
		instanceID: 7001,
		itemDefID:  "cave_blade",
		slot:       mainHandSlot,
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "cave_blade",
			DisplayName:    "Common Cave Blade",
			Rarity:         "common",
			Stats:          map[string]int{"damage_min": 2, "damage_max": 4},
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	sellable := &invItem{
		instanceID: 7002,
		itemDefID:  "cave_shield",
		slot:       offHandSlot,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "cave_shield",
			DisplayName:    "Magic Cave Shield",
			Rarity:         "magic",
			Stats:          map[string]int{"armor": 5, "block_percent": 9},
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	sim.inventory = append(sim.inventory, equipped, sellable)
	sim.equipped[mainHandSlot] = equipped.instanceID
	sim.savePlayer(sim.defaultPlayer())

	open := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_open_shop_appraisal",
		Action:    &ActionIntent{TargetID: idStr(vendor.id)},
	}})
	opened := findEvent(open.Events, "shop_opened")
	if opened == nil {
		t.Fatalf("missing shop_opened event: %+v", open)
	}
	if len(opened.SellAppraisals) != 1 {
		t.Fatalf("sell appraisals = %+v, want only unequipped sellable item", opened.SellAppraisals)
	}
	appraisal := opened.SellAppraisals[0]
	if appraisal.ItemInstanceID != idStr(sellable.instanceID) || appraisal.DisplayName != "Magic Cave Shield" || appraisal.SellPrice != 38 {
		t.Fatalf("sell appraisal = %+v", appraisal)
	}
	if !containsShopString(appraisal.SummaryLines, "Armor +5") || !containsShopString(appraisal.SummaryLines, "Block +9%") {
		t.Fatalf("sell appraisal summary lines = %+v", appraisal.SummaryLines)
	}
	redPotion := findOffer(opened.Offers, "fixed:red_potion")
	if redPotion == nil || redPotion.Category != "consumable" || !containsShopString(redPotion.SummaryLines, "Restores 5 HP") {
		t.Fatalf("red potion offer = %+v", redPotion)
	}
	blade := findGeneratedOfferByTemplate(opened.Offers, "cave_blade")
	if blade == nil {
		t.Fatalf("missing generated cave_blade offer: %+v", opened.Offers)
	}
	if blade.Slot != mainHandSlot || blade.Category != "equipment" || blade.Comparison == nil {
		t.Fatalf("generated blade appraisal = %+v", blade)
	}
	if blade.Comparison.EquippedItemInstanceID != idStr(equipped.instanceID) {
		t.Fatalf("comparison equipped id = %q, want %s", blade.Comparison.EquippedItemInstanceID, idStr(equipped.instanceID))
	}
	if len(blade.Comparison.Deltas) == 0 {
		t.Fatalf("comparison deltas empty for %+v", blade)
	}
	for _, delta := range blade.Comparison.Deltas {
		if delta.Stat == "damage_max" && delta.Equipped != 4 {
			t.Fatalf("damage_max delta = %+v, want equipped 4", delta)
		}
	}
}

func TestShopRequirementPreviewAndPurchase(t *testing.T) {
	var golden struct {
		TemplateID     string                     `json:"template_id"`
		FreshCharacter equipmentRequirementGolden `json:"fresh_character"`
		ExpectedReject string                     `json:"expected_reject"`
	}
	loadGolden(t, "equipment_requirements.json", &golden)

	rules := cloneRules(loadRules(t))
	rules.TreasureClasses["test_requirements_shop_tc"] = TreasureClassDef{Attempts: []TreasureAttemptDef{{
		AttemptID:     "requirements_offer",
		SuccessWeight: 1,
		NoDropWeight:  0,
		Entries: []TreasureClassEntry{{
			ItemTemplateID: golden.TemplateID,
			Weight:         1,
		}},
	}}}
	rules.LootTables["test_requirements_shop_drop"] = LootTable{TreasureClassID: "test_requirements_shop_tc"}
	rules.DungeonGeneration.LootBands = []DungeonLootBand{{
		MinDepth:         1,
		MonsterLootTable: "test_requirements_shop_drop",
		ChestLootTable:   "test_requirements_shop_drop",
	}}
	sim, err := NewSimWithWorldProgression("sess_shop_requirements", "v43_requirements_shop", rules, "dungeon_levels", CharacterProgressionState{
		Level:               1,
		Gold:                1000,
		DeepestDungeonDepth: 1,
		BaseStats:           rules.CharacterProgression.BaseStats,
	})
	if err != nil {
		t.Fatalf("new requirements shop sim: %v", err)
	}
	sim.SetPlayerMetadata(sim.DefaultPlayerID(), "acct_shop", "char_01H00000000000000000000043", "Hero", "host")
	vendor := townVendorEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: vendor.pos.X, Y: vendor.pos.Y - 1})
	sim.savePlayer(sim.defaultPlayer())

	open := sim.Tick([]Input{{
		Type:      "action_intent",
		MessageID: "msg_open_requirements_shop",
		Action:    &ActionIntent{TargetID: idStr(vendor.id)},
	}})
	opened := findEvent(open.Events, "shop_opened")
	if opened == nil {
		t.Fatalf("missing shop_opened event: %+v", open)
	}
	offer := findGeneratedOfferByTemplate(opened.Offers, golden.TemplateID)
	if offer == nil {
		t.Fatalf("missing generated %s offer: %+v", golden.TemplateID, opened.Offers)
	}
	assertRequirementStatus(t, offer.RequirementStatus, golden.FreshCharacter.Status)
	if offer.RequirementsMet == nil || *offer.RequirementsMet || offer.EquipPreview == nil || offer.EquipPreview.RequirementsMet {
		t.Fatalf("shop offer requirement preview = %+v", offer)
	}
	if !containsShopString(offer.SummaryLines, "Requires level 2") || !containsShopString(offer.SummaryLines, "Requires STR 6") {
		t.Fatalf("shop offer summary lines = %+v", offer.SummaryLines)
	}
	if findPreviewDelta(offer.EquipPreview.Deltas, "damage_max") == nil {
		t.Fatalf("shop offer preview missing damage_max delta: %+v", offer.EquipPreview.Deltas)
	}

	beforeGold := sim.gold
	buy := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_requirements",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: offer.OfferID},
	}})
	if !hasAck(buy, "msg_buy_requirements") {
		t.Fatalf("buy requirements offer was not acked: %+v", buy)
	}
	bought := sim.inventory[len(sim.inventory)-1]
	if bought.itemDefID != golden.TemplateID || sim.gold != beforeGold-offer.BuyPrice {
		t.Fatalf("bought item/gold = %+v/%d, want %s/%d", bought, sim.gold, golden.TemplateID, beforeGold-offer.BuyPrice)
	}
	inventoryView := sim.itemView(bought)
	assertRequirementStatus(t, inventoryView.RequirementStatus, golden.FreshCharacter.Status)
	if inventoryView.RequirementsMet == nil || *inventoryView.RequirementsMet || inventoryView.EquipPreview == nil {
		t.Fatalf("bought inventory requirement view = %+v", inventoryView)
	}
	equip := sim.Tick([]Input{{
		Type:      "equip_intent",
		MessageID: "msg_equip_unmet_shop_item",
		Equip:     &EquipIntent{ItemInstanceID: idStr(bought.instanceID), Slot: mainHandSlot},
	}})
	if !hasReject(equip, "msg_equip_unmet_shop_item", golden.ExpectedReject) || bought.equipped || sim.equipped[mainHandSlot] != 0 {
		t.Fatalf("unmet shop item equip result=%+v item=%+v equipped=%v", equip, bought, sim.equipped)
	}
	appraisals := sim.shopSellAppraisals("town_vendor")
	if len(appraisals) != 1 || appraisals[0].ItemTemplateID != golden.TemplateID {
		t.Fatalf("requirements sell appraisals = %+v", appraisals)
	}
	assertRequirementStatus(t, appraisals[0].RequirementStatus, golden.FreshCharacter.Status)
	if appraisals[0].RequirementsMet == nil || *appraisals[0].RequirementsMet || appraisals[0].EquipPreview == nil {
		t.Fatalf("requirements appraisal preview = %+v", appraisals[0])
	}
}

func TestShopBuyFailureDoesNotMutate(t *testing.T) {
	t.Run("insufficient gold", func(t *testing.T) {
		sim := newTownVendorSim(t, 0, 1)
		vendor := townVendorEntity(t, sim)
		beforeInventory := len(sim.inventory)
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_no_gold",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: "fixed:red_potion"},
		}})
		if !hasReject(res, "msg_buy_no_gold", "insufficient_gold") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
		if sim.gold != 0 || len(sim.inventory) != beforeInventory {
			t.Fatalf("insufficient gold mutated gold=%d inv=%+v", sim.gold, sim.inventory)
		}
	})

	t.Run("full inventory", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 1)
		vendor := townVendorEntity(t, sim)
		for i := 0; i < sim.inventoryCapacity(); i++ {
			sim.inventory = append(sim.inventory, &invItem{instanceID: uint64(9000 + i), itemDefID: "red_potion"})
		}
		sim.savePlayer(sim.defaultPlayer())
		beforeGold := sim.gold
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_full",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: "fixed:blue_potion"},
		}})
		if !hasReject(res, "msg_buy_full", "inventory_full") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
		if sim.gold != beforeGold || len(sim.inventory) != sim.inventoryCapacity() {
			t.Fatalf("full inventory mutated gold=%d inv=%d", sim.gold, len(sim.inventory))
		}
	})

	t.Run("out of range", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 1)
		vendor := townVendorEntity(t, sim)
		moveDefaultPlayerTo(sim, Vec2{X: 1, Y: 1})
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_far",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: "fixed:red_potion"},
		}})
		if !hasReject(res, "msg_buy_far", "out_of_range") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
	})

	t.Run("unknown generated offer", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 1)
		vendor := townVendorEntity(t, sim)
		beforeGold := sim.gold
		beforeInventory := len(sim.inventory)
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_unknown_generated",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: "generated:wp:none:999"},
		}})
		if !hasReject(res, "msg_buy_unknown_generated", "unknown_offer") {
			t.Fatalf("buy reject = %+v", res.Rejects)
		}
		if sim.gold != beforeGold || len(sim.inventory) != beforeInventory {
			t.Fatalf("unknown generated mutated gold=%d inv=%+v", sim.gold, sim.inventory)
		}
	})

	t.Run("consumed generated offer", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 3)
		vendor := townVendorEntity(t, sim)
		offer := firstGeneratedOffer(t, sim)
		buy := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_once",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: offer.OfferID},
		}})
		if !hasAck(buy, "msg_buy_once") {
			t.Fatalf("initial generated buy failed: %+v", buy)
		}
		beforeGold := sim.gold
		beforeInventory := len(sim.inventory)
		again := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_consumed",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: offer.OfferID},
		}})
		if !hasReject(again, "msg_buy_consumed", "unknown_offer") {
			t.Fatalf("consumed buy reject = %+v", again.Rejects)
		}
		if sim.gold != beforeGold || len(sim.inventory) != beforeInventory {
			t.Fatalf("consumed buy mutated gold=%d inv=%+v", sim.gold, sim.inventory)
		}
	})

	t.Run("invalid shop target", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 1)
		res := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_invalid_shop",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: "999999", OfferID: "fixed:red_potion"},
		}})
		if !hasReject(res, "msg_buy_invalid_shop", "invalid_target") {
			t.Fatalf("invalid shop reject = %+v", res.Rejects)
		}
	})

	t.Run("missing buyback after purchase", func(t *testing.T) {
		sim := newTownVendorSim(t, 1000, 3)
		vendor := townVendorEntity(t, sim)
		offer := firstGeneratedOffer(t, sim)
		buy := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buy_for_buyback_failure",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: offer.OfferID},
		}})
		item := sim.inventory[len(sim.inventory)-1]
		sell := sim.Tick([]Input{{
			Type:      "shop_sell_intent",
			MessageID: "msg_sell_for_buyback_failure",
			ShopSell:  &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: idStr(item.instanceID)},
		}})
		if !hasAck(buy, "msg_buy_for_buyback_failure") || !hasAck(sell, "msg_sell_for_buyback_failure") {
			t.Fatalf("setup buy/sell failed buy=%+v sell=%+v", buy, sell)
		}
		buybackID := "buyback:" + idStr(item.instanceID)
		first := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buyback_once",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: buybackID},
		}})
		if !hasAck(first, "msg_buyback_once") {
			t.Fatalf("buyback setup buy failed: %+v", first)
		}
		beforeGold := sim.gold
		beforeInventory := len(sim.inventory)
		again := sim.Tick([]Input{{
			Type:      "shop_buy_intent",
			MessageID: "msg_buyback_missing",
			ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(vendor.id), OfferID: buybackID},
		}})
		if !hasReject(again, "msg_buyback_missing", "unknown_offer") {
			t.Fatalf("missing buyback reject = %+v", again.Rejects)
		}
		if sim.gold != beforeGold || len(sim.inventory) != beforeInventory {
			t.Fatalf("missing buyback mutated gold=%d inv=%+v", sim.gold, sim.inventory)
		}
	})
}

func TestShopSellEquippedItemRejected(t *testing.T) {
	sim := newTownVendorSim(t, 100, 3)
	vendor := townVendorEntity(t, sim)
	offer := firstGeneratedOffer(t, sim)
	item := sim.itemFromShopOffer(offer, sim.alloc())
	item.equipped = true
	sim.inventory = append(sim.inventory, item)
	sim.equipped[item.slot] = item.instanceID
	sim.savePlayer(sim.defaultPlayer())

	res := sim.Tick([]Input{{
		Type:      "shop_sell_intent",
		MessageID: "msg_sell_equipped",
		ShopSell:  &ShopSellIntent{ShopEntityID: idStr(vendor.id), ItemInstanceID: idStr(item.instanceID)},
	}})
	if !hasReject(res, "msg_sell_equipped", "item_equipped") {
		t.Fatalf("sell reject = %+v", res.Rejects)
	}
	if sim.findItemByID(item.instanceID) == nil || sim.gold != 100 {
		t.Fatalf("equipped sell mutated item/gold: inv=%+v gold=%d", sim.inventory, sim.gold)
	}
}

func TestDeepestDungeonDepthAdvancesOnDungeonTravel(t *testing.T) {
	sim := newTownVendorSim(t, 0, 0)
	stairs := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if stairs == nil {
		t.Fatal("missing town stairs down")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	first := sim.TickResults([]Input{{
		Type:      "descend_intent",
		MessageID: "msg_descend_1",
		Descend:   &DescendIntent{},
	}})
	if sim.progression.DeepestDungeonDepth != 1 || !resultsHaveProgressionDepth(first, 1) {
		t.Fatalf("depth after first descend = %d results=%+v", sim.progression.DeepestDungeonDepth, first)
	}

	stairs = sim.findStair(sim.activeLevel(), stairsDownDefID)
	if stairs == nil {
		t.Fatal("missing level -1 stairs down")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	second := sim.TickResults([]Input{{
		Type:      "descend_intent",
		MessageID: "msg_descend_2",
		Descend:   &DescendIntent{},
	}})
	if sim.progression.DeepestDungeonDepth != 2 || !resultsHaveProgressionDepth(second, 2) {
		t.Fatalf("depth after second descend = %d results=%+v", sim.progression.DeepestDungeonDepth, second)
	}

	stairs = sim.findStair(sim.activeLevel(), stairsUpDefID)
	if stairs == nil {
		t.Fatal("missing level -2 stairs up")
	}
	moveDefaultPlayerTo(sim, stairs.pos)
	ascend := sim.TickResults([]Input{{
		Type:      "ascend_intent",
		MessageID: "msg_ascend",
		Ascend:    &AscendIntent{},
	}})
	if sim.progression.DeepestDungeonDepth != 2 || resultsHaveProgressionDepth(ascend, 1) {
		t.Fatalf("depth after ascend = %d results=%+v", sim.progression.DeepestDungeonDepth, ascend)
	}
}

type shopOfferGolden struct {
	OfferID        string         `json:"offer_id"`
	Kind           string         `json:"kind"`
	ItemTemplateID string         `json:"item_template_id"`
	DisplayName    string         `json:"display_name"`
	Rarity         string         `json:"rarity"`
	RolledStats    map[string]int `json:"rolled_stats"`
	BuyPrice       int            `json:"buy_price"`
	Source         string         `json:"source"`
	Depth          int            `json:"depth"`
}

func shopOfferGoldenFromView(offer ShopOfferView) shopOfferGolden {
	return shopOfferGolden{
		OfferID:        offer.OfferID,
		Kind:           offer.Kind,
		ItemTemplateID: offer.ItemTemplateID,
		DisplayName:    offer.DisplayName,
		Rarity:         offer.Rarity,
		RolledStats:    offer.RolledStats,
		BuyPrice:       offer.BuyPrice,
		Source:         offer.Source,
		Depth:          offer.Depth,
	}
}

func newTownVendorSim(t *testing.T, gold int, deepestDepth int) *Sim {
	t.Helper()
	return newTownVendorSimWithLevel(t, gold, deepestDepth, 1)
}

func newTownVendorSimWithLevel(t *testing.T, gold int, deepestDepth int, level int) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_shop", "v41_shop_offers", rules, "dungeon_levels", CharacterProgressionState{
		Level:               level,
		Gold:                gold,
		DeepestDungeonDepth: deepestDepth,
		BaseStats:           rules.CharacterProgression.BaseStats,
	})
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	sim.SetPlayerMetadata(sim.DefaultPlayerID(), "acct_shop", "char_01H00000000000000000000000", "Hero", "host")
	vendor := townVendorEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: vendor.pos.X, Y: vendor.pos.Y - 1})
	sim.progression.Level = level
	sim.savePlayer(sim.defaultPlayer())
	return sim
}

func newMysterySellerSim(t *testing.T, gold int, deepestDepth int) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_mystery_shop", "v51_mystery_seller", rules, "dungeon_levels", CharacterProgressionState{
		Level:               1,
		Gold:                gold,
		DeepestDungeonDepth: deepestDepth,
		BaseStats:           rules.CharacterProgression.BaseStats,
	})
	if err != nil {
		t.Fatalf("new mystery seller sim: %v", err)
	}
	sim.SetPlayerMetadata(sim.DefaultPlayerID(), "acct_shop", "char_01H00000000000000000000000", "Hero", "host")
	seller := townMysterySellerEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: seller.pos.X, Y: seller.pos.Y - 1})
	sim.savePlayer(sim.defaultPlayer())
	return sim
}

func townVendorEntity(t *testing.T, sim *Sim) *entity {
	t.Helper()
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		e := sim.activeLevel().entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == "town_vendor" {
			return e
		}
	}
	t.Fatal("missing town vendor")
	return nil
}

func townMysterySellerEntity(t *testing.T, sim *Sim) *entity {
	t.Helper()
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		e := sim.activeLevel().entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == "town_mystery_seller" {
			return e
		}
	}
	t.Fatal("missing town mystery seller")
	return nil
}

func moveDefaultPlayerTo(sim *Sim, pos Vec2) {
	player := sim.activeLevel().entities[sim.playerID]
	if player != nil {
		player.pos = pos
	}
}

func firstGeneratedOffer(t *testing.T, sim *Sim) ShopOfferView {
	t.Helper()
	offers, ok := sim.shopCatalog("town_vendor")
	if !ok {
		t.Fatal("shop catalog failed")
	}
	for _, offer := range offers {
		if offer.Kind == shopOfferKindGenerated {
			return offer
		}
	}
	t.Fatal("missing generated offer")
	return ShopOfferView{}
}

func firstMysteryOffer(t *testing.T, sim *Sim) ShopOfferView {
	t.Helper()
	for _, offer := range mysteryOffers(t, sim) {
		if offer.Kind == shopOfferKindMystery {
			return offer
		}
	}
	t.Fatal("missing mystery offer")
	return ShopOfferView{}
}

func firstMysteryOfferFrom(offers []ShopOfferView) *ShopOfferView {
	for i := range offers {
		if offers[i].Kind == shopOfferKindMystery {
			return &offers[i]
		}
	}
	return nil
}

func mysteryOffers(t *testing.T, sim *Sim) []ShopOfferView {
	t.Helper()
	offers, ok := sim.shopCatalog("town_mystery_seller")
	if !ok {
		t.Fatal("mystery shop catalog failed")
	}
	return offers
}

func firstGeneratedOfferFrom(offers []ShopOfferView) *ShopOfferView {
	for i := range offers {
		if offers[i].Kind == shopOfferKindGenerated {
			return &offers[i]
		}
	}
	return nil
}

func findOffer(offers []ShopOfferView, offerID string) *ShopOfferView {
	for i := range offers {
		if offers[i].OfferID == offerID {
			return &offers[i]
		}
	}
	return nil
}

func generatedOfferSignatures(offers []ShopOfferView) []ShopOfferView {
	out := make([]ShopOfferView, 0, len(offers))
	for _, offer := range offers {
		if offer.Kind == shopOfferKindGenerated {
			out = append(out, ShopOfferView{
				OfferID:        offer.OfferID,
				Kind:           offer.Kind,
				ItemTemplateID: offer.ItemTemplateID,
				DisplayName:    offer.DisplayName,
				Rarity:         offer.Rarity,
				RolledStats:    offer.RolledStats,
				BuyPrice:       offer.BuyPrice,
				SourceDepth:    offer.SourceDepth,
			})
		}
	}
	return out
}

func countShopOffersByKind(offers []ShopOfferView, kind string) int {
	count := 0
	for _, offer := range offers {
		if offer.Kind == kind {
			count++
		}
	}
	return count
}

func tickOne(t *testing.T, sim *Sim, in Input) TickResult {
	t.Helper()
	results := sim.TickResults([]Input{in})
	if len(results) != 1 {
		t.Fatalf("tick results = %d, want 1: %+v", len(results), results)
	}
	return results[0]
}

func hasShopStockReplace(res TickResult, shopID string) bool {
	for _, change := range res.Changes {
		if change.Op == OpShopStockReplace && change.ShopID == shopID && len(change.ShopStock) > 0 {
			return true
		}
	}
	return false
}

func hasShopStockAvailability(res TickResult, offerID string, available bool) bool {
	for _, change := range res.Changes {
		if change.Op == OpShopStockAvailability && change.OfferID == offerID && change.Available == available {
			return true
		}
	}
	return false
}

func findGeneratedOfferByTemplate(offers []ShopOfferView, templateID string) *ShopOfferView {
	for i := range offers {
		if offers[i].Kind == shopOfferKindGenerated && offers[i].ItemTemplateID == templateID {
			return &offers[i]
		}
	}
	return nil
}

func containsShopString(rows []string, want string) bool {
	for _, row := range rows {
		if row == want {
			return true
		}
	}
	return false
}

func hasAck(res TickResult, messageID string) bool {
	for _, ack := range res.Acks {
		if ack.MessageID == messageID {
			return true
		}
	}
	return false
}

func hasReject(res TickResult, messageID, reason string) bool {
	for _, rej := range res.Rejects {
		if rej.MessageID == messageID && rej.Reason == reason {
			return true
		}
	}
	return false
}

func findEvent(events []Event, eventType string) *Event {
	for i := range events {
		if events[i].EventType == eventType {
			return &events[i]
		}
	}
	return nil
}

func resultsHaveProgressionDepth(results []TickResult, depth int) bool {
	for _, res := range results {
		for _, change := range res.Changes {
			if change.Op == OpCharacterProgressionUpdate && change.Progression != nil && change.Progression.DeepestDungeonDepth == depth {
				return true
			}
		}
	}
	return false
}
