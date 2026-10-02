package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	dungeonGenerationAuditEnv       = "ARPG_DUNGEON_GENERATION_AUDIT"
	dungeonGenerationAuditFilename  = "dungeon-generation-audit.json"
	dungeonGenerationAuditSeedCount = 20
)

// TestDungeonGenerationAuditReport is an opt-in report-only probe over the current room-first
// generator. Generation errors and invariant violations are recorded in the report; this probe
// does not replace the fail-fast progression sweep in dungeon_room_corridor_sweep_test.go.
func TestDungeonGenerationAuditReport(t *testing.T) {
	if os.Getenv(dungeonGenerationAuditEnv) != "1" {
		t.Skipf("set %s=1 to write the deterministic dungeon generation audit", dungeonGenerationAuditEnv)
	}

	rules := loadRules(t).DungeonGeneration
	seeds := dungeonGenerationAuditSeeds()
	depths := dungeonGenerationAuditDepths(rules)
	report := dungeonGenerationAuditReport{
		SchemaVersion: 1,
		Seeds:         seeds,
		Depths:        depths,
		Floors:        make([]dungeonGenerationAuditFloor, 0, len(seeds)*len(depths)),
		Failures:      []dungeonGenerationAuditFailure{},
		Findings:      []dungeonGenerationAuditFinding{},
	}
	histograms := newDungeonGenerationAuditHistograms()
	for _, seed := range seeds {
		for _, levelNum := range depths {
			levelRules := rules.RulesForLevel(levelNum)
			out, err := GenerateDungeonLevel(seed, levelNum, rules)
			report.AttemptedFloors++
			if err != nil {
				report.Failures = append(report.Failures, dungeonGenerationAuditFailure{Seed: seed, Level: levelNum, Error: err.Error()})
				report.Floors = append(report.Floors, dungeonGenerationAuditFloor{
					Seed: seed, Level: levelNum, Type: dungeonGenerationAuditFloorType(levelNum, rules), GenerationError: err.Error(),
				})
				continue
			}
			report.GeneratedFloors++
			floor, findings := auditGeneratedDungeonFloor(seed, levelNum, levelRules, out)
			report.Floors = append(report.Floors, floor)
			report.Findings = append(report.Findings, findings...)
			histograms.addFloor(floor, out)
		}
	}
	report.Distributions = histograms.sorted()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal generation audit: %v", err)
	}
	data = append(data, '\n')
	if err := writeDungeonGenerationAudit(data); err != nil {
		t.Fatalf("write generation audit: %v", err)
	}
	t.Logf("dungeon generation audit: %d/%d floors generated, %d invariant finding(s), %d seed(s), %d depth(s) -> .artifacts/%s",
		report.GeneratedFloors, report.AttemptedFloors, len(report.Findings), len(seeds), len(depths), dungeonGenerationAuditFilename)
}

func auditGeneratedDungeonFloor(seed string, levelNum int, rules DungeonGenerationRules, out generatedDungeonLevel) (dungeonGenerationAuditFloor, []dungeonGenerationAuditFinding) {
	floor := dungeonGenerationAuditFloor{Seed: seed, Level: levelNum, Type: dungeonGenerationAuditFloorType(levelNum, rules)}
	findings := make([]dungeonGenerationAuditFinding, 0)
	fail := func(invariant, detail string) {
		findings = append(findings, dungeonGenerationAuditFinding{Seed: seed, Level: levelNum, Invariant: invariant, Detail: detail})
	}
	floor.RoomCount = len(out.rooms)
	floor.CorridorEdgeCount = len(out.corridorEdges)
	floor.CorridorRouteCount = len(out.corridorRoutes)
	floor.ActualMonsterCount = len(out.monsters)

	if isBossFloor(levelNum, rules) {
		floor.PopulationTarget = rules.BossFloor.MonsterCount + 1 // configured trash plus the boss entity
		if len(out.rooms) != 0 || len(out.corridorEdges) != 0 || len(out.corridorRoutes) != 0 {
			fail("boss_floor_roomless", fmt.Sprintf("boss floor generated %d rooms and %d corridor edges", len(out.rooms), len(out.corridorEdges)))
		}
		if floor.ActualMonsterCount != floor.PopulationTarget {
			fail("boss_floor_population", fmt.Sprintf("generated %d monsters; configured boss plus trash target is %d", floor.ActualMonsterCount, floor.PopulationTarget))
		}
	} else {
		auditDungeonRoomTopology(seed, levelNum, rules, out, &floor, &findings)
		auditDungeonAnchors(seed, levelNum, rules, out, &findings)
		auditDungeonPopulation(seed, levelNum, rules, out, &floor, &findings)
	}
	auditDungeonTargetReachability(seed, levelNum, rules, out, &floor, &findings)
	return floor, findings
}

func auditDungeonRoomTopology(seed string, levelNum int, rules DungeonGenerationRules, out generatedDungeonLevel, floor *dungeonGenerationAuditFloor, findings *[]dungeonGenerationAuditFinding) {
	fail := func(invariant, detail string) {
		*findings = append(*findings, dungeonGenerationAuditFinding{Seed: seed, Level: levelNum, Invariant: invariant, Detail: detail})
	}
	pcg := rules.RoomCorridorPCG
	if !pcg.Enabled {
		return
	}
	if len(out.rooms) < pcg.RoomCount.Min || len(out.rooms) > pcg.RoomCount.Max {
		fail("room_count_bounds", fmt.Sprintf("generated %d rooms; configured range is %d..%d", len(out.rooms), pcg.RoomCount.Min, pcg.RoomCount.Max))
	}
	connected, degrees, valid := auditRoomGraph(len(out.rooms), out.corridorEdges)
	floor.RoomGraphConnected = connected
	if !valid {
		fail("room_graph_edges", "room graph contains a self-edge, duplicate edge, or invalid room index")
	}
	if len(out.rooms) > 0 && !connected {
		fail("room_graph_connected", "corridor-edge graph does not connect every generated room")
	}
	if connected {
		floor.CorridorLoopCount = len(out.corridorEdges) - len(out.rooms) + 1
	}
	if connected && (floor.CorridorLoopCount < pcg.LoopEdgeCount.Min || floor.CorridorLoopCount > pcg.LoopEdgeCount.Max) {
		fail("corridor_loop_bounds", fmt.Sprintf("graph has %d loop edge(s); configured range is %d..%d", floor.CorridorLoopCount, pcg.LoopEdgeCount.Min, pcg.LoopEdgeCount.Max))
	}
	if len(out.corridorRoutes) != len(out.corridorEdges) {
		fail("corridor_route_count", fmt.Sprintf("selected %d graph edges but generated %d routes", len(out.corridorEdges), len(out.corridorRoutes)))
	}
	if err := validateGeneratedCorridorRoutes(out); err != nil {
		fail("corridor_route_validity", err.Error())
	}
	if len(out.rooms) > 0 && !allRoomCentersReachable(rules, out) {
		fail("room_center_reachability", "one or more room centers are not mutually reachable")
	}
	hubCount, hubDegree, branches := 0, 0, 0
	roleCounts := map[string]int{}
	for i, room := range out.rooms {
		if room.isHub {
			hubCount++
			if i < len(degrees) {
				hubDegree = degrees[i]
			}
		} else if i < len(degrees) && degrees[i] >= 3 {
			branches++
		}
		roleCounts[room.role]++
	}
	if pcg.HubRoomEnabled {
		if hubCount != 1 {
			fail("hub_count", fmt.Sprintf("generated %d hub rooms; expected exactly one enabled hub", hubCount))
		} else if hubDegree < pcg.HubDegree.Min || hubDegree > pcg.HubDegree.Max {
			fail("hub_degree_bounds", fmt.Sprintf("hub degree %d is outside configured range %d..%d", hubDegree, pcg.HubDegree.Min, pcg.HubDegree.Max))
		}
	}
	if branches < pcg.BranchJunctionCount.Min || branches > pcg.BranchJunctionCount.Max {
		fail("branch_junction_bounds", fmt.Sprintf("generated %d non-hub branch junctions; configured range is %d..%d", branches, pcg.BranchJunctionCount.Min, pcg.BranchJunctionCount.Max))
	}
	if pcg.RoomRoles.Enabled {
		for _, role := range pcg.RoomRoles.Roles {
			count := roleCounts[role.ID]
			if count < role.MinRooms || count > role.MaxRooms {
				fail("room_role_bounds", fmt.Sprintf("role %q has %d rooms; configured range is %d..%d", role.ID, count, role.MinRooms, role.MaxRooms))
			}
		}
	}
}

func auditDungeonAnchors(seed string, levelNum int, rules DungeonGenerationRules, out generatedDungeonLevel, findings *[]dungeonGenerationAuditFinding) {
	if !rules.RoomCorridorPCG.Enabled || !rules.RoomCorridorPCG.RoomRoles.Enabled {
		return
	}
	fail := func(invariant, detail string) {
		*findings = append(*findings, dungeonGenerationAuditFinding{Seed: seed, Level: levelNum, Invariant: invariant, Detail: detail})
	}
	roles := rules.RoomCorridorPCG.RoomRoles.PlacementRoles
	check := func(kind, expected string, pos Vec2) {
		roomIndex := dungeonAuditRoomAt(out.rooms, pos)
		if roomIndex < 0 {
			fail("anchor_inside_room", fmt.Sprintf("%s at %.2f,%.2f is outside generated rooms", kind, pos.X, pos.Y))
			return
		}
		if out.rooms[roomIndex].role != expected {
			fail("anchor_room_role", fmt.Sprintf("%s is in room %d with role %q; expected %q", kind, roomIndex, out.rooms[roomIndex].role, expected))
		}
	}
	for _, stair := range out.stairs {
		if stair.defID == stairsUpDefID {
			check(stair.defID, roles.UpStair, stair.pos)
		} else if stair.defID == stairsDownDefID {
			check(stair.defID, roles.DownStair, stair.pos)
		}
	}
	for _, teleporter := range out.teleporters {
		check(teleporter.defID, roles.Teleporter, teleporter.pos)
	}
	for _, chest := range out.chests {
		role := roles.Chest
		if chest.eliteObjective {
			role = roles.EliteObjective
		}
		check(chest.defID, role, chest.pos)
	}
}

func auditDungeonTargetReachability(seed string, levelNum int, rules DungeonGenerationRules, out generatedDungeonLevel, floor *dungeonGenerationAuditFloor, findings *[]dungeonGenerationAuditFinding) {
	nav := generatedDungeonNavigation(rules)
	blocked := buildDungeonBlockedGrid(nav, out)
	start := generatedReachabilityStart(rules, out)
	for _, target := range generatedReachabilityTargets(out) {
		kind := dungeonAuditTargetKind(target.kind)
		reachable := generatedTargetReachableFromNav(nav, blocked.blocked, start, target.pos)
		if target.kind == woodenDoorDefID {
			reachable = generatedDoorReachableFromNav(nav, blocked.blocked, start, target.pos)
		}
		if reachable {
			floor.ReachableTargetCount++
			floor.ReachableTargetKinds = append(floor.ReachableTargetKinds, kind)
			continue
		}
		floor.UnreachableTargetCount++
		floor.UnreachableTargetKinds = append(floor.UnreachableTargetKinds, kind)
		*findings = append(*findings, dungeonGenerationAuditFinding{
			Seed: seed, Level: levelNum, Invariant: "target_reachability",
			Detail: fmt.Sprintf("%s at %.2f,%.2f is not reachable from the generated start", target.kind, target.pos.X, target.pos.Y),
		})
	}
}

func auditDungeonPopulation(seed string, levelNum int, rules DungeonGenerationRules, out generatedDungeonLevel, floor *dungeonGenerationAuditFloor, findings *[]dungeonGenerationAuditFinding) {
	fail := func(invariant, detail string) {
		*findings = append(*findings, dungeonGenerationAuditFinding{Seed: seed, Level: levelNum, Invariant: invariant, Detail: detail})
	}
	placement := rules.MonsterPlacement
	baseTarget := placement.PopulationFormula.CountForSize(rules.FloorSize)
	floor.PopulationTarget = baseTarget
	if placement.Count != baseTarget {
		fail("population_formula", fmt.Sprintf("profile count %d does not match rule-derived formula count %d", placement.Count, baseTarget))
	}
	for _, chest := range out.chests {
		// Optional elite-objective chests are appended after pack placement and do not
		// participate in the chest population bonus applied before that placement.
		if !chest.eliteObjective {
			floor.PopulationTarget += rules.ChestPlacement.MonsterCountBonus
			break
		}
	}
	packs := map[string]int{}
	roomPackCounts := make([]int, len(out.rooms))
	for _, monster := range out.monsters {
		if monster.packID == "" {
			continue
		}
		packs[monster.packID]++
		floor.PackMemberCount++
		if monster.roomIndex >= 0 && monster.roomIndex < len(roomPackCounts) {
			roomPackCounts[monster.roomIndex]++
		}
	}
	floor.PackCount = len(packs)
	floor.SupplementalMonsterCount = floor.ActualMonsterCount - floor.PackMemberCount
	if floor.SupplementalMonsterCount < 0 {
		fail("population_nonnegative_supplement", fmt.Sprintf("%d pack members exceed %d generated monsters", floor.PackMemberCount, floor.ActualMonsterCount))
		floor.SupplementalMonsterCount = 0
	}
	if floor.PackMemberCount != floor.PopulationTarget {
		fail("population_target", fmt.Sprintf("generated %d pack members; rule-derived target including chest bonus is %d", floor.PackMemberCount, floor.PopulationTarget))
	}
	if floor.PackCount < placement.PackCount.Min || floor.PackCount > placement.PackCount.Max {
		fail("pack_count_bounds", fmt.Sprintf("generated %d packs; profile range is %d..%d", floor.PackCount, placement.PackCount.Min, placement.PackCount.Max))
	}
	packIDs := make([]string, 0, len(packs))
	for packID := range packs {
		packIDs = append(packIDs, packID)
	}
	sort.Strings(packIDs)
	for _, packID := range packIDs {
		count := packs[packID]
		if count < placement.PackSize.Min || count > placement.PackSize.Max {
			fail("pack_size_bounds", fmt.Sprintf("pack %q has %d members; configured range is %d..%d", packID, count, placement.PackSize.Min, placement.PackSize.Max))
		}
	}
	if rules.RoomCorridorPCG.Enabled && rules.RoomCorridorPCG.RoomRoles.Enabled {
		if len(out.roomMonsterBudgets) != len(out.rooms) {
			fail("room_population_budget_count", fmt.Sprintf("got %d room budgets for %d generated rooms", len(out.roomMonsterBudgets), len(out.rooms)))
		}
		budgetTotal := 0
		for _, budget := range out.roomMonsterBudgets {
			budgetTotal += budget.MonsterCount
			if budget.RoomIndex < 0 || budget.RoomIndex >= len(out.rooms) {
				fail("room_population_budget_index", fmt.Sprintf("budget references invalid room index %d", budget.RoomIndex))
				continue
			}
			room := out.rooms[budget.RoomIndex]
			if budget.Role != room.role {
				fail("room_population_budget_role", fmt.Sprintf("room %d budget role %q differs from room role %q", budget.RoomIndex, budget.Role, room.role))
			}
			if budget.MonsterCount != roomPackCounts[budget.RoomIndex] {
				fail("room_population_budget_count", fmt.Sprintf("room %d budget is %d but contains %d pack members", budget.RoomIndex, budget.MonsterCount, roomPackCounts[budget.RoomIndex]))
			}
		}
		if budgetTotal != floor.PopulationTarget {
			fail("room_population_budget_total", fmt.Sprintf("room budgets total %d; rule-derived target is %d", budgetTotal, floor.PopulationTarget))
		}
		if err := validateRoomEncounterPlacement(rules, out); err != nil {
			fail("encounter_placement", "generated room encounter failed the integrated placement validator")
		}
	}
}

func auditRoomGraph(roomCount int, edges []roomEdge) (bool, []int, bool) {
	degrees := make([]int, roomCount)
	if roomCount == 0 {
		return len(edges) == 0, degrees, len(edges) == 0
	}
	adjacent := make([][]int, roomCount)
	seen := make(map[roomEdge]struct{}, len(edges))
	valid := true
	for _, edge := range edges {
		edge = normalizeRoomEdge(edge)
		if edge[0] < 0 || edge[1] >= roomCount || edge[0] == edge[1] {
			valid = false
			continue
		}
		if _, duplicate := seen[edge]; duplicate {
			valid = false
			continue
		}
		seen[edge] = struct{}{}
		degrees[edge[0]]++
		degrees[edge[1]]++
		adjacent[edge[0]] = append(adjacent[edge[0]], edge[1])
		adjacent[edge[1]] = append(adjacent[edge[1]], edge[0])
	}
	visited := make([]bool, roomCount)
	queue := []int{0}
	visited[0] = true
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacent[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	for _, ok := range visited {
		if !ok {
			return false, degrees, valid
		}
	}
	return true, degrees, valid
}

func dungeonAuditRoomAt(rooms []dungeonRoom, pos Vec2) int {
	for i, room := range rooms {
		if roomContainsCircle(room, pos, 0) {
			return i
		}
	}
	return -1
}

func formatDungeonRoomAuditArea(room dungeonRoom) string {
	cellWidth := (room.innerMax.X - room.innerMin.X) / roomShapeSide
	cellHeight := (room.innerMax.Y - room.innerMin.Y) / roomShapeSide
	activeCells := 0
	for _, active := range room.shapeCells {
		if active {
			activeCells++
		}
	}
	return fmt.Sprintf("%.1f", float64(activeCells)*cellWidth*cellHeight)
}

func dungeonAuditTargetKind(kind string) string {
	if prefix, _, ok := strings.Cut(kind, ":"); ok {
		return prefix
	}
	return kind
}

func dungeonGenerationAuditFloorType(levelNum int, rules DungeonGenerationRules) string {
	if isBossFloor(levelNum, rules) {
		return "boss"
	}
	return "ordinary"
}

func dungeonGenerationAuditSeeds() []string {
	seeds := make([]string, dungeonGenerationAuditSeedCount)
	for i := range seeds {
		seeds[i] = fmt.Sprintf("generation-audit-%02d", i+1)
	}
	return seeds
}

func dungeonGenerationAuditDepths(rules DungeonGenerationRules) []int {
	deepest := 1
	for _, profile := range rules.FloorProfiles {
		deepest = maxInt(deepest, profile.MinDepth)
		if profile.MaxDepth != nil {
			deepest = maxInt(deepest, *profile.MaxDepth)
		} else if rules.BossFloor.Cadence > 0 {
			deepest = maxInt(deepest, profile.MinDepth+rules.BossFloor.Cadence)
		}
	}
	if rules.BossFloor.Cadence > 0 {
		deepest = maxInt(deepest, absInt(rules.BossFloor.FirstLevel)+rules.BossFloor.Cadence)
	}
	depths := make([]int, deepest)
	for i := range depths {
		depths[i] = -(i + 1)
	}
	return depths
}

func writeDungeonGenerationAudit(data []byte) (retErr error) {
	// The output directory and filename are fixed application-owned constants. Root-scoped
	// operations prevent an unexpected symlink below the repository from redirecting the write.
	repoRoot, err := os.OpenRoot("../../../")
	if err != nil {
		return fmt.Errorf("open repository root: %w", err)
	}
	defer func() { retErr = errors.Join(retErr, repoRoot.Close()) }()
	if err := repoRoot.Mkdir(".artifacts", 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("create repository artifact directory: %w", err)
	}
	artifactRoot, err := repoRoot.OpenRoot(".artifacts")
	if err != nil {
		return fmt.Errorf("open repository artifact directory: %w", err)
	}
	defer func() { retErr = errors.Join(retErr, artifactRoot.Close()) }()
	file, err := artifactRoot.OpenFile(dungeonGenerationAuditFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit report: %w", err)
	}
	defer func() { retErr = errors.Join(retErr, file.Close()) }()
	if written, err := file.Write(data); err != nil {
		return fmt.Errorf("write audit report: %w", err)
	} else if written != len(data) {
		return fmt.Errorf("write audit report: short write: wrote %d of %d bytes", written, len(data))
	}
	return nil
}

func addAuditCount(histogram map[string]int, value int) {
	histogram[fmt.Sprintf("%d", value)]++
}

func addAuditLabel(histogram map[string]int, label string) {
	if label != "" {
		histogram[label]++
	}
}

func TestDungeonGenerationAuditOrderingHelpers(t *testing.T) {
	connected, degrees, valid := auditRoomGraph(4, []roomEdge{{0, 1}, {1, 2}, {2, 3}, {3, 0}})
	if !connected || !valid || len(degrees) != 4 || degrees[0] != 2 || degrees[2] != 2 {
		t.Fatalf("auditRoomGraph connected=%v degrees=%v valid=%v", connected, degrees, valid)
	}
	connected, _, valid = auditRoomGraph(3, []roomEdge{{0, 1}, {0, 1}, {1, 2}})
	if !connected || valid {
		t.Fatalf("duplicate edge should invalidate graph while preserving connectivity: connected=%v valid=%v", connected, valid)
	}
	counts := auditCountBuckets(map[string]int{"10": 2, "2": 1})
	if len(counts) != 2 || counts[0].Value != 2 || counts[1].Value != 10 {
		t.Fatalf("numeric histogram order = %#v", counts)
	}
	labels := auditLabelBuckets(map[string]int{"transition": 1, "entry": 2})
	if len(labels) != 2 || labels[0].Label != "entry" || labels[1].Label != "transition" {
		t.Fatalf("label histogram order = %#v", labels)
	}
}

func TestDungeonGenerationAuditTargetKind(t *testing.T) {
	if got := dungeonAuditTargetKind("monster:bat"); got != "monster" {
		t.Fatalf("dungeonAuditTargetKind(monster:bat) = %q", got)
	}
	if got := dungeonAuditTargetKind("stairs_down"); got != "stairs_down" {
		t.Fatalf("dungeonAuditTargetKind(stairs_down) = %q", got)
	}
}
