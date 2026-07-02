package game

import "testing"

func TestRogueDualWieldMeleeReachUsesShorterOffHand(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := newRogueSkillTestSim(t, rules)
	dagger := addRolledInventoryItem(t, sim, 9201, "dagger", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_dagger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(dagger.instanceID), Slot: offHandSlot}}}), "equip_dagger")

	if got := sim.playerWeaponSlotReach(mainHandSlot); got < 1.4 || got > 1.5 {
		t.Fatalf("main hand reach = %.2f, want ~1.45", got)
	}
	if got := sim.playerWeaponSlotReach(offHandSlot); got < 1.0 || got > 1.15 {
		t.Fatalf("off hand reach = %.2f, want ~1.1", got)
	}
	if got := sim.playerMeleeReach(); got < 1.0 || got > 1.15 {
		t.Fatalf("playerMeleeReach = %.2f, want dagger reach ~1.1", got)
	}
}

func TestRogueWeaponSlotMeleeRangeOffHandShorterThanMain(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := newRogueSkillTestSim(t, rules)
	dagger := addRolledInventoryItem(t, sim, 9202, "dagger", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_dagger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(dagger.instanceID), Slot: offHandSlot}}}), "equip_dagger")
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.65, Y: player.pos.Y}, 50)

	if !sim.inWeaponSlotMeleeRange(target, mainHandSlot) {
		t.Fatal("main hand should be in range at distance 1.65")
	}
	if sim.inWeaponSlotMeleeRange(target, offHandSlot) {
		t.Fatal("off hand should be out of range at distance 1.65")
	}
	if sim.inMeleeRange(target) {
		t.Fatal("dispatch range should use shorter off-hand reach at distance 1.65")
	}
}

func TestRogueOffHandHitsAtDualWieldReach(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1)
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	dagger := addRolledInventoryItem(t, sim, 9203, "dagger", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_dagger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(dagger.instanceID), Slot: offHandSlot}}}), "equip_dagger")
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 50)

	main := sim.Tick([]Input{{MessageID: "main_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
	assertAck(t, main, "main_1")
	if !hasWeaponSlotDamageEvent(main, mainHandSlot) {
		t.Fatalf("main hand should hit at 1.2: events=%+v", main.Events)
	}
	off := sim.Tick([]Input{{MessageID: "off_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
	assertAck(t, off, "off_1")
	if !hasWeaponSlotDamageEvent(off, offHandSlot) {
		t.Fatalf("off hand should hit at dual-wield reach 1.2: events=%+v rejects=%+v", off.Events, off.Rejects)
	}
}

func TestRogueOffHandMissesOutsideSlotReachWhenForced(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	dagger := addRolledInventoryItem(t, sim, 9204, "dagger", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_dagger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(dagger.instanceID), Slot: offHandSlot}}}), "equip_dagger")
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.65, Y: player.pos.Y}, 50)

	res := &TickResult{}
	sim.emitPlayerWeaponMiss(target, sim.playerID, "corr_forced_miss", res, offHandSlot)
	if !hasWeaponSlotCombatEvent(*res, offHandSlot) {
		t.Fatalf("forced off-hand range miss should emit combat event: events=%+v", res.Events)
	}
	if hasWeaponSlotDamageEvent(*res, offHandSlot) {
		t.Fatalf("forced off-hand range miss should not deal damage: events=%+v", res.Events)
	}
}
