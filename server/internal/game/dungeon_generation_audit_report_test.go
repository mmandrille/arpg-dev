package game

import (
	"fmt"
	"sort"
	"strconv"
)

type dungeonGenerationAuditReport struct {
	SchemaVersion   int                                `json:"schema_version"`
	Seeds           []string                           `json:"seeds"`
	Depths          []int                              `json:"levels"`
	AttemptedFloors int                                `json:"attempted_floors"`
	GeneratedFloors int                                `json:"generated_floors"`
	Floors          []dungeonGenerationAuditFloor      `json:"floors"`
	Failures        []dungeonGenerationAuditFailure    `json:"generation_failures"`
	Findings        []dungeonGenerationAuditFinding    `json:"invariant_findings"`
	Distributions   dungeonGenerationAuditDistribution `json:"distributions"`
}

type dungeonGenerationAuditFloor struct {
	Seed                     string   `json:"seed"`
	Level                    int      `json:"level"`
	Type                     string   `json:"type"`
	GenerationError          string   `json:"generation_error,omitempty"`
	ReachableTargetKinds     []string `json:"reachable_target_kinds,omitempty"`
	UnreachableTargetKinds   []string `json:"unreachable_target_kinds,omitempty"`
	RoomCount                int      `json:"room_count"`
	RoomGraphConnected       bool     `json:"room_graph_connected"`
	CorridorEdgeCount        int      `json:"corridor_edge_count"`
	CorridorRouteCount       int      `json:"corridor_route_count"`
	CorridorLoopCount        int      `json:"corridor_loop_count"`
	ReachableTargetCount     int      `json:"reachable_target_count"`
	UnreachableTargetCount   int      `json:"unreachable_target_count"`
	PopulationTarget         int      `json:"population_target"`
	PackMemberCount          int      `json:"pack_member_count"`
	PackCount                int      `json:"pack_count"`
	ActualMonsterCount       int      `json:"actual_monster_count"`
	SupplementalMonsterCount int      `json:"supplemental_monster_count"`
}

type dungeonGenerationAuditFailure struct {
	Seed  string `json:"seed"`
	Level int    `json:"level"`
	Error string `json:"error"`
}

type dungeonGenerationAuditFinding struct {
	Seed      string `json:"seed"`
	Level     int    `json:"level"`
	Invariant string `json:"invariant"`
	Detail    string `json:"detail"`
}

type dungeonGenerationAuditDistribution struct {
	RoomCounts             []dungeonGenerationAuditCountBucket `json:"room_counts"`
	RoomShapes             []dungeonGenerationAuditLabelBucket `json:"room_shapes"`
	RoomRoles              []dungeonGenerationAuditLabelBucket `json:"room_roles"`
	RoomAreas              []dungeonGenerationAuditLabelBucket `json:"room_areas"`
	CorridorEdges          []dungeonGenerationAuditCountBucket `json:"corridor_edges_per_floor"`
	CorridorLoops          []dungeonGenerationAuditCountBucket `json:"corridor_loops_per_floor"`
	RoomDegrees            []dungeonGenerationAuditCountBucket `json:"room_graph_degrees"`
	RouteSegments          []dungeonGenerationAuditCountBucket `json:"corridor_route_segments"`
	ReachableTargets       []dungeonGenerationAuditCountBucket `json:"reachable_targets_per_floor"`
	UnreachableTargets     []dungeonGenerationAuditCountBucket `json:"unreachable_targets_per_floor"`
	TargetKinds            []dungeonGenerationAuditLabelBucket `json:"reachable_target_kinds"`
	UnreachableTargetKinds []dungeonGenerationAuditLabelBucket `json:"unreachable_target_kinds"`
	PopulationTargets      []dungeonGenerationAuditCountBucket `json:"population_targets"`
	PackMembers            []dungeonGenerationAuditCountBucket `json:"pack_members_per_floor"`
	PackCounts             []dungeonGenerationAuditCountBucket `json:"pack_counts_per_floor"`
	ActualMonsters         []dungeonGenerationAuditCountBucket `json:"actual_monsters_per_floor"`
	SupplementalMonsters   []dungeonGenerationAuditCountBucket `json:"supplemental_monsters_per_floor"`
}

type dungeonGenerationAuditCountBucket struct {
	Value int `json:"value"`
	Count int `json:"count"`
}

type dungeonGenerationAuditLabelBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type dungeonGenerationAuditHistograms struct {
	roomCounts, roomShapes, roomRoles, roomAreas                                     map[string]int
	corridorEdges, corridorLoops, roomDegrees, routeSegments                         map[string]int
	reachableTargets, unreachableTargets, targetKinds, unreachableTargetKinds        map[string]int
	populationTargets, packMembers, packCounts, actualMonsters, supplementalMonsters map[string]int
}

func newDungeonGenerationAuditHistograms() dungeonGenerationAuditHistograms {
	return dungeonGenerationAuditHistograms{
		roomCounts: map[string]int{}, roomShapes: map[string]int{}, roomRoles: map[string]int{}, roomAreas: map[string]int{},
		corridorEdges: map[string]int{}, corridorLoops: map[string]int{}, roomDegrees: map[string]int{}, routeSegments: map[string]int{},
		reachableTargets: map[string]int{}, unreachableTargets: map[string]int{}, targetKinds: map[string]int{}, unreachableTargetKinds: map[string]int{},
		populationTargets: map[string]int{}, packMembers: map[string]int{}, packCounts: map[string]int{}, actualMonsters: map[string]int{}, supplementalMonsters: map[string]int{},
	}
}

func (h *dungeonGenerationAuditHistograms) addFloor(f dungeonGenerationAuditFloor, out generatedDungeonLevel) {
	if f.GenerationError != "" {
		return
	}
	addAuditCount(h.roomCounts, f.RoomCount)
	addAuditCount(h.corridorEdges, f.CorridorEdgeCount)
	addAuditCount(h.corridorLoops, f.CorridorLoopCount)
	addAuditCount(h.reachableTargets, f.ReachableTargetCount)
	addAuditCount(h.unreachableTargets, f.UnreachableTargetCount)
	addAuditCount(h.populationTargets, f.PopulationTarget)
	addAuditCount(h.packMembers, f.PackMemberCount)
	addAuditCount(h.packCounts, f.PackCount)
	addAuditCount(h.actualMonsters, f.ActualMonsterCount)
	addAuditCount(h.supplementalMonsters, f.SupplementalMonsterCount)
	_, degrees, _ := auditRoomGraph(len(out.rooms), out.corridorEdges)
	for _, degree := range degrees {
		addAuditCount(h.roomDegrees, degree)
	}
	for _, route := range out.corridorRoutes {
		addAuditCount(h.routeSegments, maxInt(0, len(route.points)-1))
	}
	for _, room := range out.rooms {
		addAuditLabel(h.roomShapes, room.shapeID)
		addAuditLabel(h.roomRoles, room.role)
		addAuditLabel(h.roomAreas, formatDungeonRoomAuditArea(room))
	}
	for _, kind := range f.ReachableTargetKinds {
		addAuditLabel(h.targetKinds, kind)
	}
	for _, kind := range f.UnreachableTargetKinds {
		addAuditLabel(h.unreachableTargetKinds, kind)
	}
}

func (h dungeonGenerationAuditHistograms) sorted() dungeonGenerationAuditDistribution {
	return dungeonGenerationAuditDistribution{
		RoomCounts: auditCountBuckets(h.roomCounts), RoomShapes: auditLabelBuckets(h.roomShapes), RoomRoles: auditLabelBuckets(h.roomRoles),
		RoomAreas: auditLabelBuckets(h.roomAreas), CorridorEdges: auditCountBuckets(h.corridorEdges), CorridorLoops: auditCountBuckets(h.corridorLoops),
		RoomDegrees: auditCountBuckets(h.roomDegrees), RouteSegments: auditCountBuckets(h.routeSegments), ReachableTargets: auditCountBuckets(h.reachableTargets),
		UnreachableTargets: auditCountBuckets(h.unreachableTargets), TargetKinds: auditLabelBuckets(h.targetKinds),
		UnreachableTargetKinds: auditLabelBuckets(h.unreachableTargetKinds), PopulationTargets: auditCountBuckets(h.populationTargets),
		PackMembers: auditCountBuckets(h.packMembers), PackCounts: auditCountBuckets(h.packCounts), ActualMonsters: auditCountBuckets(h.actualMonsters),
		SupplementalMonsters: auditCountBuckets(h.supplementalMonsters),
	}
}
func auditCountBuckets(histogram map[string]int) []dungeonGenerationAuditCountBucket {
	keys := make([]int, 0, len(histogram))
	for key := range histogram {
		value, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		keys = append(keys, value)
	}
	sort.Ints(keys)
	buckets := make([]dungeonGenerationAuditCountBucket, 0, len(keys))
	for _, value := range keys {
		key := fmt.Sprintf("%d", value)
		buckets = append(buckets, dungeonGenerationAuditCountBucket{Value: value, Count: histogram[key]})
	}
	return buckets
}

func auditLabelBuckets(histogram map[string]int) []dungeonGenerationAuditLabelBucket {
	keys := make([]string, 0, len(histogram))
	for key := range histogram {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	buckets := make([]dungeonGenerationAuditLabelBucket, 0, len(keys))
	for _, key := range keys {
		buckets = append(buckets, dungeonGenerationAuditLabelBucket{Label: key, Count: histogram[key]})
	}
	return buckets
}
