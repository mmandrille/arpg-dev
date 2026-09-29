package game

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

// Seed/level pairs where anchor geometry (stairs, teleporter, chests placed before rooms) made
// every room-corridor attempt fail before v472. Pinned so the fallback cannot silently regress.
var roomCorridorAnchorConflictCases = []struct {
	seed  string
	level int
}{
	// Anchor rooms could not be placed at all.
	{"audit-09", -1}, // spawn room is fixed; nearby stair is its own cluster but cannot fit beside it
	{"audit-13", -3}, // pair-cluster room blocks a floor-edge anchor on both axes
	{"audit-17", -7}, // three single anchors ~15 apart: separate rooms need edge-exact placement
	{"audit-20", -3}, // one chained cluster taller than room_size_max
	// Anchor on the margin_from_perimeter line: its room wall traps it against the floor edge.
	{"audit-102", -4},
	{"audit-111", -8},
	{"audit-168", -9},
	{"audit-181", -3},
}

func TestRoomCorridorLayout_AnchorConflictSeedsGenerate(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range roomCorridorAnchorConflictCases {
		level, err := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if err != nil {
			t.Errorf("seed %s level %d: %v", tc.seed, tc.level, err)
			continue
		}
		assertAnchorsInsideRooms(t, tc.seed, tc.level, level)
	}
}

// TestRoomCorridorLayout_SeedSweepAlwaysGenerates is the progression-blocker gate: every
// generated floor for every seed must build, because Sim.ensureDungeonLevel has no recovery path.
// Full generation costs ~0.2s per floor, so the default sweep is sized for CI (seeds run in
// parallel); set ARPG_DUNGEON_SWEEP_SEEDS to widen it locally.
func TestRoomCorridorLayout_SeedSweepAlwaysGenerates(t *testing.T) {
	rules := loadRules(t)
	seeds := 24
	if raw := os.Getenv("ARPG_DUNGEON_SWEEP_SEEDS"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			t.Fatalf("ARPG_DUNGEON_SWEEP_SEEDS=%q: want a positive integer", raw)
		}
		seeds = n
	}
	const deepest = -10
	for i := range seeds {
		seed := fmt.Sprintf("audit-%02d", i)
		t.Run(seed, func(t *testing.T) {
			t.Parallel()
			for levelNum := -1; levelNum >= deepest; levelNum-- {
				if _, err := GenerateDungeonLevel(seed, levelNum, rules.DungeonGeneration); err != nil {
					t.Errorf("level %d: %v", levelNum, err)
				}
			}
		})
	}
}

func TestRoomCorridorLayout_AnchorFallbackDeterministic(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range roomCorridorAnchorConflictCases {
		a, errA := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		b, errB := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if errA != nil || errB != nil {
			t.Fatalf("seed %s level %d: %v / %v", tc.seed, tc.level, errA, errB)
		}
		if fmt.Sprint(a.rooms) != fmt.Sprint(b.rooms) || fmt.Sprint(a.walls) != fmt.Sprint(b.walls) {
			t.Fatalf("seed %s level %d: repeated generation produced different layouts", tc.seed, tc.level)
		}
	}
}

// assertAnchorsInsideRooms checks the pre-layout anchors that must be wrapped by a room. Chests
// are skipped: the elite-objective chest is placed after the layout and may sit in a corridor.
func assertAnchorsInsideRooms(t *testing.T, seed string, levelNum int, level generatedDungeonLevel) {
	t.Helper()
	for _, anchor := range generatedAnchorPoints(generatedDungeonLevel{stairs: level.stairs, teleporters: level.teleporters}) {
		if !pointInsideAnyRoom(anchor, level.rooms, 0) {
			t.Errorf("seed %s level %d: anchor %.1f,%.1f is outside every room", seed, levelNum, anchor.X, anchor.Y)
		}
	}
}
