package game

func (s *Sim) resetPlayerNavigationBudget() {
	s.playerPathNodesThisTick = 0
}

func (s *Sim) playerPathBudgetAvailable() bool {
	nav := s.activeNav()
	return s.playerPathNodesThisTick < nav.PlayerPathNodesPerTick
}

func (s *Sim) planPlayerPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	steps, ok, stats := s.planPlayerPathSearch(nav, start, goal, blocked)
	if stats.NodesVisited > 0 {
		s.playerPathNodesThisTick += stats.NodesVisited
	}
	if stats.LimitExceeded {
		return nil, false
	}
	if ok && len(steps) > nav.PlayerMaxAutoSteps {
		steps = append([]Vec2(nil), steps[:nav.PlayerMaxAutoSteps]...)
	}

	return steps, ok
}

func (s *Sim) planPlayerPathForApproach(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	if !s.playerPathBudgetAvailable() {
		return nil, false
	}
	steps, ok, stats := s.planPlayerPathSearch(nav, start, goal, blocked)
	charge := stats.NodesVisited
	if !ok {
		failedCap := 64
		if charge > failedCap {
			charge = failedCap
		}
	}
	if charge > 0 {
		s.playerPathNodesThisTick += charge
	}
	if stats.LimitExceeded {
		return nil, false
	}
	if ok && len(steps) > nav.PlayerMaxAutoSteps {
		steps = append([]Vec2(nil), steps[:nav.PlayerMaxAutoSteps]...)
	}

	return steps, ok
}

func (s *Sim) planPlayerPathSearch(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool, PathSearchStats) {
	if !s.playerPathBudgetAvailable() {
		return nil, false, PathSearchStats{}
	}
	remainingTick := nav.PlayerPathNodesPerTick - s.playerPathNodesThisTick
	nodeLimit := nav.PlayerPathNodesPerSearch
	startCell := worldToGrid(nav, start)
	goalCell := worldToGrid(nav, goal)
	dist := octile(startCell, goalCell)
	if dist > 0 {
		scaled := dist*128 + 64
		if scaled > nodeLimit {
			nodeLimit = scaled
		}
	}
	if nodeLimit > remainingTick {
		nodeLimit = remainingTick
	}
	if nodeLimit <= 0 {
		return nil, false, PathSearchStats{}
	}
	steps, ok, stats := s.planPathWithNodeLimit(nav, start, goal, blocked, nodeLimit)

	return steps, ok, stats
}

func (s *Sim) planPlayerPathForInteractable(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	steps, ok, stats := s.planPlayerPathSearch(nav, start, goal, blocked)
	if stats.NodesVisited > 0 {
		s.playerPathNodesThisTick += stats.NodesVisited
	}
	if stats.LimitExceeded {
		return nil, false
	}
	maxSteps := nav.MaxAutoSteps
	if maxSteps < nav.PlayerMaxAutoSteps {
		maxSteps = nav.PlayerMaxAutoSteps
	}
	if ok && len(steps) > maxSteps {
		steps = append([]Vec2(nil), steps[:maxSteps]...)
	}

	return steps, ok
}
