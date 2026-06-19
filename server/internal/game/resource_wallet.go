package game

func (s *Sim) isWalletResourceItem(itemDefID string) bool {
	return itemDefID != "" && (itemDefID == s.rules.MainConfig.Gameplay.ItemUpgradeResourceID || s.rules.isBadgeRewardResourceItem(itemDefID))
}

func (s *Sim) isAutoPickableWalletResource(e *entity) bool {
	return e != nil && e.kind == lootEntity && s.isWalletResourceItem(e.itemDefID)
}

func (s *Sim) pickUpWalletResource(e *entity, in Input, res *TickResult, ack bool) {
	ackMessageID := ""
	if ack {
		ackMessageID = in.MessageID
	}
	s.pickUpWalletResourceForPlayer(e, s.playerID, in.CorrelationID, ackMessageID, res)
}

func (s *Sim) pickUpWalletResourceForPlayer(e *entity, playerID uint64, correlationID, ackMessageID string, res *TickResult) bool {
	if !s.isAutoPickableWalletResource(e) {
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
	resourceID := e.itemDefID
	if s.resourceWallet == nil {
		s.resourceWallet = make(map[string]int)
	}
	s.resourceWallet[resourceID]++
	delete(level.entities, e.id)
	res.Changes = append(res.Changes,
		Change{Op: OpEntityRemove, EntityID: idStr(e.id)},
		Change{Op: OpResourceWalletUpdate, OwnerPlayerID: playerID, ResourceID: resourceID, ResourceAmount: intPtr(s.resourceWallet[resourceID])},
	)
	res.Events = append(res.Events, Event{
		EventType:      "resource_picked_up",
		EntityID:       idStr(playerID),
		CorrelationID:  correlationID,
		ItemInstanceID: idStr(e.id),
		ResourceID:     resourceID,
		Amount:         intPtr(1),
	})
	if ackMessageID != "" {
		res.ack(ackMessageID)
	}
	s.savePlayer(ps)
	return true
}

func (s *Sim) ResourceWalletView() []ResourceAmountView {
	keys := sortedStringKeys(s.resourceWallet)
	out := make([]ResourceAmountView, 0, len(keys))
	for _, resourceID := range keys {
		amount := s.resourceWallet[resourceID]
		if amount <= 0 {
			continue
		}
		out = append(out, ResourceAmountView{ResourceID: resourceID, Amount: amount})
	}
	return out
}
