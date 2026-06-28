package game

import "testing"

func TestActionAutoApproachQueuesWhenOutOfRange(t *testing.T) {
	rules := loadRules(t)

	t.Run("monster", func(t *testing.T) {
		sim := MustNewSim("sess_range_monster", "01", rules)
		r := sim.Tick([]Input{{MessageID: "a", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "a")
	})

	t.Run("loot", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_loot", "01", rules, "gear_before_combat")
		if err != nil {
			t.Fatalf("gear world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "p", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "p")
	})

	t.Run("door", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_door", "01", rules, "door_lab")
		if err != nil {
			t.Fatalf("door world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "d", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "d")
	})
}

func TestDoorLabClosedDoorPreventsPassageUntilActivated(t *testing.T) {
	sim, err := NewSimWithWorld("sess_door_passage", "01", loadRules(t), "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}

	sim.Tick([]Input{{MessageID: "push_closed", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 7}}})
	for i := 0; i < 6; i++ {
		sim.Tick(nil)
	}
	if got := sim.entities[sim.playerID].pos; got.X >= 4 {
		t.Fatalf("player passed closed door: pos=%+v", got)
	}
	sim.entities[sim.playerID].pos = Vec2{X: 3, Y: 2}
	open := sim.Tick([]Input{{MessageID: "open", CorrelationID: "corr_door", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, open, "open")
	if !hasEvent(open, "interactable_activated") {
		t.Fatalf("missing interactable_activated: %+v", open.Events)
	}
	door := sim.findEntity("1002")
	if door == nil || door.state != interactableOpen {
		t.Fatalf("door state = %+v, want open", door)
	}

	// 9 ticks at min momentum speed (0.75×0.6=0.45/tick) reaches X=7.05 from {3,2},
	// within loot interaction reach (1.0+0.35=1.35) of the loot at {8,2}.
	sim.Tick([]Input{{MessageID: "through", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 9}}})
	for i := 0; i < 8; i++ {
		sim.Tick(nil)
	}
	if got := sim.entities[sim.playerID].pos; got.X <= 4 {
		t.Fatalf("player did not pass open door: pos=%+v", got)
	}
	pickup := sim.Tick([]Input{{MessageID: "loot", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}}})
	assertAck(t, pickup, "loot")
	if !hasEvent(pickup, "item_picked_up") {
		t.Fatalf("missing item_picked_up after door passage: %+v", pickup.Events)
	}
}

func TestOpenDoorCanBeClosedAgain(t *testing.T) {
	sim, err := NewSimWithWorld("sess_door_toggle", "01", loadRules(t), "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}
	door := sim.findEntity("1002")
	if door == nil {
		t.Fatal("missing door")
	}
	sim.entities[sim.playerID].pos = Vec2{X: 3, Y: 2}
	open := sim.Tick([]Input{{MessageID: "open", CorrelationID: "corr_open", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, open, "open")
	if door.state != interactableOpen {
		t.Fatalf("door after open = %s, want open", door.state)
	}
	close := sim.Tick([]Input{{MessageID: "close", CorrelationID: "corr_close", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, close, "close")
	if door.state != interactableClosed {
		t.Fatalf("door after close = %s, want closed", door.state)
	}
	foundCloseEvent := false
	for _, ev := range close.Events {
		if ev.EventType == "interactable_state_changed" && ev.EntityID == "1002" && ev.State == interactableClosed {
			foundCloseEvent = true
			break
		}
	}
	if !foundCloseEvent {
		t.Fatalf("missing close state event: %+v", close.Events)
	}
}

func TestClosedDoorAutoApproachPrefersPlayerSide(t *testing.T) {
	sim, err := NewSimWithWorld("sess_door_same_side", "01", loadRules(t), "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}
	door := sim.findEntity("1002")
	if door == nil {
		t.Fatal("missing door")
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5.5, Y: 4.0}
	goal, steps, ok := sim.findMeleeApproachGoal(door)
	if !ok {
		t.Fatal("findMeleeApproachGoal ok=false")
	}
	if len(steps) == 0 {
		t.Fatal("findMeleeApproachGoal returned empty path")
	}
	if goal.Y <= door.pos.Y {
		t.Fatalf("door approach goal = %+v, want same side above closed door at %+v", goal, door.pos)
	}
}

func TestTreasureChestOpensOnceAndDropsLoot(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.EliteObjective.Enabled = false
	sim, err := NewSimWithWorld("sess_chest_open", "chest_seed_22", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	var chest *entity
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == treasureChestDefID {
			chest = e
			break
		}
	}
	if chest == nil {
		t.Fatalf("missing generated chest: %+v", sim.activeLevel().entities)
	}
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	beforeLoot := countEntitiesByKind(sim.activeLevel(), lootEntity)
	beforeGold := countLootByItemDef(sim.activeLevel(), goldItemDefID)
	open := sim.Tick([]Input{{MessageID: "open_chest", CorrelationID: "corr_chest", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_chest")
	if !hasEvent(open, "interactable_activated") || !hasEvent(open, "loot_dropped") {
		t.Fatalf("open chest events = %+v", open.Events)
	}
	if chest.state != interactableOpen {
		t.Fatalf("chest state = %s, want open", chest.state)
	}
	afterLoot := countEntitiesByKind(sim.activeLevel(), lootEntity)
	if afterLoot <= beforeLoot {
		t.Fatalf("loot count after open = %d, before %d", afterLoot, beforeLoot)
	}
	if got := countLootByItemDef(sim.activeLevel(), goldItemDefID); got != beforeGold+1 {
		t.Fatalf("gold drops after chest open = %d, want %d", got, beforeGold+1)
	}
	again := sim.Tick([]Input{{MessageID: "open_chest_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, again, "open_chest_again", "invalid_target")
	if got := countEntitiesByKind(sim.activeLevel(), lootEntity); got != afterLoot {
		t.Fatalf("reopen changed loot count = %d, want %d", got, afterLoot)
	}
}

func TestV40ObstaclesWoodenDoorActionAcks(t *testing.T) {
	sim, err := NewSimWithWorld("sess_v40_door", "v40_obstacles", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	var door *entity
	for _, e := range sim.activeLevel().entities {
		if e != nil && e.kind == interactableEntity && e.interactableDefID == woodenDoorDefID {
			door = e
			break
		}
	}
	if door == nil {
		t.Fatal("wooden_door not found on generated floor")
	}
	sim.resetPlayerNavigationBudget()
	res := sim.Tick([]Input{{
		MessageID: "msg-4",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(door.id)},
	}})
	if len(res.Rejects) > 0 {
		t.Fatalf("door action rejected: %+v", res.Rejects)
	}
}
