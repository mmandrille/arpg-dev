package game

import "testing"

func TestDungeonEliteObjectiveChestRequiresEliteLeader(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	level, err := GenerateDungeonLevel("v158_forced_elite_objective", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate forced elite objective: %v", err)
	}
	objectives := 0
	for _, chest := range level.chests {
		if chest.eliteObjective {
			objectives++
			if chest.defID != treasureChestDefID || chest.lootTable != rules.DungeonGeneration.EliteObjective.LootTable {
				t.Fatalf("elite objective chest = %+v", chest)
			}
		}
	}
	if objectives != 1 {
		t.Fatalf("elite objective chests = %d in %+v, want 1", objectives, level.chests)
	}

	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 0
	level, err = GenerateDungeonLevel("v158_forced_elite_objective", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate no elite objective: %v", err)
	}
	for _, chest := range level.chests {
		if chest.eliteObjective {
			t.Fatalf("unexpected elite objective without elite leader: %+v", level.chests)
		}
	}
}

func TestEliteObjectiveChestRequiresLeaderKill(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	sim, err := NewSimWithWorld("sess_elite_objective_gate", "v158_elite_objective_0000", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	level := sim.activeLevel()
	chest := findEliteObjectiveChestEntity(level)
	if chest == nil {
		t.Fatalf("missing elite objective chest: %+v", level.entities)
	}
	leaders := findLivingPackLeaders(level)
	if len(leaders) == 0 {
		t.Fatalf("missing living pack leader: %+v", level.entities)
	}

	sim.entities[sim.playerID].pos = chest.pos
	blocked := sim.Tick([]Input{{MessageID: "open_locked_objective", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, blocked, "open_locked_objective", "elite_objective_incomplete")
	if chest.state != interactableClosed {
		t.Fatalf("chest state after blocked open = %s, want %s", chest.state, interactableClosed)
	}
	beforeLoot := countEntitiesByKind(level, lootEntity)

	for _, leader := range leaders {
		leader.hp = 0
	}

	sim.entities[sim.playerID].pos = chest.pos
	open := sim.Tick([]Input{{MessageID: "open_completed_objective", CorrelationID: "corr_open_objective", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_completed_objective")
	if !hasEvent(open, "interactable_activated") || !hasEvent(open, "loot_dropped") {
		t.Fatalf("open completed objective events = %+v", open.Events)
	}
	if chest.state != interactableOpen {
		t.Fatalf("chest state after completed open = %s, want %s", chest.state, interactableOpen)
	}
	if got := countEntitiesByKind(level, lootEntity); got <= beforeLoot {
		t.Fatalf("loot count after completed open = %d, before %d", got, beforeLoot)
	}
}

func TestEliteObjectiveChestRequiresAllLeaderKills(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	sim, err := NewSimWithWorld("sess_elite_objective_clear_all", "v158_elite_objective_0000", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	level := sim.activeLevel()
	chest := findEliteObjectiveChestEntity(level)
	if chest == nil {
		t.Fatalf("missing elite objective chest: %+v", level.entities)
	}
	leaders := findLivingPackLeaders(level)
	if len(leaders) == 0 {
		t.Fatalf("missing living pack leader: %+v", level.entities)
	}
	extraLeader := &entity{
		id:                sim.alloc(),
		kind:              monsterEntity,
		pos:               Vec2{X: chest.pos.X + 1, Y: chest.pos.Y},
		spawnPos:          Vec2{X: chest.pos.X + 1, Y: chest.pos.Y},
		hp:                1,
		maxHP:             1,
		monsterDefID:      monsterDefID,
		monsterPackID:     "test_extra_elite_pack",
		monsterPackLeader: true,
		lootTable:         "depth_1_monsters",
	}
	level.entities[extraLeader.id] = extraLeader
	leaders = append(leaders, extraLeader)

	sim.entities[sim.playerID].pos = chest.pos
	blocked := sim.Tick([]Input{{MessageID: "open_locked_objective", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, blocked, "open_locked_objective", "elite_objective_incomplete")

	leaders[0].hp = 0
	stillBlocked := sim.Tick([]Input{{MessageID: "open_partially_cleared_objective", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, stillBlocked, "open_partially_cleared_objective", "elite_objective_incomplete")
	if chest.state != interactableClosed {
		t.Fatalf("chest state after partial clear = %s, want %s", chest.state, interactableClosed)
	}

	for _, leader := range leaders {
		leader.hp = 0
	}
	beforeLoot := countEntitiesByKind(level, lootEntity)
	open := sim.Tick([]Input{{MessageID: "open_fully_cleared_objective", CorrelationID: "corr_full_clear", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_fully_cleared_objective")
	if !hasEvent(open, "interactable_activated") || !hasEvent(open, "loot_dropped") {
		t.Fatalf("open fully cleared objective events = %+v", open.Events)
	}
	if chest.state != interactableOpen {
		t.Fatalf("chest state after full clear = %s, want %s", chest.state, interactableOpen)
	}
	if got := countEntitiesByKind(level, lootEntity); got <= beforeLoot {
		t.Fatalf("loot count after full clear = %d, before %d", got, beforeLoot)
	}
}

func findEliteObjectiveChestEntity(level *LevelState) *entity {
	for _, id := range sortedEntityIDs(level.entities) {
		if level.eliteObjectiveChestIDs[id] {
			return level.entities[id]
		}
	}
	return nil
}

func findLivingPackLeaders(level *LevelState) []*entity {
	leaders := []*entity{}
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity.kind == monsterEntity && entity.monsterPackLeader && entity.hp > 0 {
			leaders = append(leaders, entity)
		}
	}
	return leaders
}
