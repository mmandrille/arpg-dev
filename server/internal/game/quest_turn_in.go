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

	if trophy := s.firstQuestStewardTrophyItem(); trophy != nil {
		s.openQuestStewardOffers(giver, trophy, in, res, ack)
		return
	}

	itemDefID := s.rules.MainConfig.Gameplay.QuestTurnInItemDefID
	item := s.firstResourceBagItemByDef(itemDefID)
	if item == nil {
		res.reject(in.MessageID, "missing_quest_item")
		return
	}

	removedID := idStr(item.stashItemID)
	rewardGold := s.rules.MainConfig.Gameplay.QuestTurnInRewardGold
	s.removeResourceBagItemByID(item.stashItemID)
	s.gold += rewardGold
	s.progression.Gold = s.gold

	res.Changes = append(res.Changes, Change{Op: OpResourceBagItemRemove, StashItemID: removedID})
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

func (s *Sim) firstResourceBagItemByDef(itemDefID string) *stashItem {
	for _, item := range s.resourceBagItems {
		if item != nil && item.itemDefID == itemDefID {
			return item
		}
	}

	return nil
}
