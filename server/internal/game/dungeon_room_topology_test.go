package game

import (
	"reflect"
	"testing"
)

func TestRoomConnectionEdges_EnforcesConfiguredMotifs(t *testing.T) {
	tests := []struct {
		name       string
		roomCount  int
		hubEnabled bool
		hubDegree  IntRange
		branches   IntRange
		loops      IntRange
	}{
		{
			name:       "hub with a non-hub branch and one loop",
			roomCount:  6,
			hubEnabled: true,
			hubDegree:  IntRange{Min: 2, Max: 2},
			branches:   IntRange{Min: 1, Max: 1},
			loops:      IntRange{Min: 1, Max: 1},
		},
		{
			name:       "hub and loop without a non-hub branch",
			roomCount:  4,
			hubEnabled: true,
			hubDegree:  IntRange{Min: 3, Max: 3},
			branches:   IntRange{Min: 0, Max: 0},
			loops:      IntRange{Min: 1, Max: 1},
		},
		{
			name:      "branch motif with no hub",
			roomCount: 4,
			branches:  IntRange{Min: 1, Max: 1},
			loops:     IntRange{Min: 0, Max: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := RoomCorridorPCGRules{
				HubRoomEnabled:      tt.hubEnabled,
				HubDegree:           tt.hubDegree,
				BranchJunctionCount: tt.branches,
				LoopEdgeCount:       tt.loops,
			}
			rooms := make([]dungeonRoom, tt.roomCount)
			for i := range rooms {
				rooms[i].center = Vec2{X: float64(i), Y: float64(i % 2)}
			}
			if tt.hubEnabled {
				rooms[0].isHub = true
			}

			for seed := uint64(0); seed < 32; seed++ {
				rng := NewRNG(seed)
				edges := roomConnectionEdges(rng, rooms, rules)
				if edges == nil {
					t.Fatal("room graph was not constructed")
				}
				assertRoomGraphSemantics(t, rooms, edges, tt.hubEnabled, tt.hubDegree, tt.branches, tt.loops)

				again := roomConnectionEdges(NewRNG(seed), rooms, rules)
				if !reflect.DeepEqual(edges, again) {
					t.Fatalf("same seed produced different room edges: first=%v second=%v", edges, again)
				}
			}
		})
	}
}

func assertRoomGraphSemantics(t *testing.T, rooms []dungeonRoom, edges []roomEdge, hubEnabled bool, hubDegree, branchCount, loops IntRange) {
	t.Helper()
	degree := make([]int, len(rooms))
	adjacency := make([][]int, len(rooms))
	seen := make(map[roomEdge]struct{}, len(edges))
	for _, raw := range edges {
		edge := normalizeRoomEdge(raw)
		if edge[0] < 0 || edge[1] >= len(rooms) || edge[0] == edge[1] {
			t.Fatalf("invalid room edge %v for %d rooms", raw, len(rooms))
		}
		if _, ok := seen[edge]; ok {
			t.Fatalf("duplicate room edge %v", edge)
		}
		seen[edge] = struct{}{}
		degree[edge[0]]++
		degree[edge[1]]++
		adjacency[edge[0]] = append(adjacency[edge[0]], edge[1])
		adjacency[edge[1]] = append(adjacency[edge[1]], edge[0])
	}

	visited := make([]bool, len(rooms))
	queue := []int{0}
	visited[0] = true
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacency[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	for i, connected := range visited {
		if !connected {
			t.Fatalf("room %d is disconnected", i)
		}
	}

	if hubEnabled {
		hubIndex := -1
		for i := range rooms {
			if rooms[i].isHub {
				if hubIndex >= 0 {
					t.Fatal("fixture has multiple hubs")
				}
				hubIndex = i
			}
		}
		if hubIndex < 0 || degree[hubIndex] < hubDegree.Min || degree[hubIndex] > hubDegree.Max {
			t.Fatalf("hub degree = %d, want range [%d,%d]", degreeAt(degree, hubIndex), hubDegree.Min, hubDegree.Max)
		}
	}
	branches := 0
	for i := range rooms {
		if rooms[i].isHub {
			continue
		}
		if degree[i] >= 3 {
			branches++
		}
	}
	if branches < branchCount.Min || branches > branchCount.Max {
		t.Fatalf("non-hub branch junctions = %d, want range [%d,%d]", branches, branchCount.Min, branchCount.Max)
	}
	cycleRank := len(edges) - len(rooms) + 1
	if cycleRank < loops.Min || cycleRank > loops.Max {
		t.Fatalf("cycle rank = %d, want range [%d,%d]", cycleRank, loops.Min, loops.Max)
	}
}

func degreeAt(degrees []int, index int) int {
	if index < 0 || index >= len(degrees) {
		return 0
	}
	return degrees[index]
}

func TestRoomConnectionEdges_RejectsInfeasibleMotifCombination(t *testing.T) {
	rules := RoomCorridorPCGRules{
		HubRoomEnabled:      true,
		HubDegree:           IntRange{Min: 3, Max: 3},
		BranchJunctionCount: IntRange{Min: 1, Max: 1},
		LoopEdgeCount:       IntRange{Min: 0, Max: 0},
	}
	rooms := make([]dungeonRoom, 4)
	rooms[0].isHub = true
	if edges := roomConnectionEdges(NewRNG(9), rooms, rules); edges != nil {
		t.Fatalf("infeasible four-room hub+branch graph produced edges: %v", edges)
	}
}

func TestRoomWallGapsMergeOverlappingDoors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		walls []wallObstacle
		want  [][2]float64
	}{
		{
			name:  "horizontal",
			walls: horizontalRoomWall(0, 20, 5, 1, []float64{8, 10}, 4, true),
			want:  [][2]float64{{0, 6}, {12, 20}},
		},
		{
			name:  "vertical",
			walls: verticalRoomWall(0, 20, 5, 1, []float64{8, 10}, 4, false),
			want:  [][2]float64{{0, 6}, {12, 20}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.walls) != len(tc.want) {
				t.Fatalf("wall segments = %d, want %d: %+v", len(tc.walls), len(tc.want), tc.walls)
			}
			for i, wall := range tc.walls {
				lo, hi := wall.pos.X-wall.size.X/2, wall.pos.X+wall.size.X/2
				if tc.name == "vertical" {
					lo, hi = wall.pos.Y-wall.size.Y/2, wall.pos.Y+wall.size.Y/2
				}
				if got := [2]float64{lo, hi}; got != tc.want[i] {
					t.Errorf("wall %d span = %v, want %v", i, got, tc.want[i])
				}
				if wall.source != "room_wall" || wall.size.X <= 0 || wall.size.Y <= 0 {
					t.Errorf("invalid room wall segment: %+v", wall)
				}
			}
		})
	}
}

func TestValidateRoomCorridorPCGRules_RejectsImpossibleTopologyRanges(t *testing.T) {
	base := loadRules(t).DungeonGeneration.RoomCorridorPCG
	tests := []struct {
		name   string
		mutate func(*RoomCorridorPCGRules)
	}{
		{
			name: "hub degree exceeds room count",
			mutate: func(r *RoomCorridorPCGRules) {
				r.HubDegree.Max = r.RoomCount.Max
			},
		},
		{
			name: "branch count has no matching tree",
			mutate: func(r *RoomCorridorPCGRules) {
				r.BranchJunctionCount = IntRange{Min: 2, Max: 2}
			},
		},
		{
			name: "loop count exceeds minimum room graph capacity",
			mutate: func(r *RoomCorridorPCGRules) {
				r.LoopEdgeCount = IntRange{Min: 0, Max: roomTopologyLoopCapacity(r.RoomCount.Min) + 1}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := base
			tt.mutate(&rules)
			if err := validateRoomCorridorPCGRules(rules); err == nil {
				t.Fatal("expected invalid topology rules error")
			}
		})
	}
}
