package game

func (s *Sim) isWalletResourceItem(itemDefID string) bool {
	return itemDefID != "" && itemDefID == s.rules.MainConfig.Gameplay.ItemUpgradeResourceID
}

func (s *Sim) pickUpWalletResource(e *entity, in Input, res *TickResult, ack bool) {
	resourceID := e.itemDefID
	if s.resourceWallet == nil {
		s.resourceWallet = make(map[string]int)
	}
	s.resourceWallet[resourceID]++
	delete(s.activeLevel().entities, e.id)
	res.Changes = append(res.Changes,
		Change{Op: OpEntityRemove, EntityID: idStr(e.id)},
		Change{Op: OpResourceWalletUpdate, ResourceID: resourceID, ResourceAmount: intPtr(s.resourceWallet[resourceID])},
	)
	res.Events = append(res.Events, Event{
		EventType:      "resource_picked_up",
		EntityID:       idStr(s.playerID),
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: idStr(e.id),
		ResourceID:     resourceID,
		Amount:         intPtr(1),
	})
	if ack {
		res.ack(in.MessageID)
	}
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
