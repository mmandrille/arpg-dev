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
	for attempt := 0; attempt < objective.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		blocked := false
		for _, stair := range out.stairPositions() {
			if distance(pos, stair) < objective.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		for _, chest := range out.chestPositions() {
			if distance(pos, chest) < objective.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked || !generatedTargetReachable(rules, *out, pos) {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}
