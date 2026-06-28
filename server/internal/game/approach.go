package game

import (
	"math"
	"sort"
)

func (s *Sim) findApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	if target.kind == monsterEntity && s.playerAttackMode() == attackModeRanged {
		return s.findRangedApproachGoal(target)
	}
	return s.findMeleeApproachGoal(target)
}

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
	if target.kind == interactableEntity && !(target.state == interactableClosed && s.hasClosedBarrier(target)) {
		steps, ok := s.planPlayerPathForInteractable(nav, player.pos, target.pos, blocked)
		if ok {
			return target.pos, steps, true
		}
		fallbackGoal := interactableApproachMoveGoal(player.pos, target.pos, s.targetInteractionRadius(target))
		steps, ok = s.planPlayerPathForInteractable(nav, player.pos, fallbackGoal, blocked)
		if ok {
			return fallbackGoal, steps, true
		}
	}
	preferClosedBarrierSide := target.state == interactableClosed && s.hasClosedBarrier(target)
	bestGoal := Vec2{}
	bestSteps := []Vec2(nil)
	bestSide := 0
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := 0; radius <= maxRadius; radius++ {
		if !s.playerPathBudgetAvailable() {
			break
		}
		candidates := sortApproachCellsByPlayer(nav, player.pos, ringCells(targetCell, radius))
		for _, cell := range candidates {
			if !s.playerPathBudgetAvailable() {
				break
			}
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			goal := gridToWorld(nav, cell)
			if !inRange(goal, target) {
				continue
			}
			steps, ok := s.planPlayerPathForApproach(nav, player.pos, goal, blocked)
			if !ok {
				continue
			}
			if !preferClosedBarrierSide {
				return goal, steps, true
			}
			side := s.closedBarrierApproachSideScore(player.pos, target, goal)
			if side == 0 {
				return goal, steps, true
			}
			if bestSteps == nil || side < bestSide || (side == bestSide && len(steps) < len(bestSteps)) {
				bestGoal = goal
				bestSteps = steps
				bestSide = side
			}
		}
	}
	if bestSteps != nil {
		return bestGoal, bestSteps, true
	}

	return Vec2{}, nil, false
}

func (s *Sim) findRangedApproachGoal(target *entity) (Vec2, []Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		return Vec2{}, nil, false
	}
	nav := s.activeNav()
	playerCell := worldToGrid(nav, player.pos)
	blocked := s.buildBlockedFn()
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := 0; radius <= maxRadius; radius++ {
		if !s.playerPathBudgetAvailable() {
			break
		}
		candidates := ringCells(playerCell, radius)
		for _, cell := range candidates {
			if !s.playerPathBudgetAvailable() {
				break
			}
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			origin := gridToWorld(nav, cell)
			if !s.inActionRangeFrom(origin, target) || !s.hasClearRangedShot(origin, target) {
				continue
			}
			if approachPos := s.rangedApproachStopPos(player.pos, origin); !s.hasClearRangedShot(approachPos, target) {
				continue
			}
			steps, ok := s.planPlayerPathForApproach(nav, player.pos, origin, blocked)
			if ok {
				return origin, steps, true
			}
		}
	}

	return Vec2{}, nil, false
}

func interactableApproachMoveGoal(playerPos, targetPos Vec2, interactionRadius float64) Vec2 {
	stop := 1.275
	if interactionRadius*0.85 > stop {
		stop = interactionRadius * 0.85
	}
	delta := Vec2{X: targetPos.X - playerPos.X, Y: targetPos.Y - playerPos.Y}
	dist := distance(playerPos, targetPos)
	if dist <= stop || dist <= 1e-9 {
		return targetPos
	}
	dir := Vec2{X: delta.X / dist, Y: delta.Y / dist}

	return Vec2{X: targetPos.X - dir.X*stop, Y: targetPos.Y - dir.Y*stop}
}

func (s *Sim) rangedApproachStopPos(from, goal Vec2) Vec2 {
	dx := from.X - goal.X
	dy := from.Y - goal.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1e-9 {
		return goal
	}
	stop := s.activeNav().StopDistance

	return Vec2{X: goal.X + dx/dist*stop, Y: goal.Y + dy/dist*stop}
}

func (s *Sim) findSkillCastApproachGoal(target *entity, castRange float64, requireClearShot bool) (Vec2, []Vec2, bool) {
	player := s.activeLevel().entities[s.playerID]
	if player == nil || target == nil {
		return Vec2{}, nil, false
	}
	nav := s.activeNav()
	blocked := s.buildBlockedFn()
	bestGoal := Vec2{}
	bestSteps := []Vec2(nil)
	bestDistance := -1.0
	bestStepCount := 0
	maxRadius := maxInt(nav.GridBounds.MaxX-nav.GridBounds.MinX, nav.GridBounds.MaxY-nav.GridBounds.MinY) + 1
	for radius := maxRadius; radius >= 0; radius-- {
		if !s.playerPathBudgetAvailable() {
			break
		}
		for _, cell := range ringCells(worldToGrid(nav, target.pos), radius) {
			if !s.playerPathBudgetAvailable() {
				break
			}
			if !cellInBounds(nav, cell) || blocked(cell.x, cell.y) {
				continue
			}
			origin := gridToWorld(nav, cell)
			dist := distance(origin, target.pos)
			if dist > castRange+meleeRangeEpsilon {
				continue
			}
			if requireClearShot && !s.hasClearRangedShot(origin, target) {
				continue
			}
			if requireClearShot {
				if approachPos := s.rangedApproachStopPos(player.pos, origin); !s.hasClearRangedShot(approachPos, target) {
					continue
				}
			}
			steps, ok := s.planPlayerPathForApproach(nav, player.pos, origin, blocked)
			if !ok {
				continue
			}
			if dist > bestDistance+0.000001 || (math.Abs(dist-bestDistance) <= 0.000001 && (bestSteps == nil || len(steps) < bestStepCount)) {
				bestGoal = origin
				bestSteps = steps
				bestDistance = dist
				bestStepCount = len(steps)
			}
		}
		if bestSteps != nil {
			return bestGoal, bestSteps, true
		}
	}

	return Vec2{}, nil, false
}

func sortApproachCellsByPlayer(nav NavigationRules, playerPos Vec2, cells []gridCell) []gridCell {
	if len(cells) < 2 {
		return cells
	}
	playerCell := worldToGrid(nav, playerPos)
	out := append([]gridCell(nil), cells...)
	sort.Slice(out, func(i, j int) bool {
		di := octile(playerCell, out[i])
		dj := octile(playerCell, out[j])
		if di != dj {
			return di < dj
		}
		if out[i].y != out[j].y {
			return out[i].y < out[j].y
		}

		return out[i].x < out[j].x
	})

	return out
}
