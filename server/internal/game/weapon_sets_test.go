package game

import "testing"

func TestWeaponSetEquipTargetsInactiveSetAndSwapUpdatesActiveHands(t *testing.T) {
	sim, err := NewSimWithWorld("sess_weapon_sets", "weapon_sets", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	blade := addRolledInventoryItem(t, sim, 9101, "cave_blade", nil)
	bow := addRolledInventoryItem(t, sim, 9102, "cave_bow", nil)
	set2 := 1

	assertAck(t, sim.Tick([]Input{{MessageID: "blade_set1", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "blade_set1")
	assertAck(t, sim.Tick([]Input{{MessageID: "bow_set2", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot, WeaponSet: &set2}}}), "bow_set2")

	if sim.activeWeaponSet != 0 {
		t.Fatalf("active weapon set = %d, want 0", sim.activeWeaponSet)
	}
	assertEquippedTemplate(t, sim, mainHandSlot, "cave_blade")
	if got := sim.weaponSets[1][mainHandSlot]; got != bow.instanceID {
		t.Fatalf("set 2 main_hand = %d, want %d", got, bow.instanceID)
	}

	swap := sim.Tick([]Input{{MessageID: "swap", Type: "swap_weapon_set_intent", SwapWeaponSet: &SwapWeaponSetIntent{}}})
	assertAck(t, swap, "swap")
	if sim.activeWeaponSet != 1 {
		t.Fatalf("active weapon set after swap = %d, want 1", sim.activeWeaponSet)
	}
	assertEquippedTemplate(t, sim, mainHandSlot, "cave_bow")
	if got := sim.playerAttackMode(); got != attackModeRanged {
		t.Fatalf("active attack mode after bow swap = %s, want %s", got, attackModeRanged)
	}
	if !hasEvent(swap, "weapon_set_swapped") || !hasChangeOp(swap, OpWeaponSetUpdate) {
		t.Fatalf("swap result missing event/change: %+v", swap)
	}
	assertEquippedUpdateWeaponSet(t, swap, mainHandSlot, set2)

	snap := sim.Snapshot()
	if snap.ActiveWeaponSet != 1 || len(snap.WeaponSets) != weaponSetCount {
		t.Fatalf("snapshot weapon sets = active %d sets %+v", snap.ActiveWeaponSet, snap.WeaponSets)
	}
	if snap.WeaponSets[0].MainHand == nil || *snap.WeaponSets[0].MainHand != idStr(blade.instanceID) {
		t.Fatalf("snapshot set 1 main hand = %+v, want blade", snap.WeaponSets[0].MainHand)
	}
	if snap.WeaponSets[1].MainHand == nil || *snap.WeaponSets[1].MainHand != idStr(bow.instanceID) {
		t.Fatalf("snapshot set 2 main hand = %+v, want bow", snap.WeaponSets[1].MainHand)
	}
}

func assertEquippedUpdateWeaponSet(t *testing.T, res TickResult, slot string, want int) {
	t.Helper()
	for _, change := range res.Changes {
		if change.Op != OpEquippedUpdate || change.Slot != slot || change.ItemInstanceID == nil {
			continue
		}
		if change.WeaponSet == nil || *change.WeaponSet != want {
			t.Fatalf("equipped_update %s weapon_set = %+v, want %d in changes %+v", slot, change.WeaponSet, want, res.Changes)
		}
		return
	}
	t.Fatalf("missing equipped_update for %s in changes %+v", slot, res.Changes)
}

func TestLoadInventoryRestoresPersistedWeaponSets(t *testing.T) {
	sim, err := NewSimWithWorld("sess_weapon_sets_reload", "weapon_sets_reload", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	sim.LoadInventory([]PersistedItem{
		{InstanceID: "9301", ItemDefID: "cave_greatsword", Slot: mainHandSlot, Equipped: true, WeaponSet: 0},
		{InstanceID: "9302", ItemDefID: "cave_bow", Slot: mainHandSlot, Equipped: true, WeaponSet: 1},
	})

	assertEquippedItemDef(t, sim, mainHandSlot, "cave_greatsword")
	if sim.weaponSets[0][mainHandSlot] != 9301 {
		t.Fatalf("set 1 main_hand = %d, want 9301", sim.weaponSets[0][mainHandSlot])
	}
	if sim.weaponSets[1][mainHandSlot] != 9302 {
		t.Fatalf("set 2 main_hand = %d, want 9302", sim.weaponSets[1][mainHandSlot])
	}

	swap := sim.Tick([]Input{{MessageID: "swap", Type: "swap_weapon_set_intent", SwapWeaponSet: &SwapWeaponSetIntent{}}})
	assertAck(t, swap, "swap")
	assertEquippedItemDef(t, sim, mainHandSlot, "cave_bow")
}

func TestLoadInventoryRestoresFourHandSlots(t *testing.T) {
	sim, err := NewSimWithWorld("sess_weapon_sets_four_hands_reload", "weapon_sets_reload", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	sim.LoadInventory([]PersistedItem{
		{InstanceID: "9401", ItemDefID: "cave_blade", Slot: mainHandSlot, Equipped: true, WeaponSet: 0},
		{InstanceID: "9402", ItemDefID: "cave_shield", Slot: offHandSlot, Equipped: true, WeaponSet: 0},
		{InstanceID: "9403", ItemDefID: "cave_bow", Slot: mainHandSlot, Equipped: true, WeaponSet: 1},
	})

	assertEquippedItemDef(t, sim, mainHandSlot, "cave_blade")
	assertEquippedItemDef(t, sim, offHandSlot, "cave_shield")
	snap := sim.Snapshot()
	if snap.WeaponSets[0].MainHand == nil || *snap.WeaponSets[0].MainHand != "9401" ||
		snap.WeaponSets[0].OffHand == nil || *snap.WeaponSets[0].OffHand != "9402" ||
		snap.WeaponSets[1].MainHand == nil || *snap.WeaponSets[1].MainHand != "9403" ||
		snap.WeaponSets[1].OffHand != nil {
		t.Fatalf("snapshot weapon sets = %+v, want set1 sword+shield and set2 bow", snap.WeaponSets)
	}
	if len(snap.Inventory) != 3 {
		t.Fatalf("inventory count = %d, want 3", len(snap.Inventory))
	}
	for _, item := range snap.Inventory {
		if !item.Equipped {
			t.Fatalf("item %s loaded as bag item in snapshot: %+v", item.ItemDefID, item)
		}
	}

	swap := sim.Tick([]Input{{MessageID: "swap", Type: "swap_weapon_set_intent", SwapWeaponSet: &SwapWeaponSetIntent{}}})
	assertAck(t, swap, "swap")
	assertEquippedItemDef(t, sim, mainHandSlot, "cave_bow")
	if sim.equipped[offHandSlot] != 0 {
		t.Fatalf("set 2 off hand = %d, want empty with bow", sim.equipped[offHandSlot])
	}
}

func TestWeaponSetEquipFlowPersistsFourHandSlots(t *testing.T) {
	sim, err := NewSimWithWorld("sess_weapon_sets_four_hands_flow", "weapon_sets", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	blade := addRolledInventoryItem(t, sim, 9501, "cave_blade", nil)
	shield := addRolledInventoryItem(t, sim, 9502, "cave_shield", nil)
	bow := addRolledInventoryItem(t, sim, 9503, "cave_bow", nil)
	persisted := map[string]PersistedItem{
		"9501": {InstanceID: "9501", ItemDefID: "cave_blade"},
		"9502": {InstanceID: "9502", ItemDefID: "cave_shield"},
		"9503": {InstanceID: "9503", ItemDefID: "cave_bow"},
	}
	set2 := 1

	applyWeaponSetPersistence(t, persisted, sim.Tick([]Input{{MessageID: "blade_set1", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}))
	applyWeaponSetPersistence(t, persisted, sim.Tick([]Input{{MessageID: "shield_set1", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}))
	applyWeaponSetPersistence(t, persisted, sim.Tick([]Input{{MessageID: "bow_set2", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot, WeaponSet: &set2}}}))
	applyWeaponSetPersistence(t, persisted, sim.Tick([]Input{{MessageID: "swap", Type: "swap_weapon_set_intent", SwapWeaponSet: &SwapWeaponSetIntent{}}}))

	reloaded, err := NewSimWithWorld("sess_weapon_sets_four_hands_reloaded", "weapon_sets", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new reload sim: %v", err)
	}
	reloaded.LoadInventory(persistedItemsFromMap(persisted))
	assertEquippedItemDef(t, reloaded, mainHandSlot, "cave_blade")
	assertEquippedItemDef(t, reloaded, offHandSlot, "cave_shield")
	if reloaded.weaponSets[1][mainHandSlot] != bow.instanceID || reloaded.weaponSets[1][offHandSlot] != 0 {
		t.Fatalf("reloaded set 2 = %+v, want bow only", reloaded.weaponSets[1])
	}
}

func TestWeaponSetHandBlockingIsPerSet(t *testing.T) {
	sim, err := NewSimWithWorld("sess_weapon_sets_blocking", "weapon_sets", loadRules(t), "equipment_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	greatsword := addRolledInventoryItem(t, sim, 9201, "cave_greatsword", nil)
	shield := addRolledInventoryItem(t, sim, 9202, "cave_shield", nil)
	set2 := 1

	assertAck(t, sim.Tick([]Input{{MessageID: "greatsword_set1", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}}), "greatsword_set1")
	assertAck(t, sim.Tick([]Input{{MessageID: "shield_set2", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot, WeaponSet: &set2}}}), "shield_set2")
	if sim.weaponSets[1][offHandSlot] != shield.instanceID {
		t.Fatalf("set 2 off_hand = %d, want %d", sim.weaponSets[1][offHandSlot], shield.instanceID)
	}

	reject := sim.Tick([]Input{{MessageID: "shield_set1", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}})
	assertReject(t, reject, "shield_set1", "hands_blocked")
}

func applyWeaponSetPersistence(t *testing.T, persisted map[string]PersistedItem, res TickResult) {
	t.Helper()
	if len(res.Rejects) > 0 {
		t.Fatalf("unexpected rejection: %+v", res.Rejects)
	}
	for _, change := range res.Changes {
		switch change.Op {
		case OpInventoryUpdate:
			if change.Item == nil {
				continue
			}
			weaponSet := 0
			if change.WeaponSet != nil {
				weaponSet = *change.WeaponSet
			}
			persisted[change.Item.ItemInstanceID] = PersistedItem{
				InstanceID: change.Item.ItemInstanceID,
				ItemDefID:  change.Item.ItemDefID,
				Slot:       change.Item.Slot,
				Equipped:   change.Item.Equipped,
				WeaponSet:  weaponSet,
			}
		case OpEquippedUpdate:
			if change.ItemInstanceID == nil {
				continue
			}
			item := persisted[*change.ItemInstanceID]
			item.Slot = change.Slot
			item.Equipped = true
			if change.WeaponSet != nil {
				item.WeaponSet = *change.WeaponSet
			}
			persisted[*change.ItemInstanceID] = item
		}
	}
}

func persistedItemsFromMap(items map[string]PersistedItem) []PersistedItem {
	out := make([]PersistedItem, 0, len(items))
	for _, key := range []string{"9501", "9502", "9503"} {
		out = append(out, items[key])
	}
	return out
}

func assertEquippedItemDef(t *testing.T, sim *Sim, slot, itemDefID string) {
	t.Helper()
	item := sim.findItemByID(sim.equipped[slot])
	if item == nil || item.itemDefID != itemDefID {
		t.Fatalf("equipped[%s] = %+v, want item def %s", slot, item, itemDefID)
	}
}
