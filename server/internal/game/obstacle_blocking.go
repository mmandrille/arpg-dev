package game

const (
	obstacleKindWall  = "wall"
	obstacleKindWater = "water"
	obstacleKindHole  = "hole"
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
	return w.obstacleKind() == obstacleKindWall
}

func obstacleBlocksLineOfSight(w wallObstacle) bool {
	return w.obstacleKind() == obstacleKindWall
}
