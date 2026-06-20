package game

import "fmt"

type SolidObstacleKindWeights struct {
	Wall   int `json:"wall"`
	Rock   int `json:"rock"`
	Column int `json:"column"`
	Rubble int `json:"rubble"`
}

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
	case obstacleKindWall, obstacleKindRock, obstacleKindColumn, obstacleKindRubble:
		return true
	default:
		return false
	}
}

func solidObstacleLineOfSightOverride(kind string) *bool {
	switch kind {
	case obstacleKindRock, obstacleKindColumn:
		return boolPtr(true)
	default:
		return nil
	}
}
