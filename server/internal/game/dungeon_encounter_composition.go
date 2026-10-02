package game

import (
	"fmt"
	"math"
	"strings"
)

func validateDungeonEncounterGenerationRules(roomRoleCompositionEnabled bool, generationPlacement, basePlacement MonsterPlacementRules, rules *Rules) error {
	if err := validateDungeonRoomPopulationRoleWeights(generationPlacement.RoomRoleWeights); err != nil {
		return err
	}
	if roomRoleCompositionEnabled {
		if err := validateDungeonEncounterComposition(basePlacement, rules); err != nil {
			return err
		}
	}
	if generationPlacement.CompositionSearchNodeLimit <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.composition_search_node_limit: must be positive")
	}
	if generationPlacement.CompositionFormationsPerPack <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.composition_formations_per_pack: must be positive")
	}
	return nil
}

var encounterRoomRoleIDs = [...]string{dungeonRoomRoleEntry, dungeonRoomRoleTransition, dungeonRoomRoleCombat, dungeonRoomRoleRewardObjective, dungeonRoomRoleBossArena}
var encounterMonsterRoleIDs = [...]string{"frontline", "ranged", "flanker", "swarm"}

func encounterRulesForRoom(rules DungeonGenerationRules, role string) (DungeonRoomEncounterRules, bool) {
	for _, configured := range rules.MonsterPlacement.EncounterComposition.RoomRoles {
		if configured.RoomRole == role {
			return configured, true
		}
	}
	return DungeonRoomEncounterRules{}, false
}

func validateDungeonEncounterComposition(placement MonsterPlacementRules, r *Rules) error {
	configured := placement.EncounterComposition.RoomRoles
	if len(configured) != len(encounterRoomRoleIDs) {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.encounter_composition.room_roles: expected %d roles", len(encounterRoomRoleIDs))
	}
	poolRoles := map[string]bool{}
	for _, entry := range placement.MonsterPool {
		if def, ok := r.Monsters[entry.MonsterDefID]; ok {
			poolRoles[def.PackRole] = true
		}
	}
	if len(placement.MonsterPool) == 0 {
		if def, ok := r.Monsters[placement.MonsterDefID]; ok {
			poolRoles[def.PackRole] = true
		}
	}
	for i, room := range configured {
		path := fmt.Sprintf("dungeon_generation.monster_placement.encounter_composition.room_roles[%d]", i)
		if room.RoomRole != encounterRoomRoleIDs[i] {
			return fmt.Errorf("game: invalid rules %s.room_role: expected %q", path, encounterRoomRoleIDs[i])
		}
		if room.EliteChancePercent < 0 || room.EliteChancePercent > 100 || room.MinimumGuardCount < 0 || room.MinimumGuardCount > placement.PackSize.Max {
			return fmt.Errorf("game: invalid rules %s: elite and guard bounds are invalid", path)
		}
		known := map[string]bool{}
		minTotal, maxTotal := 0, 0
		for j, member := range room.MemberRoles {
			if !knownEncounterMonsterRole(member.Role) || known[member.Role] {
				return fmt.Errorf("game: invalid rules %s.member_roles[%d].role: unknown or duplicate role %q", path, j, member.Role)
			}
			known[member.Role] = true
			if member.MinCount < 0 || member.MaxCount < member.MinCount || member.MaxCount > placement.PackSize.Max || member.Weight < 0 || member.Weight > maxDungeonRoomPopulationWeight {
				return fmt.Errorf("game: invalid rules %s.member_roles[%d]: invalid count or weight bounds", path, j)
			}
			if member.MaxCount > 0 && !poolRoles[member.Role] {
				return fmt.Errorf("game: invalid rules %s.member_roles[%d].role: no monster in pool has role %q", path, j, member.Role)
			}
			minTotal += member.MinCount
			maxTotal += member.MaxCount
		}
		if !known[room.LeaderRole] || countRoleMinimum(room.MemberRoles, room.LeaderRole) < 1 {
			return fmt.Errorf("game: invalid rules %s.leader_role: must have a positive member minimum", path)
		}
		if minTotal > placement.PackSize.Min || maxTotal < placement.PackSize.Max {
			return fmt.Errorf("game: invalid rules %s.member_roles: constraints must cover configured pack size range", path)
		}
		guardSet := map[string]bool{}
		for _, role := range room.GuardRoles {
			if !known[role] || role == room.LeaderRole || guardSet[role] {
				return fmt.Errorf("game: invalid rules %s.guard_roles: unknown, duplicate, or leader role %q", path, role)
			}
			guardSet[role] = true
		}
		if room.MinimumGuardCount > 0 && len(guardSet) == 0 {
			return fmt.Errorf("game: invalid rules %s.guard_roles: required when minimum_guard_count is positive", path)
		}
	}
	return nil
}

func knownEncounterMonsterRole(role string) bool {
	for _, known := range encounterMonsterRoleIDs {
		if role == known {
			return true
		}
	}
	return false
}
func countRoleMinimum(roles []DungeonEncounterMemberRoleRule, role string) int {
	for _, member := range roles {
		if member.Role == role {
			return member.MinCount
		}
	}
	return 0
}

func partitionEncounterPackSizes(packSizes []int, budgets []dungeonRoomPopulationBudget) ([][]int, error) {
	options, err := partitionEncounterPackSizeOptions(packSizes, budgets, 1)
	if err != nil {
		return nil, err
	}
	return options[0], nil
}

func partitionEncounterPackSizeOptions(packSizes []int, budgets []dungeonRoomPopulationBudget, limit int) ([][][]int, error) {
	if limit < 1 {
		limit = 1
	}
	remaining := make([]int, len(budgets))
	for i, b := range budgets {
		if b.MonsterCount < 0 {
			return nil, fmt.Errorf("negative room population budget")
		}
		remaining[i] = b.MonsterCount
	}
	order := make([]int, len(packSizes))
	for i, size := range packSizes {
		if size <= 0 {
			return nil, fmt.Errorf("non-positive pack size %d", size)
		}
		order[i] = i
	}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			if packSizes[order[j]] < packSizes[order[i]] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	assigned := make([][]int, len(budgets))
	nodes := 0
	options := make([][][]int, 0, limit)
	var place func(int)
	place = func(at int) {
		if len(options) >= limit {
			return
		}
		nodes++
		if nodes > 100000 {
			return
		}
		if at == len(order) {
			for _, n := range remaining {
				if n != 0 {
					return
				}
			}
			copyOption := make([][]int, len(assigned))
			for i := range assigned {
				copyOption[i] = append([]int(nil), assigned[i]...)
			}
			options = append(options, copyOption)
			return
		}
		size := packSizes[order[at]]
		seen := map[int]bool{}
		for room := range remaining {
			if remaining[room] < size || seen[remaining[room]] {
				continue
			}
			seen[remaining[room]] = true
			remaining[room] -= size
			assigned[room] = append(assigned[room], size)
			place(at + 1)
			assigned[room] = assigned[room][:len(assigned[room])-1]
			remaining[room] += size
		}
	}
	place(0)
	if len(options) == 0 {
		return nil, fmt.Errorf("pack sizes cannot be partitioned exactly across fixed room budgets")
	}
	return options, nil
}

func placeRoomEncounterMonsters(rng, composeRNG, rarityRNG *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, budgets []dungeonRoomPopulationBudget, packSizes []int) error {
	baseMonsters := append([]generatedMonster(nil), out.monsters...)
	rooms := make([]encounterRoomChoice, 0, len(budgets))
	for _, budget := range budgets {
		if budget.RoomIndex < 0 || budget.RoomIndex >= len(out.rooms) || out.rooms[budget.RoomIndex].role != budget.Role {
			return fmt.Errorf("game: generate dungeon level %d: room budget index/role mismatch", out.levelNum)
		}
		weight, ok := rules.MonsterPlacement.RoomRoleWeights.weight(budget.Role)
		if !ok {
			return fmt.Errorf("game: generate dungeon level %d: unsupported room role %q", out.levelNum, budget.Role)
		}
		candidates := legalEncounterCenters(rng, rules, *out, budget.RoomIndex)
		capacity := encounterRoomOccupancyUpperBound(candidates, encounterMonsterMinSeparation())
		fits := map[int]bool{}
		for _, size := range packSizes {
			if !fits[size] {
				fits[size] = roomHasStaticEncounterLayout(rules, *out, budget.RoomIndex, candidates, size)
			}
		}
		// Candidate count expresses usable area for weighting; the cell bound
		// prevents assigning more bodies than can coexist at minimum separation.
		if weight > 0 && capacity > 0 {
			rooms = append(rooms, encounterRoomChoice{budget: budget, weight: capacity * weight, capacity: capacity, candidates: candidates, fitsPackSize: fits})
		}
	}
	if len(rooms) == 0 {
		return fmt.Errorf("game: generate dungeon level %d: no positive-weight room can fit a pack", out.levelNum)
	}
	order := make([]int, len(packSizes))
	for i := range packSizes {
		if packSizes[i] <= 0 {
			return fmt.Errorf("game: generate dungeon level %d: invalid pack size %d", out.levelNum, packSizes[i])
		}
		order[i] = i
	}
	for i := range order {
		for j := i + 1; j < len(order); j++ {
			if packSizes[order[j]] > packSizes[order[i]] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	// Retry bounded, seeded complete assignments. Each failed floor attempt is
	// rolled back before a fresh room ordering and formation search begins.
	attemptLimit := max(1, rules.MonsterPlacement.MaxAttempts)
	searchLimit := rules.MonsterPlacement.CompositionSearchNodeLimit
	var lastErr error
	searchNodes := 0
	for attempt := 0; attempt < attemptLimit; attempt++ {
		out.monsters = append([]generatedMonster(nil), baseMonsters...)
		var assignPacks func(int) bool
		assignPacks = func(orderIndex int) bool {
			if orderIndex == len(order) {
				return true
			}
			if searchNodes >= searchLimit {
				lastErr = fmt.Errorf("bounded room assignment search exceeded %d nodes", searchLimit)
				return false
			}
			capacities := make([]int, len(rooms))
			for i, room := range rooms {
				freeCandidates := 0
				for _, candidate := range room.candidates {
					if !roomEncounterMonsterBlocked(candidate, rules, *out) {
						freeCandidates++
					}
				}
				capacities[i] = minInt(room.capacity, freeCandidates)
			}
			packIndex := order[orderIndex]
			size := packSizes[packIndex]
			choices := weightedEncounterRoomOrder(rng, rooms, capacities, size)
			if len(choices) == 0 {
				roomState := make([]string, 0, len(rooms))
				for i, room := range rooms {
					roomState = append(roomState, fmt.Sprintf("room=%d role=%s remaining=%d candidates=%d feasible=%t", room.budget.RoomIndex, room.budget.Role, capacities[i], len(room.candidates), room.fitsPackSize[size]))
				}
				lastErr = fmt.Errorf("no positive-weight room has remaining body-center capacity for pack size %d at occurrence %d (%s)", size, packIndex+1, strings.Join(roomState, ", "))
				return false
			}
			for _, roomIndex := range choices {
				searchNodes++
				if searchNodes > searchLimit {
					lastErr = fmt.Errorf("bounded room assignment search exceeded %d nodes", searchLimit)
					break
				}
				roomChoice := rooms[roomIndex]
				roomRules, ok := encounterRulesForRoom(rules, roomChoice.budget.Role)
				if !ok {
					continue
				}
				roles, elite, err := chooseEncounterMemberRoles(composeRNG, roomRules, size)
				if err != nil {
					lastErr = err
					continue
				}
				defs, err := chooseEncounterMonsterDefs(composeRNG, rules.MonsterPlacement, rules, roles)
				if err != nil {
					lastErr = err
					continue
				}
				leader := -1
				if elite {
					for i, role := range roles {
						if role == roomRules.LeaderRole {
							leader = i
							break
						}
					}
				}
				packID := fmt.Sprintf("pack_%02d_%d", roomChoice.budget.RoomIndex+1, packIndex+1)
				packStart := len(out.monsters)
				placed, err := searchRoomEncounterPackLayouts(rng, rarityRNG, rules, out, roomChoice.budget.RoomIndex, roomChoice.candidates, defs, packID, leader, &searchNodes, searchLimit, func() bool {
					return assignPacks(orderIndex + 1)
				})
				if err != nil {
					lastErr = err
				}
				if placed {
					return true
				}
				out.monsters = out.monsters[:packStart]
			}
			return false
		}
		if !assignPacks(0) {
			if searchNodes >= searchLimit {
				break
			}
			continue
		}
		if out.reservedEliteObjectivePos != nil && !hasEliteLeaderNearObjective(*out, rules) {
			lastErr = fmt.Errorf("reserved elite objective has no eligible leader in its room cluster")
			continue
		}
		if err := enforceEncounterMinimumMonsters(rules, out); err != nil {
			lastErr = err
			continue
		}
		if err := placeRoomEncounterChampionMinions(rng, rules, out); err != nil {
			lastErr = err
			continue
		}
		for i := range budgets {
			budgets[i].MonsterCount = 0
			for _, monster := range out.monsters {
				if monster.packID != "" && monster.roomIndex == budgets[i].RoomIndex {
					budgets[i].MonsterCount++
				}
			}
		}
		out.roomMonsterBudgets = budgets
		return nil
	}
	out.monsters = baseMonsters
	if lastErr != nil {
		roomCapacity := make([]string, 0, len(rooms))
		for _, room := range rooms {
			roomCapacity = append(roomCapacity, fmt.Sprintf("room=%d role=%s weight=%d body_centers=%d", room.budget.RoomIndex, room.budget.Role, room.weight, room.capacity))
		}
		return fmt.Errorf("game: generate dungeon level %d: %w (floor_attempts=%d/%d search_nodes=%d/%d; %s)", out.levelNum, lastErr, attemptLimit, rules.MonsterPlacement.MaxAttempts, searchNodes, searchLimit, strings.Join(roomCapacity, ", "))
	}
	return fmt.Errorf("game: generate dungeon level %d: no exact room pack partition could be placed", out.levelNum)
}

func encounterRoomOccupancyUpperBound(candidates []Vec2, minimumSeparation float64) int {
	if minimumSeparation <= 0 {
		return len(candidates)
	}
	// A cell's diagonal is strictly smaller than the minimum separation, so at
	// most one monster center can occupy each cell. Distinct cells may still
	// conflict; complete pack placement remains the exact feasibility check.
	cellSize := minimumSeparation/math.Sqrt2 - 0.000001
	if cellSize <= 0 {
		return len(candidates)
	}
	cells := make(map[[2]int]struct{}, len(candidates))
	for _, pos := range candidates {
		cells[[2]int{int(math.Floor(pos.X / cellSize)), int(math.Floor(pos.Y / cellSize))}] = struct{}{}
	}
	return len(cells)
}

func encounterMonsterMinSeparation() float64 {
	return 2*monsterRadius + 0.05
}

func placeRoomEncounterChampionMinions(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	primaryCount := len(out.monsters)
	for i := 0; i < primaryCount; i++ {
		member := out.monsters[i]
		if member.packID != "" && member.rarityID == "champion" && !member.packLeader {
			if err := placeChampionCommonMinions(rng, rules, out, member.pos, member.roomIndex); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasLeaderInRoom(out generatedDungeonLevel, role string) bool {
	for _, monster := range out.monsters {
		if monster.packLeader && monster.roomIndex >= 0 && monster.roomIndex < len(out.rooms) && out.rooms[monster.roomIndex].role == role {
			return true
		}
	}
	return false
}

func hasEliteLeaderNearObjective(out generatedDungeonLevel, rules DungeonGenerationRules) bool {
	if out.reservedEliteObjectivePos == nil {
		return true
	}
	for _, monster := range out.monsters {
		if monster.packLeader && distance(monster.pos, *out.reservedEliteObjectivePos) <= rules.EliteObjective.RoomClusterRadius {
			return true
		}
	}
	return false
}
