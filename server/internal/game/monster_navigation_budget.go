package game

func (s *Sim) resetMonsterNavigationBudget() {
	s.monsterPathRequestsThisTick = 0
	s.monsterPathNodesThisTick = 0
}

func (s *Sim) monsterPathBudgetAvailable() bool {
	nav := s.activeNav()
	return s.monsterPathRequestsThisTick < nav.MonsterPathRequestsPerTick &&
		s.monsterPathNodesThisTick < nav.MonsterPathNodesPerTick
}

func (s *Sim) monsterCanRepath(monster *entity) bool {
	return monster == nil || monster.navNextRepathTick <= s.tick
}

func (s *Sim) planMonsterPath(monster *entity, nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	if !s.monsterCanRepath(monster) || !s.monsterPathBudgetAvailable() {
		return nil, false
	}
	remainingNodes := nav.MonsterPathNodesPerTick - s.monsterPathNodesThisTick
	if remainingNodes <= 0 {
		return nil, false
	}
	s.monsterPathRequestsThisTick++
	stats := PathSearchStats{NodeLimit: remainingNodes}
	steps, ok := s.runPathSearch(nav, start, goal, blocked, &stats)
	s.monsterPathNodesThisTick += stats.NodesVisited
	if stats.LimitExceeded {
		return nil, false
	}
	return steps, ok
}

func (s *Sim) cachedMonsterNavigationGoal(monster *entity, player *entity) (Vec2, bool) {
	if monster == nil || player == nil || !monster.navPathValid {
		return Vec2{}, false
	}
	nav := s.activeNav()
	if monster.navTargetPlayerID != player.id || nav.MonsterPathCacheTicks <= 0 {
		monster.navPathValid = false
		return Vec2{}, false
	}
	if s.tick-monster.navPathTick > uint64(nav.MonsterPathCacheTicks) {
		monster.navPathValid = false
		return Vec2{}, false
	}
	if !s.positionInNavigationBounds(nav, monster.navGoal) || s.monsterPositionBlocked(monster.navGoal, monster.id) {
		monster.navPathValid = false
		return Vec2{}, false
	}
	s.tickPerf.PathCacheHits++
	return monster.navGoal, true
}

func (s *Sim) monsterPathForMovement(monster *entity, nav NavigationRules, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	if steps, ok := s.cachedMonsterPathForGoal(monster, goal); ok {
		return steps, true
	}
	if !s.monsterCanRepath(monster) {
		return nil, false
	}
	steps, ok := s.planMonsterPath(monster, nav, monster.pos, goal, blocked)
	if ok {
		s.cacheMonsterNavigationPath(monster, monster.aiTargetPlayerID, goal, steps)
		return steps, true
	}
	s.scheduleMonsterRepath(monster)
	return nil, false
}

func (s *Sim) cachedMonsterPathForGoal(monster *entity, goal Vec2) ([]Vec2, bool) {
	if monster == nil || !monster.navPathValid || distance(monster.navGoal, goal) > 1e-9 {
		return nil, false
	}
	s.advanceCachedMonsterPath(monster)
	s.tickPerf.PathCacheHits++
	return monster.navPath, true
}

func (s *Sim) advanceCachedMonsterPath(monster *entity) {
	current := worldToGrid(s.activeNav(), monster.pos)
	for monster.navPathValid && len(monster.navPath) > 0 && current != monster.navPathCell {
		step := monster.navPath[0]
		monster.navPath = monster.navPath[1:]
		monster.navPathCell = gridCell{
			x: monster.navPathCell.x + int(step.X),
			y: monster.navPathCell.y + int(step.Y),
		}
	}
}

func (s *Sim) cacheMonsterNavigationPath(monster *entity, targetPlayerID uint64, goal Vec2, steps []Vec2) {
	if monster == nil {
		return
	}
	monster.navGoal = goal
	monster.navPath = append(monster.navPath[:0], steps...)
	monster.navPathValid = true
	monster.navPathCell = worldToGrid(s.activeNav(), monster.pos)
	monster.navTargetPlayerID = targetPlayerID
	monster.navPathTick = s.tick
	s.scheduleMonsterRepath(monster)
}

func (s *Sim) scheduleMonsterRepath(monster *entity) {
	if monster == nil {
		return
	}
	nav := s.activeNav()
	delay := nav.MonsterRepathThrottleTicks
	if nav.MonsterRepathStaggerTicks > 1 {
		delay += int(monster.id % uint64(nav.MonsterRepathStaggerTicks))
	}
	monster.navNextRepathTick = s.tick + uint64(delay)
}
