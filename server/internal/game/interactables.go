package game

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

func (s *Sim) eliteObjectiveChestLocked(e *entity) bool {
	level := s.activeLevel()
	if e == nil || level == nil || !level.eliteObjectiveChestIDs[e.id] {
		return false
	}
	for _, entityID := range sortedEntityIDs(level.entities) {
		candidate := level.entities[entityID]
		if candidate.kind == monsterEntity && candidate.monsterPackLeader && candidate.hp <= 0 {
			return false
		}
	}
	return true
}
