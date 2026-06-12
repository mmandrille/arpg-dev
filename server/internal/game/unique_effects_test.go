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
	for i := 0; i < 12; i++ {
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

func uniqueEventDamage(r TickResult, eventType string, skillID string) (int, bool) {
	return uniqueEventListDamage(r.Events, eventType, skillID)
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
