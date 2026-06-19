package game

func (s *Sim) handleBishopRespec(in Input, res *TickResult) {
	if in.BishopRespec == nil || in.BishopRespec.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(in.BishopRespec.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	goldCost := s.respecCostGold()
	if s.gold < goldCost {
		res.reject(in.MessageID, "not_enough_gold")
		return
	}
	resourceCost := s.bishopRespecResourceCost()
	if !s.canPayBishopResourceCost(resourceCost) {
		res.reject(in.MessageID, "missing_resource")
		return
	}
	s.gold -= goldCost
	s.progression.Gold = s.gold
	s.consumeBishopResourceCost(resourceCost, res)
	s.resetCharacterBuildForRespec()
	player.maxHP = s.currentMaxHP()
	player.maxMana = s.currentMaxMana()
	healed, restored := s.restorePlayerResources(player, res)
	s.skillCooldowns = make(map[string]skillCooldownState)

	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	s.appendProgressionAndSkillUpdates(res)
	s.appendSkillCooldownUpdate(res)
	res.Events = append(res.Events, Event{
		EventType:          "bishop_respec",
		EntityID:           idStr(bishopEntity.id),
		CorrelationID:      in.CorrelationID,
		Service:            "bishop",
		Heal:               intPtr(healed),
		Mana:               intPtr(restored),
		Price:              intPtr(goldCost),
		TotalGold:          intPtr(s.gold),
		ResourceID:         resourceCost.ResourceID,
		ResourceAmount:     intPtr(resourceCost.Count),
		UnspentStatPoints:  intPtr(s.progression.UnspentStatPoints),
		UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
	})
	res.ack(in.MessageID)
}
