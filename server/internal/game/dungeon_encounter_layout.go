package game

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

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
