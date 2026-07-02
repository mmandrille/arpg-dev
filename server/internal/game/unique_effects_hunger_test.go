package game

import "testing"

func TestOffensiveUniqueHungerOfTheDeepRampsAndResets(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_hunger", "hunger", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	secondary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X, Y: player.pos.Y + 1.2}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[secondary.id] = secondary
	blade := addRolledInventoryItem(t, sim, 9806, "long_sword", map[string]int{"damage_min": 25, "damage_max": 25})
	blade.rollPayload.EffectIDs = []string{hungerOfTheDeepEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_hunger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_hunger")

	first := sim.Tick([]Input{{MessageID: "hunger_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	firstDamage, ok := uniqueEventDamage(first, "monster_damaged", "")
	if !ok {
		t.Fatalf("first hunger events = %+v, want damage", first.Events)
	}
	if stack := sim.uniqueHungerStacks[player.id]; stack.Stacks != 1 || stack.TargetID != primary.id {
		t.Fatalf("after first hit stack = %+v, want 1 stack on primary", stack)
	}
	advanceBasicAttackCooldown(sim)
	second := sim.Tick([]Input{{MessageID: "hunger_2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	secondDamage, ok := uniqueEventDamage(second, "monster_damaged", "")
	if !ok {
		t.Fatalf("second hunger events = %+v, want damage", second.Events)
	}
	if stack := sim.uniqueHungerStacks[player.id]; stack.Stacks != 2 || stack.TargetID != primary.id {
		t.Fatalf("after second hit stack = %+v, want 2 stacks on primary", stack)
	}
	if secondDamage < firstDamage {
		t.Fatalf("second damage = %d, first = %d; want ramp or equal under integer rounding", secondDamage, firstDamage)
	}
	advanceBasicAttackCooldown(sim)
	other := sim.Tick([]Input{{MessageID: "hunger_reset", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(secondary.id)}}})
	otherDamage, ok := uniqueEventDamage(other, "monster_damaged", "")
	if !ok {
		t.Fatalf("target-change events = %+v, want damage", other.Events)
	}
	if stack := sim.uniqueHungerStacks[player.id]; stack.Stacks != 1 || stack.TargetID != secondary.id {
		t.Fatalf("after target change stack = %+v, want reset to 1 stack on secondary", stack)
	}
	if otherDamage > firstDamage+1 {
		t.Fatalf("target-change damage = %d, first = %d; want reset without stacked bonus", otherDamage, firstDamage)
	}
}
