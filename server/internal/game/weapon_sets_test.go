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
