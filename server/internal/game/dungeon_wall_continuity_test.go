package game

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestWallContinuity_RoomShapesSealTheirPerimeters(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	nav := generatedDungeonNavigation(rules)
	for _, shape := range rules.RoomCorridorPCG.RoomShapes {
		t.Run(shape.ID, func(t *testing.T) {
			var cells [roomShapeCellCount]bool
			copy(cells[:], shape.Cells)
			room := makeDungeonRoom(Vec2{X: 40, Y: 40}, Vec2{X: 49, Y: 49}, false, shape.ID, cells)
			walls := roomPerimeterWalls(rules, []dungeonRoom{room}, nil, nil)

			for _, edge := range roomBoundaryEdges(room) {
				for _, coordinate := range []float64{edge.lo, (edge.lo + edge.hi) / 2, edge.hi} {
					probe := boundaryProbe(edge, coordinate)
					if !pointBlockedByWalls(probe, playerRadius, walls) {
						nearest, clearance := nearestWallAt(probe, walls)
						t.Errorf("open perimeter at %s edge %.2f..%.2f probe %.2f (%.2f, %.2f); nearest %s/%s clearance %.3f, player radius %.3f",
							edge.side, edge.lo, edge.hi, coordinate, probe.X, probe.Y, nearest.source, nearest.obstacleKind(), clearance, playerRadius)
					}
				}
			}

			blocked := buildDungeonBlockedGrid(nav, generatedDungeonLevel{walls: walls})
			center := roomCenter(room)
			for y := 0; y < roomShapeSide; y++ {
				for x := 0; x < roomShapeSide; x++ {
					min, max := roomShapeCellBounds(room, x, y)
					cellCenter := Vec2{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}
					reachable := generatedTargetReachableFromNav(nav, blocked.blocked, center, cellCenter)
					if roomShapeCellActive(room, x, y) && !reachable {
						t.Errorf("active cell (%d,%d) in %s room is blocked", x, y, shape.ID)
					}
					if !roomShapeCellActive(room, x, y) && reachable {
						t.Errorf("inactive cell (%d,%d) in %s room leaks through its boundary", x, y, shape.ID)
					}
				}
			}
		})
	}
}

func TestWallContinuity_ShapeCorridorDoorsStayOpen(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	nav := generatedDungeonNavigation(rules)
	for _, shape := range rules.RoomCorridorPCG.RoomShapes {
		t.Run(shape.ID, func(t *testing.T) {
			var cells [roomShapeCellCount]bool
			copy(cells[:], shape.Cells)
			rooms := []dungeonRoom{
				makeDungeonRoom(Vec2{X: 30, Y: 40}, Vec2{X: 39, Y: 49}, false, shape.ID, cells),
				makeDungeonRoom(Vec2{X: 60, Y: 40}, Vec2{X: 69, Y: 49}, false, shape.ID, cells),
			}
			route, ok := corridorBetweenRooms(NewRNG(SeedToUint64("wall_continuity_"+shape.ID)), rules, rooms, roomEdge{0, 1}, nil)
			if !ok || len(route.zones) == 0 {
				t.Fatalf("could not form a corridor between two %s rooms", shape.ID)
			}
			walls := roomPerimeterWalls(rules, rooms, []roomDoor{route.doorA, route.doorB}, nil)
			for _, door := range []roomDoor{route.doorA, route.doorB} {
				probe := doorWallProbe(rules, door)
				if pointBlockedByWalls(probe, playerRadius, walls) {
					nearest, clearance := nearestWallAt(probe, walls)
					t.Errorf("%s corridor door at (%.2f, %.2f) blocked by %s/%s at clearance %.3f, player radius %.3f",
						door.side, probe.X, probe.Y, nearest.source, nearest.obstacleKind(), clearance, playerRadius)
				}
			}

			blocked := buildDungeonBlockedGrid(nav, generatedDungeonLevel{walls: walls})
			if !generatedTargetReachableFromNav(nav, blocked.blocked, roomCenter(rooms[0]), roomCenter(rooms[1])) {
				t.Errorf("corridor between %s rooms is not traversable; door A=%+v door B=%+v route=%+v", shape.ID, route.doorA, route.doorB, route.points)
			}
		})
	}
}

func TestWallContinuity_SolidObstacleJoinsRemainBlocking(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	var cells [roomShapeCellCount]bool
	for i := range cells {
		cells[i] = true
	}
	room := makeDungeonRoom(Vec2{X: 20, Y: 20}, Vec2{X: 29, Y: 29}, false, "rectangle", cells)
	edge := roomBoundaryEdge{side: "north", fixed: room.innerMax.Y, lo: room.innerMin.X, hi: room.innerMax.X}
	door := roomDoor{roomIndex: 0, side: edge.side, center: Vec2{X: (edge.lo + edge.hi) / 2, Y: edge.fixed}}
	walls := roomPerimeterWalls(rules, []dungeonRoom{room}, []roomDoor{door}, nil)
	nav := generatedDungeonNavigation(rules)
	outside := Vec2{X: door.center.X, Y: door.center.Y + rules.WallThickness + playerRadius + 1}
	baselineBlocked := buildDungeonBlockedGrid(nav, generatedDungeonLevel{walls: walls})
	if !generatedTargetReachableFromNav(nav, baselineBlocked.blocked, roomCenter(room), outside) {
		t.Fatalf("doorway fixture is not navigable without the obstacle: corridor_width=%.2f player_radius=%.2f", defaultCorridorWidth(rules.RoomCorridorPCG), playerRadius)
	}
	// This solid obstacle meets the outside face of the north perimeter wall beside
	// an intended doorway. The join remains blocking without consuming doorway clearance.
	gapWidth := defaultCorridorWidth(rules.RoomCorridorPCG)
	join := wallObstacle{
		pos:    Vec2{X: door.center.X - gapWidth/2 - playerRadius - rules.WallThickness, Y: door.center.Y + rules.WallThickness*1.5},
		size:   Vec2{X: rules.WallThickness, Y: rules.WallThickness},
		source: "generated",
		kind:   obstacleKindRock,
	}
	walls = append(walls, join)
	if probe := doorWallProbe(rules, door); pointBlockedByWalls(probe, playerRadius, walls) {
		t.Fatalf("room-wall / solid-obstacle join blocked intended door at (%.2f, %.2f)", probe.X, probe.Y)
	}
	if !pointBlockedByWalls(join.pos, playerRadius, walls) || !obstacleBlocksMovement(join) {
		t.Fatal("solid obstacle at the wall jamb must block authoritative grounded movement")
	}
	contact := Vec2{X: join.pos.X, Y: door.center.Y + rules.WallThickness}
	if !pointBlockedByWalls(contact, playerRadius, walls) {
		nearest, clearance := nearestWallAt(contact, walls)
		t.Fatalf("room-wall / solid-obstacle contact at (%.2f, %.2f) has a traversable seam; nearest %s/%s clearance %.3f, player radius %.3f",
			contact.X, contact.Y, nearest.source, nearest.obstacleKind(), clearance, playerRadius)
	}
	blocked := buildDungeonBlockedGrid(nav, generatedDungeonLevel{walls: walls})
	if !generatedTargetReachableFromNav(nav, blocked.blocked, roomCenter(room), outside) {
		t.Fatal("adjacent solid obstacle blocked the intended corridor opening")
	}
}

func TestWallContinuity_RepresentativeGeneratedLayoutsAreStableAndReachable(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	for _, tc := range []struct {
		seed  string
		level int
	}{
		{seed: "wall_continuity_v522_v523_a", level: -1},
		{seed: "wall_continuity_v522_v523_b", level: -3},
		{seed: "wall_continuity_v522_v523_c", level: -7},
	} {
		t.Run(fmt.Sprintf("%s_level_%d", tc.seed, tc.level), func(t *testing.T) {
			a, err := GenerateDungeonLevel(tc.seed, tc.level, rules)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			b, err := GenerateDungeonLevel(tc.seed, tc.level, rules)
			if err != nil {
				t.Fatalf("repeat generate: %v", err)
			}
			if !reflect.DeepEqual(a, b) {
				t.Fatal("same seed and level produced different wall rectangles or generated output")
			}
			if err := validateGeneratedDungeonReachability(rules.RulesForLevel(tc.level), a); err != nil {
				t.Errorf("generated targets are not reachable: %v", err)
			}
			if !allRoomCentersReachable(rules.RulesForLevel(tc.level), a) {
				t.Error("not all generated room centers are mutually reachable")
			}
		})
	}
}

func boundaryProbe(edge roomBoundaryEdge, coordinate float64) Vec2 {
	if edge.side == "north" || edge.side == "south" {
		return Vec2{X: coordinate, Y: edge.fixed}
	}
	return Vec2{X: edge.fixed, Y: coordinate}
}

func doorWallProbe(rules DungeonGenerationRules, door roomDoor) Vec2 {
	probe := door.center
	step := rules.WallThickness / 2
	switch door.side {
	case "north":
		probe.Y += step
	case "south":
		probe.Y -= step
	case "east":
		probe.X += step
	case "west":
		probe.X -= step
	}
	return probe
}

func pointBlockedByWalls(point Vec2, radius float64, walls []wallObstacle) bool {
	for _, wall := range walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(point, radius, wall.pos, wall.size) {
			return true
		}
	}
	return false
}

func nearestWallAt(point Vec2, walls []wallObstacle) (wallObstacle, float64) {
	nearest := wallObstacle{}
	clearance := math.Inf(1)
	for _, wall := range walls {
		halfX, halfY := wall.size.X/2, wall.size.Y/2
		closestX := math.Max(wall.pos.X-halfX, math.Min(point.X, wall.pos.X+halfX))
		closestY := math.Max(wall.pos.Y-halfY, math.Min(point.Y, wall.pos.Y+halfY))
		distance := distance(point, Vec2{X: closestX, Y: closestY})
		if distance < clearance {
			nearest, clearance = wall, distance
		}
	}
	return nearest, clearance
}
