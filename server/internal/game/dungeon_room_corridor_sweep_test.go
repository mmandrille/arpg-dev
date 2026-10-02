package game

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

// Seed/level pairs where pre-placed anchor geometry made every room-corridor attempt fail before
// v472, plus two rooms-first reachability regressions. Pinned to guard the new placement flow.
var roomCorridorAnchorConflictCases = []struct {
	seed  string
	level int
}{
	// Anchor room conflicts from v472's original audit.
	{"audit-09", -1},
	{"audit-13", -3},
	{"audit-17", -7},
	{"audit-20", -3},
	// Perimeter-margin anchors from v472's wider sweep.
	{"audit-102", -4},
	{"audit-111", -8},
	{"audit-168", -9},
	{"audit-181", -3},
	{"audit-137", -2}, // room-contained down stair was unreachable from the up stair
	{"audit-572", -8}, // room-contained quest chest was unreachable from the up stair
}

func TestRoomCorridorLayout_AnchorConflictSeedsGenerate(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range roomCorridorAnchorConflictCases {
		level, err := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if err != nil {
			t.Errorf("seed %s level %d: %v", tc.seed, tc.level, err)
			continue
		}
		if err := validateGeneratedDungeonReachability(rules.DungeonGeneration.RulesForLevel(tc.level), level); err != nil {
			t.Errorf("seed %s level %d: generated target reachability: %v", tc.seed, tc.level, err)
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
				if !isBossFloor(levelNum, rules.DungeonGeneration) {
					layoutOnly := generatedDungeonLevel{
						levelNum: levelNum,
						walls:    perimeterWalls(rules.DungeonGeneration.FloorSize, rules.DungeonGeneration.WallThickness),
					}
					if !tryRoomCorridorLayoutPass(seed, rules.DungeonGeneration, &layoutOnly) {
						t.Errorf("level %d: room-corridor layout could not be generated", levelNum)
					}
				}
				level, err := GenerateDungeonLevel(seed, levelNum, rules.DungeonGeneration)
				if err != nil {
					t.Errorf("level %d: %v", levelNum, err)
					continue
				}
				if !isBossFloor(levelNum, rules.DungeonGeneration) {
					assertAnchorsInsideRooms(t, seed, levelNum, level)
				}
			}
		})
	}
}

func TestRoomCorridorLayout_Deterministic(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range roomCorridorAnchorConflictCases {
		a, errA := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		b, errB := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if errA != nil || errB != nil {
			t.Fatalf("seed %s level %d: %v / %v", tc.seed, tc.level, errA, errB)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("seed %s level %d: repeated generation produced different output", tc.seed, tc.level)
		}
	}
}

func TestRoomCorridorLayout_PreLayoutAnchorsInsideRooms(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.ChestPlacement.Enabled = true
	rules.DungeonGeneration.ChestPlacement.ChanceWeight = 100
	rules.DungeonGeneration.ChestPlacement.NoChestWeight = 0
	seed := "rooms_first_guarded_chest"
	level, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate guarded chest floor: %v", err)
	}
	if len(level.chests) == 0 {
		t.Fatal("expected forced guarded chest")
	}
	assertAnchorsInsideRooms(t, seed, -1, level)

	for i := range 100 {
		seed = fmt.Sprintf("rooms_first_quest_chest_%d", i)
		if !dungeonLevelHasRandomQuestReward(seed, -1, rules.DungeonGeneration) {
			continue
		}
		level, err = GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("generate quest chest floor %s: %v", seed, err)
		}
		foundQuestReward := false
		for _, chest := range level.chests {
			foundQuestReward = foundQuestReward || chest.questReward
		}
		if !foundQuestReward {
			t.Fatalf("seed %s is quest-reward eligible but no quest chest was placed", seed)
		}
		assertAnchorsInsideRooms(t, seed, -1, level)
		return
	}
	t.Fatal("could not find deterministic quest-reward seed in search window")
}

// assertAnchorsInsideRooms checks stairs, teleporters, and pre-layout chests. The elite-objective
// chest is placed after the layout and may sit in a corridor.
func assertAnchorsInsideRooms(t *testing.T, seed string, levelNum int, level generatedDungeonLevel) {
	t.Helper()
	anchors := generatedAnchorPoints(generatedDungeonLevel{stairs: level.stairs, teleporters: level.teleporters})
	for _, chest := range level.chests {
		if !chest.eliteObjective {
			anchors = append(anchors, chest.pos)
		}
	}
	for _, anchor := range anchors {
		if !pointInsideAnyRoom(anchor, level.rooms, playerRadius+0.1) {
			t.Errorf("seed %s level %d: anchor %.1f,%.1f is outside every room", seed, levelNum, anchor.X, anchor.Y)
		}
	}
}
