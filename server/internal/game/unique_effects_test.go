package game

import "testing"

func TestUniqueBurnAppliesFromEquippedEffectAndTicksFromOriginalDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_unique_burn", "unique_burn", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[target.id] = target
	blade := addRolledInventoryItem(t, sim, 9801, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{everburningWoundEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_burn", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_burn")

	var attack TickResult
	for i := 0; i < 80; i++ {
		attack = sim.Tick([]Input{{MessageID: "burn_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
		if _, ok := uniqueEventDamage(attack, "monster_damaged", ""); ok {
			break
		}
	}
	hitDamage, ok := uniqueEventDamage(attack, "monster_damaged", "")
	if !ok || hitDamage <= 0 {
		t.Fatalf("attack events = %+v, want positive damage", attack.Events)
	}
	startDamage, ok := eventAmount(attack, "skill_effect_started", everburningWoundEffectID)
	if !ok {
		t.Fatalf("attack events = %+v, want burn start", attack.Events)
	}
	wantBurnDamage := hitDamage / 10
	if wantBurnDamage < 1 {
		wantBurnDamage = 1
	}
	if startDamage != wantBurnDamage {
		t.Fatalf("burn amount = %d, want 10%% of hit %d => %d", startDamage, hitDamage, wantBurnDamage)
	}
	if !sameStringSlice(target.effectIDs, []string{everburningWoundEffectID}) {
		t.Fatalf("target effect ids = %v, want burn", target.effectIDs)
	}

	tickEvents := []Event{}
	for i := 0; i < 13; i++ {
		for _, result := range sim.TickResults(nil) {
			tickEvents = append(tickEvents, result.Events...)
		}
	}
	tickDamage, ok := uniqueEventListDamage(tickEvents, "monster_damaged", everburningWoundEffectID)
	if !ok {
		t.Fatalf("burn tick events = %+v, want damage", tickEvents)
	}
	if tickDamage != wantBurnDamage {
		t.Fatalf("burn tick damage = %d, want %d", tickDamage, wantBurnDamage)
	}
	if !eventListHasDamageType(tickEvents, "monster_damaged", everburningWoundEffectID, damageTypeFire) {
		t.Fatalf("burn tick events = %+v, want fire damage", tickEvents)
	}
}

func TestUniqueBurnCanKillThroughExistingMonsterKillPath(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_unique_burn_kill", "unique_burn_kill", rules)
	player := sim.entities[sim.playerID]
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 1, maxHP: 1, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[target.id] = target

	sim.startUniqueBurnDot(player.id, target, rules.UniqueEffects[everburningWoundEffectID], 10, "burn_kill", &TickResult{})
	sim.savePlayer(sim.defaultPlayer())
	events := []Event{}
	for i := 0; i < 12; i++ {
		for _, result := range sim.TickResults(nil) {
			events = append(events, result.Events...)
		}
	}
	if !eventListHas(events, "monster_killed") {
		t.Fatalf("burn kill events = %+v, want monster_killed", events)
	}
	if target.hp != 0 {
		t.Fatalf("target hp = %d, want dead", target.hp)
	}
}

func TestOffensiveUniqueStormboundEchoChainsFromBasicAttack(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	rules.UniqueEffects[stormboundEchoEffectID].Params["trigger_chance_percent"] = 100
	sim := MustNewSim("sess_stormbound_echo", "stormbound_echo", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	secondary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[secondary.id] = secondary
	blade := addRolledInventoryItem(t, sim, 9802, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{stormboundEchoEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_storm", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_storm")

	attack := sim.Tick([]Input{{MessageID: "storm_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	if !eventListHasDamageType(attack.Events, "monster_damaged", stormboundEchoEffectID, damageTypeLightning) {
		t.Fatalf("storm events = %+v, want lightning secondary damage", attack.Events)
	}
	damage, ok := uniqueEventListDamage(attack.Events, "monster_damaged", stormboundEchoEffectID)
	if !ok || damage <= 0 {
		t.Fatalf("storm events = %+v, want positive storm damage", attack.Events)
	}
	if secondary.hp >= secondary.maxHP {
		t.Fatalf("secondary hp = %d, want damaged; events=%+v", secondary.hp, attack.Events)
	}
}

func TestOffensiveUniqueStormboundEchoDoesNotChainFromSkillDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.UniqueEffects[stormboundEchoEffectID].Params["trigger_chance_percent"] = 100
	sim := MustNewSim("sess_stormbound_skill", "stormbound_skill", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	secondary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[secondary.id] = secondary
	blade := addRolledInventoryItem(t, sim, 9803, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{stormboundEchoEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_storm", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_storm")

	res := &TickResult{}
	sim.damageMonsterByPlayerSkillTyped(primary, player.id, "skill_storm", res, DamageRange{Min: 10, Max: 10}, damageTypeForce)
	if eventListHasDamageType(res.Events, "monster_damaged", stormboundEchoEffectID, damageTypeLightning) {
		t.Fatalf("skill events = %+v, want no stormbound echo", res.Events)
	}
	if secondary.hp != secondary.maxHP {
		t.Fatalf("secondary hp = %d, want untouched", secondary.hp)
	}
}

func TestOffensiveUniqueExecutionersMarkPulsesOnMarkedKill(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_executioners_mark", "executioners_mark", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 15, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	secondary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, hp: 50, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[secondary.id] = secondary
	blade := addRolledInventoryItem(t, sim, 9804, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{executionersMarkEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_mark", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_mark")

	first := sim.Tick([]Input{{MessageID: "mark_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	if _, ok := eventAmount(first, "skill_effect_started", executionersMarkEffectID); !ok {
		t.Fatalf("mark events = %+v, want mark start", first.Events)
	}
	advanceBasicAttackCooldown(sim)
	second := sim.Tick([]Input{{MessageID: "mark_kill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	damage, ok := uniqueEventListDamage(second.Events, "monster_damaged", executionersMarkEffectID)
	if !ok || damage <= 0 {
		t.Fatalf("kill events = %+v, want execution pulse damage", second.Events)
	}
	if secondary.hp >= secondary.maxHP {
		t.Fatalf("secondary hp = %d, want pulse damage", secondary.hp)
	}
}

func TestOffensiveUniqueExecutionersMarkExpires(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	rules.UniqueEffects[executionersMarkEffectID].Params["mark_duration_seconds"] = 1
	sim := MustNewSim("sess_executioners_mark_expire", "executioners_mark_expire", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 15, maxHP: 50, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	blade := addRolledInventoryItem(t, sim, 9805, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{executionersMarkEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_mark", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_mark")

	sim.Tick([]Input{{MessageID: "mark_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	events := []Event{}
	for i := 0; i < 12; i++ {
		for _, result := range sim.TickResults(nil) {
			events = append(events, result.Events...)
		}
	}
	if !eventListHas(events, "skill_effect_ended") {
		t.Fatalf("expire events = %+v, want mark ended", events)
	}
	if len(primary.effectIDs) != 0 {
		t.Fatalf("primary effects = %v, want expired", primary.effectIDs)
	}
}

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
	blade := addRolledInventoryItem(t, sim, 9806, "long_sword", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{hungerOfTheDeepEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_hunger", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_hunger")

	first := sim.Tick([]Input{{MessageID: "hunger_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	firstDamage, ok := uniqueEventDamage(first, "monster_damaged", "")
	if !ok {
		t.Fatalf("first hunger events = %+v, want damage", first.Events)
	}
	advanceBasicAttackCooldown(sim)
	second := sim.Tick([]Input{{MessageID: "hunger_2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	secondDamage, ok := uniqueEventDamage(second, "monster_damaged", "")
	if !ok || secondDamage <= firstDamage {
		t.Fatalf("second damage = %d, first = %d, events=%+v; want ramp", secondDamage, firstDamage, second.Events)
	}
	advanceBasicAttackCooldown(sim)
	other := sim.Tick([]Input{{MessageID: "hunger_reset", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(secondary.id)}}})
	otherDamage, ok := uniqueEventDamage(other, "monster_damaged", "")
	if !ok || otherDamage != firstDamage {
		t.Fatalf("target-change damage = %d, first = %d, events=%+v; want reset", otherDamage, firstDamage, other.Events)
	}
}

func TestSurvivalUniqueVeilOfTheLastOathPreventsLethalHitOnce(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceUniqueTestMonsterHitChance(rules, 1)
	sim := MustNewSim("sess_veil_oath", "veil_oath", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	monster := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 50)
	equipUniqueTestEffect(t, sim, veilOfTheLastOathEffectID, 9901, "shield", offHandSlot)
	player.hp = 5
	player.maxHP = 20

	first := &TickResult{}
	outcome := sim.damagePlayerByMonster(monster, player, DamageRange{Min: 10, Max: 10}, "veil", first)
	if outcome.Damage != 0 || player.hp != 5 {
		t.Fatalf("first lethal outcome=%+v hp=%d, want prevented at 5", outcome, player.hp)
	}
	if !sameStringSlice(player.effectIDs, []string{"last_oath_veil"}) {
		t.Fatalf("player effects = %v, want last_oath_veil", player.effectIDs)
	}
	if _, active := sim.skillCooldownRemaining(veilOfTheLastOathEffectID); !active {
		t.Fatalf("veil cooldown missing after trigger; changes=%+v events=%+v", first.Changes, first.Events)
	}

	second := &TickResult{}
	outcome = sim.damagePlayerByMonster(monster, player, DamageRange{Min: 10, Max: 10}, "veil_cooldown", second)
	if outcome.Damage <= 0 || player.hp != 0 {
		t.Fatalf("second lethal outcome=%+v hp=%d, want death during cooldown", outcome, player.hp)
	}
}

func TestSurvivalUniqueFrostglassWardSlowsAndBuffsAfterLargeHit(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceUniqueTestMonsterHitChance(rules, 1)
	sim := MustNewSim("sess_frostglass", "frostglass", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	attacker := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 50)
	nearby := uniqueTestMonster(sim, Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 50)
	equipUniqueTestEffect(t, sim, frostglassWardEffectID, 9902, "shield", offHandSlot)
	player.hp = 100
	player.maxHP = 100

	res := &TickResult{}
	sim.damagePlayerByMonster(attacker, player, DamageRange{Min: 30, Max: 30}, "frost", res)
	if !sameStringSlice(player.effectIDs, []string{frostglassWardEffectID}) {
		t.Fatalf("player effects = %v, want frostglass ward armor marker", player.effectIDs)
	}
	if !sameStringSlice(nearby.effectIDs, []string{"ice_slow"}) {
		t.Fatalf("nearby effects = %v, want ice slow", nearby.effectIDs)
	}
	if _, active := sim.skillCooldownRemaining(frostglassWardEffectID); !active {
		t.Fatalf("frostglass cooldown missing; events=%+v", res.Events)
	}
	if _, ok := eventAmount(TickResult{Events: res.Events}, "skill_effect_started", frostglassWardEffectID); !ok {
		t.Fatalf("frostglass events = %+v, want skill_effect_started", res.Events)
	}
}

func TestSurvivalUniqueMirrorsteelSkinReducesProjectileAndReflects(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceUniqueTestMonsterHitChance(rules, 1)
	sim := MustNewSim("sess_mirrorsteel", "mirrorsteel", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	attacker := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 50)
	equipUniqueTestEffect(t, sim, mirrorsteelSkinEffectID, 9903, "shield", offHandSlot)
	player.hp = 100
	player.maxHP = 100

	res := &TickResult{}
	outcome := sim.damagePlayerByMonsterWithSource(attacker, player, DamageRange{Min: 10, Max: 10}, "mirror", res, uniqueIncomingDamageSource{Projectile: true})
	if outcome.Damage != 2 || player.hp != 98 {
		t.Fatalf("mirror outcome=%+v hp=%d, want 70%% reduction after armor mitigation", outcome, player.hp)
	}
	if attacker.hp != 48 {
		t.Fatalf("attacker hp=%d, want reflected damage after mitigation; events=%+v", attacker.hp, res.Events)
	}
	if _, active := sim.skillCooldownRemaining(mirrorsteelSkinEffectID); !active {
		t.Fatalf("mirrorsteel cooldown missing; events=%+v", res.Events)
	}
}

func TestSurvivalUniqueAshenReprisalPrimesAndConsumesOnNextHit(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceUniqueTestMonsterHitChance(rules, 0)
	def := rules.Monsters[monsterDefID]
	def.RetaliationDamage = nil
	rules.Monsters[monsterDefID] = def
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_ashen", "ashen", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	attacker := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 50)
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 200)
	blade := equipUniqueTestEffect(t, sim, ashenReprisalEffectID, 9904, "long_sword", mainHandSlot)
	blade.rollPayload.Stats["damage_min"] = 10
	blade.rollPayload.Stats["damage_max"] = 10

	avoid := &TickResult{}
	sim.damagePlayerByMonster(attacker, player, DamageRange{Min: 10, Max: 10}, "ashen_prime", avoid)
	if _, ok := sim.uniqueAshenReprisals[player.id]; !ok {
		t.Fatalf("ashen was not primed; events=%+v", avoid.Events)
	}

	hit := &TickResult{}
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "ashen_hit", hit, DamageRange{Min: 10, Max: 10}, damageTypeForce, mainHandSlot)
	if _, ok := sim.uniqueAshenReprisals[player.id]; ok {
		t.Fatalf("ashen remained primed after hit outcome=%+v effects=%v events=%+v equipped=%v", outcome, sim.equippedUniqueEffectIDs(player.id), hit.Events, sim.equipped)
	}
	if !eventListHasDamageType(hit.Events, "monster_damaged", ashenReprisalEffectID, damageTypeFire) {
		t.Fatalf("ashen hit events = %+v, want fire bonus damage", hit.Events)
	}
	if !sameStringSlice(target.effectIDs, []string{"burning"}) {
		t.Fatalf("target effects = %v, want burning", target.effectIDs)
	}
}

func TestResourceUniqueGravePactHealsLowHealthKiller(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_grave_pact", "grave_pact", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	player.hp = 10
	player.maxHP = 100
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 1)
	equipUniqueTestEffect(t, sim, gravePactEffectID, 9910, "ring", ringLeftSlot)
	player.hp = 10
	player.maxHP = 100

	res := &TickResult{}
	sim.finishMonsterKill(target, player.id, "grave", res)
	if player.hp != 18 {
		t.Fatalf("player hp = %d, want 8%% max hp heal to 18; events=%+v", player.hp, res.Events)
	}
	if !eventListHas(res.Events, "player_healed") {
		t.Fatalf("grave pact events = %+v, want player_healed", res.Events)
	}
}

func TestResourceUniqueBloodPricePaysMissingManaWithHP(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_blood_price", "blood_price", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 50)
	equipUniqueTestEffect(t, sim, bloodPriceEffectID, 9911, "belt", "belt")
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 10
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	player.mana = 0
	player.hp = 20
	player.maxHP = 50
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{MessageID: "blood_cast", Type: "cast_skill_intent", CastSkill: &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(target.id)}}})
	assertAck(t, cast, "blood_cast")
	if player.hp >= 20 || player.mana != 0 {
		t.Fatalf("player hp/mana = %d/%d, want hp paid and mana clamped; events=%+v", player.hp, player.mana, cast.Events)
	}
	if !eventListHas(cast.Events, "skill_cast") || !eventListHas(cast.Events, "skill_effect_started") {
		t.Fatalf("blood price cast events = %+v, want skill cast and blood price marker", cast.Events)
	}
}

func TestNamedUniqueBloodboundSigilPaysMissingManaWithHP(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_bloodbound_sigil", "bloodbound_sigil", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 50)
	payload, ok := rules.namedUniquePayload("bloodbound_sigil")
	if !ok {
		t.Fatal("missing bloodbound_sigil named payload")
	}
	sim.progression.Level = 5
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 10
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	item := &invItem{
		instanceID:  9916,
		itemDefID:   payload.ItemTemplateID,
		slot:        sim.itemSlot(payload.ItemTemplateID, &payload),
		rollPayload: cloneRollPayload(&payload),
	}
	addTestInventoryItem(sim, item)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_bloodbound", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: ringLeftSlot}}}), "equip_bloodbound")
	player.mana = 0
	player.hp = 20
	player.maxHP = 50
	manaCost := sim.effectiveSkillManaCost(rules.Skills[magicBoltSkillID], 1)
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{MessageID: "bloodbound_cast", Type: "cast_skill_intent", CastSkill: &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(target.id)}}})
	assertAck(t, cast, "bloodbound_cast")
	if player.hp != 20-manaCost || player.mana != 0 {
		t.Fatalf("player hp/mana = %d/%d, want hp paid by named unique cost %d and mana clamped; events=%+v", player.hp, player.mana, manaCost, cast.Events)
	}
	if !eventListHas(cast.Events, "skill_cast") || !eventListHas(cast.Events, "skill_effect_started") {
		t.Fatalf("bloodbound cast events = %+v, want skill cast and blood price marker", cast.Events)
	}
}

func TestUniqueSkillModifierBoostsConfiguredSkill(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_arcane_conduit", "arcane_conduit", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 100)
	equipUniqueTestEffect(t, sim, arcaneConduitEffectID, 9914, "amulet", "amulet")

	res := &TickResult{}
	outcome := sim.damageMonsterByPlayerSkillTypedWithID(target, player.id, magicBoltSkillID, "arcane_conduit", res, DamageRange{Min: 10, Max: 10}, damageTypeForce)
	if outcome.Damage != 15 {
		t.Fatalf("magic bolt unique-modified damage = %d, want 15; events=%+v", outcome.Damage, res.Events)
	}
}

func TestUniqueSkillModifierDoesNotBoostOtherSkills(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_arcane_conduit_other", "arcane_conduit_other", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 100)
	equipUniqueTestEffect(t, sim, arcaneConduitEffectID, 9915, "amulet", "amulet")

	res := &TickResult{}
	outcome := sim.damageMonsterByPlayerSkillTypedWithID(target, player.id, "ice_shard", "arcane_conduit_other", res, DamageRange{Min: 10, Max: 10}, damageTypeForce)
	if outcome.Damage != 10 {
		t.Fatalf("non-target skill damage = %d, want 10; events=%+v", outcome.Damage, res.Events)
	}
}

func TestResourceUniquePilgrimsMomentumChargesAndKnocksBack(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := MustNewSim("sess_pilgrim", "pilgrim", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 200)
	equipUniqueTestEffect(t, sim, pilgrimsMomentumEffectID, 9912, "boots", "boots")
	for i := 0; i < 20; i++ {
		sim.updatePilgrimMomentumMovement(player, true, &TickResult{})
	}
	before := target.pos

	res := &TickResult{}
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "pilgrim_hit", res, DamageRange{Min: 10, Max: 10}, damageTypeForce, mainHandSlot)
	if outcome.Damage <= 10 {
		t.Fatalf("pilgrim damage = %d, want bonus over 10; events=%+v", outcome.Damage, res.Events)
	}
	if _, ok := sim.uniquePilgrimMomentum[player.id]; ok {
		t.Fatalf("pilgrim charge not consumed")
	}
	if target.pos == before {
		t.Fatalf("target position = %+v, want knockback from %+v", target.pos, before)
	}
}

func TestResourceUniqueLanternOfTheFallenHealsLowestNearbyHero(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_lantern", "lantern", rules)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	player.hp = 20
	player.maxHP = 100
	target := uniqueTestMonster(sim, Vec2{X: player.pos.X + 1, Y: player.pos.Y}, 1)
	equipUniqueTestEffect(t, sim, lanternOfTheFallenEffectID, 9913, "amulet", "amulet")
	player.hp = 20
	player.maxHP = 100

	res := &TickResult{}
	sim.finishMonsterKill(target, player.id, "lantern", res)
	if player.hp != 26 {
		t.Fatalf("player hp = %d, want lantern 6%% max hp heal to 26; events=%+v", player.hp, res.Events)
	}
	if !eventListHas(res.Events, "player_healed") || !eventListHas(res.Events, "skill_effect_started") {
		t.Fatalf("lantern events = %+v, want heal and wisp marker", res.Events)
	}
}

func uniqueEventDamage(r TickResult, eventType string, skillID string) (int, bool) {
	return uniqueEventListDamage(r.Events, eventType, skillID)
}

func forceUniqueTestHeroHitChance(sim *Sim) {
	minOne := 1.0
	maxOne := 1.0
	maxZero := 0.0
	sim.rules.CharacterProgression.DerivedStats["hit_chance"] = LinearStatFormula{Type: "linear", Base: 1, Min: &minOne, Max: &maxOne}
	sim.rules.CharacterProgression.DerivedStats["crit_chance"] = LinearStatFormula{Type: "linear", Base: 0, Min: &maxZero, Max: &maxZero}
}

func clearUniqueTestMonsters(sim *Sim) {
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
}

func forceUniqueTestMonsterHitChance(rules *Rules, chance float64) {
	def := rules.Monsters[monsterDefID]
	def.HitChance = &chance
	def.CritChance = floatPtr(0)
	rules.Monsters[monsterDefID] = def
}

func uniqueTestMonster(sim *Sim, pos Vec2, hp int) *entity {
	monster := &entity{id: sim.alloc(), kind: monsterEntity, pos: pos, hp: hp, maxHP: hp, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[monster.id] = monster
	return monster
}

func equipUniqueTestEffect(t *testing.T, sim *Sim, effectID string, instanceID uint64, templateID string, slot string) *invItem {
	t.Helper()
	item := addRolledInventoryItem(t, sim, instanceID, templateID, map[string]int{})
	item.rollPayload.EffectIDs = []string{effectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_" + effectID, Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: slot}}}), "equip_"+effectID)
	return item
}

func uniqueEventListDamage(events []Event, eventType string, skillID string) (int, bool) {
	for _, ev := range events {
		if ev.EventType != eventType || ev.Damage == nil {
			continue
		}
		if skillID != "" && ev.SkillID != skillID {
			continue
		}
		return *ev.Damage, true
	}
	return 0, false
}

func eventAmount(r TickResult, eventType string, skillID string) (int, bool) {
	for _, ev := range r.Events {
		if ev.EventType == eventType && ev.SkillID == skillID && ev.Amount != nil {
			return *ev.Amount, true
		}
	}
	return 0, false
}

func eventHasDamageType(r TickResult, eventType string, skillID string, damageType string) bool {
	return eventListHasDamageType(r.Events, eventType, skillID, damageType)
}

func eventListHasDamageType(events []Event, eventType string, skillID string, damageType string) bool {
	for _, ev := range events {
		if ev.EventType == eventType && ev.SkillID == skillID && ev.DamageType == damageType {
			return true
		}
	}
	return false
}
