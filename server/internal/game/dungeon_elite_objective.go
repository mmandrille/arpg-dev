package game

import (
	"math"
	"sort"
)

func maybePlaceEliteObjectiveChest(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	objective := rules.EliteObjective
	if !objective.Enabled || !generatedLevelHasEliteLeader(*out) {
		return nil
	}
	if out.eliteObjectiveChanceResolved {
		if !out.eliteObjectiveChancePassed || out.reservedEliteObjectivePos == nil {
			return nil
		}
		if !hasEliteLeaderNearObjective(*out, rules) {
			return nil
		}
		out.chests = append(out.chests, generatedChest{defID: objective.InteractableDefID, lootTable: objective.LootTable, pos: *out.reservedEliteObjectivePos, eliteObjective: true})
		return nil
	}
	if !eliteObjectiveFloorChancePasses(rng, objective.FloorChancePercent) {
		return nil
	}
	pos, ok := randomObjectiveChestPosition(rng, rules, objective, out)
	if !ok {
		return nil
	}
	out.chests = append(out.chests, generatedChest{
		defID:          objective.InteractableDefID,
		lootTable:      objective.LootTable,
		pos:            pos,
		eliteObjective: true,
	})
	return nil
}

func reserveEliteObjectiveChestPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) {
	if !rules.EliteObjective.Enabled || !rules.RoomCorridorPCG.Enabled || !rules.RoomCorridorPCG.RoomRoles.Enabled {
		return
	}
	out.eliteObjectiveChanceResolved = true
	if !eliteObjectiveFloorChancePasses(rng, rules.EliteObjective.FloorChancePercent) {
		return
	}
	out.eliteObjectiveChancePassed = true
	pos, ok := randomObjectiveChestPosition(rng, rules, rules.EliteObjective, out)
	if ok {
		reserved := pos
		out.reservedEliteObjectivePos = &reserved
	}
}

func eliteObjectiveFloorChancePasses(rng *RNG, chancePercent int) bool {
	if chancePercent <= 0 {
		return false
	}

	return rng.IntN(100) < chancePercent
}

func generatedLevelHasEliteLeader(out generatedDungeonLevel) bool {
	for _, monster := range out.monsters {
		if monster.packLeader {
			return true
		}
	}
	return false
}

func randomObjectiveChestPosition(rng *RNG, rules DungeonGenerationRules, objective EliteObjectiveRules, out *generatedDungeonLevel) (Vec2, bool) {
	minX := int(math.Ceil(rules.MonsterPlacement.MarginFromWall))
	maxX := int(math.Floor(rules.FloorSize.Width - rules.MonsterPlacement.MarginFromWall))
	minY := int(math.Ceil(rules.MonsterPlacement.MarginFromWall))
	maxY := int(math.Floor(rules.FloorSize.Height - rules.MonsterPlacement.MarginFromWall))
	if maxX < minX || maxY < minY {
		return Vec2{}, false
	}
	monsterClearance := math.Max(rules.MonsterPlacement.MarginFromWall, rules.MonsterPlacement.PackMemberRadius*2)
	clusterRadius := objective.RoomClusterRadius
	role := rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective
	roomIndices := roomIndicesWithRole(out.rooms, role)
	roleEnabled := rules.RoomCorridorPCG.Enabled && rules.RoomCorridorPCG.RoomRoles.Enabled
	if !rules.RoomCorridorPCG.RoomRoles.Enabled {
		roomIndices = allRoomIndices(out.rooms)
	}
	if roleEnabled && len(roomIndices) == 0 {
		return Vec2{}, false
	}
	leaderPositions := make([]Vec2, 0)
	for _, monster := range out.monsters {
		if !monster.packLeader {
			continue
		}
		leaderPositions = append(leaderPositions, monster.pos)
	}
	shuffleDungeonPositions(rng, leaderPositions)
	tryPosition := func(pos Vec2, requireCluster, requireObjectiveRole bool) (Vec2, bool) {
		roomClearance := math.Max(playerRadius+0.1, rules.ObstacleGeneration.Clearance.Chest)
		if generatedPositionInCorridorZone(pos, rules.MonsterPlacement.PackMemberRadius, *out) {
			return Vec2{}, false
		}
		if requireObjectiveRole && roleEnabled && !pointInsideRoomWithRole(pos, out.rooms, role, roomClearance) {
			return Vec2{}, false
		}
		if requireCluster {
			nearLeader := false
			for _, leaderPos := range leaderPositions {
				if distance(pos, leaderPos) <= clusterRadius {
					nearLeader = true
					break
				}
			}
			if !nearLeader {
				return Vec2{}, false
			}
		}
		if !roleEnabled && generatedPositionInCorridorZone(pos, rules.MonsterPlacement.PackMemberRadius, *out) {
			return Vec2{}, false
		}
		blocked := false
		for _, stair := range out.stairPositions() {
			if distance(pos, stair) < objective.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked {
			return Vec2{}, false
		}
		for _, chest := range out.chestPositions() {
			if distance(pos, chest) < objective.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked {
			return Vec2{}, false
		}
		clusterPacks := map[string]bool{}
		if requireCluster {
			for _, monster := range out.monsters {
				if monster.packLeader && distance(pos, monster.pos) <= clusterRadius {
					clusterPacks[monster.packID] = true
				}
			}
		}
		for _, monster := range out.monsters {
			if !clusterPacks[monster.packID] && distance(pos, monster.pos) < monsterClearance {
				blocked = true
				break
			}
		}
		if blocked {
			return Vec2{}, false
		}
		chestClearance := rules.ObstacleGeneration.Clearance.Chest
		for _, wall := range out.walls {
			if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, chestClearance, wall.pos, wall.size) {
				blocked = true
				break
			}
		}
		if blocked || !generatedTargetReachable(rules, *out, pos) {
			return Vec2{}, false
		}

		return pos, true
	}
	var roomPositions []Vec2
	if roleEnabled {
		roomPositions = roomInteriorCandidates(out.rooms, roomIndices, math.Max(playerRadius+0.1, rules.ObstacleGeneration.Clearance.Chest), rules.FloorSize)
		shuffleDungeonPositions(rng, roomPositions)
	}
	distanceToLeader := func(pos Vec2) float64 {
		nearest := math.MaxFloat64
		for _, leaderPos := range leaderPositions {
			if d := distance(pos, leaderPos); d < nearest {
				nearest = d
			}
		}
		return nearest
	}
	if len(leaderPositions) > 0 {
		sort.SliceStable(roomPositions, func(i, j int) bool {
			return distanceToLeader(roomPositions[i]) < distanceToLeader(roomPositions[j])
		})
		for _, pos := range roomPositions {
			if distanceToLeader(pos) > clusterRadius {
				break
			}
			if placed, ok := tryPosition(pos, true, roleEnabled); ok {
				return placed, true
			}
		}
	}
	roomPositionAttempts := objective.MaxAttempts
	if roleEnabled {
		roomPositionAttempts = len(roomPositions)
	}
	for attempt := 0; attempt < roomPositionAttempts; attempt++ {
		var pos Vec2
		if roleEnabled {
			pos = roomPositions[attempt]
		} else {
			pos = Vec2{
				X: float64(minX + rng.IntN(maxX-minX+1)),
				Y: float64(minY + rng.IntN(maxY-minY+1)),
			}
		}
		if placed, ok := tryPosition(pos, false, roleEnabled); ok {
			return placed, true
		}
	}
	return Vec2{}, false
}

func elitePackLeaderPosition(out generatedDungeonLevel) (Vec2, bool) {
	for _, monster := range out.monsters {
		if monster.packLeader {
			return monster.pos, true
		}
	}
	return Vec2{}, false
}
