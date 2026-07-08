package game

import "math"

func (s *Sim) monsterAlongDirectionBeyondRange(player *entity, dir Vec2, castRange float64) *entity {
	if player == nil {
		return nil
	}
	dir = normalize(dir)
	if dir.X == 0 && dir.Y == 0 {
		return nil
	}
	level := s.activeLevel()
	if level == nil {
		return nil
	}

	var nearest *entity
	nearestAlong := math.MaxFloat64
	aimTolerance := playerRadius + meleeRangeEpsilon
	for _, candidate := range level.entities {
		if candidate == nil || candidate.kind != monsterEntity || candidate.hp <= 0 {
			continue
		}
		offset := Vec2{X: candidate.pos.X - player.pos.X, Y: candidate.pos.Y - player.pos.Y}
		along := offset.X*dir.X + offset.Y*dir.Y
		if along <= meleeRangeEpsilon {
			continue
		}
		perp := math.Abs(offset.X*dir.Y - offset.Y*dir.X)
		if perp > aimTolerance {
			continue
		}
		if along < nearestAlong {
			nearest = candidate
			nearestAlong = along
		}
	}
	if nearest == nil {
		return nil
	}
	if distance(player.pos, nearest.pos) <= castRange+meleeRangeEpsilon {
		return nil
	}

	return nearest
}

func (s *Sim) beginDirectionalSkillAutoNav(in Input, res *TickResult, target *entity, castRange float64, requireClearShot bool) {
	if in.CastSkill == nil || in.CastSkill.Direction == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	goal, steps, ok := s.findSkillCastApproachGoal(target, castRange, requireClearShot)
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().PlayerMaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = s.newAutoNavState(
		steps, goal, nil,
		&CastSkillIntent{
			SkillID:   in.CastSkill.SkillID,
			Direction: cloneVec2Ptr(in.CastSkill.Direction),
		},
		in.MessageID, in.CorrelationID,
	)
	res.ack(in.MessageID)
}

func (s *Sim) maybeBeginDirectionalSkillAutoNav(in Input, res *TickResult, player *entity, def SkillDef, dir Vec2, castRange float64, requireClearShot bool) bool {
	if in.CastSkill == nil || in.CastSkill.TargetID != "" || def.Targeting != "direction" {
		return false
	}
	target := s.monsterAlongDirectionBeyondRange(player, dir, castRange)
	if target == nil {
		return false
	}
	s.beginDirectionalSkillAutoNav(in, res, target, castRange, requireClearShot)
	return true
}
