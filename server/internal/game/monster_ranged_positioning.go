package game

import "math"

func (s *Sim) monsterRangedRetreatGoal(monster *entity, player *entity, def MonsterDef) (Vec2, bool) {
	if def.effectiveAttackMode() != attackModeRanged || def.PreferredMinRange <= 0 {
		return Vec2{}, false
	}
	currentDistance := distance(monster.pos, player.pos)
	preferredMin := maxFloat(def.PreferredMinRange, playerRadius+monsterRadius+0.05)
	attackMax := s.monsterAttackReach(def) + playerRadius + meleeRangeEpsilon
	if preferredMin > attackMax {
		preferredMin = attackMax
	}
	if currentDistance >= preferredMin {
		return Vec2{}, false
	}
	nav := s.activeNav()
	if goal, ok := s.cachedMonsterNavigationGoal(monster, player); ok {
		goalDistance := distance(goal, player.pos)
		if goalDistance >= preferredMin && goalDistance <= attackMax && goalDistance > currentDistance+nav.StopDistance && s.hasClearMonsterRangedShot(goal, player) {
			return goal, true
		}
		monster.navPathValid = false
		monster.navNextRepathTick = s.tick
	}
	blocked := s.buildMonsterBlockedFn(monster.id)
	goal, steps, ok := s.findMonsterRangedRetreatGoal(monster, player, def, blocked, currentDistance, preferredMin, attackMax)
	if !ok {
		return Vec2{}, false
	}
	s.cacheMonsterNavigationPath(monster, player.id, goal, steps)
	return goal, true
}

func (s *Sim) findMonsterRangedRetreatGoal(monster *entity, player *entity, def MonsterDef, blocked func(gx, gy int) bool, currentDistance float64, preferredMin float64, attackMax float64) (Vec2, []Vec2, bool) {
	nav := s.activeNav()
	var (
		bestGoal           Vec2
		bestSteps          []Vec2
		bestPathLen        int
		bestMonsterDist    float64
		bestPlayerDistance float64
		found              bool
	)
	for _, goal := range monsterRangedRetreatCandidates(monster, player, preferredMin, attackMax) {
		if !s.monsterPathBudgetAvailable() {
			return Vec2{}, nil, false
		}
		if !s.positionInNavigationBounds(nav, goal) || s.monsterPositionBlocked(goal, monster.id) {
			continue
		}
		playerDistance := distance(goal, player.pos)
		if playerDistance < preferredMin || playerDistance <= currentDistance+nav.StopDistance || playerDistance > attackMax {
			continue
		}
		if !s.hasClearMonsterRangedShot(goal, player) {
			continue
		}
		steps, ok := s.planMonsterPath(monster, nav, monster.pos, goal, blocked)
		if !ok {
			continue
		}
		if len(steps) > 0 && s.resolveMonsterMovement(monster, s.monsterMoveDelta(monster.pos, goal, steps, s.monsterMoveSpeed(monster, def, nav))) == monster.pos {
			continue
		}
		monsterDist := distance(monster.pos, goal)
		if !found || len(steps) < bestPathLen ||
			(len(steps) == bestPathLen && monsterDist < bestMonsterDist-1e-9) ||
			(len(steps) == bestPathLen && monsterDist <= bestMonsterDist+1e-9 && bestMonsterDist <= monsterDist+1e-9 && playerDistance > bestPlayerDistance+1e-9) ||
			(len(steps) == bestPathLen && monsterDist <= bestMonsterDist+1e-9 && bestMonsterDist <= monsterDist+1e-9 && playerDistance <= bestPlayerDistance+1e-9 && bestPlayerDistance <= playerDistance+1e-9 && vecLess(goal, bestGoal)) {
			bestGoal = goal
			bestSteps = steps
			bestPathLen = len(steps)
			bestMonsterDist = monsterDist
			bestPlayerDistance = playerDistance
			found = true
		}
	}
	return bestGoal, bestSteps, found
}

func monsterRangedRetreatCandidates(monster *entity, player *entity, preferredMin float64, attackMax float64) []Vec2 {
	distances := []float64{preferredMin}
	if attackMax > preferredMin+1.0 {
		distances = append(distances, attackMax)
	}
	directions := []Vec2{}
	addDirection := func(dir Vec2) {
		if dir.X == 0 && dir.Y == 0 {
			return
		}
		normalized := normalize(dir)
		for _, existing := range directions {
			if math.Abs(existing.X-normalized.X) <= 1e-6 && math.Abs(existing.Y-normalized.Y) <= 1e-6 {
				return
			}
		}
		directions = append(directions, normalized)
	}
	addDirection(Vec2{X: monster.pos.X - player.pos.X, Y: monster.pos.Y - player.pos.Y})
	for i := 0; i < 16; i++ {
		angle := (2 * math.Pi * float64(i)) / 16
		addDirection(Vec2{X: math.Cos(angle), Y: math.Sin(angle)})
	}
	candidates := make([]Vec2, 0, len(directions)*len(distances))
	for _, dist := range distances {
		for _, dir := range directions {
			candidates = append(candidates, Vec2{X: player.pos.X + dir.X*dist, Y: player.pos.Y + dir.Y*dist})
		}
	}
	return candidates
}

func (s *Sim) monsterFallbackChaseGoal(monster *entity, player *entity, def MonsterDef, blocked func(gx, gy int) bool, attempted bool) (Vec2, bool) {
	if def.effectiveAttackMode() == attackModeRanged {
		if goal, steps, ok := s.findCloserMonsterRangedChaseGoal(monster, player, def, blocked); ok {
			s.cacheMonsterNavigationPath(monster, player.id, goal, steps)
			return goal, true
		}
	}
	if attempted || !s.monsterPathBudgetAvailable() {
		s.scheduleMonsterRepath(monster)
	}
	return Vec2{}, false
}

func (s *Sim) findCloserMonsterRangedChaseGoal(monster *entity, player *entity, def MonsterDef, blocked func(gx, gy int) bool) (Vec2, []Vec2, bool) {
	nav := s.activeNav()
	playerCell := worldToGrid(nav, player.pos)
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	minSeparation := playerRadius + monsterRadius + 0.05
	for radius := 1; radius <= maxRadius; radius++ {
		for _, cell := range ringCells(playerCell, radius) {
			if !s.monsterPathBudgetAvailable() {
				return Vec2{}, nil, false
			}
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			playerDistance := distance(goal, player.pos)
			if playerDistance < minSeparation || playerDistance > s.monsterAttackReach(def)+playerRadius+meleeRangeEpsilon {
				continue
			}
			if !s.hasClearMonsterRangedShot(goal, player) {
				continue
			}
			steps, ok := s.planMonsterPath(monster, nav, monster.pos, goal, blocked)
			if !ok {
				continue
			}
			if len(steps) > 0 && s.resolveMonsterMovement(monster, s.monsterMoveDelta(monster.pos, goal, steps, s.monsterMoveSpeed(monster, def, nav))) == monster.pos {
				continue
			}
			return goal, steps, true
		}
	}
	return Vec2{}, nil, false
}
