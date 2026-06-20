package game

import (
	"fmt"
	"math"
	"strconv"
)

type HoleGenerationRules struct {
	Enabled     bool     `json:"enabled"`
	MaxAttempts int      `json:"max_attempts"`
	TargetCount IntRange `json:"target_count"`
	MinSize     Vec2     `json:"min_size"`
	MaxSize     Vec2     `json:"max_size"`
}

func validateHoleGenerationRules(holes HoleGenerationRules, floor DungeonFloorSize) error {
	if !holes.Enabled {
		return nil
	}
	if holes.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.holes.max_attempts: must be positive")
	}
	if holes.TargetCount.Min < 0 || holes.TargetCount.Max < holes.TargetCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.holes.target_count: invalid min/max")
	}
	if holes.MinSize.X <= 0 || holes.MinSize.Y <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.holes.min_size: must be positive")
	}
	if holes.MaxSize.X < holes.MinSize.X || holes.MaxSize.Y < holes.MinSize.Y {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.holes: invalid min/max size")
	}
	maxSpan := math.Max(holes.MaxSize.X, holes.MaxSize.Y)
	if maxSpan >= math.Min(floor.Width, floor.Height) {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.holes: largest hole obstacle must fit inside floor")
	}
	return nil
}

func placeDungeonHoles(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	holes := rules.ObstacleGeneration.Holes
	if !holes.Enabled || holes.TargetCount.Max == 0 {
		return nil
	}
	baseWalls := append([]wallObstacle(nil), out.walls...)
	for attempt := 0; attempt < holes.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + "|holes|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		generated, ok := randomHoleObstacles(rng, rules, *out)
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
	return fmt.Errorf("game: generate dungeon level %d: could not place reachable hole obstacles after %d attempts", out.levelNum, holes.MaxAttempts)
}

func randomHoleObstacles(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel) ([]wallObstacle, bool) {
	holes := rules.ObstacleGeneration.Holes
	target := holes.TargetCount.Min
	if holes.TargetCount.Max > holes.TargetCount.Min {
		target += rng.IntN(holes.TargetCount.Max - holes.TargetCount.Min + 1)
	}
	generated := make([]wallObstacle, 0, target)
	for len(generated) < target {
		placed := false
		for try := 0; try < 32; try++ {
			obstacle := randomHoleObstacle(rng, rules)
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

func randomHoleObstacle(rng *RNG, rules DungeonGenerationRules) wallObstacle {
	holes := rules.ObstacleGeneration.Holes
	width := float64(randomIntRange(rng, int(math.Ceil(holes.MinSize.X)), int(math.Floor(holes.MaxSize.X))))
	height := float64(randomIntRange(rng, int(math.Ceil(holes.MinSize.Y)), int(math.Floor(holes.MaxSize.Y))))
	size := Vec2{X: width, Y: height}
	return wallObstacle{
		pos:         randomWallCenter(rng, rules, size),
		size:        size,
		source:      "generated",
		shapeFamily: obstacleKindHole,
		kind:        obstacleKindHole,
	}
}
