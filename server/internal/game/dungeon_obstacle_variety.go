package game

import (
	"fmt"
	"math"
)

type SolidObstacleKindWeights struct {
	Wall   int `json:"wall"`
	Rock   int `json:"rock"`
	Column int `json:"column"`
	Rubble int `json:"rubble"`
}

var solidObstacleLineOfSightTrue = true

func (w SolidObstacleKindWeights) total() int {
	return w.Wall + w.Rock + w.Column + w.Rubble
}

func validateSolidObstacleKindWeights(weights SolidObstacleKindWeights) error {
	if weights.Wall < 0 || weights.Rock < 0 || weights.Column < 0 || weights.Rubble < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.solid_kind_weights: must be non-negative")
	}
	if weights.total() <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.solid_kind_weights: at least one solid kind must be enabled")
	}
	return nil
}

func chooseSolidObstacleKind(rng *RNG, weights SolidObstacleKindWeights) string {
	draw := rng.IntN(weights.total())
	if draw < weights.Wall {
		return obstacleKindWall
	}
	draw -= weights.Wall
	if draw < weights.Rock {
		return obstacleKindRock
	}
	draw -= weights.Rock
	if draw < weights.Column {
		return obstacleKindColumn
	}
	return obstacleKindRubble
}

func solidObstacleBlocksProjectiles(kind string) bool {
	switch kind {
	case obstacleKindWall, obstacleKindWood, obstacleKindRock, obstacleKindColumn, obstacleKindRubble:
		return true
	default:
		return false
	}
}

func solidObstacleLineOfSightOverride(kind string) *bool {
	switch kind {
	case obstacleKindRock, obstacleKindColumn:
		return &solidObstacleLineOfSightTrue
	default:
		return nil
	}
}

func validateObstacleGenerationRules(o ObstacleGenerationRules, floor DungeonFloorSize) error {
	if o.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.max_attempts: must be positive")
	}
	if o.WallSegment.MinLength <= 0 || o.WallSegment.MaxLength < o.WallSegment.MinLength {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.wall_segment: invalid min/max length")
	}
	if o.WallSegment.Thickness <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.wall_segment.thickness: must be positive")
	}
	if o.SolidBlock.MinSize.X <= 0 || o.SolidBlock.MinSize.Y <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.solid_block.min_size: must be positive")
	}
	if o.SolidBlock.MaxSize.X < o.SolidBlock.MinSize.X || o.SolidBlock.MaxSize.Y < o.SolidBlock.MinSize.Y {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.solid_block: invalid min/max size")
	}
	if o.ShapeWeights.Line < 0 || o.ShapeWeights.L < 0 || o.ShapeWeights.T < 0 || o.ShapeWeights.Block < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.shape_weights: must be non-negative")
	}
	if o.ShapeWeights.total() <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.shape_weights: at least one shape must be enabled")
	}
	if err := validateSolidObstacleKindWeights(o.SolidKindWeights); err != nil {
		return err
	}
	for label, value := range map[string]float64{
		"player_spawn": o.Clearance.PlayerSpawn,
		"stairs":       o.Clearance.Stairs,
		"teleporter":   o.Clearance.Teleporter,
		"chest":        o.Clearance.Chest,
		"monster":      o.Clearance.Monster,
		"loot":         o.Clearance.Loot,
	} {
		if value < 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.clearance.%s: must be non-negative", label)
		}
	}
	maxSpan := math.Max(float64(o.WallSegment.MaxLength), math.Max(o.SolidBlock.MaxSize.X, o.SolidBlock.MaxSize.Y))
	if o.Water.Enabled {
		maxSpan = math.Max(maxSpan, math.Max(o.Water.MaxSize.X, o.Water.MaxSize.Y))
	}
	if o.Holes.Enabled {
		maxSpan = math.Max(maxSpan, math.Max(o.Holes.MaxSize.X, o.Holes.MaxSize.Y))
	}
	if maxSpan >= math.Min(floor.Width, floor.Height) {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation: largest obstacle must fit inside floor")
	}
	return nil
}
