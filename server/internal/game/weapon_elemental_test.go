package game

import "testing"

func TestElementalAffixRollsAreMutuallyExclusive(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rng := NewRNG(SeedToUint64("elemental_exclusive_roll"))
	template := rules.ItemTemplates["long_sword"]
	rarity := rules.Rarities["rare"]

	stats := cloneIntMap(template.BaseStats)
	rollableStats := rules.rollableStatsForRarity(template.RollableStats, "rare", 1)
	rollCount := rarity.StatRollsMax
	rollAffixStatsOntoMap(stats, rollableStats, rng, rollCount)

	elementalCount := 0
	for _, stat := range []string{"bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage"} {
		if stats[stat] > 0 {
			elementalCount++
		}
	}
	if elementalCount > 1 {
		t.Fatalf("stats = %+v, want at most one elemental affix", stats)
	}
}

func TestWeaponElementalDamageSplitsFromPhysical(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_weapon_elemental", "weapon_elemental_seed", rules)
	sim.progression.BaseStats.Dex = 20
	sim.progression.CharacterClass = "barbarian"
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 60, maxHP: 60, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	sword := addRolledInventoryItem(t, sim, 9001, "long_sword", map[string]int{
		"damage_min":        5,
		"damage_max":        8,
		"bonus_cold_damage": 4,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_cold", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "equip_cold")

	var res TickResult
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "cold_split", &res, DamageRange{Min: 10, Max: 10}, damageTypeForce, mainHandSlot)
	if !outcome.Hit || outcome.Damage <= 0 {
		t.Fatalf("physical outcome = %+v, want hit with damage", outcome)
	}
	if outcome.DamageType != damageTypeForce {
		t.Fatalf("physical damage type = %q, want force", outcome.DamageType)
	}

	forceDamage := 0
	coldDamage := 0
	for _, ev := range res.Events {
		if ev.EventType != "monster_damaged" || ev.TargetEntityID != idStr(target.id) {
			continue
		}
		if ev.DamageType == damageTypeForce && ev.Damage != nil {
			forceDamage = *ev.Damage
		}
		if ev.DamageType == damageTypeCold && ev.Damage != nil {
			coldDamage = *ev.Damage
		}
	}
	if forceDamage <= 0 {
		t.Fatalf("events = %+v, want force monster_damaged", res.Events)
	}
	if coldDamage != 4 {
		t.Fatalf("cold damage = %d, want 4", coldDamage)
	}
}

func TestWeaponElementalDamageRespectsResistance(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_weapon_elemental_resist", "weapon_elemental_resist_seed", rules)
	sim.progression.BaseStats.Dex = 20
	sim.progression.CharacterClass = "barbarian"
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 60, maxHP: 60, monsterDefID: "combat_lab_cold_resistant", lootTable: "no_drop"}
	sim.entities[target.id] = target

	sword := addRolledInventoryItem(t, sim, 9002, "long_sword", map[string]int{
		"damage_min":        4,
		"damage_max":        4,
		"bonus_cold_damage": 10,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_cold_resist", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "equip_cold_resist")

	var res TickResult
	sim.applyWeaponElementalDamageFromSlot(target, player.id, "cold_resist", mainHandSlot, 0, &res)

	coldDamage := 0
	for _, ev := range res.Events {
		if ev.EventType == "monster_damaged" && ev.DamageType == damageTypeCold && ev.Damage != nil {
			coldDamage = *ev.Damage
		}
	}
	if coldDamage != 5 {
		t.Fatalf("cold damage = %d, want 5 after 50%% resistance", coldDamage)
	}
}

func TestItemSummaryShowsElementalDamageLine(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_elemental_summary", "01", rules)
	lines := sim.itemSummaryLines("equipment", mainHandSlot, "one_handed", map[string]int{
		"damage_min":        4,
		"damage_max":        7,
		"bonus_cold_damage": 4,
	}, nil, nil, "long_sword")
	found := false
	for _, line := range lines {
		if line == "+4 Cold Damage" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("summary lines = %v, want +4 Cold Damage", lines)
	}
}
