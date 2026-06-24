package game

import "testing"

func TestPlaceRoomLayout_DividersPresent(t *testing.T) {
	rules := loadRules(t)
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
	corridorWidth := rules.DungeonGeneration.RoomLayout.CorridorWidth
	seeds := []string{"room_layout_gap_a", "room_layout_gap_b", "room_layout_gap_c"}
	for _, seed := range seeds {
		level, err := GenerateDungeonLevel(seed, -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("seed %s generate: %v", seed, err)
		}
		// For each pair of room_divider walls at the same Y (horizontal), verify
		// consecutive X extents leave at least corridorWidth gap.
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
			// Sort by minX
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
	rules.DungeonGeneration.RoomLayout.Enabled = false
	level, err := GenerateDungeonLevel("room_layout_disabled", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, w := range level.walls {
		if w.source == "room_divider" {
			t.Fatalf("found room_divider wall when disabled: %+v", w)
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
		if w.source == "room_divider" {
			t.Fatalf("boss floor has room_divider wall: %+v", w)
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
		if !generatedTargetReachableFrom(rules.DungeonGeneration.RulesForLevel(-1), level, start, target.pos) {
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
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	rules.DungeonGeneration.ChestPlacement.Enabled = false
	clusterRadius := rules.DungeonGeneration.EliteObjective.RoomClusterRadius
	level, err := GenerateDungeonLevel("room_spawn_elite_cluster", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	leaderPos, ok := elitePackLeaderPosition(level)
	if !ok {
		t.Fatal("expected elite pack leader")
	}
	var objective *generatedChest
	for i := range level.chests {
		if level.chests[i].eliteObjective {
			objective = &level.chests[i]
			break
		}
	}
	if objective == nil {
		t.Fatalf("missing elite objective chest: %+v", level.chests)
	}
	if generatedPositionInCorridorZone(objective.pos, rules.DungeonGeneration.MonsterPlacement.PackMemberRadius, level) {
		t.Fatalf("elite objective chest at %+v overlaps corridor", objective.pos)
	}
	if distance(objective.pos, leaderPos) > clusterRadius {
		t.Fatalf("elite chest distance %.2f from leader at %+v, want <= %.2f", distance(objective.pos, leaderPos), leaderPos, clusterRadius)
	}
}
