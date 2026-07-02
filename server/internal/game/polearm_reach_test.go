package game

import "testing"

func newPolearmReachTestSim(t *testing.T, rules *Rules) *Sim {
	t.Helper()
	sim := MustNewSim("sess_polearm_reach", "polearm_reach_seed", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.BaseStats.Str = 12
	sim.savePlayer(sim.defaultPlayer())
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}

	return sim
}

func addPolearmReachTarget(sim *Sim, pos Vec2, hp int) *entity {
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: pos, hp: hp, maxHP: hp, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[target.id] = target

	return target
}

func TestPolearmSpearReachFromTemplate(t *testing.T) {
	rules := loadRules(t)
	sim := newPolearmReachTestSim(t, rules)
	spear := addRolledInventoryItem(t, sim, 9410, "spear", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_spear", Type: "equip_intent", Equip: &EquipIntent{
		ItemInstanceID: idStr(spear.instanceID),
		Slot:         mainHandSlot,
	}}}), "equip_spear")

	got := sim.playerWeaponSlotReach(mainHandSlot)
	if got < 2.9 {
		t.Fatalf("spear reach = %.2f, want >= 2.9", got)
	}
}

func TestPolearmSpearHitsBeyondLongSwordRange(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1)
	rules.Combat.BaseCritChance = 0
	sim := newPolearmReachTestSim(t, rules)
	spear := addRolledInventoryItem(t, sim, 9411, "spear", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_spear", Type: "equip_intent", Equip: &EquipIntent{
		ItemInstanceID: idStr(spear.instanceID),
		Slot:         mainHandSlot,
	}}}), "equip_spear")

	player := sim.entities[sim.playerID]
	target := addPolearmReachTarget(sim, Vec2{X: player.pos.X + 2.5, Y: player.pos.Y}, 50)

	if !sim.inWeaponSlotMeleeRange(target, mainHandSlot) {
		t.Fatal("spear should be in range at distance 2.5")
	}

	longSword := addRolledInventoryItem(t, sim, 9412, "long_sword", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_long_sword", Type: "equip_intent", Equip: &EquipIntent{
		ItemInstanceID: idStr(longSword.instanceID),
		Slot:         mainHandSlot,
	}}}), "equip_long_sword")

	if sim.inWeaponSlotMeleeRange(target, mainHandSlot) {
		t.Fatal("long sword should be out of range at distance 2.5")
	}
}
