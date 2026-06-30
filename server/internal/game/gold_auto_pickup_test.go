package game

import "testing"

func TestGoldAutoPickupWalkingIntoTownGold(t *testing.T) {
	sim := MustNewSim("sess_gold_auto_town", "v49_gold_auto_town", loadRules(t))
	player := sim.entities[sim.playerID]
	// Gold at 1.6 units east: beyond 1.5 reach before movement, within reach
	// after one tick at min momentum speed (0.75×0.6=0.45/tick → 1.15 dist).
	gold := addTestGoldLoot(sim, Vec2{X: player.pos.X + 1.6, Y: player.pos.Y}, 7)

	results := sim.TickResults([]Input{{MessageID: "move_gold", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})

	assertAckInResults(t, results, "move_gold")
	if sim.gold != 7 || sim.players[sim.playerID].Gold != 7 || sim.players[sim.playerID].Progression.Gold != 7 {
		t.Fatalf("gold after auto-pickup = sim %d player %d progression %d, want 7", sim.gold, sim.players[sim.playerID].Gold, sim.players[sim.playerID].Progression.Gold)
	}
	if sim.entities[gold.id] != nil {
		t.Fatalf("gold entity %d still present after auto-pickup", gold.id)
	}
	if !hasEventInResults(results, "gold_picked_up") {
		t.Fatalf("missing gold_picked_up in results: %+v", results)
	}
	if owners := changeOwnersForOpInResults(results, OpGoldUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("gold_update owners = %v, want player %d", owners, sim.playerID)
	}
	if owners := changeOwnersForOpInResults(results, OpCharacterProgressionUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("progression owners = %v, want player %d", owners, sim.playerID)
	}
}

func TestGoldAutoPickupWorksOnDungeonLevelWallet(t *testing.T) {
	sim, err := NewSimWithWorld("sess_gold_auto_dungeon", "v49_gold_auto_dungeon", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	hostID := sim.playerID
	descendFromCurrentLevel(t, sim, "descend_1")
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	player := sim.entities[hostID]
	gold := addTestGoldLoot(sim, player.pos, 13)

	res := sim.Tick(nil)

	if sim.entities[gold.id] != nil {
		t.Fatalf("dungeon gold entity %d still present", gold.id)
	}
	if sim.players[hostID].Gold != 13 || sim.players[hostID].Progression.Gold != 13 {
		t.Fatalf("dungeon wallet gold=%d progression=%d, want 13", sim.players[hostID].Gold, sim.players[hostID].Progression.Gold)
	}
	if !hasEvent(res, "gold_picked_up") || !hasChange(res, OpGoldUpdate) || !hasProgressionChange(res) {
		t.Fatalf("dungeon pickup result missing wallet/progression: %+v", res)
	}
}

func TestNonGoldLootDoesNotAutoPickup(t *testing.T) {
	sim := MustNewSim("sess_item_no_auto", "v49_item_no_auto", loadRules(t))
	player := sim.entities[sim.playerID]
	loot := addTestFloorLoot(sim, "quest_leaf", player.pos)

	res := sim.Tick(nil)

	if sim.entities[loot.id] == nil {
		t.Fatalf("non-gold loot %d was auto-picked", loot.id)
	}
	if len(sim.inventory) != 0 || hasEvent(res, "item_picked_up") {
		t.Fatalf("non-gold loot auto-mutated inventory/events: inv=%+v result=%+v", sim.inventory, res)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_leaf", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertAck(t, pickup, "pick_leaf")
	if len(sim.inventory) != 1 || sim.inventory[0].itemDefID != "quest_leaf" {
		t.Fatalf("explicit item pickup inventory = %+v, want quest_leaf", sim.inventory)
	}
}

func TestResourceWalletAutoPickupWalkingIntoBadge(t *testing.T) {
	sim := MustNewSim("sess_resource_auto_town", "v229_resource_auto_town", loadRules(t))
	player := sim.entities[sim.playerID]
	badge := addTestWalletResourceLoot(sim, Vec2{X: player.pos.X + 1.6, Y: player.pos.Y})
	resourceID := "respec_badge"

	results := sim.TickResults([]Input{{MessageID: "move_badge", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})

	assertAckInResults(t, results, "move_badge")
	if sim.entities[badge.id] != nil {
		t.Fatalf("resource entity %d still present after auto-pickup", badge.id)
	}
	if got := sim.resourceWallet[resourceID]; got != 1 {
		t.Fatalf("resource wallet %s = %d, want 1", resourceID, got)
	}
	if !hasEventInResults(results, "resource_picked_up") {
		t.Fatalf("missing resource_picked_up in results: %+v", results)
	}
	if owners := changeOwnersForOpInResults(results, OpResourceWalletUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("resource_wallet_update owners = %v, want player %d", owners, sim.playerID)
	}
}

func TestResourceWalletAutoPickupCoopLowestPlayerIDWins(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_resource_coop_winner", "v229_resource_coop_winner", rules)
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	badge := addTestWalletResourceLoot(sim, Vec2{X: 6, Y: 6})
	resourceID := "respec_badge"
	sim.entities[hostID].pos = badge.pos
	sim.entities[guestID].pos = badge.pos

	results := sim.TickResults(nil)

	if sim.players[hostID].ResourceWallet[resourceID] != 1 || sim.players[guestID].ResourceWallet[resourceID] != 0 {
		t.Fatalf("coop resource wallet host=%d guest=%d, want host winner",
			sim.players[hostID].ResourceWallet[resourceID], sim.players[guestID].ResourceWallet[resourceID])
	}
	if owners := changeOwnersForOpInResults(results, OpResourceWalletUpdate); !sameUint64Set(owners, []uint64{hostID}) {
		t.Fatalf("coop resource_wallet_update owners = %v, want host %d", owners, hostID)
	}
	if !eventForEntityInResults(results, "resource_picked_up", hostID) {
		t.Fatalf("coop resource pickup event missing host winner: %+v", results)
	}
}

func TestManualGoldPickupStillWorksInRange(t *testing.T) {
	sim := MustNewSim("sess_gold_manual", "v49_gold_manual", loadRules(t))
	player := sim.entities[sim.playerID]
	gold := addTestGoldLoot(sim, player.pos, 9)

	res := sim.Tick([]Input{{MessageID: "pick_gold", CorrelationID: "corr_gold", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(gold.id)}}})

	assertAck(t, res, "pick_gold")
	if sim.gold != 9 || sim.entities[gold.id] != nil {
		t.Fatalf("manual gold pickup gold=%d entity=%+v, want gold 9 and removed", sim.gold, sim.entities[gold.id])
	}
	if owners := changeOwnersForOp(res, OpGoldUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("manual gold_update owners = %v, want player %d", owners, sim.playerID)
	}
	foundCorr := false
	for _, ev := range res.Events {
		if ev.EventType == "gold_picked_up" && ev.CorrelationID == "corr_gold" {
			foundCorr = true
		}
	}
	if !foundCorr {
		t.Fatalf("manual gold pickup missing correlation: %+v", res.Events)
	}
}

func TestGoldAutoPickupCoopLowestPlayerIDWins(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_gold_coop_winner", "v49_gold_coop_winner", rules)
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 11)
	sim.entities[hostID].pos = gold.pos
	sim.entities[guestID].pos = gold.pos

	results := sim.TickResults(nil)

	if sim.players[hostID].Gold != 11 || sim.players[guestID].Gold != 0 {
		t.Fatalf("coop gold host=%d guest=%d, want host winner", sim.players[hostID].Gold, sim.players[guestID].Gold)
	}
	if owners := changeOwnersForOpInResults(results, OpGoldUpdate); !sameUint64Set(owners, []uint64{hostID}) {
		t.Fatalf("coop gold_update owners = %v, want host %d", owners, hostID)
	}
	if !eventForEntityInResults(results, "gold_picked_up", hostID) {
		t.Fatalf("coop pickup event missing host winner: %+v", results)
	}
}

func TestGoldAutoPickupSkipsDeadAndDisconnectedPlayers(t *testing.T) {
	t.Run("dead lower player", func(t *testing.T) {
		rules := loadRules(t)
		sim := MustNewSim("sess_gold_dead_skip", "v49_gold_dead_skip", rules)
		hostID := sim.playerID
		guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
		if err != nil {
			t.Fatalf("add guest: %v", err)
		}
		gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 5)
		sim.entities[hostID].pos = gold.pos
		sim.entities[guestID].pos = gold.pos
		sim.entities[hostID].hp = 0

		sim.TickResults(nil)

		if sim.players[hostID].Gold != 0 || sim.players[guestID].Gold != 5 {
			t.Fatalf("dead-player skip gold host=%d guest=%d, want guest", sim.players[hostID].Gold, sim.players[guestID].Gold)
		}
	})
	t.Run("disconnected lower player", func(t *testing.T) {
		rules := loadRules(t)
		sim := MustNewSim("sess_gold_disconnected_skip", "v49_gold_disconnected_skip", rules)
		hostID := sim.playerID
		guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
		if err != nil {
			t.Fatalf("add guest: %v", err)
		}
		gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 6)
		sim.entities[hostID].pos = gold.pos
		sim.entities[guestID].pos = gold.pos
		sim.players[hostID].Connected = false

		sim.TickResults(nil)

		if sim.players[hostID].Gold != 0 || sim.players[guestID].Gold != 6 {
			t.Fatalf("disconnected-player skip gold host=%d guest=%d, want guest", sim.players[hostID].Gold, sim.players[guestID].Gold)
		}
	})
}

func TestGoldAutoPickupMultipleGoldStableEntityOrder(t *testing.T) {
	sim := MustNewSim("sess_gold_order", "v49_gold_order", loadRules(t))
	player := sim.entities[sim.playerID]
	first := addTestGoldLoot(sim, player.pos, 3)
	second := addTestGoldLoot(sim, player.pos, 5)
	if first.id >= second.id {
		t.Fatalf("test setup ids not increasing: first=%d second=%d", first.id, second.id)
	}

	res := sim.Tick(nil)

	if sim.gold != 8 {
		t.Fatalf("gold after multiple pickup = %d, want 8", sim.gold)
	}
	amounts := goldPickupAmounts(res)
	if len(amounts) != 2 || amounts[0] != 3 || amounts[1] != 5 {
		t.Fatalf("gold pickup amounts = %v, want [3 5]", amounts)
	}
}

func TestGoldAutoPickupPendingAutoNavDoesNotDuplicate(t *testing.T) {
	sim := MustNewSim("sess_gold_pending_nav", "v49_gold_pending_nav", loadRules(t))
	hostID := sim.playerID
	sim.entities[hostID].pos = Vec2{X: 2, Y: 2}
	gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 2}, 10)

	queue := sim.TickResults([]Input{{MessageID: "queue_gold", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(gold.id)}}})
	assertAckInResults(t, queue, "queue_gold")
	if sim.players[hostID].Gold != 0 {
		t.Fatalf("out-of-range click picked gold immediately: %d", sim.players[hostID].Gold)
	}

	pickupEvents := countEventsInResults(queue, "gold_picked_up")
	for i := 0; i < 20 && sim.players[hostID].Gold == 0; i++ {
		results := sim.TickResults(nil)
		pickupEvents += countEventsInResults(results, "gold_picked_up")
	}
	if sim.players[hostID].Gold != 10 {
		t.Fatalf("pending auto-nav gold = %d, want 10", sim.players[hostID].Gold)
	}
	if sim.entities[gold.id] != nil {
		t.Fatalf("pending auto-nav gold entity %d still present", gold.id)
	}
	for i := 0; i < 5; i++ {
		results := sim.TickResults(nil)
		pickupEvents += countEventsInResults(results, "gold_picked_up")
		if sim.players[hostID].Gold != 10 {
			t.Fatalf("gold duplicated after pickup: %d", sim.players[hostID].Gold)
		}
	}
	if pickupEvents != 1 {
		t.Fatalf("gold_picked_up events = %d, want exactly 1", pickupEvents)
	}
}

func addTestGoldLoot(sim *Sim, pos Vec2, amount int) *entity {
	loot := addTestFloorLoot(sim, goldItemDefID, pos)
	loot.goldAmount = amount
	return loot
}

func addTestWalletResourceLoot(sim *Sim, pos Vec2) *entity {
	return addTestFloorLoot(sim, "respec_badge", pos)
}
