package game

import (
	"fmt"
	"math"
	"strconv"
)

// FloorFeatureGenerationRules tunes a hard-blocking, reachability-validated floor
// hazard (water or holes). Water and holes share one generation algorithm and
// differ only in their rules block, deterministic seed token, and obstacle kind.
type FloorFeatureGenerationRules struct {
	Enabled     bool     `json:"enabled"`
	MaxAttempts int      `json:"max_attempts"`
	TargetCount IntRange `json:"target_count"`
	MinSize     Vec2     `json:"min_size"`
	MaxSize     Vec2     `json:"max_size"`
}

// floorFeatureSpec binds a feature's rules to its identity: the config-path leaf
// used in validation messages, the deterministic seed token, and the obstacle
// kind stamped onto generated walls.
type floorFeatureSpec struct {
	rules     FloorFeatureGenerationRules
	field     string
	seedToken string
	kind      string
}

func validateFloorFeatureRules(spec floorFeatureSpec, floor DungeonFloorSize) error {
	r := spec.rules
	if !r.Enabled {
		return nil
	}
	base := "game: invalid rules dungeon_generation.obstacle_generation." + spec.field
	if r.MaxAttempts <= 0 {
		return fmt.Errorf("%s.max_attempts: must be positive", base)
	}
	if r.TargetCount.Min < 0 || r.TargetCount.Max < r.TargetCount.Min {
		return fmt.Errorf("%s.target_count: invalid min/max", base)
	}
	if r.MinSize.X <= 0 || r.MinSize.Y <= 0 {
		return fmt.Errorf("%s.min_size: must be positive", base)
	}
	if r.MaxSize.X < r.MinSize.X || r.MaxSize.Y < r.MinSize.Y {
		return fmt.Errorf("%s: invalid min/max size", base)
	}
	maxSpan := math.Max(r.MaxSize.X, r.MaxSize.Y)
	if maxSpan >= math.Min(floor.Width, floor.Height) {
		return fmt.Errorf("%s: largest %s obstacle must fit inside floor", base, spec.kind)
	}
	return nil
}

func placeDungeonFloorFeature(seed string, spec floorFeatureSpec, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	if !spec.rules.Enabled || spec.rules.TargetCount.Max == 0 {
		return nil
	}
	baseWalls := append([]wallObstacle(nil), out.walls...)
	for attempt := 0; attempt < spec.rules.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + spec.seedToken + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		generated, ok := randomFloorFeatureObstacles(rng, spec, rules, *out)
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
	return fmt.Errorf("game: generate dungeon level %d: could not place reachable %s obstacles after %d attempts", out.levelNum, spec.kind, spec.rules.MaxAttempts)
}

func randomFloorFeatureObstacles(rng *RNG, spec floorFeatureSpec, rules DungeonGenerationRules, out generatedDungeonLevel) ([]wallObstacle, bool) {
	target := spec.rules.TargetCount.Min
	if spec.rules.TargetCount.Max > spec.rules.TargetCount.Min {
		target += rng.IntN(spec.rules.TargetCount.Max - spec.rules.TargetCount.Min + 1)
	}
	generated := make([]wallObstacle, 0, target)
	for len(generated) < target {
		placed := false
		for try := 0; try < 32; try++ {
			obstacle := randomFloorFeatureObstacle(rng, spec, rules)
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

func randomFloorFeatureObstacle(rng *RNG, spec floorFeatureSpec, rules DungeonGenerationRules) wallObstacle {
	r := spec.rules
	width := float64(randomIntRange(rng, int(math.Ceil(r.MinSize.X)), int(math.Floor(r.MaxSize.X))))
	height := float64(randomIntRange(rng, int(math.Ceil(r.MinSize.Y)), int(math.Floor(r.MaxSize.Y))))
	size := Vec2{X: width, Y: height}
	return wallObstacle{
		pos:         randomWallCenter(rng, rules, size),
		size:        size,
		source:      "generated",
		shapeFamily: spec.kind,
		kind:        spec.kind,
	}
}
