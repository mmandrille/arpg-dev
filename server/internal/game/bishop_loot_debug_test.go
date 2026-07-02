package game

import "testing"

func setBishopTestDeepestDepth(sim *Sim, depth int) {
	sim.progression.DeepestDungeonDepth = depth
	sim.savePlayer(sim.defaultPlayer())
}

func TestBishopLootCatalogCapsDepthAtDeepestReached(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_loot_catalog", "v_bishop_loot_catalog", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	setBishopTestDeepestDepth(sim, 2)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:              "loot_catalog",
		CorrelationID:          "corr_loot_catalog",
		Type:                   "bishop_debug_loot_catalog_intent",
		BishopDebugLootCatalog: &BishopDebugLootCatalogIntent{BishopEntityID: idStr(bishop.id)},
	}})

	assertAck(t, res, "loot_catalog")
	ev := findEvent(res.Events, "bishop_debug_loot_catalog")
	if ev == nil || ev.BishopLootDepthCatalog == nil {
		t.Fatalf("bishop_debug_loot_catalog event = %+v", ev)
	}
	if ev.BishopLootDepthCatalog.MaxReachableDepth != 2 {
		t.Fatalf("max reachable depth = %d, want 2", ev.BishopLootDepthCatalog.MaxReachableDepth)
	}
	for _, depth := range ev.BishopLootDepthCatalog.Depths {
		if depth.Depth > 2 {
			t.Fatalf("catalog depth %d exceeds deepest reached depth 2", depth.Depth)
		}
	}
}

func TestBishopForceLootRejectsDepthAboveDeepestReached(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_force_depth", "v_bishop_force_depth", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	setBishopTestDeepestDepth(sim, 1)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID: "force_depth",
		Type:      "bishop_debug_force_loot_intent",
		BishopDebugForceLoot: &BishopDebugForceLootIntent{
			BishopEntityID: idStr(bishop.id),
			Depth:          3,
			SourceType:     "monster",
			DropKind:       "wallet_item",
			ItemDefID:      "respec_badge",
		},
	}})

	assertReject(t, res, "force_depth", "invalid_depth")
}

func TestBishopForceLootSpawnsGoldFromMonsterTable(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_force_gold", "v_bishop_force_gold", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	setBishopTestDeepestDepth(sim, 3)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}
	startGold := sim.gold

	res := sim.Tick([]Input{{
		MessageID:     "force_gold",
		CorrelationID: "corr_force_gold",
		Type:          "bishop_debug_force_loot_intent",
		BishopDebugForceLoot: &BishopDebugForceLootIntent{
			BishopEntityID: idStr(bishop.id),
			Depth:          1,
			SourceType:     "monster",
			DropKind:       "treasure_entry",
			AttemptID:      "primary",
			EntryIndex:     0,
			ItemLevel:      1,
		},
	}})

	assertAck(t, res, "force_gold")
	if sim.gold <= startGold {
		t.Fatalf("gold = %d, want > %d after forced gold drop", sim.gold, startGold)
	}
	if findEvent(res.Events, "bishop_debug_loot_dropped") == nil {
		t.Fatalf("missing bishop_debug_loot_dropped: %+v", res.Events)
	}
}

func TestBishopForceLootSpawnsRenewStoneAtItemLevel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_force_renew", "v_bishop_force_renew", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	setBishopTestDeepestDepth(sim, 25)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:     "force_renew",
		CorrelationID: "corr_force_renew",
		Type:          "bishop_debug_force_loot_intent",
		BishopDebugForceLoot: &BishopDebugForceLootIntent{
			BishopEntityID: idStr(bishop.id),
			Depth:          21,
			SourceType:     "monster",
			DropKind:       "resource_pool",
			ItemDefID:      RenewStoneItemDefID,
			ItemLevel:      3,
		},
	}})

	assertAck(t, res, "force_renew")
	ev := findEvent(res.Events, "bishop_debug_loot_dropped")
	if ev == nil || ev.Amount == nil || *ev.Amount != 3 {
		t.Fatalf("bishop_debug_loot_dropped event = %+v", ev)
	}
}

func TestBishopForceLootSpawnsWalletBadge(t *testing.T) {
	sim, err := NewSimWithWorld("sess_bishop_force_badge", "v_bishop_force_badge", loadRules(t), "vendor_lab")
	if err != nil {
		t.Fatal(err)
	}
	sim.SetGameplayDebug(true)
	bishop := findInteractableByDefID(t, sim, "town_bishop")
	sim.activeLevel().entities[sim.playerID].pos = Vec2{X: bishop.pos.X - 0.5, Y: bishop.pos.Y}

	res := sim.Tick([]Input{{
		MessageID:     "force_badge",
		CorrelationID: "corr_force_badge",
		Type:          "bishop_debug_force_loot_intent",
		BishopDebugForceLoot: &BishopDebugForceLootIntent{
			BishopEntityID: idStr(bishop.id),
			Depth:          1,
			SourceType:     "monster",
			DropKind:       "wallet_item",
			ItemDefID:      "respec_badge",
		},
	}})

	assertAck(t, res, "force_badge")
	if findEvent(res.Events, "bishop_debug_loot_dropped") == nil {
		t.Fatalf("missing bishop_debug_loot_dropped: %+v", res.Events)
	}
}
