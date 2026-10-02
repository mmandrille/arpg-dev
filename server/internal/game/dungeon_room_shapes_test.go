package game

import (
	"reflect"
	"testing"
)

func TestDungeonRoomShapes_RectangleOnlyPreservesLegacyFootprintAndRNG(t *testing.T) {
	rules := loadRules(t)
	for i := range rules.DungeonGeneration.RoomCorridorPCG.RoomShapes {
		shape := &rules.DungeonGeneration.RoomCorridorPCG.RoomShapes[i]
		shape.Weight = 0
		if shape.ID == "rectangle" {
			shape.Weight = 1
		}
	}

	selector := NewRNG(SeedToUint64("rectangle_only_shape_stream"))
	control := NewRNG(SeedToUint64("rectangle_only_shape_stream"))
	id, cells, ok := selectDungeonRoomShape(selector, rules.DungeonGeneration.RoomCorridorPCG.RoomShapes)
	if !ok || id != "rectangle" {
		t.Fatalf("only positive shape selection = %q, %t; want rectangle", id, ok)
	}
	if got, want := selector.IntN(1<<16), control.IntN(1<<16); got != want {
		t.Fatalf("single-shape selection consumed the room-generation RNG stream: next draw = %d, want %d", got, want)
	}
	for i, active := range cells {
		if !active {
			t.Fatalf("rectangle shape cell %d is inactive", i)
		}
	}

	room := makeDungeonRoom(Vec2{X: 10, Y: 10}, Vec2{X: 22, Y: 19}, false, id, cells)
	if edges := roomBoundaryEdges(room); len(edges) != 4 {
		t.Fatalf("rectangle-only footprint has %d boundary edges, want four outer sides", len(edges))
	}
	walls := roomPerimeterWalls(rules.DungeonGeneration, []dungeonRoom{room}, nil, nil)
	if len(walls) != 4 {
		t.Fatalf("rectangle-only footprint emitted %d room-wall segments, want one per outer side", len(walls))
	}
}

func TestDungeonRoomShapes_ForcedFootprintsStayBoundedAndReachable(t *testing.T) {
	for _, shapeID := range requiredDungeonRoomShapeIDs {
		t.Run(shapeID, func(t *testing.T) {
			rules := loadRules(t)
			rules.DungeonGeneration.RoomCorridorPCG.RoomCount = IntRange{Min: 2, Max: 2}
			rules.DungeonGeneration.RoomCorridorPCG.HubRoomEnabled = false
			rules.DungeonGeneration.RoomCorridorPCG.LoopEdgeCount = IntRange{Min: 0, Max: 0}
			// This slice owns footprint geometry; semantic placement constraints are
			// exercised independently by v528's room-role tests.
			rules.DungeonGeneration.RoomCorridorPCG.RoomRoles.Enabled = false
			for i := range rules.DungeonGeneration.RoomCorridorPCG.RoomShapes {
				shape := &rules.DungeonGeneration.RoomCorridorPCG.RoomShapes[i]
				shape.Weight = 0
				if shape.ID == shapeID {
					shape.Weight = 1
				}
			}

			level, err := GenerateDungeonLevel("v523_shape_"+shapeID, -1, rules.DungeonGeneration)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if len(level.rooms) != 2 {
				t.Fatalf("room count = %d, want 2", len(level.rooms))
			}
			foundShape := false
			for _, room := range level.rooms {
				foundShape = foundShape || room.shapeID == shapeID
				assertRoomBounds(t, rules.DungeonGeneration, room)
			}
			for _, wall := range level.walls {
				if wall.source != "room_wall" {
					continue
				}
				if wall.pos.X-wall.size.X/2 < 0 || wall.pos.Y-wall.size.Y/2 < 0 ||
					wall.pos.X+wall.size.X/2 > rules.DungeonGeneration.FloorSize.Width ||
					wall.pos.Y+wall.size.Y/2 > rules.DungeonGeneration.FloorSize.Height {
					t.Errorf("room wall exceeds floor bounds: %+v", wall)
				}
			}
			if !foundShape {
				t.Fatalf("generated rooms do not include forced shape %q", shapeID)
			}
			assertAnchorsInsideRooms(t, "v523_shape_"+shapeID, -1, level)

			nav := generatedDungeonNavigation(rules.DungeonGeneration)
			blocked := buildDungeonBlockedGrid(nav, level)
			for _, room := range level.rooms {
				center := roomCenter(room)
				for y := 0; y < roomShapeSide; y++ {
					for x := 0; x < roomShapeSide; x++ {
						if !roomShapeCellActive(room, x, y) {
							continue
						}
						min, max := roomShapeCellBounds(room, x, y)
						cellCenter := Vec2{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}
						if !generatedTargetReachableFromNav(nav, blocked.blocked, center, cellCenter) {
							t.Errorf("room %q cell (%d,%d) is not reachable from its center", room.shapeID, x, y)
						}
					}
				}
			}
		})
	}
}

func TestDungeonRoomShapes_Deterministic(t *testing.T) {
	rules := loadRules(t)
	// This test owns shape geometry determinism. Role-aware encounter placement
	// has its own tests and should not make a narrow shape fail this assertion.
	rules.DungeonGeneration.RoomCorridorPCG.RoomRoles.Enabled = false
	for _, shapeID := range requiredDungeonRoomShapeIDs {
		for i := range rules.DungeonGeneration.RoomCorridorPCG.RoomShapes {
			shape := &rules.DungeonGeneration.RoomCorridorPCG.RoomShapes[i]
			shape.Weight = 0
			if shape.ID == shapeID {
				shape.Weight = 1
			}
		}
		seed := "v523_shape_repeat_" + shapeID
		a, errA := GenerateDungeonLevel(seed, -2, rules.DungeonGeneration)
		b, errB := GenerateDungeonLevel(seed, -2, rules.DungeonGeneration)
		if errA != nil || errB != nil {
			t.Fatalf("shape %q: %v / %v", shapeID, errA, errB)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("shape %q: repeated generation produced different output", shapeID)
		}
	}
}

func TestDungeonRoomShapes_RuleValidation(t *testing.T) {
	rules := loadRules(t)
	base := rules.DungeonGeneration.RoomCorridorPCG
	tests := []struct {
		name   string
		mutate func(*RoomCorridorPCGRules)
	}{
		{
			name: "duplicate id",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].ID = r.RoomShapes[0].ID
			},
		},
		{
			name: "disconnected footprint",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].Cells = []bool{true, false, false, false, false, false, false, false, true}
			},
		},
		{
			name: "undersized footprint",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].Cells = []bool{true, true, false, false, false, false, false, false, false}
			},
		},
		{
			name: "zero total weight",
			mutate: func(r *RoomCorridorPCGRules) {
				for i := range r.RoomShapes {
					r.RoomShapes[i].Weight = 0
				}
			},
		},
		{
			name: "malformed mask",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].Cells = []bool{true, true, true}
			},
		},
		{
			name: "unknown id",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].ID = "donut"
			},
		},
		{
			name: "out of range weight",
			mutate: func(r *RoomCorridorPCGRules) {
				r.RoomShapes[1].Weight = 101
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := base
			candidate.RoomShapes = append([]DungeonRoomShapeRule(nil), base.RoomShapes...)
			for i := range candidate.RoomShapes {
				candidate.RoomShapes[i].Cells = append([]bool(nil), base.RoomShapes[i].Cells...)
			}
			tc.mutate(&candidate)
			if err := validateRoomCorridorPCGRules(candidate); err == nil {
				t.Fatal("expected invalid room-shape rules to be rejected")
			}
		})
	}
}

func TestDungeonRoomShapes_ClearanceUsesConnectedFootprint(t *testing.T) {
	cross := roomForTestShape("cross", [roomShapeCellCount]bool{false, true, false, true, true, true, false, true, false})
	if !roomContainsCircle(cross, Vec2{X: 6, Y: 4.5}, 0.9) {
		t.Fatal("clearance circle spanning adjacent active cells should fit the cross footprint")
	}
	arena := roomForTestShape("arena", [roomShapeCellCount]bool{true, true, true, true, false, true, true, true, true})
	if roomContainsCircle(arena, Vec2{X: 4.5, Y: 4.5}, 0) {
		t.Fatal("arena center void must not be treated as walkable room interior")
	}
}

func roomForTestShape(id string, cells [roomShapeCellCount]bool) dungeonRoom {
	return makeDungeonRoom(Vec2{X: 0, Y: 0}, Vec2{X: 9, Y: 9}, false, id, cells)
}

func assertRoomBounds(t *testing.T, rules DungeonGenerationRules, room dungeonRoom) {
	t.Helper()
	margin := rules.RoomCorridorPCG.MarginFromPerimeter
	thickness := rules.WallThickness / 2
	if room.innerMin.X-thickness < margin-1e-6 || room.innerMin.Y-thickness < margin-1e-6 ||
		room.innerMax.X+thickness > rules.FloorSize.Width-margin+1e-6 ||
		room.innerMax.Y+thickness > rules.FloorSize.Height-margin+1e-6 {
		t.Fatalf("room %q bounds %+v..%+v exceed floor margins %+v", room.shapeID, room.innerMin, room.innerMax, rules.FloorSize)
	}
}
