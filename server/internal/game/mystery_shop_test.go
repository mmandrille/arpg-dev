package game

import "testing"

func TestMysterySellerSpecialCandidatesRespectCapAndDepth(t *testing.T) {
	sim := newMysterySellerSim(t, 5000, 5)
	sim.usePlayer(sim.defaultPlayer())
	mystery := sim.rules.Shops["town_mystery_seller"].MysteryOffers

	rareOnly := mystery
	rareOnly.MaxRarity = "rare"
	if candidates := sim.mysterySpecialCandidatesForSlot("main_hand", 5, rareOnly); len(candidates) != 0 {
		t.Fatalf("rare-capped mystery candidates = %+v, want none", candidates)
	}
	if candidates := sim.mysterySpecialCandidatesForSlot("main_hand", 4, mystery); len(candidates) != 0 {
		t.Fatalf("under-level mystery candidates = %+v, want none", candidates)
	}

	candidates := sim.mysterySpecialCandidatesForSlot("main_hand", 5, mystery)
	if len(candidates) == 0 {
		t.Fatal("main_hand mystery candidates empty, want set/unique candidates")
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		payload := candidate.Payload
		seen[payload.Rarity] = true
		template := sim.rules.ItemTemplates[payload.ItemTemplateID]
		if template.Slot != "main_hand" {
			t.Fatalf("candidate slot = %s, want main_hand: %+v", template.Slot, payload)
		}
		if payload.Requirements["level"] > 5 {
			t.Fatalf("candidate level = %d, want <= 5: %+v", payload.Requirements["level"], payload)
		}
	}
	if !seen["unique"] || !seen["set"] {
		t.Fatalf("candidate rarities = %+v, want unique and set", seen)
	}
}

func TestMysterySellerSpecialPurchaseRevealsOwnedPayload(t *testing.T) {
	sim := newMysterySellerSim(t, 5000, 5)
	sim.usePlayer(sim.defaultPlayer())
	seller := townMysterySellerEntity(t, sim)
	shopID := "town_mystery_seller"
	shop := sim.rules.Shops[shopID]
	payload := firstMysterySpecialPayload(t, sim, "main_hand", 5)
	refreshKey := sim.shopRefreshKey()
	row, ok := sim.mysteryStockItemFromPayload(shop, refreshKey, "main_hand", 0, 5, payload)
	if !ok {
		t.Fatalf("mysteryStockItemFromPayload failed for %+v", payload)
	}
	sim.shopStock = map[string]*shopStockState{
		shopID: {RefreshKey: refreshKey, Generated: []*shopStockItem{row}},
	}
	sim.savePlayer(sim.defaultPlayer())
	beforeGold := sim.gold

	buy := sim.Tick([]Input{{
		Type:      "shop_buy_intent",
		MessageID: "msg_buy_special_mystery",
		ShopBuy:   &ShopBuyIntent{ShopEntityID: idStr(seller.id), OfferID: row.OfferID},
	}})

	assertAck(t, buy, "msg_buy_special_mystery")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory count = %d, want 1", len(sim.inventory))
	}
	bought := sim.inventory[0]
	if bought.rollPayload == nil || bought.rollPayload.Rarity != payload.Rarity || bought.rollPayload.DisplayName != payload.DisplayName {
		t.Fatalf("bought payload = %+v, want %+v", bought.rollPayload, payload)
	}
	ev := findEvent(buy.Events, "shop_purchase")
	if ev == nil || ev.Item == nil || ev.Item.Rarity != payload.Rarity || ev.Item.DisplayName != payload.DisplayName {
		t.Fatalf("shop_purchase reveal = %+v, want %s %s", ev, payload.Rarity, payload.DisplayName)
	}
	if ev.Price == nil || sim.gold != beforeGold-*ev.Price {
		t.Fatalf("gold after special mystery buy = %d, event = %+v", sim.gold, ev)
	}
}

func firstMysterySpecialPayload(t *testing.T, sim *Sim, slot string, sourceDepth int) ItemRollPayload {
	t.Helper()
	mystery := sim.rules.Shops["town_mystery_seller"].MysteryOffers
	candidates := sim.mysterySpecialCandidatesForSlot(slot, sourceDepth, mystery)
	if len(candidates) == 0 {
		t.Fatalf("no mystery special candidates for slot=%s sourceDepth=%d", slot, sourceDepth)
	}
	return candidates[0].Payload
}
