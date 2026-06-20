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

// obstacleBlocksMovement is the grounded baseline: every obstacle kind blocks
// ground movement. Per-monster trait and per-skill exceptions (e.g. flying over
// water/holes, barbarian leap ignoring kinds) live in monsterObstacleBlocksMovement
// and skillMobilityIgnoresObstacleKind, not here.
func obstacleBlocksMovement(w wallObstacle) bool {
	return true
}

func obstacleBlocksProjectiles(w wallObstacle) bool {
	return solidObstacleBlocksProjectiles(w.obstacleKind())
}

func obstacleBlocksLineOfSight(w wallObstacle) bool {
	if w.blocksLOS != nil {
		return *w.blocksLOS
	}
	return w.obstacleKind() == obstacleKindWall
}
