package game

import "testing"

func weaponElementalProcRules(t *testing.T) *Rules {
	t.Helper()
	rules := cloneRules(loadRules(t))
	rules.MainConfig.WeaponElementalProcs.Cold.ProcChancePercent = 100
	rules.MainConfig.WeaponElementalProcs.Fire.ProcChancePercent = 100
	rules.MainConfig.WeaponElementalProcs.Lightning.ProcChancePercent = 100
	rules.MainConfig.WeaponElementalProcs.Poison.ProcChancePercent = 100

	return rules
}

func TestWeaponElementalProcFreezeAppliesSlow(t *testing.T) {
	rules := weaponElementalProcRules(t)
	sim := MustNewSim("sess_weapon_proc_freeze", "weapon_proc_freeze_seed", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: player.pos, hp: 40, maxHP: 40, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	var res TickResult
	sim.tryWeaponElementalProcs(target, player.id, "freeze", damageTypeCold, 4, 10, &res)

	if !containsStringValue(target.effectIDs, "weapon_freeze") {
		t.Fatalf("effect ids = %v, want weapon_freeze", target.effectIDs)
	}
	if sim.monsterMoveSpeed(target, rules.Monsters["combat_lab_soft_target"], rules.Navigation) >= rules.Monsters["combat_lab_soft_target"].effectiveMoveSpeed(rules.Navigation) {
		t.Fatalf("monster move speed = %v, want slowed", sim.monsterMoveSpeed(target, rules.Monsters["combat_lab_soft_target"], rules.Navigation))
	}
}

func TestWeaponElementalProcBurnStartsDot(t *testing.T) {
	rules := weaponElementalProcRules(t)
	sim := MustNewSim("sess_weapon_proc_burn", "weapon_proc_burn_seed", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: player.pos, hp: 40, maxHP: 40, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	var res TickResult
	sim.tryWeaponElementalProcs(target, player.id, "burn", damageTypeFire, 4, 20, &res)

	if _, ok := sim.uniqueBurnDots[uniqueBurnDotKey("weapon_burn", target.id)]; !ok {
		t.Fatalf("burn dot missing for target %d", target.id)
	}
}

func TestWeaponElementalProcStunRootsTarget(t *testing.T) {
	rules := weaponElementalProcRules(t)
	sim := MustNewSim("sess_weapon_proc_stun", "weapon_proc_stun_seed", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: player.pos, hp: 40, maxHP: 40, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	var res TickResult
	sim.tryWeaponElementalProcs(target, player.id, "stun", damageTypeLightning, 4, 10, &res)

	if !sim.monsterRooted(target) {
		t.Fatalf("target should be stunned/rooted")
	}
}

func TestWeaponElementalProcPoisonRefreshesDot(t *testing.T) {
	rules := weaponElementalProcRules(t)
	sim := MustNewSim("sess_weapon_proc_poison", "weapon_proc_poison_seed", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: player.pos, hp: 40, maxHP: 40, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	var res TickResult
	sim.tryWeaponElementalProcs(target, player.id, "poison1", damageTypePoison, 8, 8, &res)
	sim.tryWeaponElementalProcs(target, player.id, "poison2", damageTypePoison, 4, 4, &res)
	second := sim.poisonDots[target.id]
	if second.DamagePerTick != 1 {
		t.Fatalf("refreshed poison tick damage = %d, want 1 (25%% of 4)", second.DamagePerTick)
	}
}

func TestWeaponElementalProcOnBasicAttackHit(t *testing.T) {
	rules := weaponElementalProcRules(t)
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_weapon_proc_attack", "weapon_elemental_seed", rules)
	sim.progression.BaseStats.Dex = 20
	sim.progression.CharacterClass = "barbarian"
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 150, maxHP: 150, monsterDefID: "combat_lab_soft_target", lootTable: "no_drop"}
	sim.entities[target.id] = target

	sword := addRolledInventoryItem(t, sim, 9010, "long_sword", map[string]int{
		"damage_min":        5,
		"damage_max":        8,
		"bonus_cold_damage": 4,
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_proc", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "equip_proc")

	var res TickResult
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "proc_hit", &res, DamageRange{Min: 10, Max: 10}, damageTypeForce, mainHandSlot)
	if !outcome.Hit || outcome.Damage <= 0 {
		t.Fatalf("physical outcome = %+v, want hit", outcome)
	}
	freezeStarted := false
	for _, ev := range res.Events {
		if ev.EventType == "skill_effect_started" && ev.SkillID == weaponElementalFreezeSkillID {
			freezeStarted = true
		}
	}
	if !freezeStarted {
		t.Fatalf("events = %+v, want weapon freeze proc", res.Events)
	}
}
