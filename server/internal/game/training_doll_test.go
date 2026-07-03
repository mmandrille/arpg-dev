package game

import "testing"

func TestTrainingDollMirrorsHostDefensiveStats(t *testing.T) {
	sim, err := NewSimWithWorld("sess_training_doll_mirror", "town_training_doll_mirror", loadRules(t), "town_training_doll_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}

	player := sim.levels[townLevel].entities[sim.playerID]
	doll := findMonsterByDefID(t, sim, townTrainingDollDefID)
	stats, _ := sim.playerEffectiveCombatStats()
	dollDef := sim.rules.Monsters[townTrainingDollDefID]
	wantHP := trainingDollMaxHPFromHost(player.maxHP, dollDef)

	if doll.maxHP != wantHP || doll.hp != wantHP {
		t.Fatalf("doll hp = %d/%d, want host max %d x%d = %d", doll.hp, doll.maxHP, player.maxHP, dollDef.effectiveTrainingHPMultiplier(), wantHP)
	}

	if doll.monsterArmor != stats.Armor {
		t.Fatalf("doll armor = %v, want %v", doll.monsterArmor, stats.Armor)
	}

	if doll.monsterBlockPercent != stats.BlockPercent {
		t.Fatalf("doll block = %v, want %v", doll.monsterBlockPercent, stats.BlockPercent)
	}
}

func TestTrainingDollCombatEventIncludesDamageBreakdown(t *testing.T) {
	sim, err := NewSimWithWorld("sess_training_doll_breakdown", "town_training_doll_breakdown", loadRules(t), "town_training_doll_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}

	doll := findMonsterByDefID(t, sim, townTrainingDollDefID)
	player := sim.levels[townLevel].entities[sim.playerID]
	player.pos = Vec2{X: doll.pos.X - 1, Y: doll.pos.Y}

	res := sim.Tick([]Input{{
		MessageID: "attack",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(doll.id)},
	}})

	event := findTrainingDollEvent(res.Events, "monster_damaged")
	if event == nil {
		t.Fatalf("missing monster_damaged: %+v", res.Events)
	}

	if len(event.DamageBreakdown) < 3 {
		t.Fatalf("damage_breakdown = %+v, want at least 3 lines", event.DamageBreakdown)
	}
}

func TestTrainingDollRevivesAfterDelay(t *testing.T) {
	sim, err := NewSimWithWorld("sess_training_doll_revive", "town_training_doll_revive", loadRules(t), "town_training_doll_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}

	doll := findMonsterByDefID(t, sim, townTrainingDollDefID)
	doll.hp = 1
	player := sim.levels[townLevel].entities[sim.playerID]
	player.pos = Vec2{X: doll.pos.X - 1, Y: doll.pos.Y}

	kill := sim.Tick([]Input{{
		MessageID: "kill",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(doll.id)},
	}})
	if !hasEvent(kill, "monster_killed") {
		t.Fatalf("missing monster_killed: %+v", kill.Events)
	}

	if doll.hp != 0 || doll.trainingDollReviveAt == 0 {
		t.Fatalf("doll down state = hp %d reviveAt %d, want hp 0 with revive scheduled", doll.hp, doll.trainingDollReviveAt)
	}

	delay := sim.trainingDollReviveDelayTicks(doll)
	for i := 0; i < delay && !hasEvent(sim.Tick(nil), "training_doll_revived"); i++ {
	}

	if doll.hp != doll.maxHP {
		t.Fatalf("doll hp after revive = %d, want %d", doll.hp, doll.maxHP)
	}
}

func TestCompanionIgnoresTrainingDoll(t *testing.T) {
	sim, err := NewSimWithWorld("sess_training_doll_companion", "town_training_doll_companion", loadRules(t), "town_training_doll_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}

	player := sim.levels[townLevel].entities[sim.playerID]
	doll := findMonsterByDefID(t, sim, townTrainingDollDefID)
	startHP := doll.hp
	companion := &entity{
		id:                    sim.alloc(),
		kind:                  companionEntity,
		pos:                   Vec2{X: doll.pos.X + 0.5, Y: doll.pos.Y},
		hp:                    20,
		maxHP:                 20,
		ownerID:               player.id,
		monsterDefID:          characterMercenaryMonsterDefID,
		monsterAttackDamage:   &DamageRange{Min: 5, Max: 5},
		monsterAttackCooldown: 1,
		aiMode:                monsterAIModeIdle,
		speed:                 1,
	}
	sim.levels[townLevel].entities[companion.id] = companion

	for i := 0; i < 80; i++ {
		sim.Tick(nil)
	}

	if doll.hp != startHP {
		t.Fatalf("companion damaged training doll hp=%d, want %d", doll.hp, startHP)
	}
	if companion.targetID != 0 {
		t.Fatalf("companion target_id=%d, want 0 for training doll", companion.targetID)
	}
}

func findTrainingDollEvent(events []Event, eventType string) *Event {
	for i := range events {
		if events[i].EventType == eventType {
			return &events[i]
		}
	}
	return nil
}

func findMonsterByDefID(t *testing.T, sim *Sim, defID string) *entity {
	t.Helper()
	level := sim.levels[townLevel]
	for _, id := range sortedEntityIDs(level.entities) {
		monster := level.entities[id]
		if monster != nil && monster.kind == monsterEntity && monster.monsterDefID == defID {
			return monster
		}
	}
	t.Fatalf("missing monster %s", defID)
	return nil
}
