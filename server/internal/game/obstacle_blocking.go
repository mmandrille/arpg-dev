package game

const (
	obstacleKindWall   = "wall"
	obstacleKindWater  = "water"
	obstacleKindHole   = "hole"
	obstacleKindRock   = "rock"
	obstacleKindColumn = "column"
	obstacleKindRubble = "rubble"
)

func (w wallObstacle) obstacleKind() string {
	if w.kind == "" {
		return obstacleKindWall
	}
	return w.kind
}

func obstacleBlocksMovement(w wallObstacle) bool {
	switch w.obstacleKind() {
	case obstacleKindWater, obstacleKindHole:
		return true
	default:
		return true
	}
}

func obstacleBlocksProjectiles(w wallObstacle) bool {
	return solidObstacleBlocksProjectiles(w.obstacleKind())
}

func obstacleBlocksLineOfSight(w wallObstacle) bool {
	return w.obstacleKind() == obstacleKindWall
}
