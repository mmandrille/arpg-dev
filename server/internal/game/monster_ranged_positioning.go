package game

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
