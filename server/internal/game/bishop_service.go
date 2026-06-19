package game

func (s *Sim) openBishopService(e *entity, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	healed, restored := s.restorePlayerResources(player, res)
	cost := s.respecCostGold()
	resourceCost := s.bishopRespecResourceCost()
	affordable := s.gold >= cost && s.canPayBishopResourceCost(resourceCost)
	res.Events = append(res.Events, Event{
		EventType:      "bishop_service_opened",
		EntityID:       idStr(e.id),
		CorrelationID:  in.CorrelationID,
		Service:        "bishop",
		Heal:           intPtr(healed),
		Mana:           intPtr(restored),
		Price:          intPtr(cost),
		Affordable:     boolPtr(affordable),
		TotalGold:      intPtr(s.gold),
		ResourceID:     resourceCost.ResourceID,
		ResourceAmount: intPtr(resourceCost.Count),
	})
	if ack {
		res.ack(in.MessageID)
	}
}
