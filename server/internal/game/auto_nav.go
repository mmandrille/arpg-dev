package game

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
