package game

import "testing"

func TestOffensiveUniqueReplicatingBlightCopiesBurnDebuffOnly(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	rules.UniqueEffects[replicatingBlightEffectID].Params["replicate_radius_tiles"] = 5
	sim := MustNewSim("sess_replicating_blight", "replicating_blight", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	nearA := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	nearB := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X, Y: primary.pos.Y + 4.9}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	far := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 5.5, Y: primary.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[nearA.id] = nearA
	sim.entities[nearB.id] = nearB
	sim.entities[far.id] = far
	blade := addRolledInventoryItem(t, sim, 9807, "cave_blade", map[string]int{"damage_min": 10, "damage_max": 10})
	blade.rollPayload.EffectIDs = []string{everburningWoundEffectID, replicatingBlightEffectID}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_replicate", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "equip_replicate")

	attack := sim.Tick([]Input{{MessageID: "replicate_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(primary.id)}}})
	primaryDamage, ok := uniqueEventDamage(attack, "monster_damaged", "")
	if !ok || primaryDamage <= 0 {
		t.Fatalf("replicate events = %+v, want primary damage", attack.Events)
	}
	if nearA.hp != nearA.maxHP || nearB.hp != nearB.maxHP || far.hp != far.maxHP {
		t.Fatalf("replicate hit damaged secondaries: nearA=%d nearB=%d far=%d", nearA.hp, nearB.hp, far.hp)
	}
	if !containsStringValue(nearA.effectIDs, everburningWoundEffectID) || !containsStringValue(nearB.effectIDs, everburningWoundEffectID) {
		t.Fatalf("nearby effects = %v / %v, want replicated burn", nearA.effectIDs, nearB.effectIDs)
	}
	if containsStringValue(far.effectIDs, everburningWoundEffectID) {
		t.Fatalf("far effects = %v, want no replicated burn", far.effectIDs)
	}
	if _, ok := sim.uniqueBurnDots[uniqueBurnDotKey(everburningWoundEffectID, nearA.id)]; !ok {
		t.Fatalf("nearA missing replicated burn dot")
	}
	if _, ok := sim.uniqueBurnDots[uniqueBurnDotKey(everburningWoundEffectID, nearB.id)]; !ok {
		t.Fatalf("nearB missing replicated burn dot")
	}

	events := []Event{}
	for i := 0; i < 13; i++ {
		for _, result := range sim.TickResults(nil) {
			events = append(events, result.Events...)
		}
	}
	if !eventListHasTargetDamage(events, everburningWoundEffectID, nearA.id) || !eventListHasTargetDamage(events, everburningWoundEffectID, nearB.id) {
		t.Fatalf("burn tick events = %+v, want ticks on replicated targets", events)
	}
	if eventListHasTargetDamage(events, everburningWoundEffectID, far.id) {
		t.Fatalf("far target received burn tick: %+v", events)
	}
}

func TestOffensiveUniqueReplicatingBlightCopiesPinningRootDuration(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_replicating_root", "replicating_root", rules)
	forceUniqueTestHeroHitChance(sim)
	clearUniqueTestMonsters(sim)
	player := sim.entities[sim.playerID]
	primary := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	nearby := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: primary.pos.X + 1, Y: primary.pos.Y}, hp: 80, maxHP: 80, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[primary.id] = primary
	sim.entities[nearby.id] = nearby
	equipUniqueTestEffect(t, sim, replicatingBlightEffectID, 9808, "cave_blade", mainHandSlot)
	root := SkillRootDef{EffectID: "pinning_root", DurationTicks: 30}

	sim.applyMonsterRoot(primary, player.id, "pinning_shot", root, "pin_corr", &TickResult{})
	if !containsStringValue(primary.effectIDs, "pinning_root") || !containsStringValue(nearby.effectIDs, "pinning_root") {
		t.Fatalf("root effects primary=%v nearby=%v, want replicated pin", primary.effectIDs, nearby.effectIDs)
	}
	if !sim.monsterRooted(nearby) {
		t.Fatalf("nearby target is not rooted")
	}

	for i := 0; i < root.DurationTicks+1; i++ {
		sim.TickResults(nil)
	}
	if containsStringValue(primary.effectIDs, "pinning_root") || containsStringValue(nearby.effectIDs, "pinning_root") {
		t.Fatalf("root effects after expiry primary=%v nearby=%v, want expired together", primary.effectIDs, nearby.effectIDs)
	}
}

func eventListHasTargetDamage(events []Event, skillID string, targetID uint64) bool {
	for _, ev := range events {
		if ev.EventType == "monster_damaged" && ev.SkillID == skillID && ev.TargetEntityID == idStr(targetID) && ev.Damage != nil && *ev.Damage > 0 {
			return true
		}
	}
	return false
}
