package game

import (
	"fmt"
	"math"
	"sort"
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

type encounterRoomChoice struct {
	budget       dungeonRoomPopulationBudget
	weight       int
	capacity     int
	candidates   []Vec2
	fitsPackSize map[int]bool
}

func weightedEncounterRoomOrder(rng *RNG, rooms []encounterRoomChoice, remaining []int, packSize int) []int {
	available := make([]int, 0, len(rooms))
	for i := range rooms {
		if remaining[i] >= packSize && rooms[i].weight > 0 && rooms[i].fitsPackSize[packSize] {
			available = append(available, i)
		}
	}
	order := make([]int, 0, len(available))
	for len(available) > 0 {
		total := 0
		for _, i := range available {
			total += remainingEncounterRoomWeight(rooms[i], remaining[i])
		}
		if total <= 0 {
			break
		}
		draw := rng.IntN(total)
		selected := 0
		for at, i := range available {
			weight := remainingEncounterRoomWeight(rooms[i], remaining[i])
			if weight <= 0 {
				continue
			}
			if draw < weight {
				selected = at
				break
			}
			draw -= weight
		}
		order = append(order, available[selected])
		available = append(available[:selected], available[selected+1:]...)
	}
	return order
}

func remainingEncounterRoomWeight(room encounterRoomChoice, remaining int) int {
	if remaining <= 0 || room.weight <= 0 {
		return 0
	}
	capacity := maxInt(1, room.capacity)
	// Weight by remaining usable area squared to reduce overfilling the initially
	// largest room while retaining the configured role preference.
	return maxInt(1, room.weight*remaining*remaining/(capacity*capacity))
}

func roomHasStaticEncounterLayout(rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, candidates []Vec2, packSize int) bool {
	if packSize <= 0 || roomIndex < 0 || roomIndex >= len(out.rooms) {
		return false
	}
	corridorClearance := rules.MonsterPlacement.PackMemberRadius
	for _, center := range candidates {
		if roomEncounterPositionBlocked(center, rules, out, roomIndex) || roomEncounterMonsterBlocked(center, rules, out) || generatedPositionInCorridorZone(center, corridorClearance, out) {
			continue
		}
		positions := []Vec2{center}
		memberCandidates := append([]Vec2(nil), candidates...)
		sort.SliceStable(memberCandidates, func(i, j int) bool {
			return distance(center, memberCandidates[i]) < distance(center, memberCandidates[j])
		})
		var fit func(int) bool
		fit = func(start int) bool {
			if len(positions) >= packSize {
				return true
			}
			for i := start; i < len(memberCandidates); i++ {
				pos := memberCandidates[i]
				if distance(center, pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 ||
					roomEncounterMonsterBlocked(pos, rules, out) ||
					generatedPositionInCorridorZone(pos, corridorClearance, out) {
					continue
				}
				separated := true
				for _, existing := range positions {
					if distance(existing, pos) < encounterMonsterMinSeparation() {
						separated = false
						break
					}
				}
				if !separated {
					continue
				}
				positions = append(positions, pos)
				if fit(i + 1) {
					return true
				}
				positions = positions[:len(positions)-1]
			}
			return false
		}
		if fit(0) {
			return true
		}
	}
	return false
}

func legalEncounterCenters(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int) []Vec2 {
	if roomIndex < 0 || roomIndex >= len(out.rooms) {
		return nil
	}
	centers := roomInteriorCandidates(out.rooms, []int{roomIndex}, rules.MonsterPlacement.MarginFromWall, rules.FloorSize)
	legal := make([]Vec2, 0, len(centers))
	for _, pos := range centers {
		if !roomEncounterPositionBlocked(pos, rules, out, roomIndex) &&
			generatedTargetReachable(rules, out, pos) {
			legal = append(legal, pos)
		}
	}
	return legal
}

func chooseEncounterMemberRoles(rng *RNG, rules DungeonRoomEncounterRules, size int) ([]string, bool, error) {
	roles := make([]string, 0, size)
	counts := map[string]int{}
	for _, member := range rules.MemberRoles {
		for i := 0; i < member.MinCount; i++ {
			roles = append(roles, member.Role)
			counts[member.Role]++
		}
	}
	if len(roles) > size {
		return nil, false, fmt.Errorf("room role %s minimum member roles exceed pack size %d", rules.RoomRole, size)
	}
	for len(roles) < size {
		total := 0
		for _, member := range rules.MemberRoles {
			if counts[member.Role] < member.MaxCount {
				total += member.Weight
			}
		}
		if total <= 0 {
			return nil, false, fmt.Errorf("room role %s has no available member role for pack size %d", rules.RoomRole, size)
		}
		draw := rng.IntN(total)
		for _, member := range rules.MemberRoles {
			if counts[member.Role] >= member.MaxCount || member.Weight == 0 {
				continue
			}
			if draw < member.Weight {
				roles = append(roles, member.Role)
				counts[member.Role]++
				break
			}
			draw -= member.Weight
		}
	}
	elite := rules.EliteChancePercent > 0 && rng.IntN(100) < rules.EliteChancePercent
	if elite {
		guards := 0
		for _, role := range roles {
			for _, guardRole := range rules.GuardRoles {
				if role == guardRole {
					guards++
					break
				}
			}
		}
		for guards < rules.MinimumGuardCount {
			replaced := false
			leaderIndex := -1
			for i, role := range roles {
				if role == rules.LeaderRole {
					leaderIndex = i
					break
				}
			}
			for i := len(roles) - 1; i >= 0; i-- {
				if i == leaderIndex {
					continue
				}
				for _, guardRole := range rules.GuardRoles {
					if counts[guardRole] < roleMaxCount(rules.MemberRoles, guardRole) {
						counts[roles[i]]--
						roles[i] = guardRole
						counts[guardRole]++
						guards++
						replaced = true
						break
					}
				}
				if replaced {
					break
				}
			}
			if !replaced {
				return nil, false, fmt.Errorf("room role %s cannot fit elite guard requirements in size %d roles %v (guards %v, min %d, member rules %+v)", rules.RoomRole, size, roles, rules.GuardRoles, rules.MinimumGuardCount, rules.MemberRoles)
			}
		}
	}
	return roles, elite, nil
}

func roleMaxCount(roles []DungeonEncounterMemberRoleRule, role string) int {
	for _, member := range roles {
		if member.Role == role {
			return member.MaxCount
		}
	}
	return 0
}

func chooseEncounterMonsterDefs(rng *RNG, placement MonsterPlacementRules, rules DungeonGenerationRules, roles []string) ([]string, error) {
	defs := make([]string, len(roles))
	for i, role := range roles {
		total := 0
		for _, entry := range placement.MonsterPool {
			if rules.MonsterRole(entry.MonsterDefID) == role {
				total += entry.Weight
			}
		}
		if total <= 0 {
			return nil, fmt.Errorf("monster pool has no member for pack role %s", role)
		}
		draw := rng.IntN(total)
		for _, entry := range placement.MonsterPool {
			if rules.MonsterRole(entry.MonsterDefID) != role {
				continue
			}
			if draw < entry.Weight {
				defs[i] = entry.MonsterDefID
				break
			}
			draw -= entry.Weight
		}
	}
	return defs, nil
}

func enforceEncounterMinimumMonsters(rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	used := map[int]bool{}
	for _, minimum := range rules.MonsterPlacement.MinimumMonsters {
		wantedRole := rules.MonsterRole(minimum.MonsterDefID)
		present := 0
		for _, monster := range out.monsters {
			if monster.packID != "" && rules.MonsterRole(monster.defID) == wantedRole {
				present++
			}
		}
		for present < minimum.Count {
			replaced := false
			for i := range out.monsters {
				monster := &out.monsters[i]
				if monster.packID == "" || monster.packLeader || used[i] || monster.roomIndex < 0 || monster.roomIndex >= len(out.rooms) {
					continue
				}
				oldRole := rules.MonsterRole(monster.defID)
				roomRules, ok := encounterRulesForRoom(rules, out.rooms[monster.roomIndex].role)
				if !ok || oldRole == wantedRole {
					continue
				}
				members := packMembers(out.monsters, monster.packID)
				roleCounts := map[string]int{}
				for _, member := range members {
					roleCounts[rules.MonsterRole(member.defID)]++
				}
				if roleCounts[oldRole] <= countRoleMinimum(roomRules.MemberRoles, oldRole) || roleCounts[wantedRole] >= roleMaxCount(roomRules.MemberRoles, wantedRole) {
					continue
				}
				if monsterIsRequiredGuard(roomRules, rules, members, oldRole) {
					continue
				}
				monster.defID = minimum.MonsterDefID
				used[i] = true
				present++
				replaced = true
				break
			}
			if !replaced {
				return fmt.Errorf("game: generate dungeon level %d: room-role composition cannot satisfy minimum monster %s count %d", out.levelNum, minimum.MonsterDefID, minimum.Count)
			}
		}
	}
	return nil
}

func packMembers(monsters []generatedMonster, packID string) []generatedMonster {
	var members []generatedMonster
	for _, monster := range monsters {
		if monster.packID == packID {
			members = append(members, monster)
		}
	}
	return members
}

func monsterIsRequiredGuard(room DungeonRoomEncounterRules, rules DungeonGenerationRules, members []generatedMonster, candidateRole string) bool {
	if room.MinimumGuardCount == 0 {
		return false
	}
	guards := 0
	for _, member := range members {
		for _, guardRole := range room.GuardRoles {
			if rules.MonsterRole(member.defID) == guardRole {
				guards++
				break
			}
		}
	}
	return guards <= room.MinimumGuardCount && containsEncounterRole(room.GuardRoles, candidateRole)
}

func containsEncounterRole(roles []string, role string) bool {
	for _, value := range roles {
		if value == role {
			return true
		}
	}
	return false
}

func searchRoomEncounterPackLayouts(rng, rarityRNG *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, roomIndex int, candidates []Vec2, defs []string, packID string, leaderIndex int, nodes *int, nodeLimit int, continueSearch func() bool) (bool, error) {
	candidates = append([]Vec2(nil), candidates...)
	shuffleDungeonPositions(rng, candidates)
	if leaderIndex >= 0 && out.reservedEliteObjectivePos != nil && out.rooms[roomIndex].role == rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective {
		sort.SliceStable(candidates, func(i, j int) bool {
			return distance(candidates[i], *out.reservedEliteObjectivePos) < distance(candidates[j], *out.reservedEliteObjectivePos)
		})
	}
	baseMonsterCount := len(out.monsters)
	positions := make([]Vec2, len(defs))
	// Bound complete formations per center while still allowing later centers
	// to be explored when earlier pack layouts conflict with other packs.
	formationAlternatives := 0
	localNodeLimit := rules.MonsterPlacement.MaxAttempts * maxInt(1, len(candidates)) * maxInt(1, len(defs))
	corridorClearance := rules.MonsterPlacement.PackMemberRadius
	localNodes := 0
	var placementErr error
	allowReservedObjectiveCluster := leaderIndex >= 0 && out.reservedEliteObjectivePos != nil && out.rooms[roomIndex].role == rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective
	var assignMembers func(Vec2, []Vec2, int) bool
	assignMembers = func(center Vec2, memberCandidates []Vec2, memberIndex int) bool {
		if formationAlternatives >= rules.MonsterPlacement.CompositionFormationsPerPack {
			return false
		}
		if memberIndex >= len(positions) {
			formationAlternatives++
			placeholders := append([]generatedMonster(nil), out.monsters[baseMonsterCount:]...)
			out.monsters = out.monsters[:baseMonsterCount]
			for i, defID := range defs {
				if err := appendGeneratedMonster(rng, rarityRNG, rules, out, defID, packID, i == leaderIndex, positions[i], roomIndex, false); err != nil {
					placementErr = err
					out.monsters = append(out.monsters[:baseMonsterCount], placeholders...)
					return false
				}
			}
			if continueSearch() {
				return true
			}
			out.monsters = append(out.monsters[:baseMonsterCount], placeholders...)
			return false
		}
		for _, pos := range memberCandidates {
			localNodes++
			*nodes = *nodes + 1
			if *nodes > nodeLimit || localNodes > localNodeLimit {
				return false
			}
			if distance(center, pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 ||
				roomEncounterPositionBlockedForPackWithObjectivePermission(pos, rules, *out, roomIndex, "__placement__", allowReservedObjectiveCluster) ||
				roomEncounterMonsterBlocked(pos, rules, *out) ||
				generatedPositionInCorridorZone(pos, corridorClearance, *out) {
				continue
			}
			if memberIndex == leaderIndex && out.reservedEliteObjectivePos != nil && out.rooms[roomIndex].role == rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective && distance(pos, *out.reservedEliteObjectivePos) > rules.EliteObjective.RoomClusterRadius {
				continue
			}
			positions[memberIndex] = pos
			out.monsters = append(out.monsters, generatedMonster{packID: "__placement__", roomIndex: roomIndex, pos: pos})
			if assignMembers(center, memberCandidates, memberIndex+1) {
				return true
			}
			out.monsters = out.monsters[:len(out.monsters)-1]
		}
		return false
	}
	for _, center := range candidates {
		formationAlternatives = 0
		localNodes++
		*nodes = *nodes + 1
		if *nodes > nodeLimit {
			return false, fmt.Errorf("bounded room pack layout search exceeded %d nodes", nodeLimit)
		}
		if localNodes > localNodeLimit {
			break
		}
		if roomEncounterPositionBlockedForPackWithObjectivePermission(center, rules, *out, roomIndex, "__placement__", allowReservedObjectiveCluster) || roomEncounterMonsterBlocked(center, rules, *out) || generatedPositionInCorridorZone(center, corridorClearance, *out) {
			continue
		}
		if leaderIndex == 0 && out.reservedEliteObjectivePos != nil && out.rooms[roomIndex].role == rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective && distance(center, *out.reservedEliteObjectivePos) > rules.EliteObjective.RoomClusterRadius {
			continue
		}
		memberCandidates := append([]Vec2(nil), candidates...)
		sort.SliceStable(memberCandidates, func(i, j int) bool {
			return distance(center, memberCandidates[i]) < distance(center, memberCandidates[j])
		})
		positions[0] = center
		out.monsters = append(out.monsters, generatedMonster{packID: "__placement__", roomIndex: roomIndex, pos: center})
		formationAlternatives = 0
		if assignMembers(center, memberCandidates, 1) {
			return true, nil
		}
		out.monsters = out.monsters[:baseMonsterCount]
	}
	if placementErr != nil {
		return false, placementErr
	}
	return false, fmt.Errorf("game: generate dungeon level %d: could not place complete %s (size=%d centers=%d formations=%d role=%s bounds=%v..%v budget=%d)", out.levelNum, packID, len(defs), len(candidates), formationAlternatives, out.rooms[roomIndex].role, out.rooms[roomIndex].innerMin, out.rooms[roomIndex].innerMax, roomPopulationForRoom(out.roomMonsterBudgets, roomIndex))
}

func distanceToNearestMonster(pos Vec2, monsters []generatedMonster) float64 {
	if len(monsters) == 0 {
		return math.MaxFloat64
	}
	nearest := math.MaxFloat64
	for _, monster := range monsters {
		if d := distance(pos, monster.pos); d < nearest {
			nearest = d
		}
	}
	return nearest
}

func roomPopulationForRoom(budgets []dungeonRoomPopulationBudget, roomIndex int) int {
	for _, b := range budgets {
		if b.RoomIndex == roomIndex {
			return b.MonsterCount
		}
	}
	return 0
}

func validateRoomEncounterPlacement(rules DungeonGenerationRules, out generatedDungeonLevel) error {
	if len(out.roomMonsterBudgets) == 0 {
		return nil
	}
	counts := make([]int, len(out.roomMonsterBudgets))
	packs := map[string][]generatedMonster{}
	for _, monster := range out.monsters {
		if monster.roomIndex < 0 {
			continue
		}
		if monster.roomIndex >= len(out.rooms) || !roomContainsCircle(out.rooms[monster.roomIndex], monster.pos, rules.MonsterPlacement.MarginFromWall) {
			return fmt.Errorf("game: generated encounter member escaped its assigned room")
		}
		if monster.packID != "" {
			packs[monster.packID] = append(packs[monster.packID], monster)
			for i, budget := range out.roomMonsterBudgets {
				if budget.RoomIndex == monster.roomIndex {
					counts[i]++
					break
				}
			}
		}
		if roomEncounterPositionBlockedForPack(monster.pos, rules, out, monster.roomIndex, monster.packID) {
			return fmt.Errorf("game: generated encounter member in room %d violates clearance at %v (blockers=%s)", monster.roomIndex, monster.pos, strings.Join(roomEncounterBlockReasons(monster.pos, rules, out, monster.roomIndex, monster.packID), ","))
		}
		if !generatedTargetReachable(rules, out, monster.pos) {
			return fmt.Errorf("game: generated encounter member in room %d is unreachable at %v", monster.roomIndex, monster.pos)
		}
		if generatedPositionInCorridorZone(monster.pos, rules.MonsterPlacement.PackMemberRadius, out) {
			return fmt.Errorf("game: generated encounter member in room %d intersects a corridor at %v", monster.roomIndex, monster.pos)
		}
	}
	for i, budget := range out.roomMonsterBudgets {
		if counts[i] != budget.MonsterCount {
			return fmt.Errorf("game: room %d encounter count %d does not match fixed budget %d", budget.RoomIndex, counts[i], budget.MonsterCount)
		}
	}
	for _, packID := range sortedStringKeys(packs) {
		members := packs[packID]
		if len(members) < rules.MonsterPlacement.PackSize.Min || len(members) > rules.MonsterPlacement.PackSize.Max {
			return fmt.Errorf("game: generated pack %s has invalid member count %d", packID, len(members))
		}
		center := members[0].pos
		roomIndex := members[0].roomIndex
		for i, member := range members {
			if member.roomIndex != roomIndex || distance(center, member.pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 {
				return fmt.Errorf("game: generated pack %s crosses room or spread boundary", packID)
			}
			for j := i + 1; j < len(members); j++ {
				if distance(member.pos, members[j].pos) < encounterMonsterMinSeparation() {
					return fmt.Errorf("game: generated pack %s has overlapping members", packID)
				}
			}
		}
	}
	return nil
}

func roomEncounterBlockReasons(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, packID string) []string {
	var reasons []string
	if roomIndex < 0 || roomIndex >= len(out.rooms) || !roomContainsCircle(out.rooms[roomIndex], pos, rules.MonsterPlacement.MarginFromWall) {
		reasons = append(reasons, "room")
	}
	if distance(pos, rules.PlayerSpawn) < rules.MonsterPlacement.MinSpawnDistance {
		reasons = append(reasons, "player_spawn")
	}
	for _, stair := range out.stairPositions() {
		if distance(pos, stair) < math.Max(rules.ObstacleGeneration.Clearance.Monster, rules.ObstacleGeneration.Clearance.Stairs) {
			reasons = append(reasons, "stair")
			break
		}
	}
	for _, teleporter := range out.teleporterPositions() {
		if distance(pos, teleporter) < math.Max(rules.ObstacleGeneration.Clearance.Monster, rules.ObstacleGeneration.Clearance.Teleporter) {
			reasons = append(reasons, "teleporter")
			break
		}
	}
	for _, chest := range out.chests {
		if chest.eliteObjective && monsterPackGuardsObjectiveAt(packID, out, chest.pos, rules.EliteObjective.RoomClusterRadius) {
			continue
		}
		if distance(pos, chest.pos) < math.Max(rules.ObstacleGeneration.Clearance.Monster, rules.ObstacleGeneration.Clearance.Chest) {
			reasons = append(reasons, "chest")
			break
		}
	}
	if out.reservedEliteObjectivePos != nil && distance(pos, *out.reservedEliteObjectivePos) < math.Max(math.Max(rules.ObstacleGeneration.Clearance.Monster, rules.ObstacleGeneration.Clearance.Chest), rules.MonsterPlacement.PackMemberRadius*2) {
		reasons = append(reasons, "reserved_objective")
	}
	for _, door := range out.doorPositions() {
		if distance(pos, door) < rules.ObstacleGeneration.Clearance.Monster {
			reasons = append(reasons, "door")
			break
		}
	}
	for _, monster := range out.monsters {
		if !((packID != "" && monster.packID == packID) || (packID == "" && monster.pos == pos)) && distance(pos, monster.pos) < encounterMonsterMinSeparation() {
			reasons = append(reasons, "monster")
			break
		}
	}
	for _, wall := range out.walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, rules.ObstacleGeneration.Clearance.Monster, wall.pos, wall.size) {
			reasons = append(reasons, "wall")
			break
		}
	}
	return reasons
}

func roomPackMemberPosition(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, center Vec2) (Vec2, bool) {
	if roomIndex < 0 || roomIndex >= len(out.rooms) {
		return Vec2{}, false
	}
	candidates := roomInteriorCandidates(out.rooms, []int{roomIndex}, rules.MonsterPlacement.MarginFromWall, rules.FloorSize)
	shuffleDungeonPositions(rng, candidates)
	for _, pos := range candidates {
		if distance(center, pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 || !insideDungeonFloor(pos, rules) || roomEncounterPositionBlocked(pos, rules, out, roomIndex) || !generatedTargetReachable(rules, out, pos) {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

func roomEncounterPositionBlocked(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int) bool {
	return roomEncounterPositionBlockedForPack(pos, rules, out, roomIndex, "")
}

func roomEncounterPositionBlockedForPack(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, packID string) bool {
	return roomEncounterPositionBlockedForPackWithObjectivePermission(pos, rules, out, roomIndex, packID, false)
}

func roomEncounterPositionBlockedForPackWithObjectivePermission(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, packID string, allowReservedObjectiveCluster bool) bool {
	p := rules.MonsterPlacement
	clearance := math.Max(0, rules.ObstacleGeneration.Clearance.Monster)
	interactableClearance := math.Max(clearance, math.Max(p.MarginFromWall, p.PackMemberRadius*2))
	if roomIndex < 0 || roomIndex >= len(out.rooms) || !roomContainsCircle(out.rooms[roomIndex], pos, p.MarginFromWall) || distance(pos, rules.PlayerSpawn) < p.MinSpawnDistance {
		return true
	}
	for _, anchor := range out.stairPositions() {
		if distance(pos, anchor) < math.Max(interactableClearance, rules.ObstacleGeneration.Clearance.Stairs) {
			return true
		}
	}
	for _, anchor := range out.teleporterPositions() {
		if distance(pos, anchor) < math.Max(interactableClearance, rules.ObstacleGeneration.Clearance.Teleporter) {
			return true
		}
	}
	for _, chest := range out.chests {
		if chest.eliteObjective && packID != "" && monsterPackGuardsObjectiveAt(packID, out, chest.pos, rules.EliteObjective.RoomClusterRadius) {
			continue
		}
		if distance(pos, chest.pos) < math.Max(interactableClearance, rules.ObstacleGeneration.Clearance.Chest) {
			return true
		}
	}
	if !allowReservedObjectiveCluster && out.reservedEliteObjectivePos != nil && distance(pos, *out.reservedEliteObjectivePos) < math.Max(math.Max(clearance, rules.ObstacleGeneration.Clearance.Chest), rules.MonsterPlacement.PackMemberRadius*2) {
		return true
	}
	for _, door := range out.doorPositions() {
		if distance(pos, door) < interactableClearance {
			return true
		}
	}
	for _, monster := range out.monsters {
		if (packID != "" && monster.packID == packID) || (packID == "" && monster.pos == pos) {
			continue
		}
		if distance(pos, monster.pos) < encounterMonsterMinSeparation() {
			return true
		}
	}
	for _, wall := range out.walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, rules.ObstacleGeneration.Clearance.Monster, wall.pos, wall.size) {
			return true
		}
	}
	return false
}

func monsterPackGuardsObjectiveAt(packID string, out generatedDungeonLevel, objectivePos Vec2, clusterRadius float64) bool {
	for _, monster := range out.monsters {
		if !monster.packLeader || monster.packID != packID || distance(monster.pos, objectivePos) > clusterRadius {
			continue
		}
		return true
	}
	return false
}

func roomEncounterMonsterBlocked(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	for _, monster := range out.monsters {
		if distance(pos, monster.pos) < encounterMonsterMinSeparation() {
			return true
		}
	}
	return false
}
