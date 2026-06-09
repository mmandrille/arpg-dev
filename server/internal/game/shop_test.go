package game

import (
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
			sim := NewSim(golden.ShopID+"_"+c.Name, golden.Seed, rules)
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
	if !containsShopString(appraisal.SummaryLines, "Armor +5") || !containsShopString(appraisal.SummaryLines, "Block +9") {
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
	moveDefaultPlayerTo(sim, Vec2{X: 6, Y: 12})
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
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_shop", "v41_shop_offers", rules, "dungeon_levels", CharacterProgressionState{
		Level:               1,
		Gold:                gold,
		DeepestDungeonDepth: deepestDepth,
		BaseStats:           rules.CharacterProgression.BaseStats,
	})
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	sim.SetPlayerMetadata(sim.DefaultPlayerID(), "acct_shop", "char_01H00000000000000000000000", "Hero", "host")
	moveDefaultPlayerTo(sim, Vec2{X: 6, Y: 12})
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

func findOffer(offers []ShopOfferView, offerID string) *ShopOfferView {
	for i := range offers {
		if offers[i].OfferID == offerID {
			return &offers[i]
		}
	}
	return nil
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
