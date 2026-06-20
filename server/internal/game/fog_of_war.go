package game

type fogVisibilityContext struct {
	player *entity
	level  *LevelState
	radius float64
}

func (s *Sim) visibleEntityViewsForPlayer(ps *playerState) []EntityView {
	level := s.activeLevel()
	if level == nil {
		return []EntityView{}
	}
	ctx, ok := s.fogVisibilityContext(ps.PlayerID, level.levelNum)
	if !ok {
		return s.entityViewsForLevel(level)
	}
	s.resetFogVisibility(ps, level.levelNum)
	views := make([]EntityView, 0, len(level.entities))
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if s.livingMonsterHiddenByFog(ctx, e) {
			continue
		}
		if e != nil && e.kind == monsterEntity && e.hp > 0 {
			ps.VisibleMonsterIDs[e.id] = true
		}
		views = append(views, s.entityView(e))
	}
	return views
}

func (s *Sim) SetFogOfWarEnabled(enabled bool) {
	s.fogOfWarEnabled = enabled
}

func (s *Sim) entityViewsForLevel(level *LevelState) []EntityView {
	views := make([]EntityView, 0, len(level.entities))
	for _, id := range sortedEntityIDs(level.entities) {
		views = append(views, s.entityView(level.entities[id]))
	}
	return views
}

func (s *Sim) FilterChangesForPlayer(playerID uint64, levelNum int, changes []Change) []Change {
	ps := s.players[playerID]
	ctx, ok := s.fogVisibilityContext(playerID, levelNum)
	if !ok || ps == nil {
		return changes
	}
	s.resetFogVisibility(ps, levelNum)
	seenBefore := cloneBoolMap(ps.VisibleMonsterIDs)
	touched := make(map[uint64]bool)
	out := make([]Change, 0, len(changes))
	for _, change := range changes {
		filtered, keep, id := s.filterChangeForFog(ctx, ps, change)
		if id != 0 {
			touched[id] = true
		}
		if keep {
			out = append(out, filtered)
		}
	}
	out = s.appendFogVisibilityTransitions(ctx, ps, seenBefore, touched, out)
	return out
}

func (s *Sim) FilterEventsForPlayer(playerID uint64, levelNum int, events []Event) []Event {
	ctx, ok := s.fogVisibilityContext(playerID, levelNum)
	if !ok {
		return events
	}
	out := make([]Event, 0, len(events))
	for _, event := range events {
		if s.eventReferencesHiddenMonster(ctx, event) {
			continue
		}
		out = append(out, event)
	}
	return out
}

func (s *Sim) FilterTickResultForPlayer(playerID uint64, res TickResult) TickResult {
	res.Changes = s.FilterChangesForPlayer(playerID, res.Level, res.Changes)
	res.Events = s.FilterEventsForPlayer(playerID, res.Level, res.Events)
	return res
}

func (s *Sim) filterChangeForFog(ctx fogVisibilityContext, ps *playerState, change Change) (Change, bool, uint64) {
	switch change.Op {
	case OpEntitySpawn, OpEntityUpdate:
		if change.Entity == nil || change.Entity.Type != monsterEntity {
			return change, true, 0
		}
		id, ok := ParseEntityID(change.Entity.ID)
		if !ok {
			return change, true, 0
		}
		monster := ctx.level.entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			return change, true, id
		}
		if !s.monsterVisibleByFog(ctx, monster) {
			if ps.VisibleMonsterIDs[id] {
				delete(ps.VisibleMonsterIDs, id)
				return Change{Op: OpEntityRemove, EntityID: idStr(id)}, true, id
			}
			return change, false, id
		}
		if !ps.VisibleMonsterIDs[id] && change.Op == OpEntityUpdate {
			change.Op = OpEntitySpawn
		}
		ps.VisibleMonsterIDs[id] = true
		return change, true, id
	case OpEntityRemove:
		id, ok := ParseEntityID(change.EntityID)
		if !ok {
			return change, true, 0
		}
		if ps.VisibleMonsterIDs[id] {
			delete(ps.VisibleMonsterIDs, id)
			return change, true, id
		}
		return change, true, id
	default:
		return change, true, 0
	}
}

func (s *Sim) appendFogVisibilityTransitions(ctx fogVisibilityContext, ps *playerState, seenBefore map[uint64]bool, touched map[uint64]bool, out []Change) []Change {
	for _, id := range sortedEntityIDs(ctx.level.entities) {
		if touched[id] {
			continue
		}
		monster := ctx.level.entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		visible := s.monsterVisibleByFog(ctx, monster)
		wasVisible := seenBefore[id]
		switch {
		case visible && !wasVisible:
			ps.VisibleMonsterIDs[id] = true
			out = append(out, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(monster))})
		case !visible && wasVisible:
			delete(ps.VisibleMonsterIDs, id)
			out = append(out, Change{Op: OpEntityRemove, EntityID: idStr(id)})
		}
	}
	return out
}

func (s *Sim) fogVisibilityContext(playerID uint64, levelNum int) (fogVisibilityContext, bool) {
	if !s.fogOfWarEnabled {
		return fogVisibilityContext{}, false
	}
	ps := s.players[playerID]
	if ps == nil || ps.CurrentLevel != levelNum {
		return fogVisibilityContext{}, false
	}
	level := s.levels[levelNum]
	if level == nil {
		return fogVisibilityContext{}, false
	}
	player := level.entities[playerID]
	if player == nil || player.kind != playerEntity {
		return fogVisibilityContext{}, false
	}
	return fogVisibilityContext{
		player: player,
		level:  level,
		radius: s.lightRadiusForPlayer(ps),
	}, true
}

func (s *Sim) lightRadiusForPlayer(ps *playerState) float64 {
	if ps == nil {
		return 0
	}
	previous := s.players[s.playerID]
	s.usePlayer(ps)
	stats, _ := s.playerEffectiveCombatStats()
	if previous != nil {
		s.usePlayer(previous)
	}
	return stats.LightRadius
}

func (s *Sim) resetFogVisibility(ps *playerState, levelNum int) {
	if ps.VisibleMonsterIDs == nil || ps.FogVisibleLevel != levelNum {
		ps.VisibleMonsterIDs = make(map[uint64]bool)
		ps.FogVisibleLevel = levelNum
	}
}

func (s *Sim) livingMonsterHiddenByFog(ctx fogVisibilityContext, e *entity) bool {
	return e != nil && e.kind == monsterEntity && e.hp > 0 && !s.monsterVisibleByFog(ctx, e)
}

func (s *Sim) monsterVisibleByFog(ctx fogVisibilityContext, monster *entity) bool {
	if ctx.player == nil || monster == nil {
		return false
	}
	if distance(ctx.player.pos, monster.pos) > ctx.radius+meleeRangeEpsilon {
		return false
	}
	return !s.blocksFogLineOfSight(ctx, monster)
}

func (s *Sim) blocksFogLineOfSight(ctx fogVisibilityContext, monster *entity) bool {
	if ctx.level == nil {
		return false
	}
	for _, wall := range ctx.level.walls {
		if !obstacleBlocksLineOfSight(wall) {
			continue
		}
		if _, ok := segmentIntersectsInflatedAABB(ctx.player.pos, monster.pos, wall.pos, wall.size, 0); ok {
			return true
		}
	}
	for _, id := range sortedEntityIDs(ctx.level.entities) {
		e := ctx.level.entities[id]
		if e == nil || e.id == ctx.player.id || e.id == monster.id {
			continue
		}
		if e.kind != interactableEntity || e.state != interactableClosed {
			continue
		}
		def, ok := s.rules.Interactables[e.interactableDefID]
		if !ok || def.BarrierWhenClosed == nil {
			continue
		}
		if _, ok := segmentIntersectsInflatedAABB(ctx.player.pos, monster.pos, e.pos, def.BarrierWhenClosed.Size, 0); ok {
			return true
		}
	}
	return false
}

func (s *Sim) eventReferencesHiddenMonster(ctx fogVisibilityContext, event Event) bool {
	for _, rawID := range []string{event.EntityID, event.SourceEntityID, event.TargetEntityID} {
		if rawID == "" {
			continue
		}
		id, ok := ParseEntityID(rawID)
		if !ok {
			continue
		}
		e := ctx.level.entities[id]
		if s.livingMonsterHiddenByFog(ctx, e) {
			return true
		}
	}
	return false
}

func cloneBoolMap(in map[uint64]bool) map[uint64]bool {
	out := make(map[uint64]bool, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
