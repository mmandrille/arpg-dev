package game

import (
	"encoding/json"
	"math"
	"testing"
)

func TestMercenaryEffectiveLevelCapsHighSource(t *testing.T) {
	if got := mercenaryEffectiveLevel(25, 10); got != 10 {
		t.Fatalf("effective level = %d, want 10", got)
	}
}

func TestMercenaryEffectiveLevelDoesNotScaleLowSourceUp(t *testing.T) {
	if got := mercenaryEffectiveLevel(10, 20); got != 10 {
		t.Fatalf("effective level = %d, want 10", got)
	}
}

func TestMercenaryHiringListsCandidatesAndHiresCharacter(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_character_merc_hire")
	sim.progression.Level = 10
	cost := sim.mercenaryHireCostGold()
	if cost != 100 {
		t.Fatalf("hire cost = %d, want 100", cost)
	}
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_alt",
			Name:           "Alt Hero",
			CharacterClass: "barbarian",
			Level:          25,
			Progression: CharacterProgressionState{
				CharacterClass: "barbarian",
				Level:          25,
				BaseStats:      BaseStatsView{Str: 20, Dex: 10, Vit: 20, Magic: 5},
			},
			Items: []PersistedItem{
				{
					InstanceID: "9001",
					ItemDefID:  "long_sword",
					Slot:       mainHandSlot,
					Equipped:   true,
				},
			},
		},
	})
	sim.savePlayer(sim.defaultPlayer())

	open := sim.Tick([]Input{mercenaryHireInput(board, "open_board")})
	opened := findEvent(open.Events, "mercenary_board_opened")
	if opened == nil || len(opened.MercenaryCandidates) != 1 || opened.MercenaryCandidates[0].CharacterID != "char_alt" {
		t.Fatalf("mercenary_board_opened = %+v", opened)
	}

	hire := sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_alt", "char_alt")})
	assertAck(t, hire, "hire_alt")
	hired := findEvent(hire.Events, "mercenary_hired")
	mercenary := hiredMercenary(sim)
	if hired == nil || mercenary == nil || hired.SourceCharacterID != "char_alt" || mercenary.sourceCharacterID != "char_alt" {
		t.Fatalf("mercenary_hired=%+v mercenary=%+v", hired, mercenary)
	}
	if sim.progression.HiredMercenaryCharacterID != "char_alt" {
		t.Fatalf("hired mercenary id = %q, want char_alt", sim.progression.HiredMercenaryCharacterID)
	}
	if mercenary.maxHP >= 25*10 {
		t.Fatalf("mercenary max hp = %d, want scaled below naive lvl25 hp", mercenary.maxHP)
	}
}

func TestMercenaryHiringKeepsLowLevelSourceUnscaled(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_character_merc_low")
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	lowSnap := MercenaryCharacterSnapshot{
		CharacterID:    "char_low",
		Name:           "Low Hero",
		CharacterClass: "sorcerer",
		Level:          10,
		Progression: CharacterProgressionState{
			CharacterClass: "sorcerer",
			Level:          10,
			BaseStats:      BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18},
		},
	}
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{lowSnap})
	equipped := equippedItemsFromPersisted(lowSnap.Items)
	baseStats, _ := mercenaryCombatStats(sim.rules, lowSnap.Progression, equipped)

	sim.savePlayer(sim.defaultPlayer())
	hire := sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_low", "char_low")})
	assertAck(t, hire, "hire_low")
	mercenary := hiredMercenary(sim)
	if mercenary == nil {
		t.Fatal("missing hired mercenary")
	}
	wantHP := scaleMercenaryInt(int(math.Round(baseStats.MaxHP)), lowSnap.Level, mercenaryEffectiveLevel(lowSnap.Level, 20))
	if mercenary.maxHP != wantHP {
		t.Fatalf("mercenary max hp = %d, want %d", mercenary.maxHP, wantHP)
	}
}

func TestMercenaryRangedCompanionUsesProjectile(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_character_merc_ranged")
	sim.progression.Level = 10
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_ranger",
			Name:           "Ranger Alt",
			CharacterClass: "ranger",
			Level:          10,
			Progression: CharacterProgressionState{
				CharacterClass: "ranger",
				Level:          10,
				BaseStats:      BaseStatsView{Str: 8, Dex: 16, Vit: 10, Magic: 6},
			},
			Items: []PersistedItem{mustMercenaryBowItem(t, "9002")},
		},
	})
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_ranger", "char_ranger")}), "hire_ranger")
	mercenary := hiredMercenary(sim)
	if mercenary == nil || mercenary.companionAttackMode != attackModeRanged {
		t.Fatalf("ranged mercenary = %+v", mercenary)
	}
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if target == nil {
		t.Fatal("missing target")
	}
	mercenary.targetID = target.id
	var spawn TickResult
	var projectile *EntityView
	for i := 0; i < 40; i++ {
		spawn = sim.Tick(nil)
		projectile = firstChangeEntityByType(spawn, projectileEntity)
		if projectile != nil {
			break
		}
	}
	if projectile == nil {
		t.Fatalf("ranged mercenary did not spawn projectile: %+v", spawn.Events)
	}
	mercenary = hiredMercenary(sim)
	var damaged bool
	for i := 0; i < 80; i++ {
		for _, res := range sim.TickResults(nil) {
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(mercenary.id) {
					damaged = true
				}
			}
		}
		if damaged {
			break
		}
	}
	if !damaged {
		t.Fatalf("ranged mercenary projectile did not damage monster")
	}
}

func TestCharacterMercenaryScenarioSeedDamagesMonster(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v403_character_mercenary_hire")
	placePlayerWithinMercenaryLabDefendRadius(sim)
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{{
		CharacterID: "char_seed", Name: "Merc Alt Sorc", CharacterClass: "sorcerer", Level: 10,
		Progression: CharacterProgressionState{
			CharacterClass: "sorcerer",
			Level:          10,
			BaseStats:      BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18},
		},
		Items: []PersistedItem{mustMercenaryStaffItem(t, "9003")},
	}})
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_seed", "char_seed")}), "hire_seed")
	mercenary := hiredMercenary(sim)
	setCompanionAssistStance(mercenary)
	var damaged bool
	for i := 0; i < 120; i++ {
		for _, res := range sim.TickResults(nil) {
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(mercenary.id) {
					damaged = true
				}
			}
		}
	}
	if !damaged {
		t.Fatalf("scenario seed did not produce companion damage")
	}
}

func TestCharacterMercenaryRealtimeShapedTicksEmitCombat(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v404_character_merc_realtime")
	placePlayerWithinMercenaryLabDefendRadius(sim)
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{{
		CharacterID: "char_rt", Name: "RT Alt", CharacterClass: "sorcerer", Level: 10,
		Progression: CharacterProgressionState{CharacterClass: "sorcerer", Level: 10, BaseStats: BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18}},
	}})
	sim.savePlayer(sim.defaultPlayer())
	results := sim.TickResults([]Input{mercenaryHireCharacterInput(board, "hire_rt", "char_rt")})
	assertAckResult(t, results, "hire_rt")
	mercenary := hiredMercenary(sim)
	setCompanionAssistStance(mercenary)
	if mercenary == nil {
		t.Fatal("missing hired mercenary")
	}
	var damaged bool
	for i := 0; i < 120; i++ {
		for _, res := range sim.TickResults(nil) {
			for _, ev := range res.Events {
				if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(mercenary.id) {
					damaged = true
				}
			}
		}
	}
	if !damaged {
		target := findMonsterByDef(sim, "combat_lab_soft_target")
		t.Fatalf("no companion damage in tick results; mercenary=%+v target=%+v", mercenary.pos, target.pos)
	}
}

func assertAckResult(t *testing.T, results []TickResult, messageID string) {
	t.Helper()
	for _, res := range results {
		for _, ack := range res.Acks {
			if ack.MessageID == messageID {
				return
			}
		}
	}
	t.Fatalf("missing ack for %s in %+v", messageID, results)
}

func TestCharacterMercenaryMovesFromPlayerAtWorldSpawn(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v404_character_merc_world_spawn")
	placePlayerWithinMercenaryLabDefendRadius(sim)
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{{
		CharacterID: "char_spawn", Name: "Spawn Alt", CharacterClass: "sorcerer", Level: 10,
		Progression: CharacterProgressionState{CharacterClass: "sorcerer", Level: 10, BaseStats: BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18}},
	}})
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_spawn", "char_spawn")}), "hire_spawn")
	mercenary := hiredMercenary(sim)
	setCompanionAssistStance(mercenary)
	start := mercenary.pos
	for i := 0; i < 120; i++ {
		sim.Tick(nil)
	}
	if distance(mercenary.pos, start) < 0.5 {
		t.Fatalf("mercenary at %+v did not move from %+v", mercenary.pos, start)
	}
}

func TestCharacterMercenaryDamagesMonsterInLab(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v404_character_merc_combat")
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_combat",
			Name:           "Combat Alt",
			CharacterClass: "sorcerer",
			Level:          10,
			Progression: CharacterProgressionState{
				CharacterClass: "sorcerer",
				Level:          10,
				BaseStats:      BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18},
			},
		},
	})
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_combat", "char_combat")}), "hire_combat")
	var damaged bool
	for i := 0; i < 120; i++ {
		res := sim.Tick(nil)
		for _, ev := range res.Events {
			if ev.EventType != "monster_damaged" {
				continue
			}
			sourceID, ok := ParseEntityID(ev.SourceEntityID)
			if !ok {
				continue
			}
			mercenary := hiredMercenary(sim)
			if mercenary != nil && sourceID == mercenary.id {
				damaged = true
				break
			}
		}
		if damaged {
			break
		}
	}
	if !damaged {
		mercenary := hiredMercenary(sim)
		target := findMonsterByDef(sim, "combat_lab_soft_target")
		t.Fatalf("companion did not damage target; mercenary=%+v target=%+v mercenary.pos=%+v target.pos=%+v", mercenary, target, mercenary.pos, target.pos)
	}
}

func TestCharacterMercenaryMovesTowardMonster(t *testing.T) {
	sim, board := newMercenaryHiringSim(t, "v404_character_merc_move")
	sim.progression.Level = 20
	cost := sim.mercenaryHireCostGold()
	sim.gold = cost
	sim.progression.Gold = cost
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{
		{
			CharacterID:    "char_move",
			Name:           "Move Alt",
			CharacterClass: "sorcerer",
			Level:          10,
			Progression: CharacterProgressionState{
				CharacterClass: "sorcerer",
				Level:          10,
				BaseStats:      BaseStatsView{Str: 5, Dex: 8, Vit: 8, Magic: 18},
			},
		},
	})
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{mercenaryHireCharacterInput(board, "hire_move", "char_move")}), "hire_move")
	mercenary := hiredMercenary(sim)
	if mercenary == nil {
		t.Fatal("missing hired mercenary")
	}
	if mercenary.speed <= 0 {
		t.Fatalf("mercenary speed = %v, want > 0", mercenary.speed)
	}
	start := mercenary.pos
	for i := 0; i < 120; i++ {
		sim.Tick(nil)
	}
	if distance(mercenary.pos, start) < 0.5 {
		t.Fatalf("mercenary moved %.3f from start, want >= 0.5 (start=%+v end=%+v speed=%v)", distance(mercenary.pos, start), start, mercenary.pos, mercenary.speed)
	}
}

func mercenaryHireCharacterInput(board *entity, msgID, characterID string) Input {
	return Input{
		MessageID:     msgID,
		CorrelationID: "corr_" + msgID,
		Type:          "action_intent",
		Action: &ActionIntent{
			TargetID:             idStr(board.id),
			MercenaryCharacterID: characterID,
		},
	}
}

func TestRestoreHiredMercenaryCompanionRespawnsWithoutGold(t *testing.T) {
	sim, _ := newMercenaryHiringSim(t, "v404_restore_hired_mercenary")
	sim.progression.HiredMercenaryCharacterID = "char_restore"
	sim.LoadMercenaryRoster([]MercenaryCharacterSnapshot{{
		CharacterID: "char_restore", Name: "Restore Alt", CharacterClass: "barbarian", Level: 8,
		Progression: CharacterProgressionState{
			CharacterClass: "barbarian", Level: 8,
			BaseStats: BaseStatsView{Str: 12, Dex: 8, Vit: 10, Magic: 5},
		},
	}})
	sim.RestoreHiredMercenaryCompanion(sim.playerID)
	if hiredMercenary(sim) == nil {
		t.Fatal("expected restored hired mercenary companion")
	}
	if sim.progression.HiredMercenaryCharacterID != "char_restore" {
		t.Fatalf("hired mercenary id = %q, want char_restore", sim.progression.HiredMercenaryCharacterID)
	}
}

func TestClearHiredMercenaryForOwnerClearsProgressionField(t *testing.T) {
	sim, _ := newMercenaryHiringSim(t, "v404_clear_hired_mercenary")
	sim.progression.HiredMercenaryCharacterID = "char_clear"
	sim.clearHiredMercenaryForOwner(sim.playerID, nil)
	if sim.progression.HiredMercenaryCharacterID != "" {
		t.Fatalf("hired mercenary id = %q, want cleared", sim.progression.HiredMercenaryCharacterID)
	}
}

func mustMercenaryStaffItem(t *testing.T, instanceID string) PersistedItem {
	t.Helper()
	raw, err := json.Marshal(ItemRollPayload{
		ItemTemplateID: "sorcerer_staff",
		DisplayName:    "Starter Staff",
		Rarity:         "common",
		Stats:          map[string]int{"damage_min": 1, "damage_max": 3, "magic": 2},
		Requirements:   map[string]int{"level": 1, "magic": 5},
	})
	if err != nil {
		t.Fatal(err)
	}

	return PersistedItem{
		InstanceID:  instanceID,
		ItemDefID:   "sorcerer_staff",
		Slot:        mainHandSlot,
		Equipped:    true,
		RolledStats: raw,
	}
}

func mustMercenaryBowItem(t *testing.T, instanceID string) PersistedItem {
	t.Helper()
	raw, err := json.Marshal(ItemRollPayload{
		ItemTemplateID: "hunter_bow",
		DisplayName:    "Hunter Bow",
		Rarity:         "common",
		Stats:          map[string]int{"damage_min": 1, "damage_max": 2},
		Requirements:   map[string]int{"level": 1, "dex": 7},
	})
	if err != nil {
		t.Fatal(err)
	}

	return PersistedItem{
		InstanceID:  instanceID,
		ItemDefID:   "hunter_bow",
		Slot:        mainHandSlot,
		Equipped:    true,
		RolledStats: raw,
	}
}

func placePlayerWithinMercenaryLabDefendRadius(sim *Sim) {
	player := sim.activeLevel().entities[sim.playerID]
	if player == nil {
		return
	}
	// mercenary_hiring_lab soft target is at x=7.5; defend stance requires owner within assist radius.
	player.pos = Vec2{X: 3, Y: 5}
}

func setCompanionAssistStance(companion *entity) {
	if companion != nil {
		companion.companionStance = CompanionStanceAssist
	}
}
