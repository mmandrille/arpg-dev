package game

import "testing"

func TestRoguePoisonStabAppliesDamageOverTime(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 20)

	cast := sim.Tick([]Input{{
		MessageID:     "poison",
		CorrelationID: "corr_poison",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "poison_stab", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "poison")
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_effect_started") || !hasEvent(cast, "monster_damaged") {
		t.Fatalf("poison cast events = %+v", cast.Events)
	}
	hpAfterHit := target.hp
	var tick TickResult
	for i := 0; i < 12; i++ {
		tick = sim.Tick(nil)
		if hasPoisonDamageEvent(tick, "poison_stab") {
			break
		}
	}
	if !hasPoisonDamageEvent(tick, "poison_stab") {
		t.Fatalf("missing poison tick; last events=%+v", tick.Events)
	}
	if target.hp >= hpAfterHit {
		t.Fatalf("poison target hp = %d, want below %d", target.hp, hpAfterHit)
	}
}

func TestRoguePoisonStabPoisonsImmuneUndeadForZeroDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 20)
	target.monsterDefID = "dungeon_undead"

	cast := sim.Tick([]Input{{
		MessageID:     "poison_undead",
		CorrelationID: "corr_poison_undead",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "poison_stab", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "poison_undead")
	if !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("poison immune cast events = %+v, want skill_effect_started", cast.Events)
	}
	if !hasZeroDamageTypeEvent(cast, damageTypePoison) {
		t.Fatalf("poison immune cast events = %+v, want zero poison hit", cast.Events)
	}
	hpAfterHit := target.hp
	var tick TickResult
	for i := 0; i < 12; i++ {
		tick = sim.Tick(nil)
		if hasPoisonZeroDamageEvent(tick, "poison_stab") {
			break
		}
	}
	if !hasPoisonZeroDamageEvent(tick, "poison_stab") {
		t.Fatalf("missing zero poison tick; last events=%+v", tick.Events)
	}
	if target.hp != hpAfterHit {
		t.Fatalf("poison immune target hp = %d, want unchanged %d", target.hp, hpAfterHit)
	}
}

func TestRogueDashMovesThroughAndDamagesTarget(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 20)
	startX := player.pos.X

	cast := sim.Tick([]Input{{
		MessageID:     "dash",
		CorrelationID: "corr_dash",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "dash", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "dash")
	if !hasEvent(cast, "skill_cast") || !hasSkillDamageEvent(cast, "dash") {
		t.Fatalf("dash events = %+v", cast.Events)
	}
	if player.pos.X <= target.pos.X {
		t.Fatalf("dash player x = %.2f, want past target %.2f", player.pos.X, target.pos.X)
	}
	if player.pos.X <= startX {
		t.Fatalf("dash player x = %.2f, want moved from %.2f", player.pos.X, startX)
	}
	if target.hp >= 20 {
		t.Fatalf("dash target hp = %d, want damaged", target.hp)
	}
}

func TestRogueOffHandBasicAttackCanFireBetweenMainHandAttacks(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 50)

	main := sim.Tick([]Input{{MessageID: "main_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
	assertAck(t, main, "main_1")
	if !hasWeaponSlotDamageEvent(main, mainHandSlot) {
		t.Fatalf("first attack events = %+v, want main hand", main.Events)
	}
	off := sim.Tick([]Input{{MessageID: "off_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
	assertAck(t, off, "off_1")
	if !hasWeaponSlotCombatEvent(off, offHandSlot) {
		t.Fatalf("second attack events = %+v, want off hand", off.Events)
	}
	var secondMain TickResult
	for i := 0; i < 30; i++ {
		secondMain = sim.Tick([]Input{{MessageID: "main_2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
		if hasWeaponSlotCombatEvent(secondMain, mainHandSlot) {
			break
		}
	}
	if !hasWeaponSlotCombatEvent(secondMain, mainHandSlot) {
		t.Fatalf("missing second main-hand hit; last events=%+v rejects=%+v", secondMain.Events, secondMain.Rejects)
	}
}

func newRogueSkillTestSim(t *testing.T, rules *Rules) *Sim {
	t.Helper()
	sim := MustNewSim("sess_rogue_skills", "rogue_skills_seed", rules)
	sim.progression.CharacterClass = "rogue"
	sim.progression.BaseStats.Dex = 8
	sim.progression.BaseStats.Magic = 4
	sim.progression.SkillRanks["poison_stab"] = 1
	sim.progression.SkillRanks["dash"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.mana = player.maxMana
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	main := addRolledInventoryItem(t, sim, 9101, "starter_rogue_sword", nil)
	off := addRolledInventoryItem(t, sim, 9102, "starter_rogue_sword", nil)
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_main", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(main.instanceID), Slot: mainHandSlot}}}), "equip_main")
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_off", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(off.instanceID), Slot: offHandSlot}}}), "equip_off")
	player.mana = player.maxMana
	return sim
}

func addRogueSkillTarget(sim *Sim, pos Vec2, hp int) *entity {
	target := &entity{id: sim.alloc(), kind: monsterEntity, pos: pos, hp: hp, maxHP: hp, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[target.id] = target
	return target
}

func hasPoisonDamageEvent(r TickResult, skillID string) bool {
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.SkillID == skillID && ev.Damage != nil && *ev.Damage > 0 {
			return true
		}
	}
	return false
}

func hasPoisonZeroDamageEvent(r TickResult, skillID string) bool {
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.SkillID == skillID && ev.DamageType == damageTypePoison && ev.Damage != nil && *ev.Damage == 0 {
			return true
		}
	}
	return false
}

func hasZeroDamageTypeEvent(r TickResult, damageType string) bool {
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.DamageType == damageType && ev.Damage != nil && *ev.Damage == 0 {
			return true
		}
	}
	return false
}

func hasSkillDamageEvent(r TickResult, skillID string) bool {
	return hasPoisonDamageEvent(r, skillID)
}

func hasWeaponSlotDamageEvent(r TickResult, slot string) bool {
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.WeaponSlot == slot && ev.Damage != nil && *ev.Damage > 0 {
			return true
		}
	}
	return false
}

func hasWeaponSlotCombatEvent(r TickResult, slot string) bool {
	for _, ev := range r.Events {
		if (ev.EventType == "monster_damaged" || ev.EventType == "attack_missed" || ev.EventType == "attack_blocked") && ev.WeaponSlot == slot {
			return true
		}
	}
	return false
}
