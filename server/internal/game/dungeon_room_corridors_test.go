package game

import "testing"

func TestPlaceRoomCorridorLayout_RoomWallsPresent(t *testing.T) {
	rules := loadRules(t)
	if !rules.DungeonGeneration.RoomCorridorPCG.Enabled {
		t.Fatal("expected room_corridor_pcg enabled in rules")
	}
	level, err := GenerateDungeonLevel("room_corridor_walls", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	roomWallCount := 0
	for _, w := range level.walls {
		if w.source == "room_wall" {
			roomWallCount++
		}
		if w.source == "room_divider" {
			t.Fatalf("unexpected room_divider when PCG enabled: %+v", w)
		}
	}
	if roomWallCount < 8 {
		t.Fatalf("room_wall count = %d, want at least 8", roomWallCount)
	}
	if len(level.rooms) < rules.DungeonGeneration.RoomCorridorPCG.RoomCount.Min {
		t.Fatalf("room count = %d, want at least %d", len(level.rooms), rules.DungeonGeneration.RoomCorridorPCG.RoomCount.Min)
	}
}

func TestPlaceRoomCorridorLayout_NoInteriorScatter(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("room_corridor_no_scatter", -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, w := range level.walls {
		if w.source == "generated" && w.shapeFamily != "" && w.obstacleKind() != obstacleKindWater && w.obstacleKind() != obstacleKindHole {
			t.Fatalf("unexpected interior scatter wall: %+v", w)
		}
	}
}

func TestPlaceRoomCorridorLayout_Reachability(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range []struct {
		seed  string
		level int
	}{
		{"room_corridor_reach_a", -1},
		{"room_corridor_reach_b", -2},
		{"room_corridor_reach_c", -3},
		{"room_corridor_reach_d", -4},
	} {
		_, err := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if err != nil {
			t.Errorf("seed %s level %d: %v", tc.seed, tc.level, err)
		}
	}
}

func TestPlaceRoomCorridorLayout_BossFloorUnaffected(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("room_corridor_boss", -5, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor: %v", err)
	}
	for _, w := range level.walls {
		if w.source == "room_wall" || w.source == "room_divider" {
			t.Fatalf("boss floor has structured room wall: %+v", w)
		}
	}
}

func TestPlaceRoomCorridorLayout_LegacyDividersWhenDisabled(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.RoomCorridorPCG.Enabled = false
	rules.DungeonGeneration.RoomLayout.Enabled = true
	level, err := GenerateDungeonLevel("room_corridor_legacy", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	found := false
	for _, w := range level.walls {
		if w.source == "room_divider" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected room_divider walls when legacy layout enabled")
	}
}
