package game

import (
	"fmt"
	"testing"
)

func TestPlaceRoomLayout_DividersPresent(t *testing.T) {
	rules := loadRules(t)
	if rules.DungeonGeneration.RoomCorridorPCG.Enabled {
		t.Skip("room_corridor_pcg enabled; see TestPlaceRoomCorridorLayout_RoomWallsPresent")
	}
	level, err := GenerateDungeonLevel("room_layout_test_dividers", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	count := 0
	for _, w := range level.walls {
		if w.source == "room_divider" {
			count++
		}
	}
	if count < 2 {
		t.Fatalf("room_divider wall count = %d, want at least 2", count)
	}
}

func TestPlaceRoomLayout_CorridorGapWidth(t *testing.T) {
	rules := loadRules(t)
	if rules.DungeonGeneration.RoomCorridorPCG.Enabled {
		t.Skip("room_corridor_pcg enabled; corridor gaps covered by room_wall doors")
	}
	corridorWidth := rules.DungeonGeneration.RoomLayout.CorridorWidth
	seeds := []string{"room_layout_gap_a", "room_layout_gap_b", "room_layout_gap_c"}
	for _, seed := range seeds {
		level, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("seed %s generate: %v", seed, err)
		}
		type yGroup struct {
			minX, maxX float64
		}
		byY := map[float64][]yGroup{}
		for _, w := range level.walls {
			if w.source != "room_divider" {
				continue
			}
			if w.size.Y < w.size.X {
				y := w.pos.Y
				lo := w.pos.X - w.size.X/2
				hi := w.pos.X + w.size.X/2
				byY[y] = append(byY[y], yGroup{lo, hi})
			}
		}
		for y, segs := range byY {
			if len(segs) < 2 {
				continue
			}
			for i := 0; i < len(segs); i++ {
				for j := i + 1; j < len(segs); j++ {
					if segs[j].minX < segs[i].minX {
						segs[i], segs[j] = segs[j], segs[i]
					}
				}
			}
			for i := 0; i+1 < len(segs); i++ {
				gap := segs[i+1].minX - segs[i].maxX
				if gap < corridorWidth-0.001 {
					t.Errorf("seed %s Y=%.1f: gap between segments = %.2f, want >= %.2f", seed, y, gap, corridorWidth)
				}
			}
		}
	}
}

func TestPlaceRoomLayout_Reachability(t *testing.T) {
	rules := loadRules(t)
	for _, tc := range []struct {
		seed  string
		level int
	}{
		{"reachability_test_a", -1},
		{"reachability_test_b", -2},
		{"reachability_test_c", -3},
	} {
		_, err := GenerateDungeonLevel(tc.seed, tc.level, rules.DungeonGeneration)
		if err != nil {
			t.Errorf("seed %s level %d: %v", tc.seed, tc.level, err)
		}
	}
}

func TestPlaceRoomLayout_Disabled(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.RoomCorridorPCG.Enabled = false
	rules.DungeonGeneration.RoomLayout.Enabled = false
	rules.DungeonGeneration.ObstacleGeneration.Enabled = false
	level, err := GenerateDungeonLevel("room_layout_disabled", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, w := range level.walls {
		if w.source == "room_divider" || w.source == "room_wall" {
			t.Fatalf("found structured room wall when disabled: %+v", w)
		}
	}
}

func TestPlaceRoomLayout_BossFloorUnaffected(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("boss_floor_no_rooms", -5, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor: %v", err)
	}
	for _, w := range level.walls {
		if w.source == "room_divider" || w.source == "room_wall" {
			t.Fatalf("boss floor has structured room wall: %+v", w)
		}
	}
}

func TestFinalizeGeneratedDungeonLevel_MonsterPresent(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("finalize_helper_test", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(level.monsters) == 0 {
		t.Fatal("expected at least one monster")
	}
	if len(level.stairs) == 0 {
		t.Fatal("expected stairs")
	}
	start := generatedReachabilityStart(rules.DungeonGeneration.RulesForLevel(-1), level)
	for _, target := range generatedReachabilityTargets(level) {
		reachable := generatedTargetReachableFrom(rules.DungeonGeneration.RulesForLevel(-1), level, start, target.pos)
		if target.kind == woodenDoorDefID {
			rulesForLevel := rules.DungeonGeneration.RulesForLevel(-1)
			nav := generatedDungeonNavigation(rulesForLevel)
			blockedGrid := buildDungeonBlockedGrid(nav, level)
			reachable = generatedDoorReachableFromNav(nav, blockedGrid.blocked, start, target.pos)
		}
		if !reachable {
			t.Errorf("target %s at %+v unreachable from %+v", target.kind, target.pos, start)
		}
	}
}

func TestRoomSpawnAwareness_MonstersAvoidCorridors(t *testing.T) {
	rules := loadRules(t)
	packRadius := rules.DungeonGeneration.MonsterPlacement.PackMemberRadius
	seeds := []string{"room_spawn_a", "room_spawn_b", "room_spawn_c"}
	for _, seed := range seeds {
		level, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("seed %s generate: %v", seed, err)
		}
		if len(level.corridorZones) == 0 {
			t.Fatalf("seed %s: expected corridor zones when room layout enabled", seed)
		}
		for _, monster := range level.monsters {
			if generatedPositionInCorridorZone(monster.pos, packRadius, level) {
				t.Fatalf("seed %s monster %s at %+v overlaps corridor zone", seed, monster.defID, monster.pos)
			}
		}
	}
}

func TestRoomSpawnAwareness_EliteChestClustersNearLeader(t *testing.T) {
	rules := forceEliteObjectiveGenerationRules(t)
	clusterRadius := rules.DungeonGeneration.EliteObjective.RoomClusterRadius
	// The objective chest is optional per layout (rooms may lack chest clearance), so the contract is:
	// some seed in the sweep reserves it, and every reserved chest stays near its leader and off corridors.
	reserved := 0
	for i := range 8 {
		seed := fmt.Sprintf("room_spawn_elite_cluster_%d", i)
		level, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("generate %s: %v", seed, err)
		}
		leaderPos, ok := elitePackLeaderPosition(level)
		if !ok {
			continue
		}
		var objective *generatedChest
		for j := range level.chests {
			if level.chests[j].eliteObjective {
				objective = &level.chests[j]
				break
			}
		}
		if objective == nil {
			continue
		}
		reserved++
		if generatedPositionInCorridorZone(objective.pos, rules.DungeonGeneration.MonsterPlacement.PackMemberRadius, level) {
			t.Fatalf("%s: elite objective chest at %+v overlaps corridor", seed, objective.pos)
		}
		if distance(objective.pos, leaderPos) > clusterRadius {
			t.Fatalf("%s: elite chest distance %.2f from leader at %+v, want <= %.2f", seed, distance(objective.pos, leaderPos), leaderPos, clusterRadius)
		}
	}
	if reserved == 0 {
		t.Fatal("no seed in the sweep reserved an elite objective chest")
	}
}
