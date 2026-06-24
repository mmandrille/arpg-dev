package game

import "math"

func maybePlaceEliteObjectiveChest(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	objective := rules.EliteObjective
	if !objective.Enabled || !generatedLevelHasEliteLeader(*out) {
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
	leaderPos, hasLeader := elitePackLeaderPosition(*out)
	clusterRadius := objective.RoomClusterRadius
	tryPosition := func(pos Vec2, requireCluster bool) (Vec2, bool) {
		if requireCluster && hasLeader && distance(pos, leaderPos) > clusterRadius {
			return Vec2{}, false
		}
		if generatedPositionInCorridorZone(pos, rules.MonsterPlacement.PackMemberRadius, *out) {
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
		for _, monster := range out.monsters {
			if distance(pos, monster.pos) < monsterClearance {
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
	if hasLeader {
		diameter := int(math.Ceil(clusterRadius * 2))
		if diameter < 1 {
			diameter = 1
		}
		for attempt := 0; attempt < objective.MaxAttempts; attempt++ {
			offsetX := rng.IntN(diameter+1) - diameter/2
			offsetY := rng.IntN(diameter+1) - diameter/2
			pos := Vec2{X: leaderPos.X + float64(offsetX), Y: leaderPos.Y + float64(offsetY)}
			if distance(pos, leaderPos) > clusterRadius {
				continue
			}
			if placed, ok := tryPosition(pos, true); ok {
				return placed, true
			}
		}
	}
	for attempt := 0; attempt < objective.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		if placed, ok := tryPosition(pos, false); ok {
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
