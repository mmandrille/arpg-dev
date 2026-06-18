package game

import "math"

func (s *Sim) activateInteractable(e *entity, in Input, res *TickResult, ack bool) {
	if e.interactableDefID == teleporterDefID {
		s.activateTeleporter(e, in, res, ack)
		return
	}
	if shopID := s.shopIDForInteractable(e); shopID != "" {
		s.openShop(e, shopID, in, res, ack)
		return
	}
	if stashID := s.stashIDForInteractable(e); stashID != "" {
		s.openStash(e, stashID, in, res, ack)
		return
	}
	if service := s.serviceForInteractable(e); service == "bishop" {
		s.openBishopService(e, in, res, ack)
		return
	}
	if service := s.serviceForInteractable(e); service == "market" {
		s.openMarketService(e, in, res, ack)
		return
	}
	if service := s.serviceForInteractable(e); service == "mercenary" {
		s.hireMercenaryFromBoard(e, in, res, ack)
		return
	}
	if service := s.serviceForInteractable(e); service == "blacksmith" {
		s.openBlacksmithService(e, in, res, ack)
		return
	}
	if service := s.serviceForInteractable(e); service == uniqueTestChestService {
		if !s.gameplayDebug {
			res.reject(in.MessageID, "debug_disabled")
			return
		}
		s.openUniqueTestChest(e, in, res, ack)
		return
	}
	if e.state == interactableOpen && s.hasClosedBarrier(e) {
		e.state = interactableClosed
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
		res.Events = append(res.Events, Event{EventType: "interactable_state_changed", EntityID: idStr(e.id), State: interactableClosed, CorrelationID: in.CorrelationID})
		if ack {
			res.ack(in.MessageID)
		}
		return
	}
	if e.state != interactableClosed {
		res.reject(in.MessageID, "already_open")
		return
	}
	if s.eliteObjectiveChestLocked(e) {
		res.reject(in.MessageID, "elite_objective_incomplete")
		return
	}
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
	res.Events = append(res.Events, Event{EventType: "interactable_activated", EntityID: idStr(e.id), CorrelationID: in.CorrelationID})
	if e.interactableDefID == treasureChestDefID && e.lootTable != "" {
		drops := s.rules.LootDrops(e.lootTable, s.rng)
		drops = append(drops, LootDrop{ItemDefID: goldItemDefID})
		s.spawnLootDrops(drops, e.pos, s.targetInteractionRadius(e), in.CorrelationID, res, goldRollContext{levelNum: s.activeLevel().levelNum})
	}
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) hasClosedBarrier(e *entity) bool {
	if e == nil || e.kind != interactableEntity {
		return false
	}
	def, ok := s.rules.Interactables[e.interactableDefID]
	return ok && def.BarrierWhenClosed != nil
}

func (s *Sim) closedBarrierApproachSideScore(playerPos Vec2, target *entity, goal Vec2) int {
	if target == nil || target.state != interactableClosed || !s.hasClosedBarrier(target) {
		return 0
	}
	def := s.rules.Interactables[target.interactableDefID]
	barrier := def.BarrierWhenClosed.Size
	playerDelta := playerPos.Y - target.pos.Y
	goalDelta := goal.Y - target.pos.Y
	if barrier.Y > barrier.X {
		playerDelta = playerPos.X - target.pos.X
		goalDelta = goal.X - target.pos.X
	}
	if math.Abs(playerDelta) <= 0.000001 || math.Abs(goalDelta) <= 0.000001 {
		return 1
	}
	if (playerDelta < 0 && goalDelta < 0) || (playerDelta > 0 && goalDelta > 0) {
		return 0
	}
	return 2
}

func (s *Sim) eliteObjectiveChestLocked(e *entity) bool {
	level := s.activeLevel()
	if e == nil || level == nil || !level.eliteObjectiveChestIDs[e.id] {
		return false
	}
	foundLeader := false
	for _, entityID := range sortedEntityIDs(level.entities) {
		candidate := level.entities[entityID]
		if candidate.kind == monsterEntity && candidate.monsterPackLeader {
			foundLeader = true
			if candidate.hp > 0 {
				return true
			}
		}
	}
	return !foundLeader
}
