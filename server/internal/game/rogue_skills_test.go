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
	if !containsStringValue(target.effectIDs, "dash_bleed") {
		t.Fatalf("dash target effects = %+v, want dash_bleed", target.effectIDs)
	}
	if !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("dash events = %+v, want bleed skill_effect_started", cast.Events)
	}
}

func TestRogueDashBleedTicksPercentMaxHPAndRefreshes(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 100)
	target.maxHP = 100

	cast := sim.Tick([]Input{{
		MessageID:     "dash_bleed",
		CorrelationID: "corr_dash_bleed",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "dash", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "dash_bleed")
	dot, ok := sim.bleedDots[target.id]
	if !ok || dot.RemainingTicks != 50 {
		t.Fatalf("bleed dot = %+v ok=%v, want 50 remaining ticks", dot, ok)
	}

	sim.tick += 10
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceBleedDots(&res)
	if damage := eventDamage(res, "monster_damaged"); damage != 5 {
		t.Fatalf("first bleed tick damage = %d, want 5 (5%% of 100 max hp)", damage)
	}

	dot.RemainingTicks = 10
	sim.bleedDots[target.id] = dot
	refresh := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.startBleedDot(player, target, "dash", rules.Skills["dash"].Dash, "corr_dash_refresh", &refresh)
	refreshed, ok := sim.bleedDots[target.id]
	if !ok || refreshed.RemainingTicks != 50 {
		t.Fatalf("refreshed bleed dot = %+v ok=%v, want replenished 50 ticks", refreshed, ok)
	}
}

func TestRoguePoisonStabMarkIncreasesAllPlayerDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	poison := rules.Skills["poison_stab"]
	poison.Poison.MarkDamageBonusPercent = 100
	poison.Poison.MarkDurationTicks = 40
	poison.Poison.MarkEffectID = "test_rogue_mark"
	rules.Skills["poison_stab"] = poison
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 100)

	cast := sim.Tick([]Input{{
		MessageID:     "mark",
		CorrelationID: "corr_mark",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "poison_stab", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "mark")
	if _, ok := sim.rogueMarks[target.id]; !ok || !containsStringValue(target.effectIDs, "test_rogue_mark") {
		t.Fatalf("mark state/effects = %+v / %+v, want active mark", sim.rogueMarks, target.effectIDs)
	}

	var hit TickResult
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "marked_hit", &hit, DamageRange{Min: 4, Max: 4}, damageTypeForce, mainHandSlot)
	if !outcome.Hit || outcome.Damage < 8 {
		t.Fatalf("marked basic outcome=%+v events=%+v, want doubled damage", outcome, hit.Events)
	}
}

func TestRogueMarkIncreasesPoisonTickDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseCritChance = 0
	sim := newRogueSkillTestSim(t, rules)
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 100)
	sim.rogueMarks[target.id] = rogueMarkState{
		SourcePlayerID:     player.id,
		TargetID:           target.id,
		SkillID:            "poison_stab",
		Rank:               1,
		DamageBonusPercent: 100,
		EndsTick:           sim.tick + 20,
		TotalTicks:         20,
		EffectID:           "test_rogue_mark",
		CorrelationID:      "corr_poison_mark",
	}
	sim.poisonDots[target.id] = poisonDotState{
		SourcePlayerID: player.id,
		TargetID:       target.id,
		SkillID:        "poison_stab",
		Rank:           1,
		DamagePerTick:  4,
		NextTick:       sim.tick,
		RemainingTicks: 10,
		CorrelationID:  "corr_poison_mark",
	}

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advancePoisonDots(&res)
	if damage := eventDamage(res, "monster_damaged"); damage < 8 {
		t.Fatalf("poison mark tick events=%+v, want at least 8 damage", res.Events)
	}
}

func TestRogueExecutionerPassiveExecutesLowHealthTarget(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.BaseHitChance = 1
	rules.Combat.BaseCritChance = 0
	executioner := rules.Skills["executioner"]
	executioner.Execute.ChancePercent = 100
	rules.Skills["executioner"] = executioner
	sim := newRogueSkillTestSim(t, rules)
	sim.progression.SkillRanks["executioner"] = 1
	player := sim.entities[sim.playerID]
	target := addRogueSkillTarget(sim, Vec2{X: player.pos.X + 1.2, Y: player.pos.Y}, 100)
	target.hp = 11

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	outcome := sim.damageMonsterByPlayerWithSlot(target, player.id, "execute_hit", &res, DamageRange{Min: 1, Max: 1}, damageTypeForce, mainHandSlot)
	if !outcome.Hit || target.hp != 0 {
		t.Fatalf("executioner outcome=%+v hp=%d events=%+v, want executed target", outcome, target.hp, res.Events)
	}
	if !hasSkillDamageEvent(res, "executioner") || !hasEvent(res, "monster_killed") {
		t.Fatalf("executioner events=%+v, want execute damage and kill", res.Events)
	}
}

func TestRogueExecutionerCannotBeCast(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := newRogueSkillTestSim(t, rules)
	sim.progression.BaseStats.Dex = 12
	sim.progression.SkillRanks["executioner"] = 1
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{
		MessageID:     "cast_executioner",
		CorrelationID: "corr_cast_executioner",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "executioner"},
	}})

	assertReject(t, cast, "cast_executioner", "passive_skill_not_castable")
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
