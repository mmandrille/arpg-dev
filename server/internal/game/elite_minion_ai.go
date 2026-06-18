package game

import "math"

func (s *Sim) eliteMinionLeader(minion *entity) *entity {
	if minion == nil || minion.kind != monsterEntity || minion.monsterPackID == "" || minion.monsterPackLeader {
		return nil
	}
	level := s.activeLevel()
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate == nil || candidate.kind != monsterEntity || candidate.hp <= 0 {
			continue
		}
		if candidate.monsterPackLeader && candidate.monsterPackID == minion.monsterPackID {
			return candidate
		}
	}
	return nil
}

func (s *Sim) eliteMinionTargetPlayer(level *LevelState, minion *entity, leader *entity) *playerState {
	if minion == nil || leader == nil || leader.aiMode != monsterAIModeChase || leader.aiTargetPlayerID == 0 {
		return nil
	}
	minion.aiTargetPlayerID = leader.aiTargetPlayerID
	minion.aiMode = monsterAIModeChase
	return s.nearestLivingPlayerForMonster(level, minion)
}

func (s *Sim) advanceEliteMinionMovement(minion *entity, leader *entity, def MonsterDef, res *TickResult) bool {
	level := s.activeLevel()
	if targetPlayer := s.eliteMinionTargetPlayer(level, minion, leader); targetPlayer != nil {
		player := level.entities[targetPlayer.PlayerID]
		if player == nil {
			return true
		}
		s.usePlayer(targetPlayer)
		s.moveMonsterTowardGoal(minion, player, def, res)
		return true
	}

	minion.aiTargetPlayerID = 0
	minion.aiMode = monsterAIModeIdle
	if distance(minion.pos, leader.pos) <= s.rules.MainConfig.Gameplay.CompanionFollowStop {
		return true
	}
	goal := s.eliteMinionFollowGoal(minion, leader)
	if distance(minion.pos, goal) <= s.activeNav().StopDistance {
		return true
	}
	s.moveMonsterToPoint(minion, def, goal, res)
	return true
}

func (s *Sim) eliteMinionFollowGoal(minion *entity, leader *entity) Vec2 {
	nav := s.activeNav()
	slotIndex := int(minion.id % 6)
	angle := (math.Pi * 2 * float64(slotIndex)) / 6.0
	goal := Vec2{
		X: leader.pos.X + math.Cos(angle)*s.rules.MainConfig.Gameplay.CompanionFollowDistance,
		Y: leader.pos.Y + math.Sin(angle)*s.rules.MainConfig.Gameplay.CompanionFollowDistance,
	}
	if s.positionInNavigationBounds(nav, goal) && !s.monsterPositionBlocked(goal, minion.id) {
		return goal
	}
	return leader.pos
}

func (s *Sim) moveMonsterTowardGoal(monster *entity, player *entity, def MonsterDef, res *TickResult) {
	goal, hasGoal := s.monsterMovementGoal(monster, player, def)
	if !hasGoal {
		return
	}
	if distance(monster.pos, goal) <= s.activeNav().StopDistance && s.monsterInAttackRange(monster, player, def) {
		return
	}
	s.moveMonsterToPoint(monster, def, goal, res)
}

func (s *Sim) moveMonsterToPoint(monster *entity, def MonsterDef, goal Vec2, res *TickResult) {
	nav := s.activeNav()
	moveSpeed := s.monsterMoveSpeed(monster, def, nav)
	if moveSpeed <= 0 {
		return
	}
	before := monster.pos
	blocked := s.buildMonsterBlockedFn(monster.id)
	steps, ok := s.planPath(nav, monster.pos, goal, blocked)
	if !ok || len(steps) == 0 {
		if distance(monster.pos, goal) > nav.CellSize+nav.StopDistance {
			return
		}
	}
	monster.pos = s.resolveMonsterMovement(monster, s.monsterMoveDelta(monster.pos, goal, steps, moveSpeed))
	if monster.pos != before {
		s.recordMonsterMoved()
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(monster))})
	}
}

func (s *Sim) eliteMinionAttackTarget(level *LevelState, minion *entity) *playerState {
	leader := s.eliteMinionLeader(minion)
	if leader == nil {
		return s.nearestLivingPlayerForMonster(level, minion)
	}
	targetPlayer := s.eliteMinionTargetPlayer(level, minion, leader)
	if targetPlayer == nil {
		minion.aiTargetPlayerID = 0
		minion.aiMode = monsterAIModeIdle
	}
	return targetPlayer
}
