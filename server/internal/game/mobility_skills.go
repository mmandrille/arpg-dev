package game

import "math"

func mobilityRange(def SkillDef, rank int) float64 {
	if rank < 1 {
		rank = 1
	}
	return def.Mobility.RangeBase + def.Mobility.RangePerRank*float64(rank-1)
}

func (s *Sim) handleMobilitySkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	rng := mobilityRange(def, rank)
	dir, targetID, rejectReason := s.skillCastDirectionWithRange(def, in.CastSkill, player, rng)
	if rejectReason != "" {
		res.reject(in.MessageID, rejectReason)
		return
	}
	end := s.resolveDashEndpoint(player.pos, dir, rng)
	if end == player.pos {
		res.reject(in.MessageID, "blocked")
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	start := player.pos
	player.pos = end
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendMobilitySkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, start, dir, rng, def.Mobility.Visual)
	impactPos := player.pos
	if def.Mobility.Mode == "disengage" {
		impactPos = start
	}
	if def.Mobility.Mode == "charge" {
		s.applyChargeLineImpact(player, start, end, dir, skillID, def, rank, in.CorrelationID, res)
	} else {
		s.applyMobilityImpact(player, impactPos, skillID, def, rank, in.CorrelationID, res)
	}
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) appendMobilitySkillCastEvent(res *TickResult, player *entity, skillID string, rank int, manaCost int, correlationID string, targetID uint64, start Vec2, dir Vec2, rng float64, visual string) {
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, correlationID, targetID, visual)
	if len(res.Events) == 0 {
		return
	}
	event := &res.Events[len(res.Events)-1]
	event.Position = cloneVec2Ptr(&start)
	event.Direction = cloneVec2Ptr(&dir)
	event.Range = floatPtr(rng)
}

func (s *Sim) applyMobilityImpact(player *entity, impactPos Vec2, skillID string, def SkillDef, rank int, correlationID string, res *TickResult) {
	impactRadius := def.Mobility.ImpactRadius
	if impactRadius <= 0 {
		return
	}
	damageRange := mobilityDamageRange(s.resolvePlayerAttackDamage(), def, rank)
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || distance(impactPos, target.pos) > impactRadius {
			continue
		}
		if damageRange.Max > 0 {
			beforeEvents := len(res.Events)
			s.damageMonsterByPlayerSkillTypedWithID(target, player.id, skillID, correlationID, res, damageRange, s.skillDamageType(def))
			for i := beforeEvents; i < len(res.Events); i++ {
				if res.Events[i].EventType == "monster_damaged" && res.Events[i].TargetEntityID == idStr(target.id) {
					res.Events[i].SkillID = skillID
				}
			}
		}
		if target.hp > 0 && def.Mobility.StunDurationTicks > 0 {
			s.applyMonsterRoot(target, player.id, skillID, SkillRootDef{EffectID: def.Mobility.StunEffectID, DurationTicks: def.Mobility.StunDurationTicks}, correlationID, res)
		}
		if target.hp > 0 && def.Mobility.RootDurationTicks > 0 {
			s.applyMonsterRoot(target, player.id, skillID, SkillRootDef{EffectID: def.Mobility.RootEffectID, DurationTicks: def.Mobility.RootDurationTicks}, correlationID, res)
		}
	}
}

func (s *Sim) applyChargeLineImpact(player *entity, start Vec2, end Vec2, dir Vec2, skillID string, def SkillDef, rank int, correlationID string, res *TickResult) {
	impactRadius := def.Mobility.ImpactRadius
	if impactRadius <= 0 {
		return
	}
	damageRange := mobilityDamageRange(s.resolvePlayerAttackDamage(), def, rank)
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || distancePointToSegment(target.pos, start, end) > impactRadius {
			continue
		}
		if damageRange.Max > 0 {
			beforeEvents := len(res.Events)
			s.damageMonsterByPlayerSkillTypedWithID(target, player.id, skillID, correlationID, res, damageRange, s.skillDamageType(def))
			for i := beforeEvents; i < len(res.Events); i++ {
				if res.Events[i].EventType == "monster_damaged" && res.Events[i].TargetEntityID == idStr(target.id) {
					res.Events[i].SkillID = skillID
				}
			}
		}
		if target.hp > 0 && def.Mobility.StunDurationTicks > 0 {
			s.applyMonsterRoot(target, player.id, skillID, SkillRootDef{EffectID: def.Mobility.StunEffectID, DurationTicks: def.Mobility.StunDurationTicks}, correlationID, res)
		}
		if target.hp > 0 {
			s.pushMobilityTarget(player, target, dir, skillID, def, correlationID, res)
		}
	}
}

func (s *Sim) pushMobilityTarget(player *entity, target *entity, dir Vec2, skillID string, def SkillDef, correlationID string, res *TickResult) {
	push := s.rollFloatRange(def.Mobility.PushMin, def.Mobility.PushMax)
	if push <= 0 {
		return
	}
	away := normalize(dir)
	if away.X == 0 && away.Y == 0 {
		away = normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
	}
	if away.X == 0 && away.Y == 0 {
		away = Vec2{X: 1}
	}
	before := target.pos
	target.pos = s.resolveMonsterMovement(target, Vec2{X: away.X * push, Y: away.Y * push})
	if target.pos == before {
		return
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "monster_pushed",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(player.id),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Amount:         intPtr(int(math.Round(push))),
	})
}

func distancePointToSegment(point Vec2, start Vec2, end Vec2) float64 {
	segment := Vec2{X: end.X - start.X, Y: end.Y - start.Y}
	lengthSq := segment.X*segment.X + segment.Y*segment.Y
	if lengthSq <= 0 {
		return distance(point, start)
	}
	t := ((point.X-start.X)*segment.X + (point.Y-start.Y)*segment.Y) / lengthSq
	t = math.Max(0, math.Min(1, t))
	closest := Vec2{X: start.X + segment.X*t, Y: start.Y + segment.Y*t}
	return distance(point, closest)
}

func mobilityDamageRange(base DamageRange, def SkillDef, rank int) DamageRange {
	percent := def.Mobility.DamagePercentBase + def.Mobility.DamagePercentPerRank*max(0, rank-1)
	if percent <= 0 {
		return DamageRange{}
	}
	return DamageRange{
		Min: max(1, int(math.Round(float64(base.Min)*float64(percent)/100.0))),
		Max: max(1, int(math.Round(float64(base.Max)*float64(percent)/100.0))),
	}
}
