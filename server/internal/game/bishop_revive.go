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
	res.Events = append(res.Events, Event{
		EventType:     "bishop_revive_all",
		EntityID:      idStr(bishopEntity.id),
		CorrelationID: in.CorrelationID,
		Service:       "bishop",
		Amount:        intPtr(0),
	})
	res.ack(in.MessageID)
}
