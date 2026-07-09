package game

const questTurnInService = "quest_turn_in"

func (s *Sim) turnInTownQuest(giver *entity, in Input, res *TickResult, ack bool) {
	if giver == nil || giver.kind != interactableEntity || s.serviceForInteractable(giver) != questTurnInService {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if giver.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	if trophy := s.firstQuestTurnInTrophy(); trophy != nil {
		s.openQuestStewardOffers(giver, trophy, in, res, ack)
		return
	}

	itemDefID := s.rules.MainConfig.Gameplay.QuestTurnInItemDefID
	if bagItem := s.firstResourceBagItemByDef(itemDefID); bagItem != nil {
		s.turnInQuestLeafFromResourceBag(giver, bagItem, in, res, ack)
		return
	}
	if invItem := s.firstInventoryItemByDef(itemDefID); invItem != nil {
		s.turnInQuestLeafFromInventory(giver, invItem, in, res, ack)
		return
	}

	res.reject(in.MessageID, "missing_quest_item")
}

func (s *Sim) turnInQuestLeafFromResourceBag(giver *entity, item *stashItem, in Input, res *TickResult, ack bool) {
	removedID := idStr(item.stashItemID)
	rewardGold := s.rules.MainConfig.Gameplay.QuestTurnInRewardGold
	s.removeResourceBagItemByID(item.stashItemID)
	s.gold += rewardGold
	s.progression.Gold = s.gold

	res.Changes = append(res.Changes, Change{Op: OpResourceBagItemRemove, OwnerPlayerID: s.playerID, StashItemID: removedID})
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	s.appendCharacterProgressionUpdate(res)
	res.Events = append(res.Events, Event{
		EventType:      "quest_turn_in_completed",
		EntityID:       idStr(giver.id),
		CorrelationID:  in.CorrelationID,
		Service:        questTurnInService,
		ItemInstanceID: removedID,
		StashItemID:    removedID,
		Amount:         intPtr(1),
		Price:          intPtr(rewardGold),
		TotalGold:      intPtr(s.gold),
	})
	s.grantQuestTurnInBadgeRewards(giver, in.CorrelationID, res)
	if ack {
		res.ack(in.MessageID)
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) turnInQuestLeafFromInventory(giver *entity, item *invItem, in Input, res *TickResult, ack bool) {
	removedID := idStr(item.instanceID)
	rewardGold := s.rules.MainConfig.Gameplay.QuestTurnInRewardGold
	s.removeItemByID(item.instanceID)
	s.gold += rewardGold
	s.progression.Gold = s.gold

	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	s.appendCharacterProgressionUpdate(res)
	res.Events = append(res.Events, Event{
		EventType:      "quest_turn_in_completed",
		EntityID:       idStr(giver.id),
		CorrelationID:  in.CorrelationID,
		Service:        questTurnInService,
		ItemInstanceID: removedID,
		Amount:         intPtr(1),
		Price:          intPtr(rewardGold),
		TotalGold:      intPtr(s.gold),
	})
	s.grantQuestTurnInBadgeRewards(giver, in.CorrelationID, res)
	if ack {
		res.ack(in.MessageID)
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) firstResourceBagItemByDef(itemDefID string) *stashItem {
	for _, item := range s.resourceBagItems {
		if item != nil && item.itemDefID == itemDefID {
			return item
		}
	}

	return nil
}

func (s *Sim) firstInventoryItemByDef(itemDefID string) *invItem {
	for _, item := range s.inventory {
		if item != nil && item.itemDefID == itemDefID {
			return item
		}
	}

	return nil
}
