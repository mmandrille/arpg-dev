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
	blade := addRolledInventoryItem(t, sim, 9801, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
	blade := addRolledInventoryItem(t, sim, 9802, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
	blade := addRolledInventoryItem(t, sim, 9803, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
	blade := addRolledInventoryItem(t, sim, 9804, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
	blade := addRolledInventoryItem(t, sim, 9805, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
	blade := addRolledInventoryItem(t, sim, 9806, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
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
