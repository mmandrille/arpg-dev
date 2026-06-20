package game

import "math"

func mobilityRange(def SkillDef, rank int) float64 {
	if rank < 1 {
		rank = 1
	}
	return def.Mobility.RangeBase + def.Mobility.RangePerRank*float64(rank-1)
}

type playerMobilityBlockKind int

const (
	playerMobilityNotBlocked playerMobilityBlockKind = iota
	playerMobilityHardBlocked
	playerMobilityIgnoredObstacle
)

func (s *Sim) resolveDashEndpoint(start Vec2, dir Vec2, dashRange float64) Vec2 {
	return s.resolvePlayerMobilityEndpoint(start, dir, dashRange, SkillMobilityDef{})
}

func (s *Sim) resolveSkillMobilityEndpoint(start Vec2, dir Vec2, mobilityRange float64, mobility SkillMobilityDef) Vec2 {
	return s.resolvePlayerMobilityEndpoint(start, dir, mobilityRange, mobility)
}

func (s *Sim) resolvePlayerMobilityEndpoint(start Vec2, dir Vec2, mobilityRange float64, mobility SkillMobilityDef) Vec2 {
	dir = normalize(dir)
	sweep := start
	landing := start
	steps := int(math.Ceil(mobilityRange / 0.25))
	if steps < 1 {
		steps = 1
	}
	step := mobilityRange / float64(steps)
	for i := 0; i < steps; i++ {
		candidate := Vec2{X: sweep.X + dir.X*step, Y: sweep.Y + dir.Y*step}
		switch s.playerMobilityPositionBlockKind(candidate, mobility) {
		case playerMobilityHardBlocked:
			return landing
		case playerMobilityIgnoredObstacle:
			sweep = candidate
		default:
			sweep = candidate
			landing = candidate
		}
	}
	return landing
}

func (s *Sim) playerMobilityPositionBlockKind(pos Vec2, mobility SkillMobilityDef) playerMobilityBlockKind {
	for _, wall := range s.activeWalls() {
		if !obstacleBlocksMovement(wall) || !circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			continue
		}
		if skillMobilityIgnoresObstacleKind(mobility, wall.obstacleKind()) {
			return playerMobilityIgnoredObstacle
		}
		return playerMobilityHardBlocked
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.kind != interactableEntity || e.state != interactableClosed {
			continue
		}
		if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
			if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
				return playerMobilityHardBlocked
			}
		}
	}
	return playerMobilityNotBlocked
}

func skillMobilityIgnoresObstacleKind(mobility SkillMobilityDef, kind string) bool {
	for _, ignoredKind := range mobility.IgnoreObstacleKinds {
		if kind == ignoredKind {
			return true
		}
	}
	return false
}

func (s *Sim) handleMobilitySkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	if def.Mobility.Mode == "charge" {
		res.reject(in.MessageID, "use_channel_skill_intent")
		return
	}
	rng := mobilityRange(def, rank)
	dir, targetID, rejectReason := s.skillCastDirectionWithRange(def, in.CastSkill, player, rng)
	if rejectReason != "" {
		res.reject(in.MessageID, rejectReason)
		return
	}
	end := s.resolveSkillMobilityEndpoint(player.pos, dir, rng, def.Mobility)
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

func (s *Sim) handleChannelSkill(in Input, res *TickResult) {
	if in.ChannelSkill == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	skillID := in.ChannelSkill.SkillID
	def, ok := s.rules.Skills[skillID]
	if !ok {
		res.reject(in.MessageID, "unknown_skill")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	rank := s.effectiveSkillRank(skillID)
	if rank <= 0 {
		res.reject(in.MessageID, "skill_not_learned")
		return
	}
	if !s.skillClassAllowed(def) {
		res.reject(in.MessageID, "skill_class_not_allowed")
		return
	}
	if !s.skillRequirementsMet(def, rank) {
		res.reject(in.MessageID, "skill_requirements_not_met")
		return
	}
	if def.Kind != "mobility" || def.Mobility.Mode != "charge" {
		res.reject(in.MessageID, "unsupported_channel_skill")
		return
	}
	switch in.ChannelSkill.Phase {
	case "start":
		s.startChargeChannel(in, res, player, skillID, def, rank)
	case "update":
		s.updateChargeChannel(in, res, player, skillID)
	case "stop":
		s.stopChargeChannel(res, player, skillID, in.CorrelationID, "released")
		res.ack(in.MessageID)
	default:
		res.reject(in.MessageID, "invalid_phase")
	}
}

func (s *Sim) startChargeChannel(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int) {
	if s.activeLevel().activeChannel != nil {
		res.reject(in.MessageID, "channel_already_active")
		return
	}
	if in.ChannelSkill.Direction == nil || !finiteVec2(*in.ChannelSkill.Direction) {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	dir := normalize(*in.ChannelSkill.Direction)
	if dir.X == 0 && dir.Y == 0 {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	if player.mana <= 0 {
		res.reject(in.MessageID, "not_enough_mana")
		return
	}
	s.activeLevel().move = nil
	s.clearAutoNav()
	s.activeLevel().activeChannel = &activeSkillChannel{
		skillID:            skillID,
		rank:               rank,
		dir:                dir,
		correlationID:      in.CorrelationID,
		manaPer10Seconds:   def.Mobility.ChannelManaPer10Sec,
		impactedMonsterIDs: make(map[uint64]bool),
	}
	res.Events = append(res.Events, Event{
		EventType:     "skill_channel_started",
		EntityID:      idStr(player.id),
		CorrelationID: in.CorrelationID,
		SkillID:       skillID,
		Rank:          intPtr(rank),
		Position:      cloneVec2Ptr(&player.pos),
		Direction:     cloneVec2Ptr(&dir),
	})
	res.ack(in.MessageID)
}

func (s *Sim) updateChargeChannel(in Input, res *TickResult, player *entity, skillID string) {
	channel := s.activeLevel().activeChannel
	if channel == nil || channel.skillID != skillID {
		res.reject(in.MessageID, "channel_not_active")
		return
	}
	if in.ChannelSkill.Direction == nil || !finiteVec2(*in.ChannelSkill.Direction) {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	dir := normalize(*in.ChannelSkill.Direction)
	if dir.X == 0 && dir.Y == 0 {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	channel.dir = dir
	res.Events = append(res.Events, Event{
		EventType:     "skill_channel_updated",
		EntityID:      idStr(player.id),
		CorrelationID: in.CorrelationID,
		SkillID:       skillID,
		Position:      cloneVec2Ptr(&player.pos),
		Direction:     cloneVec2Ptr(&dir),
	})
	res.ack(in.MessageID)
}

func (s *Sim) stopChargeChannel(res *TickResult, player *entity, skillID string, correlationID string, reason string) {
	channel := s.activeLevel().activeChannel
	if channel == nil || channel.skillID != skillID {
		return
	}
	if correlationID == "" {
		correlationID = channel.correlationID
	}
	s.activeLevel().activeChannel = nil
	res.Events = append(res.Events, Event{
		EventType:     "skill_channel_ended",
		EntityID:      idStr(player.id),
		CorrelationID: correlationID,
		SkillID:       skillID,
		Position:      cloneVec2Ptr(&player.pos),
		Reason:        reason,
	})
}

func (s *Sim) applyActiveSkillChannel(res *TickResult) bool {
	channel := s.activeLevel().activeChannel
	if channel == nil {
		return false
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		s.activeLevel().activeChannel = nil
		return true
	}
	def, ok := s.rules.Skills[channel.skillID]
	if !ok || def.Mobility.Mode != "charge" {
		s.stopChargeChannel(res, player, channel.skillID, channel.correlationID, "invalid_skill")
		return true
	}
	if !s.spendChannelMana(player, channel, res) {
		s.stopChargeChannel(res, player, channel.skillID, channel.correlationID, "insufficient_mana")
		return true
	}
	step := s.channelMovementStep(def)
	if step <= 0 {
		s.stopChargeChannel(res, player, channel.skillID, channel.correlationID, "blocked")
		return true
	}
	start := player.pos
	end := s.resolveSkillMobilityEndpoint(start, channel.dir, step, def.Mobility)
	if end == start {
		s.stopChargeChannel(res, player, channel.skillID, channel.correlationID, "blocked")
		return true
	}
	player.pos = end
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.applyChargeLineImpactOnce(player, start, end, channel.dir, channel.skillID, def, channel.rank, channel.correlationID, channel.impactedMonsterIDs, res)
	return true
}

func (s *Sim) channelMovementStep(def SkillDef) float64 {
	if def.Mobility.SpeedMultiplier > 0 {
		return s.playerMoveSpeed() * def.Mobility.SpeedMultiplier
	}
	return def.Mobility.SpeedTilesPerSecond * tickDuration
}

func (s *Sim) spendChannelMana(player *entity, channel *activeSkillChannel, res *TickResult) bool {
	if channel.manaPer10Seconds <= 0 {
		return true
	}
	ticksPer10Seconds := int(math.Round(10.0 / tickDuration))
	if ticksPer10Seconds <= 0 {
		ticksPer10Seconds = 1
	}
	channel.manaAccumulator += channel.manaPer10Seconds
	cost := channel.manaAccumulator / ticksPer10Seconds
	if cost <= 0 {
		return true
	}
	if player.mana < cost {
		return false
	}
	channel.manaAccumulator -= cost * ticksPer10Seconds
	player.mana -= cost
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	return true
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
	s.applyChargeLineImpactOnce(player, start, end, dir, skillID, def, rank, correlationID, nil, res)
}

func (s *Sim) applyChargeLineImpactOnce(player *entity, start Vec2, end Vec2, dir Vec2, skillID string, def SkillDef, rank int, correlationID string, impacted map[uint64]bool, res *TickResult) {
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
		if impacted != nil && impacted[target.id] {
			continue
		}
		if impacted != nil {
			impacted[target.id] = true
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
