package game

func (s *Sim) findMeleeApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	return s.findApproachGoalMatching(target, func(pos Vec2, target *entity) bool {
		return meleeInRange(distance(pos, target.pos), s.playerMeleeReach(), s.targetInteractionRadius(target))
	})
}

func (s *Sim) findApproachGoalMatching(target *entity, inRange func(Vec2, *entity) bool) (Vec2, []Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.activeNav()
	targetCell := worldToGrid(nav, target.pos)
	blocked := s.buildBlockedFn()
	preferClosedBarrierSide := target.state == interactableClosed && s.hasClosedBarrier(target)
	bestGoal := Vec2{}
	bestSteps := []Vec2(nil)
	bestSide := 0
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := 0; radius <= maxRadius; radius++ {
		candidates := ringCells(targetCell, radius)
		for _, cell := range candidates {
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			if !inRange(goal, target) {
				continue
			}
			steps, ok := PlanPath(nav, player.pos, goal, blocked)
			if !ok {
				continue
			}
			if !preferClosedBarrierSide {
				return goal, steps, true
			}
			side := s.closedBarrierApproachSideScore(player.pos, target, goal)
			if bestSteps == nil || side < bestSide || (side == bestSide && len(steps) < len(bestSteps)) {
				bestGoal = goal
				bestSteps = steps
				bestSide = side
			}
		}
	}
	if bestSteps == nil {
		return Vec2{}, nil, false
	}
	return bestGoal, bestSteps, true
}
