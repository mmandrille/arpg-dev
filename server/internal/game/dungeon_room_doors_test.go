package game

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRoomThresholdDoorCandidatesExcludeUnsupportedOrientation(t *testing.T) {
	routes := []roomCorridorRoute{
		{
			edge:  roomEdge{2, 1},
			width: 2,
			doorA: roomDoor{side: "north", center: Vec2{X: 10, Y: 20}, width: 2},
			doorB: roomDoor{side: "south", center: Vec2{X: 10, Y: 40}, width: 2},
		},
		{
			edge:  roomEdge{0, 1},
			width: 1.5,
			doorA: roomDoor{side: "east", center: Vec2{X: 20, Y: 30}, width: 1.5},
			doorB: roomDoor{side: "west", center: Vec2{X: 40, Y: 30}, width: 1.5},
		},
		{
			edge:  roomEdge{0, 2},
			width: 1.5,
			doorA: roomDoor{side: "east", center: Vec2{X: 20, Y: 40}, width: 1.5},
			doorB: roomDoor{side: "west", center: Vec2{X: 40, Y: 40}, width: 1.5},
		},
	}
	candidates := roomThresholdDoorCandidates(routes)
	if len(candidates) != 1 {
		t.Fatalf("eligible room door routes = %d, want one north/south route", len(candidates))
	}
	if candidates[0].edge != (roomEdge{1, 2}) {
		t.Fatalf("candidate edge = %v, want normalized edge [1 2]", candidates[0].edge)
	}
	if got := roomThresholdDoorPosition(candidates[0].doorA, 1); got != (Vec2{X: 10, Y: 20.5}) {
		t.Fatalf("north threshold door position = %+v, want the wall center %+v", got, Vec2{X: 10, Y: 20.5})
	}
}

func TestRoomThresholdDoorCandidatesRequireAlternateRoute(t *testing.T) {
	routes := []roomCorridorRoute{{
		edge:  roomEdge{0, 1},
		width: 2,
		doorA: roomDoor{side: "north", width: 2},
		doorB: roomDoor{side: "south", width: 2},
	}, {
		edge:  roomEdge{1, 2},
		width: 2,
		doorA: roomDoor{side: "north", width: 2},
		doorB: roomDoor{side: "south", width: 2},
	}}
	candidates := roomThresholdDoorCandidates(routes)
	if len(candidates) != 0 {
		t.Fatalf("tree-only corridor routes produced %d door candidates, want none", len(candidates))
	}
}

func TestRoomThresholdDoorRulesRejectInvalidConfiguration(t *testing.T) {
	rules := loadRules(t)
	config := rules.DungeonGeneration.RoomCorridorPCG
	config.Doors.MaxCount = -1
	if err := validateRoomThresholdDoorRules(config.Doors, config, rules); err == nil {
		t.Fatal("negative room-door max_count was accepted")
	}
	config = rules.DungeonGeneration.RoomCorridorPCG
	config.Doors.MaxCount = config.RoomCount.Max + config.LoopEdgeCount.Max
	if err := validateRoomThresholdDoorRules(config.Doors, config, rules); err == nil {
		t.Fatal("max_count above the possible room-route count was accepted")
	}
	config = rules.DungeonGeneration.RoomCorridorPCG
	config.Doors.Enabled = true
	config.Enabled = false
	if err := validateRoomThresholdDoorRules(config.Doors, config, rules); err == nil {
		t.Fatal("room doors were enabled while room-corridor generation was disabled")
	}
}

func TestGeneratedRoomThresholdDoorsAreDeterministicAndOpenable(t *testing.T) {
	rules := loadRules(t)
	if !rules.DungeonGeneration.RoomCorridorPCG.Doors.Enabled {
		t.Fatal("room threshold doors must be enabled in the default shared rules")
	}
	var seed string
	var generated generatedDungeonLevel
	var selected []roomThresholdDoorCandidate
	for attempt := range 32 {
		candidateSeed := fmt.Sprintf("v526_room_threshold_%02d", attempt)
		level, err := GenerateDungeonLevel(candidateSeed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("generate %q: %v", candidateSeed, err)
		}
		candidates := roomThresholdDoorCandidates(level.corridorRoutes)
		if len(candidates) > 0 {
			seed, generated, selected = candidateSeed, level, candidates
			break
		}
	}
	if seed == "" {
		t.Fatal("no deterministic seed produced a room route with a supported door threshold")
	}
	t.Logf("room-door bot seed: %s", seed)
	repeat, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("repeat generate %q: %v", seed, err)
	}
	if !reflect.DeepEqual(generated, repeat) {
		t.Fatal("same seed and rules produced different room routes or doors")
	}
	roomDoorPositions := make([]Vec2, 0, len(selected)*2)
	for _, candidate := range selected {
		roomDoorPositions = append(roomDoorPositions,
			roomThresholdDoorPosition(candidate.doorA, rules.DungeonGeneration.WallThickness),
			roomThresholdDoorPosition(candidate.doorB, rules.DungeonGeneration.WallThickness))
	}
	var roomDoorCount int
	for _, door := range generated.doors {
		for _, position := range roomDoorPositions {
			if door.defID == woodenDoorDefID && door.pos == position && door.state == interactableClosed {
				roomDoorCount++
				break
			}
		}
	}
	wantAtMost := minInt(rules.DungeonGeneration.RoomCorridorPCG.Doors.MaxCount, len(selected))
	if roomDoorCount == 0 || roomDoorCount > wantAtMost {
		t.Fatalf("generated room doors = %d, want 1..%d (candidates=%d)", roomDoorCount, wantAtMost, len(selected))
	}
	if err := validateGeneratedDungeonReachability(rules.DungeonGeneration.RulesForLevel(-1), generated); err != nil {
		t.Fatalf("room threshold door made a generated target unreachable: %v", err)
	}

	sim, err := NewSimWithWorld("sess_v526_room_threshold", seed, rules, "generated_wall_reachability_lab")
	if err != nil {
		t.Fatalf("generated dungeon world: %v", err)
	}
	var door *entity
	var threshold roomDoor
	for _, candidate := range selected {
		for _, endpoint := range []roomDoor{candidate.doorA, candidate.doorB} {
			position := roomThresholdDoorPosition(endpoint, rules.DungeonGeneration.WallThickness)
			for _, entityID := range sortedEntityIDs(sim.activeLevel().entities) {
				candidateEntity := sim.activeLevel().entities[entityID]
				if candidateEntity.kind == interactableEntity &&
					candidateEntity.interactableDefID == woodenDoorDefID &&
					candidateEntity.pos == position {
					door, threshold = candidateEntity, endpoint
					break
				}
			}
			if door != nil {
				break
			}
		}
		if door != nil {
			break
		}
	}
	if door == nil {
		t.Fatal("generated room threshold door was not populated as an interactable entity")
	}
	player := sim.entities[sim.playerID]
	direction := Vec2{Y: 1}
	player.pos = Vec2{X: door.pos.X, Y: door.pos.Y - 1}
	if threshold.side == "south" {
		direction.Y = -1
		player.pos.Y = door.pos.Y + 1
	}
	start := player.pos
	sim.Tick([]Input{{MessageID: "closed_move", Type: "move_intent", Move: &MoveIntent{Direction: direction, DurationTicks: 8}}})
	for range 7 {
		sim.Tick(nil)
	}
	if (direction.Y > 0 && player.pos.Y > door.pos.Y) || (direction.Y < 0 && player.pos.Y < door.pos.Y) {
		t.Fatalf("player crossed closed room door from %+v to %+v", start, player.pos)
	}
	player.pos = start
	opened := sim.Tick([]Input{{MessageID: "open_room_door", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(door.id)}}})
	assertAck(t, opened, "open_room_door")
	if door.state != interactableOpen || !hasEvent(opened, "interactable_activated") {
		t.Fatalf("authoritative door action state=%q events=%+v", door.state, opened.Events)
	}
	sim.Tick([]Input{{MessageID: "open_move", Type: "move_intent", Move: &MoveIntent{Direction: direction, DurationTicks: 8}}})
	for range 7 {
		sim.Tick(nil)
	}
	if (direction.Y > 0 && player.pos.Y <= door.pos.Y) || (direction.Y < 0 && player.pos.Y >= door.pos.Y) {
		t.Fatalf("player did not cross open room door from %+v to %+v", start, player.pos)
	}
}
