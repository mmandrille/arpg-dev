package game

func (s *Sim) autoPickUpCurrencyLoot(resultFor func(level int, actor uint64) *TickResult) {
	for _, levelNum := range s.sortedLevelNums() {
		level := s.levels[levelNum]
		if level == nil {
			continue
		}
		for _, entityID := range sortedEntityIDs(level.entities) {
			loot := level.entities[entityID]
			winnerID := s.autoPickupWinner(levelNum, loot)
			if winnerID == 0 {
				continue
			}
			res := resultFor(levelNum, 0)
			if isAutoPickableGold(loot) {
				s.pickUpGoldForPlayer(loot, winnerID, "", "", res)
			} else if s.isAutoPickableWalletResource(loot) {
				s.pickUpWalletResourceForPlayer(loot, winnerID, "", "", res)
			}
		}
	}
}

func (s *Sim) autoPickupWinner(levelNum int, loot *entity) uint64 {
	level := s.levels[levelNum]
	if level == nil || (!isAutoPickableGold(loot) && !s.isAutoPickableWalletResource(loot)) {
		return 0
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != levelNum {
			continue
		}
		player := level.entities[playerID]
		if player == nil || player.hp <= 0 {
			continue
		}
		s.usePlayer(ps)
		if s.inLootPickupRangeFrom(player.pos, loot) {
			return playerID
		}
	}
	return 0
}

func (s *Sim) inLootPickupRangeFrom(pos Vec2, target *entity) bool {
	return meleeInRange(distance(pos, target.pos), s.playerMeleeReach(), s.targetInteractionRadius(target))
}

func isAutoPickableGold(e *entity) bool {
	return e != nil && e.kind == lootEntity && e.itemDefID == goldItemDefID && e.goldAmount > 0
}

func (s *Sim) pickUpGoldForPlayer(e *entity, playerID uint64, correlationID, ackMessageID string, res *TickResult) bool {
	if !isAutoPickableGold(e) {
		return false
	}
	ps := s.players[playerID]
	if ps == nil {
		return false
	}
	s.usePlayer(ps)
	level := s.activeLevel()
	if level == nil || level.entities[e.id] != e {
		return false
	}
	delete(level.entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})
	amount := e.goldAmount
	s.gold += amount
	s.progression.Gold = s.gold
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, OwnerPlayerID: playerID, Gold: intPtr(s.gold)})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, OwnerPlayerID: playerID, Progression: &view})
	res.Events = append(res.Events, Event{
		EventType:     "gold_picked_up",
		EntityID:      idStr(playerID),
		CorrelationID: correlationID,
		Amount:        intPtr(amount),
		TotalGold:     intPtr(s.gold),
	})
	if ackMessageID != "" {
		res.ack(ackMessageID)
	}
	s.savePlayer(ps)
	return true
}
