package game

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRoomCorridorRouting_ValidatesWidthChoices(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RoomCorridorPCG
	if err := validateRoomCorridorPCGRules(rules); err != nil {
		t.Fatalf("configured widths rejected: %v", err)
	}

	tests := []struct {
		name   string
		widths []float64
	}{
		{name: "empty", widths: nil},
		{name: "too narrow for player", widths: []float64{playerRadius}},
		{name: "too wide for room cell", widths: []float64{minFloat(rules.RoomSizeMin.X, rules.RoomSizeMin.Y)/roomShapeSide + 1}},
		{name: "duplicate", widths: []float64{rules.CorridorWidths[0], rules.CorridorWidths[0]}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			invalid := rules
			invalid.CorridorWidths = tc.widths
			if err := validateRoomCorridorPCGRules(invalid); err == nil {
				t.Fatalf("invalid corridor widths %v were accepted", tc.widths)
			}
		})
	}
}

func TestRoomCorridorRouting_DetoursAroundBlockerDeterministically(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	var rectangle [roomShapeCellCount]bool
	for i := range rectangle {
		rectangle[i] = true
	}
	rooms := []dungeonRoom{
		makeDungeonRoom(Vec2{X: 10, Y: 10}, Vec2{X: 20, Y: 20}, false, "rectangle", rectangle),
		makeDungeonRoom(Vec2{X: 40, Y: 10}, Vec2{X: 50, Y: 20}, false, "rectangle", rectangle),
	}
	blocker := wallObstacle{
		pos:         Vec2{X: 30, Y: 15},
		size:        Vec2{X: 6, Y: 5},
		source:      "generated",
		shapeFamily: "rectangle",
		kind:        obstacleKindWall,
	}
	doorA, doorB, ok := roomDoorsBetweenRooms(rooms, roomEdge{0, 1})
	if !ok {
		t.Fatal("could not resolve fixture doors")
	}
	widths := rules.RoomCorridorPCG.CorridorWidths
	var seed uint64
	var width float64
	var variantStart int
	foundBlockedFirstChoice := false
	for attempt := range 100 {
		candidateSeed := SeedToUint64(fmt.Sprintf("v525_corridor_detour_%d", attempt))
		selector := NewRNG(candidateSeed)
		widthStart := selector.IntN(len(widths))
		candidateWidth := widths[widthStart]
		candidateDoorA, candidateDoorB := doorA, doorB
		candidateDoorA.width, candidateDoorB.width = candidateWidth, candidateWidth
		candidates := orthogonalCorridorCandidates(rules, rooms, candidateDoorA, candidateDoorB, candidateWidth)
		if len(candidates) == 0 {
			continue
		}
		start := selector.IntN(len(candidates))
		blockers := append([]wallObstacle{blocker}, roomPerimeterWalls(rules, rooms, []roomDoor{candidateDoorA, candidateDoorB}, nil)...)
		if corridorRouteBlocked(candidates[start], playerRadius, blockers) {
			seed, width, variantStart = candidateSeed, candidateWidth, start
			foundBlockedFirstChoice = true
			break
		}
	}
	if !foundBlockedFirstChoice {
		t.Fatal("could not find a seeded first-choice corridor candidate blocked by fixture")
	}
	first, ok := corridorBetweenRooms(NewRNG(seed), rules, rooms, roomEdge{0, 1}, []wallObstacle{blocker})
	if !ok {
		t.Fatal("corridorBetweenRooms failed to route around the blocker")
	}
	second, ok := corridorBetweenRooms(NewRNG(seed), rules, rooms, roomEdge{0, 1}, []wallObstacle{blocker})
	if !ok {
		t.Fatal("repeat corridorBetweenRooms failed to route around the blocker")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different corridor routes:\nfirst:  %+v\nsecond: %+v", first, second)
	}
	if len(first.points) < 4 {
		t.Fatalf("blocker route has %d points, want a detour with at least four", len(first.points))
	}
	if first.width == width && first.variant == variantStart {
		t.Fatalf("blocked first seeded candidate %d was selected", variantStart)
	}
	if !isOrthogonalPath(first.points) {
		t.Fatalf("route is not orthogonal: %+v", first.points)
	}
	if corridorRouteBlocked(first.points, playerRadius, []wallObstacle{blocker}) {
		t.Fatalf("selected route intersects blocker: %+v", first.points)
	}
	if first.edge != (roomEdge{0, 1}) {
		t.Fatalf("route edge = %v, want [0 1]", first.edge)
	}
	if !containsFloat(rules.RoomCorridorPCG.CorridorWidths, first.width) {
		t.Fatalf("route width %v is not in configured choices %v", first.width, rules.RoomCorridorPCG.CorridorWidths)
	}
}

func TestRoomCorridorRouting_LayoutRoutesMatchSelectedEdges(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	for _, seed := range []string{"v525_route_edges_one", "v525_route_edges_two", "v525_route_edges_three"} {
		level, err := GenerateDungeonLevel(seed, -1, rules)
		if err != nil {
			t.Fatalf("generate %q: %v", seed, err)
		}
		if len(level.corridorEdges) == 0 {
			t.Fatalf("seed %q generated no corridor edges", seed)
		}
		if err := validateGeneratedCorridorRoutes(level); err != nil {
			t.Fatalf("seed %q generated invalid corridor routes: %v", seed, err)
		}
		if len(level.corridorRoutes) != len(level.corridorEdges) {
			t.Fatalf("seed %q has %d selected edges and %d routes", seed, len(level.corridorEdges), len(level.corridorRoutes))
		}
		for i, route := range level.corridorRoutes {
			if route.edge != normalizeRoomEdge(level.corridorEdges[i]) {
				t.Errorf("seed %q route %d edge = %v, selected edge = %v", seed, i, route.edge, level.corridorEdges[i])
			}
			if !containsFloat(rules.RoomCorridorPCG.CorridorWidths, route.width) {
				t.Errorf("seed %q route %d width %v is not configured", seed, i, route.width)
			}
			if len(route.points) < 2 || route.points[0] != route.doorA.center || route.points[len(route.points)-1] != route.doorB.center {
				t.Errorf("seed %q route %d endpoints do not match its selected doors: %+v", seed, i, route.points)
			}
		}
	}
}

func TestRoomCorridorRouting_ImpossibleLayoutFailsWithinConfiguredAttempts(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	rules.RoomCorridorPCG.MaxAttempts = 1
	out := generatedDungeonLevel{
		levelNum: -1,
		walls: append(perimeterWalls(rules.FloorSize, rules.WallThickness), wallObstacle{
			pos:         Vec2{X: rules.FloorSize.Width / 2, Y: rules.FloorSize.Height / 2},
			size:        Vec2{X: rules.FloorSize.Width + 10, Y: rules.FloorSize.Height + 10},
			source:      "test_blocker",
			shapeFamily: "rectangle",
			kind:        obstacleKindWall,
		}),
	}
	if err := placeRoomCorridorLayout("v525_impossible_layout", rules, &out); err == nil || !strings.Contains(err.Error(), "after 1 attempts") {
		t.Fatalf("impossible layout error = %v, want failure after the configured single attempt", err)
	}
	if len(out.corridorEdges) != 0 || len(out.corridorRoutes) != 0 {
		t.Fatalf("failed layout retained partial corridors: edges=%v routes=%v", out.corridorEdges, out.corridorRoutes)
	}
}

func containsFloat(values []float64, want float64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
