package game

import "math"

func (s *Sim) continueAutoNavToGoal(player *entity, res *TickResult) bool {
	nav := s.activeLevel().autoNav
	if nav == nil || !nav.hasGoal {
		return false
	}
	if distance(player.pos, nav.goal) <= s.activeNav().StopDistance {
		return false
	}
	delta := s.autoNavGoalDelta(player.pos, nav.goal)
	if delta == (Vec2{}) {
		return false
	}
	before := player.pos
	player.pos = s.resolveMovement(player.pos, delta)
	if player.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}
	if player.pos != before {
		// If the player moved, check if progress toward the goal is negligible.
		// When the Y-only fallback slides the player perpendicular to the goal
		// (Zeno convergence), the dot product of movement direction and goal
		// direction approaches 0. Treat this as stuck to trigger a re-plan.
		moved := Vec2{X: player.pos.X - before.X, Y: player.pos.Y - before.Y}
		movedDist := math.Sqrt(moved.X*moved.X + moved.Y*moved.Y)
		toGoal := Vec2{X: nav.goal.X - before.X, Y: nav.goal.Y - before.Y}
		goalDist := math.Sqrt(toGoal.X*toGoal.X + toGoal.Y*toGoal.Y)
		if movedDist > 1e-9 && goalDist > 1e-9 {
			dot := (moved.X*toGoal.X + moved.Y*toGoal.Y) / (movedDist * goalDist)
			if dot < 0.1 { // nearly perpendicular to goal direction — sliding, not approaching
				return false
			}
		}
	}
	return player.pos != before && distance(player.pos, nav.goal) > s.activeNav().StopDistance
}

func (s *Sim) autoNavGoalDelta(pos Vec2, goal Vec2) Vec2 {
	speed := s.playerMoveSpeed()
	toGoal := Vec2{X: goal.X - pos.X, Y: goal.Y - pos.Y}
	dist := distance(pos, goal)
	if dist <= 1e-9 || speed <= 0 {
		return Vec2{}
	}
	if dist <= speed+1e-9 {
		return toGoal
	}
	dir := normalize(toGoal)
	return Vec2{X: dir.X * speed, Y: dir.Y * speed}
}
