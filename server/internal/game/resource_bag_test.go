package game

import "testing"

func TestPickUpResourceBagItemRoutesQuestLeafToBag(t *testing.T) {
	sim := MustNewSim("sess_resource_bag_pickup", "resource_bag_pickup", loadRules(t))
	player := sim.activeLevel().entities[sim.playerID]
	if player == nil {
		t.Fatal("missing player")
	}
	level := sim.activeLevel()
	loot := sim.newLootEntity("quest_leaf", player.pos, nil, goldRollContext{levelNum: level.levelNum})
	loot.id = sim.alloc()
	level.entities[loot.id] = loot

	res := sim.Tick([]Input{{MessageID: "pickup", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	if len(res.Rejects) != 0 {
		t.Fatalf("pickup rejected: %+v", res.Rejects)
	}
	if len(sim.resourceBagItems) != 1 {
		t.Fatalf("resource bag count = %d, want 1", len(sim.resourceBagItems))
	}
	if sim.resourceBagItems[0].itemDefID != "quest_leaf" {
		t.Fatalf("resource bag item = %q, want quest_leaf", sim.resourceBagItems[0].itemDefID)
	}
	if sim.bagOccupancyCount() != 0 {
		t.Fatalf("inventory occupancy = %d, want 0", sim.bagOccupancyCount())
	}
}

func TestResourceBagDepositWithdrawRoundTrip(t *testing.T) {
	sim := MustNewSim("sess_resource_bag_transfer", "resource_bag_transfer", loadRules(t))
	item := &invItem{
		instanceID:  sim.alloc(),
		itemDefID:   "upgrade_shard",
		rollPayload: NewUpgradeShardRollPayload(2),
	}
	sim.inventory = append(sim.inventory, item)
	sim.savePlayer(sim.defaultPlayer())

	deposit := sim.Tick([]Input{{
		MessageID: "deposit",
		Type:      "resource_bag_deposit_item_intent",
		ResourceBagDepositItem: &ResourceBagDepositItemIntent{
			ItemInstanceID: idStr(item.instanceID),
		},
	}})
	if len(deposit.Rejects) != 0 {
		t.Fatalf("deposit rejected: %+v", deposit.Rejects)
	}
	if len(sim.resourceBagItems) != 1 || sim.findItemByID(item.instanceID) != nil {
		t.Fatalf("after deposit bag=%d inventory=%v", len(sim.resourceBagItems), sim.findItemByID(item.instanceID))
	}

	bagID := idStr(sim.resourceBagItems[0].stashItemID)
	withdraw := sim.Tick([]Input{{
		MessageID: "withdraw",
		Type:      "resource_bag_withdraw_item_intent",
		ResourceBagWithdrawItem: &ResourceBagWithdrawItemIntent{
			BagItemID: bagID,
		},
	}})
	if len(withdraw.Rejects) != 0 {
		t.Fatalf("withdraw rejected: %+v", withdraw.Rejects)
	}
	if len(sim.resourceBagItems) != 0 || sim.bagOccupancyCount() != 1 {
		t.Fatalf("after withdraw bag=%d occupancy=%d", len(sim.resourceBagItems), sim.bagOccupancyCount())
	}
}

func TestIsResourceBagItemClassification(t *testing.T) {
	sim := MustNewSim("sess_resource_bag_class", "resource_bag_class", loadRules(t))
	if !sim.isResourceBagItem("upgrade_shard") {
		t.Fatal("upgrade_shard should be resource bag item")
	}
	if !sim.isResourceBagItem("quest_trophy_wolf_heart") {
		t.Fatal("quest trophy should be resource bag item")
	}
	if sim.isResourceBagItem("respec_badge") {
		t.Fatal("respec_badge should stay wallet item")
	}
	if sim.isResourceBagItem("long_sword") {
		t.Fatal("equipment should not be resource bag item")
	}
}

func TestResourceBagDepositFromStash(t *testing.T) {
	sim, err := NewSimWithWorld("sess_resource_bag_stash", "resource_bag_stash", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})

	stashItemID := sim.alloc()
	sim.stashItems = []*stashItem{{
		stashItemID: stashItemID,
		itemDefID:   "upgrade_shard",
		rollPayload: NewUpgradeShardRollPayload(3),
	}}
	sim.savePlayer(sim.defaultPlayer())

	deposit := sim.Tick([]Input{{
		MessageID: "deposit_stash",
		Type:      "resource_bag_deposit_stash_item_intent",
		ResourceBagDepositStashItem: &ResourceBagDepositStashItemIntent{
			StashEntityID: idStr(stash.id),
			StashItemID:   idStr(stashItemID),
		},
	}})
	if len(deposit.Rejects) != 0 {
		t.Fatalf("deposit rejected: %+v", deposit.Rejects)
	}
	if len(sim.stashItems) != 0 {
		t.Fatalf("stash item count = %d, want 0", len(sim.stashItems))
	}
	if len(sim.resourceBagItems) != 1 {
		t.Fatalf("resource bag count = %d, want 1", len(sim.resourceBagItems))
	}
	if sim.resourceBagItems[0].itemDefID != "upgrade_shard" {
		t.Fatalf("bag item = %q, want upgrade_shard", sim.resourceBagItems[0].itemDefID)
	}
	if !hasChangeOp(deposit, OpStashItemRemove) || !hasChangeOp(deposit, OpResourceBagItemAdd) {
		t.Fatalf("deposit changes missing stash_item_remove/resource_bag_item_add: %+v", deposit.Changes)
	}
}

func TestStashDepositResourceBagItem(t *testing.T) {
	sim, err := NewSimWithWorld("sess_stash_resource_bag", "stash_resource_bag", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	stash := townStashEntity(t, sim)
	moveDefaultPlayerTo(sim, Vec2{X: stash.pos.X, Y: stash.pos.Y - 0.25})

	bagItemID := sim.alloc()
	sim.resourceBagItems = []*stashItem{{
		stashItemID: bagItemID,
		itemDefID:   "renew_stone",
		rollPayload: NewUpgradeShardRollPayload(2),
	}}
	sim.savePlayer(sim.defaultPlayer())

	deposit := sim.Tick([]Input{{
		MessageID: "deposit_bag_to_stash",
		Type:      "stash_deposit_resource_bag_item_intent",
		StashDepositResourceBagItem: &StashDepositResourceBagItemIntent{
			StashEntityID: idStr(stash.id),
			BagItemID:     idStr(bagItemID),
		},
	}})
	if len(deposit.Rejects) != 0 {
		t.Fatalf("deposit rejected: %+v", deposit.Rejects)
	}
	if len(sim.resourceBagItems) != 0 {
		t.Fatalf("resource bag count = %d, want 0", len(sim.resourceBagItems))
	}
	if len(sim.stashItems) != 1 {
		t.Fatalf("stash item count = %d, want 1", len(sim.stashItems))
	}
	if sim.stashItems[0].itemDefID != "renew_stone" {
		t.Fatalf("stash item = %q, want renew_stone", sim.stashItems[0].itemDefID)
	}
	if !hasChangeOp(deposit, OpResourceBagItemRemove) || !hasChangeOp(deposit, OpStashItemAdd) {
		t.Fatalf("deposit changes missing resource_bag_item_remove/stash_item_add: %+v", deposit.Changes)
	}
}
