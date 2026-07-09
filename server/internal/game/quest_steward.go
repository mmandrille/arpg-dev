package game

import (
	"fmt"
	"strconv"
)

func (s *Sim) applyGeneratedStewardHunt(level *LevelState, gen generatedDungeonLevel) {
	if gen.stewardHunt == nil {
		return
	}
	monsterName := gen.stewardHunt.MonsterDefID
	if def, ok := s.rules.Monsters[gen.stewardHunt.MonsterDefID]; ok && def.Name != "" {
		monsterName = def.Name
	}
	level.stewardHunt = &stewardHuntLevelState{
		Active:          true,
		MonsterDefID:    gen.stewardHunt.MonsterDefID,
		MonsterName:     monsterName,
		TrophyItemDefID: gen.stewardHunt.TrophyItemDefID,
		TrophyLabel:     gen.stewardHunt.TrophyLabel,
		SourceDepth:     absInt(level.levelNum),
	}
}

func (s *Sim) bindStewardHuntTargetEntity(level *LevelState) {
	if level == nil || level.stewardHunt == nil || !level.stewardHunt.Active {
		return
	}
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity == nil || entity.kind != monsterEntity || !entity.stewardHuntTarget {
			continue
		}
		level.stewardHunt.TargetEntityID = entity.id
		return
	}
}

func (s *Sim) appendStewardHuntStartedEvent(level *LevelState, corr string, res *TickResult) {
	if level == nil || level.stewardHunt == nil || !level.stewardHunt.Active || level.stewardHunt.Complete {
		return
	}
	res.Events = append(res.Events, Event{
		EventType:       "steward_hunt_started",
		CorrelationID:   corr,
		MonsterDefID:    level.stewardHunt.MonsterDefID,
		MonsterName:     level.stewardHunt.MonsterName,
		TrophyItemDefID: level.stewardHunt.TrophyItemDefID,
		TrophyLabel:     level.stewardHunt.TrophyLabel,
		SourceDepth:     intPtr(level.stewardHunt.SourceDepth),
	})
}

func (s *Sim) maybeDropStewardHuntTrophy(monster *entity, corr string, res *TickResult) {
	if monster == nil || !monster.stewardHuntTarget {
		return
	}
	level := s.activeLevel()
	if level == nil || level.stewardHunt == nil || level.stewardHunt.Complete {
		return
	}
	trophyDefID := level.stewardHunt.TrophyItemDefID
	if trophyDefID == "" {
		return
	}
	level.stewardHunt.Complete = true
	drops := []LootDrop{{ItemDefID: trophyDefID}}
	s.spawnLootDrops(drops, monster.pos, s.targetInteractionRadius(monster), corr, res, goldRollContext{levelNum: level.levelNum})
}

func (s *Sim) questTurnInSourceDepth(depth int) int {
	if depth > 0 {
		return depth
	}
	if s.progression.DeepestDungeonDepth > 0 {
		return s.progression.DeepestDungeonDepth
	}

	return 1
}

func (s *Sim) isQuestTurnInItemDef(itemDefID string) bool {
	if itemDefID == "" {
		return false
	}
	if _, ok := s.rules.questStewardTrophyForItem(itemDefID); ok {
		return true
	}

	return itemDefID == s.rules.MainConfig.Gameplay.QuestTurnInItemDefID
}

func (s *Sim) countQuestTurnInItems() int {
	count := 0
	for _, item := range s.inventory {
		if item != nil && s.isQuestTurnInItemDef(item.itemDefID) {
			count++
		}
	}
	for _, item := range s.resourceBagItems {
		if item != nil && s.isQuestTurnInItemDef(item.itemDefID) {
			count++
		}
	}

	return count
}

func (s *Sim) firstQuestTurnInTrophy() *questTurnInTrophyRef {
	for _, item := range s.inventory {
		if item == nil {
			continue
		}
		if _, ok := s.rules.questStewardTrophyForItem(item.itemDefID); !ok {
			continue
		}

		return &questTurnInTrophyRef{
			instanceID:  item.instanceID,
			itemDefID:   item.itemDefID,
			sourceDepth: s.questTurnInSourceDepth(item.questSourceDepth),
			fromBag:     false,
		}
	}
	for _, item := range s.resourceBagItems {
		if item == nil {
			continue
		}
		if _, ok := s.rules.questStewardTrophyForItem(item.itemDefID); !ok {
			continue
		}

		return &questTurnInTrophyRef{
			instanceID:  item.stashItemID,
			itemDefID:   item.itemDefID,
			sourceDepth: s.questTurnInSourceDepth(0),
			fromBag:     true,
		}
	}

	return nil
}

func (s *Sim) openQuestStewardOffers(giver *entity, trophy *questTurnInTrophyRef, in Input, res *TickResult, ack bool) {
	if trophy == nil {
		res.reject(in.MessageID, "missing_quest_item")
		return
	}
	offers := s.rollQuestStewardOffers(trophy.instanceID, trophy.sourceDepth)
	s.pendingQuestStewardOffers = &questStewardOffersState{
		GiverEntityID:         giver.id,
		TrophyInstanceID:      trophy.instanceID,
		TrophyFromResourceBag: trophy.fromBag,
		SourceDepth:           trophy.sourceDepth,
		Offers:                offers,
	}
	offerViews := make([]QuestStewardOfferView, 0, len(offers))
	for _, offer := range offers {
		offerViews = append(offerViews, QuestStewardOfferView{
			OfferID:  offer.OfferID,
			FamilyID: offer.FamilyID,
			Label:    offer.Label,
		})
	}
	var trophyView ItemView
	if trophy.fromBag {
		bagItem := s.findResourceBagItem(idStr(trophy.instanceID))
		if bagItem == nil {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		trophyView = bagItem.previewItem().view()
		trophyView.QuestSourceDepth = trophy.sourceDepth
	} else {
		invItem := s.inventoryItemByID(trophy.instanceID)
		if invItem == nil {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		trophyView = s.itemView(invItem)
	}
	res.Events = append(res.Events, Event{
		EventType:          "quest_steward_offers_opened",
		EntityID:           idStr(giver.id),
		CorrelationID:      in.CorrelationID,
		Service:            questTurnInService,
		ItemInstanceID:     idStr(trophy.instanceID),
		Item:               ptrItemView(trophyView),
		SourceDepth:        intPtr(trophy.sourceDepth),
		QuestStewardOffers: offerViews,
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) handleQuestStewardPick(in Input, res *TickResult) {
	if in.QuestStewardPick == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	pending := s.pendingQuestStewardOffers
	if pending == nil {
		res.reject(in.MessageID, "no_open_offers")
		return
	}
	giverID, err := strconv.ParseUint(in.QuestStewardPick.QuestGiverEntityID, 10, 64)
	if err != nil || giverID != pending.GiverEntityID {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	giver := s.activeLevel().entities[pending.GiverEntityID]
	if giver == nil || s.serviceForInteractable(giver) != questTurnInService {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	var chosen *questStewardOffer
	for idx := range pending.Offers {
		if pending.Offers[idx].OfferID == in.QuestStewardPick.OfferID {
			chosen = &pending.Offers[idx]
			break
		}
	}
	if chosen == nil {
		res.reject(in.MessageID, "invalid_offer")
		return
	}
	var removedID = idStr(pending.TrophyInstanceID)
	if pending.TrophyFromResourceBag {
		trophy := s.findResourceBagItem(idStr(pending.TrophyInstanceID))
		if trophy == nil {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		if _, ok := s.rules.questStewardTrophyForItem(trophy.itemDefID); !ok {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		s.removeResourceBagItemByID(trophy.stashItemID)
		res.Changes = append(res.Changes, Change{Op: OpResourceBagItemRemove, OwnerPlayerID: s.playerID, StashItemID: removedID})
	} else {
		trophy := s.inventoryItemByID(pending.TrophyInstanceID)
		if trophy == nil {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		if _, ok := s.rules.questStewardTrophyForItem(trophy.itemDefID); !ok {
			res.reject(in.MessageID, "missing_quest_item")
			return
		}
		s.removeItemByID(trophy.instanceID)
		res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})
	}
	payload, ok := s.rollQuestStewardReward(chosen.FamilyID, pending.SourceDepth, pending.TrophyInstanceID)
	if !ok || itemRarityRank(payload.Rarity) < itemRarityRank(s.rules.QuestSteward.HuntQuest.MinRarity) {
		res.reject(in.MessageID, "reward_roll_failed")
		return
	}
	s.pendingQuestStewardOffers = nil
	item := s.grantInventoryItem(payload.ItemTemplateID, &payload, s.itemSlot(payload.ItemTemplateID, &payload))
	if item == nil {
		res.reject(in.MessageID, "inventory_full")
		return
	}
	itemView := s.itemView(item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(itemView)})
	res.Events = append(res.Events, Event{
		EventType:      "quest_steward_reward_granted",
		EntityID:       idStr(giver.id),
		CorrelationID:  in.CorrelationID,
		Service:        questTurnInService,
		ItemInstanceID: idStr(item.instanceID),
		Item:           ptrItemView(itemView),
		OfferID:        chosen.OfferID,
		FamilyID:       chosen.FamilyID,
		SourceDepth:    intPtr(pending.SourceDepth),
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) inventoryItemByID(instanceID uint64) *invItem {
	for _, item := range s.inventory {
		if item != nil && item.instanceID == instanceID {
			return item
		}
	}

	return nil
}

func (s *Sim) grantInventoryItem(itemDefID string, payload *ItemRollPayload, slot string) *invItem {
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		return nil
	}
	item := &invItem{
		instanceID:  s.alloc(),
		itemDefID:   itemDefID,
		rollPayload: cloneRollPayload(payload),
		slot:        slot,
	}
	s.inventory = append(s.inventory, item)

	return item
}

func stewardHuntBannerText(hunt *stewardHuntLevelState) string {
	if hunt == nil || !hunt.Active || hunt.Complete {
		return ""
	}
	return fmt.Sprintf("Hunt the %s. Retrieve its %s.", hunt.MonsterName, hunt.TrophyLabel)
}
