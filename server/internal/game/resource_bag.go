package game

import (
	"strconv"
)

func (s *Sim) isResourceBagItem(itemDefID string) bool {
	if itemDefID == "" || itemDefID == goldItemDefID {
		return false
	}
	if s.isWalletResourceItem(itemDefID) {
		return false
	}
	def, ok := s.rules.Items[itemDefID]
	if !ok {
		return false
	}

	return def.Category == "currency" || def.Category == "quest"
}

func (s *Sim) findResourceBagItem(bagItemID string) *stashItem {
	id, err := strconv.ParseUint(bagItemID, 10, 64)
	if err != nil {
		return nil
	}
	for _, item := range s.resourceBagItems {
		if item != nil && item.stashItemID == id {
			return item
		}
	}

	return nil
}

func (s *Sim) removeResourceBagItemByID(id uint64) {
	for i, item := range s.resourceBagItems {
		if item != nil && item.stashItemID == id {
			s.resourceBagItems = append(s.resourceBagItems[:i], s.resourceBagItems[i+1:]...)

			return
		}
	}
}

func (s *Sim) resourceBagItemViews() []StashItemView {
	out := make([]StashItemView, 0, len(s.resourceBagItems))
	for _, item := range s.resourceBagItems {
		if item == nil {
			continue
		}
		out = append(out, s.stashItemView(item))
	}

	return out
}

func (s *Sim) grantResourceBagItem(playerID uint64, itemDefID string, payload *ItemRollPayload, questSourceDepth int, res *TickResult) *stashItem {
	bagItemID := s.alloc()
	stored := &stashItem{
		stashItemID: bagItemID,
		itemDefID:   itemDefID,
		rollPayload: cloneRollPayload(payload),
	}
	s.resourceBagItems = append(s.resourceBagItems, stored)
	res.Changes = append(res.Changes, Change{
		Op:            OpResourceBagItemAdd,
		OwnerPlayerID: playerID,
		StashItem:     ptrStashItemView(s.stashItemView(stored)),
	})

	return stored
}

func (s *Sim) pickUpResourceBagItem(e *entity, in Input, res *TickResult, ack bool) {
	ackMessageID := ""
	if ack {
		ackMessageID = in.MessageID
	}
	s.pickUpResourceBagItemForPlayer(e, s.playerID, in.CorrelationID, ackMessageID, res)
}

func (s *Sim) pickUpResourceBagItemForPlayer(e *entity, playerID uint64, correlationID, ackMessageID string, res *TickResult) bool {
	if e == nil || e.kind != lootEntity || !s.isResourceBagItem(e.itemDefID) {
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

	questDepth := 0
	if _, ok := s.rules.questStewardTrophyForItem(e.itemDefID); ok {
		if hunt := level.stewardHunt; hunt != nil {
			questDepth = hunt.SourceDepth
		}
	}
	stored := s.grantResourceBagItem(playerID, e.itemDefID, e.rollPayload, questDepth, res)
	delete(level.entities, e.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(e.id)})
	res.Events = append(res.Events, Event{
		EventType:      "resource_bag_item_picked_up",
		EntityID:       idStr(playerID),
		CorrelationID:  correlationID,
		ItemInstanceID: idStr(stored.stashItemID),
		StashItemID:    idStr(stored.stashItemID),
	})
	if ackMessageID != "" {
		res.ack(ackMessageID)
	}
	s.savePlayer(ps)

	return true
}

func (s *Sim) handleResourceBagDepositItem(in Input, res *TickResult) {
	if in.ResourceBagDepositItem == nil || in.ResourceBagDepositItem.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	item := s.findItem(in.ResourceBagDepositItem.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	if !s.isResourceBagItem(item.itemDefID) {
		res.reject(in.MessageID, "not_resource_item")
		return
	}
	if item.equipped && s.bagOccupancyCount() > s.inventoryCapacityWithItemUnequipped(item) {
		res.reject(in.MessageID, "capacity_would_overflow")
		return
	}
	if s.hotbarHasItem(item.instanceID) {
		res.reject(in.MessageID, "item_hotbar_assigned")
		return
	}

	bagItemID := s.alloc()
	deposited := &stashItem{
		stashItemID: bagItemID,
		itemDefID:   item.itemDefID,
		rollPayload: cloneRollPayload(item.rollPayload),
	}
	removedID := idStr(item.instanceID)
	transferID := "resource_bag_deposit_item:" + idStr(bagItemID)
	wasEquipped := item.equipped
	if wasEquipped {
		for _, slot := range s.clearEquippedItem(item.instanceID) {
			res.Changes = append(res.Changes, Change{
				Op:             OpEquippedUpdate,
				Slot:           slot,
				ItemInstanceID: nil,
				HotbarCapacity: intPtr(s.hotbarCapacity()),
				InventoryRows:  intPtr(s.inventoryRows()),
				InventoryCap:   intPtr(s.inventoryCapacity()),
			})
		}
		s.appendEquipmentProgressionChanges(res)
	}
	s.removeItemByID(item.instanceID)
	s.resourceBagItems = append(s.resourceBagItems, deposited)
	res.Changes = append(res.Changes,
		Change{Op: OpInventoryRemove, ItemInstanceID: &removedID, StashTransferID: transferID},
		Change{Op: OpResourceBagItemAdd, OwnerPlayerID: s.playerID, StashItem: ptrStashItemView(s.stashItemView(deposited)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:      "resource_bag_item_deposited",
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: removedID,
		StashItemID:    idStr(bagItemID),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleResourceBagDepositStashItem(in Input, res *TickResult) {
	if in.ResourceBagDepositStashItem == nil || in.ResourceBagDepositStashItem.StashEntityID == "" || in.ResourceBagDepositStashItem.StashItemID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.ResourceBagDepositStashItem.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	stored := s.findStashItem(in.ResourceBagDepositStashItem.StashItemID)
	if stored == nil {
		res.reject(in.MessageID, "stash_item_not_found")
		return
	}
	if !s.isResourceBagItem(stored.itemDefID) {
		res.reject(in.MessageID, "not_resource_item")
		return
	}

	bagItemID := s.alloc()
	deposited := &stashItem{
		stashItemID: bagItemID,
		itemDefID:   stored.itemDefID,
		rollPayload: cloneRollPayload(stored.rollPayload),
	}
	sourceStashItemID := idStr(stored.stashItemID)
	transferID := "resource_bag_deposit_stash_item:" + idStr(bagItemID)
	s.removeStashItemByID(stored.stashItemID)
	s.resourceBagItems = append(s.resourceBagItems, deposited)
	res.Changes = append(res.Changes,
		Change{Op: OpStashItemRemove, StashItemID: sourceStashItemID, StashTransferID: transferID},
		Change{Op: OpResourceBagItemAdd, OwnerPlayerID: s.playerID, StashItem: ptrStashItemView(s.stashItemView(deposited)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:     "resource_bag_stash_item_deposited",
		EntityID:      idStr(stashEntity.id),
		CorrelationID: in.CorrelationID,
		StashID:       stashID,
		StashItemID:   sourceStashItemID,
		ItemInstanceID: idStr(bagItemID),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleResourceBagWithdrawItem(in Input, res *TickResult) {
	if in.ResourceBagWithdrawItem == nil || in.ResourceBagWithdrawItem.BagItemID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stored := s.findResourceBagItem(in.ResourceBagWithdrawItem.BagItemID)
	if stored == nil {
		res.reject(in.MessageID, "bag_item_not_found")
		return
	}
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		res.reject(in.MessageID, "inventory_full")
		return
	}

	item := &invItem{
		instanceID:  s.alloc(),
		itemDefID:   stored.itemDefID,
		rollPayload: cloneRollPayload(stored.rollPayload),
	}
	bagItemID := idStr(stored.stashItemID)
	transferID := "resource_bag_withdraw_item:" + bagItemID
	s.removeResourceBagItemByID(stored.stashItemID)
	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes,
		Change{Op: OpResourceBagItemRemove, OwnerPlayerID: s.playerID, StashItemID: bagItemID, StashTransferID: transferID},
		Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:      "resource_bag_item_withdrawn",
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: idStr(item.instanceID),
		StashItemID:    bagItemID,
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleStashDepositResourceBagItem(in Input, res *TickResult) {
	if in.StashDepositResourceBagItem == nil || in.StashDepositResourceBagItem.StashEntityID == "" || in.StashDepositResourceBagItem.BagItemID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.StashDepositResourceBagItem.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	stored := s.findResourceBagItem(in.StashDepositResourceBagItem.BagItemID)
	if stored == nil {
		res.reject(in.MessageID, "bag_item_not_found")
		return
	}
	if !s.isResourceBagItem(stored.itemDefID) {
		res.reject(in.MessageID, "not_resource_item")
		return
	}
	if len(s.stashItems) >= s.stashCapacity {
		res.reject(in.MessageID, "stash_full")
		return
	}

	stashItemID := s.alloc()
	deposited := &stashItem{
		stashItemID: stashItemID,
		itemDefID:   stored.itemDefID,
		rollPayload: cloneRollPayload(stored.rollPayload),
	}
	sourceBagItemID := idStr(stored.stashItemID)
	transferID := "stash_deposit_resource_bag_item:" + idStr(stashItemID)
	s.removeResourceBagItemByID(stored.stashItemID)
	s.stashItems = append(s.stashItems, deposited)
	res.Changes = append(res.Changes,
		Change{Op: OpResourceBagItemRemove, OwnerPlayerID: s.playerID, StashItemID: sourceBagItemID, StashTransferID: transferID},
		Change{Op: OpStashItemAdd, StashItem: ptrStashItemView(s.stashItemView(deposited)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:      "stash_resource_bag_item_deposited",
		EntityID:       idStr(stashEntity.id),
		CorrelationID:  in.CorrelationID,
		StashID:        stashID,
		StashItemID:    idStr(stashItemID),
		ItemInstanceID: sourceBagItemID,
	})
	res.ack(in.MessageID)
}
