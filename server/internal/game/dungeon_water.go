package game

import (
	"fmt"
	"math"
	"strconv"
)

type WaterGenerationRules struct {
	Enabled     bool     `json:"enabled"`
	MaxAttempts int      `json:"max_attempts"`
	TargetCount IntRange `json:"target_count"`
	MinSize     Vec2     `json:"min_size"`
	MaxSize     Vec2     `json:"max_size"`
}

func validateWaterGenerationRules(water WaterGenerationRules, floor DungeonFloorSize) error {
	if !water.Enabled {
		return nil
	}
	if water.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.water.max_attempts: must be positive")
	}
	if water.TargetCount.Min < 0 || water.TargetCount.Max < water.TargetCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.water.target_count: invalid min/max")
	}
	if water.MinSize.X <= 0 || water.MinSize.Y <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.water.min_size: must be positive")
	}
	if water.MaxSize.X < water.MinSize.X || water.MaxSize.Y < water.MinSize.Y {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.water: invalid min/max size")
	}
	maxSpan := math.Max(water.MaxSize.X, water.MaxSize.Y)
	if maxSpan >= math.Min(floor.Width, floor.Height) {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.water: largest water obstacle must fit inside floor")
	}
	return nil
}

func placeDungeonWater(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	water := rules.ObstacleGeneration.Water
	if !water.Enabled || water.TargetCount.Max == 0 {
		return nil
	}
	baseWalls := append([]wallObstacle(nil), out.walls...)
	for attempt := 0; attempt < water.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + "|water|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		generated, ok := randomWaterObstacles(rng, rules, *out)
		if !ok {
			continue
		}
		candidate := *out
		candidate.walls = append(append([]wallObstacle(nil), baseWalls...), generated...)
		if err := validateGeneratedDungeonReachability(rules, candidate); err != nil {
			continue
		}
		*out = candidate
		return nil
	}
	out.walls = baseWalls
	return fmt.Errorf("game: generate dungeon level %d: could not place reachable water obstacles after %d attempts", out.levelNum, water.MaxAttempts)
}

func randomWaterObstacles(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel) ([]wallObstacle, bool) {
	water := rules.ObstacleGeneration.Water
	target := water.TargetCount.Min
	if water.TargetCount.Max > water.TargetCount.Min {
		target += rng.IntN(water.TargetCount.Max - water.TargetCount.Min + 1)
	}
	generated := make([]wallObstacle, 0, target)
	for len(generated) < target {
		placed := false
		for try := 0; try < 32; try++ {
			obstacle := randomWaterObstacle(rng, rules)
			if !obstacleGroupAllowed([]wallObstacle{obstacle}, rules, out, generated) {
				continue
			}
			if !floorFeatureClearsGeneratedDoors(obstacle, rules, out) {
				continue
			}
			generated = append(generated, obstacle)
			placed = true
			break
		}
		if !placed {
			return nil, false
		}
	}
	return generated, true
}

func floorFeatureClearsGeneratedDoors(feature wallObstacle, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	clearance := maxFloat(0.75, rules.ObstacleGeneration.WallSegment.Thickness)
	for _, door := range out.doors {
		if circleIntersectsAABB(door.pos, clearance, feature.pos, feature.size) {
			return false
		}
	}
	return true
}

func randomWaterObstacle(rng *RNG, rules DungeonGenerationRules) wallObstacle {
	water := rules.ObstacleGeneration.Water
	width := float64(randomIntRange(rng, int(math.Ceil(water.MinSize.X)), int(math.Floor(water.MaxSize.X))))
	height := float64(randomIntRange(rng, int(math.Ceil(water.MinSize.Y)), int(math.Floor(water.MaxSize.Y))))
	size := Vec2{X: width, Y: height}
	return wallObstacle{
		pos:         randomWallCenter(rng, rules, size),
		size:        size,
		source:      "generated",
		shapeFamily: obstacleKindWater,
		kind:        obstacleKindWater,
	}
}
