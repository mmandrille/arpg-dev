package game

import "testing"

func TestAccountStashOpenAndTransfers(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_stash", "v50_stash", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})

	item := &invItem{instanceID: sim.alloc(), itemDefID: "red_potion"}
	sim.inventory = append(sim.inventory, item)
	sim.gold = 20
	sim.progression.Gold = 20
	sim.savePlayer(sim.defaultPlayer())

	open := sim.Tick([]Input{{MessageID: "open_stash", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(stash.id)}}})
	if !hasAck(open, "open_stash") {
		t.Fatalf("open stash ack missing: %+v", open)
	}
	opened := findEvent(open.Events, "stash_opened")
	if opened == nil || opened.StashID != accountStashID || opened.StashCapacity == nil || *opened.StashCapacity != defaultStashCapacity {
		t.Fatalf("stash_opened event = %+v", opened)
	}

	deposit := sim.Tick([]Input{{MessageID: "deposit_item", Type: "stash_deposit_item_intent", StashDepositItem: &StashDepositItemIntent{StashEntityID: idStr(stash.id), ItemInstanceID: idStr(item.instanceID)}}})
	if !hasAck(deposit, "deposit_item") {
		t.Fatalf("deposit ack missing: %+v", deposit)
	}
	depositEvent := findEvent(deposit.Events, "stash_item_deposited")
	if depositEvent == nil || depositEvent.ItemInstanceID != idStr(item.instanceID) || depositEvent.StashItemID == "" {
		t.Fatalf("stash_item_deposited event = %+v", depositEvent)
	}
	if sim.findItemByID(item.instanceID) != nil {
		t.Fatal("deposited item remained in inventory")
	}
	if len(sim.stashItems) != 1 {
		t.Fatalf("stash item count = %d, want 1", len(sim.stashItems))
	}
	if !hasChangeOp(deposit, OpInventoryRemove) || !hasChangeOp(deposit, OpStashItemAdd) {
		t.Fatalf("deposit changes missing inventory_remove/stash_item_add: %+v", deposit.Changes)
	}

	withdraw := sim.Tick([]Input{{MessageID: "withdraw_item", Type: "stash_withdraw_item_intent", StashWithdrawItem: &StashWithdrawItemIntent{StashEntityID: idStr(stash.id), StashItemID: depositEvent.StashItemID}}})
	if !hasAck(withdraw, "withdraw_item") {
		t.Fatalf("withdraw ack missing: %+v", withdraw)
	}
	withdrawEvent := findEvent(withdraw.Events, "stash_item_withdrawn")
	if withdrawEvent == nil || withdrawEvent.StashItemID != depositEvent.StashItemID || withdrawEvent.ItemInstanceID == "" {
		t.Fatalf("stash_item_withdrawn event = %+v", withdrawEvent)
	}
	if len(sim.stashItems) != 0 {
		t.Fatalf("stash item count after withdraw = %d, want 0", len(sim.stashItems))
	}
	if sim.findItem(withdrawEvent.ItemInstanceID) == nil {
		t.Fatalf("withdrawn item %s missing from inventory", withdrawEvent.ItemInstanceID)
	}
	if !hasChangeOp(withdraw, OpStashItemRemove) || !hasChangeOp(withdraw, OpInventoryAdd) {
		t.Fatalf("withdraw changes missing stash_item_remove/inventory_add: %+v", withdraw.Changes)
	}

	depositGold := sim.Tick([]Input{{MessageID: "deposit_gold", Type: "stash_deposit_gold_intent", StashDepositGold: &StashDepositGoldIntent{StashEntityID: idStr(stash.id), Amount: 7}}})
	if !hasAck(depositGold, "deposit_gold") || sim.gold != 13 || sim.stashGold != 7 {
		t.Fatalf("deposit gold result gold=%d stash=%d res=%+v", sim.gold, sim.stashGold, depositGold)
	}
	if ev := findEvent(depositGold.Events, "stash_gold_deposited"); ev == nil || ev.Amount == nil || *ev.Amount != 7 || ev.TotalGold == nil || *ev.TotalGold != 13 || ev.StashGold == nil || *ev.StashGold != 7 {
		t.Fatalf("stash_gold_deposited event = %+v", ev)
	}

	withdrawGold := sim.Tick([]Input{{MessageID: "withdraw_gold", Type: "stash_withdraw_gold_intent", StashWithdrawGold: &StashWithdrawGoldIntent{StashEntityID: idStr(stash.id), Amount: 5}}})
	if !hasAck(withdrawGold, "withdraw_gold") || sim.gold != 18 || sim.stashGold != 2 {
		t.Fatalf("withdraw gold result gold=%d stash=%d res=%+v", sim.gold, sim.stashGold, withdrawGold)
	}
	if ev := findEvent(withdrawGold.Events, "stash_gold_withdrawn"); ev == nil || ev.Amount == nil || *ev.Amount != 5 || ev.TotalGold == nil || *ev.TotalGold != 18 || ev.StashGold == nil || *ev.StashGold != 2 {
		t.Fatalf("stash_gold_withdrawn event = %+v", ev)
	}
}

func TestAccountStashRejectsHotbarAssignedItem(t *testing.T) {
	sim, err := NewSimWithWorld("sess_stash_hotbar", "v50_stash_hotbar", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})

	item := &invItem{instanceID: sim.alloc(), itemDefID: "red_potion"}
	sim.inventory = append(sim.inventory, item)
	sim.hotbar[0] = item.instanceID
	sim.savePlayer(sim.defaultPlayer())

	res := sim.Tick([]Input{{MessageID: "deposit_hotbar", Type: "stash_deposit_item_intent", StashDepositItem: &StashDepositItemIntent{StashEntityID: idStr(stash.id), ItemInstanceID: idStr(item.instanceID)}}})
	if !hasReject(res, "deposit_hotbar", "item_hotbar_assigned") {
		t.Fatalf("deposit hotbar reject missing: %+v", res)
	}
	if len(sim.stashItems) != 0 || sim.findItemByID(item.instanceID) == nil {
		t.Fatalf("hotbar deposit mutated state: stash=%d item=%v", len(sim.stashItems), sim.findItemByID(item.instanceID))
	}
}

func TestAccountStashRejectsInvalidTransfers(t *testing.T) {
	t.Run("equipped item deposits and clears slot", func(t *testing.T) {
		sim, stash := newReadyStashSim(t, "equipped")
		item := &invItem{instanceID: sim.alloc(), itemDefID: "long_sword", slot: mainHandSlot, equipped: true}
		sim.inventory = append(sim.inventory, item)
		sim.equipped[mainHandSlot] = item.instanceID
		sim.savePlayer(sim.defaultPlayer())

		res := sim.Tick([]Input{{MessageID: "deposit_equipped", Type: "stash_deposit_item_intent", StashDepositItem: &StashDepositItemIntent{StashEntityID: idStr(stash.id), ItemInstanceID: idStr(item.instanceID)}}})
		if !hasAck(res, "deposit_equipped") {
			t.Fatalf("deposit equipped ack missing: %+v", res)
		}
		if len(sim.stashItems) != 1 || sim.findItemByID(item.instanceID) != nil || sim.equipped[mainHandSlot] != 0 {
			t.Fatalf("equipped deposit state stash=%d item=%v equipped=%d", len(sim.stashItems), sim.findItemByID(item.instanceID), sim.equipped[mainHandSlot])
		}
		if !hasChangeOp(res, OpEquippedUpdate) || !hasChangeOp(res, OpInventoryRemove) || !hasChangeOp(res, OpStashItemAdd) {
			t.Fatalf("equipped deposit changes missing equipped/inventory/stash updates: %+v", res.Changes)
		}
	})

	t.Run("stash full", func(t *testing.T) {
		sim, stash := newReadyStashSim(t, "stash_full")
		sim.stashCapacity = 1
		sim.stashItems = []*stashItem{{stashItemID: sim.alloc(), itemDefID: "red_potion"}}
		item := &invItem{instanceID: sim.alloc(), itemDefID: "red_potion"}
		sim.inventory = append(sim.inventory, item)
		sim.savePlayer(sim.defaultPlayer())

		res := sim.Tick([]Input{{MessageID: "deposit_full", Type: "stash_deposit_item_intent", StashDepositItem: &StashDepositItemIntent{StashEntityID: idStr(stash.id), ItemInstanceID: idStr(item.instanceID)}}})
		if !hasReject(res, "deposit_full", "stash_full") {
			t.Fatalf("deposit full reject missing: %+v", res)
		}
		if len(sim.stashItems) != 1 || sim.findItemByID(item.instanceID) == nil {
			t.Fatalf("full stash deposit mutated state: stash=%d item=%v", len(sim.stashItems), sim.findItemByID(item.instanceID))
		}
	})

	t.Run("inventory full", func(t *testing.T) {
		sim, stash := newReadyStashSim(t, "inventory_full")
		stashItemID := sim.alloc()
		sim.stashItems = []*stashItem{{stashItemID: stashItemID, itemDefID: "red_potion"}}
		sim.inventory = nil
		for i := 0; i < sim.inventoryCapacity(); i++ {
			sim.inventory = append(sim.inventory, &invItem{instanceID: sim.alloc(), itemDefID: "red_potion"})
		}
		sim.savePlayer(sim.defaultPlayer())

		res := sim.Tick([]Input{{MessageID: "withdraw_full", Type: "stash_withdraw_item_intent", StashWithdrawItem: &StashWithdrawItemIntent{StashEntityID: idStr(stash.id), StashItemID: idStr(stashItemID)}}})
		if !hasReject(res, "withdraw_full", "inventory_full") {
			t.Fatalf("withdraw full reject missing: %+v", res)
		}
		if len(sim.stashItems) != 1 || sim.findStashItem(idStr(stashItemID)) == nil {
			t.Fatalf("full inventory withdraw mutated stash: %+v", sim.stashItems)
		}
	})

	t.Run("gold balances and amount", func(t *testing.T) {
		sim, stash := newReadyStashSim(t, "gold_rejects")
		sim.gold = 2
		sim.progression.Gold = 2
		sim.stashGold = 1
		sim.savePlayer(sim.defaultPlayer())

		depositInvalid := sim.Tick([]Input{{MessageID: "deposit_invalid", Type: "stash_deposit_gold_intent", StashDepositGold: &StashDepositGoldIntent{StashEntityID: idStr(stash.id), Amount: 0}}})
		if !hasReject(depositInvalid, "deposit_invalid", "invalid_amount") {
			t.Fatalf("deposit invalid reject missing: %+v", depositInvalid)
		}
		depositTooMuch := sim.Tick([]Input{{MessageID: "deposit_too_much", Type: "stash_deposit_gold_intent", StashDepositGold: &StashDepositGoldIntent{StashEntityID: idStr(stash.id), Amount: 3}}})
		if !hasReject(depositTooMuch, "deposit_too_much", "insufficient_gold") {
			t.Fatalf("deposit insufficient reject missing: %+v", depositTooMuch)
		}
		withdrawTooMuch := sim.Tick([]Input{{MessageID: "withdraw_too_much", Type: "stash_withdraw_gold_intent", StashWithdrawGold: &StashWithdrawGoldIntent{StashEntityID: idStr(stash.id), Amount: 2}}})
		if !hasReject(withdrawTooMuch, "withdraw_too_much", "insufficient_stash_gold") {
			t.Fatalf("withdraw insufficient reject missing: %+v", withdrawTooMuch)
		}
		if sim.gold != 2 || sim.stashGold != 1 {
			t.Fatalf("rejected gold transfer mutated balances gold=%d stash=%d", sim.gold, sim.stashGold)
		}
	})

	t.Run("out of range", func(t *testing.T) {
		sim, stash := newReadyStashSim(t, "out_of_range")
		moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X + 99, Y: stash.pos.Y + 99})
		sim.gold = 1
		sim.progression.Gold = 1
		sim.savePlayer(sim.defaultPlayer())

		res := sim.Tick([]Input{{MessageID: "deposit_far", Type: "stash_deposit_gold_intent", StashDepositGold: &StashDepositGoldIntent{StashEntityID: idStr(stash.id), Amount: 1}}})
		if !hasReject(res, "deposit_far", "out_of_range") {
			t.Fatalf("deposit far reject missing: %+v", res)
		}
	})
}

func newReadyStashSim(t *testing.T, name string) (*Sim, *entity) {
	t.Helper()
	sim, err := NewSimWithWorld("sess_stash_"+name, "v50_stash_"+name, loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})
	return sim, stash
}

func townStashEntity(t *testing.T, sim *Sim) *entity {
	t.Helper()
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		e := sim.activeLevel().entities[id]
		if e != nil && e.kind == interactableEntity && e.interactableDefID == townStashDefID {
			return e
		}
	}
	t.Fatal("missing town stash")
	return nil
}

func hasChangeOp(res TickResult, op string) bool {
	for _, change := range res.Changes {
		if change.Op == op {
			return true
		}
	}
	return false
}
