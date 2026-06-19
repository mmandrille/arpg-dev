package game

func (s *Sim) handleBishopReviveAll(in Input, res *TickResult) {
	if in.BishopReviveAll == nil || in.BishopReviveAll.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(in.BishopReviveAll.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	resourceCost := s.bishopReviveResourceCost()
	if !s.canPayBishopResourceCost(resourceCost) {
		res.reject(in.MessageID, "missing_resource")
		return
	}
	s.consumeBishopResourceCost(resourceCost, res)
	res.Events = append(res.Events, Event{
		EventType:      "bishop_revive_all",
		EntityID:       idStr(bishopEntity.id),
		CorrelationID:  in.CorrelationID,
		Service:        "bishop",
		Amount:         intPtr(0),
		ResourceID:     resourceCost.ResourceID,
		ResourceAmount: intPtr(resourceCost.Count),
	})
	res.ack(in.MessageID)
}
