package game

import "testing"

func TestLightRadiusClassBaselines(t *testing.T) {
	rules := loadRules(t)
	wantByClass := map[string]float64{
		"barbarian": 9,
		"paladin":   9,
		"rogue":     9,
		"sorcerer":  10,
		"ranger":    12,
	}
	for classID, want := range wantByClass {
		t.Run(classID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = classID
			state.BaseStats = rules.CharacterProgression.Classes[classID].BaseStats
			sim, err := NewSimWithWorldProgression("sess_light_radius_"+classID, "light_radius_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			view := sim.CharacterProgressionView()
			if view.DerivedStats.LightRadius != want {
				t.Fatalf("light_radius = %v, want %v", view.DerivedStats.LightRadius, want)
			}
			breakdown := findStatBreakdown(view.StatBreakdowns, "light_radius")
			if breakdown == nil || !hasBreakdownSource(breakdown.Sources, "character_formula") {
				t.Fatalf("light_radius breakdown = %+v, all=%+v", breakdown, view.StatBreakdowns)
			}
		})
	}
}

func TestLightRadiusUsesEquippedItemRolls(t *testing.T) {
	sim := MustNewSim("sess_light_radius_item", "light_radius_item_seed", loadRules(t))
	base := sim.CharacterProgressionView().DerivedStats.LightRadius
	ring := addRolledInventoryItem(t, sim, 6900, "ring", map[string]int{"light_radius": 2})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_light", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "equip_light")

	view := sim.CharacterProgressionView()
	if view.DerivedStats.LightRadius != base+2 {
		t.Fatalf("light_radius with ring = %v, want %v", view.DerivedStats.LightRadius, base+2)
	}
	breakdown := findStatBreakdown(view.StatBreakdowns, "light_radius")
	if breakdown == nil || !hasBreakdownSource(breakdown.Sources, "equipment_roll") {
		t.Fatalf("light_radius roll breakdown = %+v, all=%+v", breakdown, view.StatBreakdowns)
	}
}

func TestFogOfWarOptIn(t *testing.T) {
	sim := MustNewSim("sess_fog_default", "fog_default_seed", loadRules(t))
	if sim.fogOfWarEnabled {
		t.Fatalf("fog of war should default off until the fog slice is enabled")
	}
	sim.SetFogOfWarEnabled(true)
	if !sim.fogOfWarEnabled {
		t.Fatalf("fog of war opt-in did not enable filtering")
	}
}

func TestFogOfWarSnapshotHidesLivingMonstersOutsideLightRadius(t *testing.T) {
	sim := newFogTestSim(t)
	radius := sim.CharacterProgressionView().DerivedStats.LightRadius
	near := addTestMonster(sim, monsterDefID, Vec2{X: radius - 1, Y: 0}, 10)
	far := addTestMonster(sim, monsterDefID, Vec2{X: radius + 1, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(near.id)) {
		t.Fatalf("snapshot missing visible monster %d: %+v", near.id, snap.Entities)
	}
	if snapshotHasEntity(snap, idStr(far.id)) {
		t.Fatalf("snapshot leaked hidden monster %d beyond radius %v: %+v", far.id, radius, snap.Entities)
	}
	visible := sim.players[sim.playerID].VisibleMonsterIDs
	if !visible[near.id] || visible[far.id] {
		t.Fatalf("visible monster memory = %+v, want near only", visible)
	}
}

func TestFogOfWarSnapshotHidesLivingMonstersBehindWallsInsideLightRadius(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	level.walls = append(level.walls, wallObstacle{pos: Vec2{X: 4, Y: 0}, size: Vec2{X: 1, Y: 4}, source: "generated"})
	clear := addTestMonster(sim, monsterDefID, Vec2{X: 0, Y: 4}, 10)
	blocked := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(clear.id)) {
		t.Fatalf("snapshot missing unobstructed monster %d: %+v", clear.id, snap.Entities)
	}
	if snapshotHasEntity(snap, idStr(blocked.id)) {
		t.Fatalf("snapshot leaked wall-occluded monster %d: %+v", blocked.id, snap.Entities)
	}
	visible := sim.players[sim.playerID].VisibleMonsterIDs
	if !visible[clear.id] || visible[blocked.id] {
		t.Fatalf("visible monster memory = %+v, want clear only", visible)
	}
}

func TestFogOfWarSnapshotShowsLivingMonstersBehindWaterInsideLightRadius(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	level.walls = append(level.walls, wallObstacle{pos: Vec2{X: 4, Y: 0}, size: Vec2{X: 3, Y: 4}, source: "generated", kind: obstacleKindWater})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(monster.id)) {
		t.Fatalf("snapshot hid monster behind water %d: %+v", monster.id, snap.Entities)
	}
	if !sim.players[sim.playerID].VisibleMonsterIDs[monster.id] {
		t.Fatalf("visible monster memory missing monster behind water %d", monster.id)
	}
}

func TestFogOfWarSnapshotShowsLivingMonstersBehindHoleInsideLightRadius(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	level.walls = append(level.walls, wallObstacle{pos: Vec2{X: 4, Y: 0}, size: Vec2{X: 3, Y: 4}, source: "generated", kind: obstacleKindHole})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(monster.id)) {
		t.Fatalf("snapshot hid monster behind hole %d: %+v", monster.id, snap.Entities)
	}
	if !sim.players[sim.playerID].VisibleMonsterIDs[monster.id] {
		t.Fatalf("visible monster memory missing monster behind hole %d", monster.id)
	}
}

func TestFogOfWarSnapshotHidesLivingMonstersBehindTallObstacleKindsInsideLightRadius(t *testing.T) {
	for _, kind := range []string{obstacleKindRock, obstacleKindColumn} {
		t.Run(kind, func(t *testing.T) {
			sim := newFogTestSim(t)
			level := sim.activeLevel()
			level.walls = append(level.walls, wallObstacle{
				pos:       Vec2{X: 4, Y: 0},
				size:      Vec2{X: 1, Y: 4},
				source:    "generated",
				kind:      kind,
				blocksLOS: boolPtr(true),
			})
			monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

			snap := sim.SnapshotForPlayer(sim.playerID)
			if snapshotHasEntity(snap, idStr(monster.id)) {
				t.Fatalf("snapshot leaked %s-occluded monster %d: %+v", kind, monster.id, snap.Entities)
			}
			if sim.players[sim.playerID].VisibleMonsterIDs[monster.id] {
				t.Fatalf("visible monster memory contains %s-occluded monster %d", kind, monster.id)
			}
		})
	}
}

func TestFogOfWarSnapshotShowsLivingMonstersBehindLowRubbleInsideLightRadius(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	level.walls = append(level.walls, wallObstacle{pos: Vec2{X: 4, Y: 0}, size: Vec2{X: 3, Y: 4}, source: "generated", kind: obstacleKindRubble})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(monster.id)) {
		t.Fatalf("snapshot hid monster behind rubble %d: %+v", monster.id, snap.Entities)
	}
	if !sim.players[sim.playerID].VisibleMonsterIDs[monster.id] {
		t.Fatalf("visible monster memory missing monster behind rubble %d", monster.id)
	}
}

func TestFogOfWarDeltasRevealMonstersWhenTallObstacleLineOfSightClears(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	player := level.entities[sim.playerID]
	level.walls = append(level.walls, wallObstacle{
		pos:       Vec2{X: 3.5, Y: 0},
		size:      Vec2{X: 1, Y: 4},
		source:    "preset",
		kind:      obstacleKindColumn,
		blocksLOS: boolPtr(true),
	})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 6, Y: 0}, 10)
	sim.SnapshotForPlayer(sim.playerID)

	player.pos = Vec2{X: 6, Y: 4}
	reveal := TickResult{Changes: sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(player))},
	})}
	if !hasEntitySpawn(reveal, idStr(monster.id)) {
		t.Fatalf("clearing tall obstacle line of sight did not spawn monster %d: %+v", monster.id, reveal.Changes)
	}
}

func TestFogOfWarPresetLineOfSightBlockerLabHidesAndMarksTallWalls(t *testing.T) {
	sim, err := NewSimWithWorld("sess_los_blocker_lab", "fog_of_war_line_of_sight_blocker_seed", loadRules(t), "line_of_sight_blocker_lab")
	if err != nil {
		t.Fatalf("new line of sight blocker sim: %v", err)
	}
	snap := sim.SnapshotForPlayer(sim.playerID)
	if snapshotEntityCountByMonsterDef(snap, monsterDefID) != 0 {
		t.Fatalf("snapshot leaked occluded lab monster: %+v", snap.Entities)
	}
	if !snapshotHasLOSBlockingKind(snap, obstacleKindColumn) {
		t.Fatalf("snapshot wall layout missing LOS-blocking column: %+v", snap.Walls)
	}
	if !snapshotHasLOSBlockingKind(snap, obstacleKindRock) {
		t.Fatalf("snapshot wall layout missing LOS-blocking rock: %+v", snap.Walls)
	}
	if snapshotHasLOSBlockingKind(snap, obstacleKindRubble) {
		t.Fatalf("snapshot wall layout marked rubble as LOS-blocking: %+v", snap.Walls)
	}
}

func TestFogOfWarDeltasRevealAndConcealIdleMonsters(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	player := level.entities[sim.playerID]
	radius := sim.CharacterProgressionView().DerivedStats.LightRadius
	monster := addTestMonster(sim, monsterDefID, Vec2{X: radius + 1, Y: 0}, 10)
	sim.SnapshotForPlayer(sim.playerID)

	player.pos = Vec2{X: 2, Y: 0}
	reveal := TickResult{Changes: sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(player))},
	})}
	if !hasEntitySpawn(reveal, idStr(monster.id)) {
		t.Fatalf("movement into light radius did not spawn monster %d: %+v", monster.id, reveal.Changes)
	}

	player.pos = Vec2{X: 0, Y: 0}
	conceal := TickResult{Changes: sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(player))},
	})}
	if !hasEntityRemove(conceal, idStr(monster.id)) {
		t.Fatalf("movement out of light radius did not remove monster %d: %+v", monster.id, conceal.Changes)
	}
}

func TestFogOfWarDeltasRevealMonstersWhenWallLineOfSightClears(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	player := level.entities[sim.playerID]
	level.walls = append(level.walls, wallObstacle{pos: Vec2{X: 4, Y: 0}, size: Vec2{X: 1, Y: 4}, source: "generated"})
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)
	sim.SnapshotForPlayer(sim.playerID)

	level.walls = nil
	reveal := TickResult{Changes: sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(player))},
	})}
	if !hasEntitySpawn(reveal, idStr(monster.id)) {
		t.Fatalf("clearing wall line of sight did not spawn monster %d: %+v", monster.id, reveal.Changes)
	}
}

func TestFogOfWarDeltasRevealMonstersWhenClosedDoorOpens(t *testing.T) {
	sim := newFogTestSim(t)
	level := sim.activeLevel()
	player := level.entities[sim.playerID]
	door := addFogTestInteractable(sim, woodenDoorDefID, Vec2{X: 4, Y: 0}, interactableClosed)
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if snapshotHasEntity(snap, idStr(monster.id)) {
		t.Fatalf("snapshot leaked closed-door-occluded monster %d: %+v", monster.id, snap.Entities)
	}
	if sim.players[sim.playerID].VisibleMonsterIDs[monster.id] {
		t.Fatalf("visible monster memory contains closed-door-occluded monster %d", monster.id)
	}

	door.state = interactableOpen
	reveal := TickResult{Changes: sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(player))},
	})}
	if !hasEntitySpawn(reveal, idStr(monster.id)) {
		t.Fatalf("opening door line of sight did not spawn monster %d: %+v", monster.id, reveal.Changes)
	}
}

func TestFogOfWarIgnoresOpenAndNonBarrierInteractables(t *testing.T) {
	sim := newFogTestSim(t)
	if def := sim.rules.Interactables[treasureChestDefID]; def.BarrierWhenClosed != nil {
		t.Fatalf("%s unexpectedly has barrier_when_closed", treasureChestDefID)
	}
	addFogTestInteractable(sim, woodenDoorDefID, Vec2{X: 4, Y: 0}, interactableOpen)
	addFogTestInteractable(sim, treasureChestDefID, Vec2{X: 4, Y: 0}, interactableClosed)
	monster := addTestMonster(sim, monsterDefID, Vec2{X: 8, Y: 0}, 10)

	snap := sim.SnapshotForPlayer(sim.playerID)
	if !snapshotHasEntity(snap, idStr(monster.id)) {
		t.Fatalf("snapshot hid monster behind open/non-barrier interactables %d: %+v", monster.id, snap.Entities)
	}
}

func TestFogOfWarSuppressesHiddenMonsterChangesAndEvents(t *testing.T) {
	sim := newFogTestSim(t)
	radius := sim.CharacterProgressionView().DerivedStats.LightRadius
	monster := addTestMonster(sim, monsterDefID, Vec2{X: radius + 1, Y: 0}, 10)
	sim.SnapshotForPlayer(sim.playerID)

	changes := sim.FilterChangesForPlayer(sim.playerID, sim.currentLevel, []Change{
		{Op: OpEntityUpdate, Entity: ptrEntityView(sim.entityView(monster))},
	})
	if len(changes) != 0 {
		t.Fatalf("hidden monster change leaked: %+v", changes)
	}
	events := sim.FilterEventsForPlayer(sim.playerID, sim.currentLevel, []Event{
		{EventType: "monster_aggro", EntityID: idStr(monster.id)},
	})
	if len(events) != 0 {
		t.Fatalf("hidden monster event leaked: %+v", events)
	}
}

func newFogTestSim(t *testing.T) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_fog_of_war", "fog_of_war_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new fog sim: %v", err)
	}
	sim.SetFogOfWarEnabled(true)
	town := sim.activeLevel()
	player := town.entities[sim.playerID]
	delete(town.entities, sim.playerID)
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatalf("ensure fog dungeon: %v", err)
	}
	player.pos = Vec2{}
	level.entities[sim.playerID] = player
	sim.currentLevel = level.levelNum
	if ps := sim.players[sim.playerID]; ps != nil {
		ps.CurrentLevel = level.levelNum
	}
	for _, id := range sortedEntityIDs(level.entities) {
		if level.entities[id].kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	level.walls = nil
	player = level.entities[sim.playerID]
	player.pos = Vec2{}
	sim.syncCompatibilityFields()
	return sim
}

func addFogTestInteractable(sim *Sim, defID string, pos Vec2, state string) *entity {
	interactable := &entity{
		id:                sim.alloc(),
		kind:              interactableEntity,
		pos:               pos,
		interactableDefID: defID,
		state:             state,
	}
	sim.activeLevel().entities[interactable.id] = interactable
	sim.syncCompatibilityFields()
	return interactable
}

func snapshotHasEntity(snap Snapshot, entityID string) bool {
	for _, entity := range snap.Entities {
		if entity.ID == entityID {
			return true
		}
	}
	return false
}

func snapshotEntityCountByMonsterDef(snap Snapshot, monsterDefID string) int {
	count := 0
	for _, entity := range snap.Entities {
		if entity.Type == monsterEntity && entity.MonsterDefID == monsterDefID {
			count++
		}
	}
	return count
}

func snapshotHasLOSBlockingKind(snap Snapshot, kind string) bool {
	for _, wall := range snap.Walls {
		if wall.Kind == kind && wall.BlocksLineOfSight != nil && *wall.BlocksLineOfSight {
			return true
		}
	}
	return false
}
