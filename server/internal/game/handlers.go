package game

import "fmt"

// Input handlers — one function per intent type, each with the same
// signature func(*Sim, Input, *TickResult).
//
// To add a new intent: implement the handler here, then register it in
// inputHandlers below.  The dispatcher in applyInput needs no edits.

// inputHandlerFunc is the common signature for all intent handlers.
type inputHandlerFunc func(*Sim, Input, *TickResult)

// wrapLevelTravel adapts the three travel intents (descend/ascend/teleport),
// whose shared handler returns a sub-result, to the flat inputHandlerFunc
// signature expected by the registry.
func wrapLevelTravel(s *Sim, in Input, res *TickResult) {
	if arrival := s.handleLevelTravel(in, res); arrival != nil {
		res.Changes = append(res.Changes, arrival.Changes...)
		res.Events = append(res.Events, arrival.Events...)
	}
}

// inputHandlers is the intent-type → handler registry.  New intents register
// here; applyInput in sim.go is the sole consumer and never needs editing.
var inputHandlers = map[string]inputHandlerFunc{
	"client_ready":                    (*Sim).handleClientReady,
	"move_intent":                     (*Sim).handleMove,
	"move_to_intent":                  (*Sim).handleMoveTo,
	"directional_attack_intent":       (*Sim).handleDirectionalAttack,
	"action_intent":                   (*Sim).handleAction,
	"descend_intent":                  wrapLevelTravel,
	"ascend_intent":                   wrapLevelTravel,
	"teleport_intent":                 wrapLevelTravel,
	"equip_intent":                    (*Sim).handleEquip,
	"unequip_intent":                  (*Sim).handleUnequip,
	"drop_intent":                     (*Sim).handleDrop,
	"use_intent":                      (*Sim).handleUse,
	"assign_hotbar_intent":            (*Sim).handleAssignHotbar,
	"use_hotbar_intent":               (*Sim).handleUseHotbar,
	"allocate_stat_intent":            (*Sim).handleAllocateStat,
	"allocate_skill_point_intent":     (*Sim).handleAllocateSkillPoint,
	"cast_skill_intent":               (*Sim).handleCastSkill,
	"set_skill_bindings_intent":       (*Sim).handleSetSkillBindings,
	"shop_buy_intent":                 (*Sim).handleShopBuy,
	"shop_sell_intent":                (*Sim).handleShopSell,
	"shop_reroll_intent":              (*Sim).handleShopReroll,
	"bishop_respec_intent":            (*Sim).handleBishopRespec,
	"bishop_debug_level_intent":       (*Sim).handleBishopDebugLevel,
	"bishop_debug_skill_point_intent": (*Sim).handleBishopDebugSkillPoint,
	"bishop_debug_stat_point_intent":  (*Sim).handleBishopDebugStatPoint,
	"stash_deposit_item_intent":       (*Sim).handleStashDepositItem,
	"stash_withdraw_item_intent":      (*Sim).handleStashWithdrawItem,
	"stash_deposit_gold_intent":       (*Sim).handleStashDepositGold,
	"stash_withdraw_gold_intent":      (*Sim).handleStashWithdrawGold,
	"corpse_withdraw_item_intent":     (*Sim).handleCorpseWithdrawItem,
	"unique_chest_take_item_intent":   (*Sim).handleUniqueChestTakeItem,
}

// handleClientReady acknowledges the client_ready handshake.
func (s *Sim) handleClientReady(in Input, res *TickResult) {
	res.ack(in.MessageID)
}

func (s *Sim) handleMove(in Input, res *TickResult) {
	if in.Move == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	dir := normalize(in.Move.Direction)
	dur := in.Move.DurationTicks
	if dur < 1 {
		dur = 1
	}
	if dir.X == 0 && dir.Y == 0 {
		s.activeLevel().move = nil
		s.clearAutoNav()
		res.ack(in.MessageID)
		return
	}
	s.clearAutoNav()
	s.activeLevel().move = &activeMove{dir: dir, remaining: dur}
	res.ack(in.MessageID)
}

func (s *Sim) handleDirectionalAttack(in Input, res *TickResult) {
	if in.DirectionalAttack == nil || !finiteVec2(in.DirectionalAttack.Direction) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	dir := normalize(in.DirectionalAttack.Direction)
	if dir.X == 0 && dir.Y == 0 {
		res.reject(in.MessageID, "invalid_direction")
		return
	}
	s.activeLevel().move = nil
	s.clearAutoNav()
	if s.playerAttackMode() == attackModeRanged {
		s.fireProjectileInDirection(dir, 0, in, res, true)
		return
	}
	weaponSlot, ok := s.consumeBasicAttack(in, res)
	if !ok {
		return
	}
	res.ack(in.MessageID)
	target := s.directionalMeleeTarget(dir)
	if target == nil {
		return
	}
	s.damageMonsterByPlayerWithSlot(target, s.playerID, in.CorrelationID, res, s.resolvePlayerAttackDamageForSlot(weaponSlot), s.playerWeaponDamageTypeForSlot(weaponSlot), weaponSlot)
}

func (s *Sim) handleMoveTo(in Input, res *TickResult) {
	if in.MoveTo == nil || !finiteVec2(in.MoveTo.Position) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return
	}
	if distance(player.pos, in.MoveTo.Position) <= s.activeNav().StopDistance {
		s.clearAutoNav()
		res.ack(in.MessageID)
		return
	}
	steps, ok := PlanPath(s.activeNav(), player.pos, in.MoveTo.Position, s.buildBlockedFn())
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = &autoNavState{steps: steps, sourceMsgID: in.MessageID, sourceCorrID: in.CorrelationID}
	res.ack(in.MessageID)
}

func (s *Sim) handleAction(in Input, res *TickResult) {
	if in.Action == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	target := s.findEntity(in.Action.TargetID)
	if target == nil || !s.actionable(target) {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if target.kind == interactableEntity &&
		target.interactableDefID == teleporterDefID &&
		(target.state == interactableDisabled || target.state == interactableLocked) {
		s.activateTeleporter(target, in, res, true)
		return
	}
	if s.inDispatchRange(target) {
		s.dispatchAction(target, in, res, true)
		return
	}

	_, steps, ok := s.findApproachGoal(target)
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = &autoNavState{
		steps:         steps,
		pendingAction: &ActionIntent{TargetID: in.Action.TargetID},
		sourceMsgID:   in.MessageID,
		sourceCorrID:  in.CorrelationID,
	}
	res.ack(in.MessageID)
}

func (s *Sim) handleShopBuy(in Input, res *TickResult) {
	if in.ShopBuy == nil || in.ShopBuy.ShopEntityID == "" || in.ShopBuy.OfferID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	shopEntity, shopID, ok, reason := s.resolveShopIntentTarget(in.ShopBuy.ShopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	entry, ok := s.findShopOffer(shopID, in.ShopBuy.OfferID, res)
	if !ok {
		res.reject(in.MessageID, "unknown_offer")
		return
	}
	offer := entry.Offer
	if s.gold < offer.BuyPrice {
		res.reject(in.MessageID, "insufficient_gold")
		return
	}
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		res.reject(in.MessageID, "inventory_full")
		return
	}

	var item *invItem
	if entry.Buyback != nil {
		row := s.removeShopBuyback(shopID, offer.OfferID)
		if row == nil || row.Item == nil {
			res.reject(in.MessageID, "unknown_offer")
			return
		}
		item = cloneInvItem(row.Item)
	} else if entry.Generated != nil && offer.Kind == shopOfferKindMystery {
		item = s.itemFromShopStock(entry.Generated, s.alloc())
		if item == nil || item.rollPayload == nil {
			res.reject(in.MessageID, "invalid_offer")
			return
		}
	} else {
		item = s.itemFromShopOffer(offer, s.alloc())
	}
	if entry.Generated != nil {
		entry.Generated.Available = false
		res.Changes = append(res.Changes, Change{Op: OpShopStockAvailability, ShopID: shopID, OfferID: entry.Generated.OfferID, Available: false})
	}
	s.inventory = append(s.inventory, item)
	s.gold -= offer.BuyPrice
	s.progression.Gold = s.gold

	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	offers, _ := s.shopCatalogWithChanges(shopID, res)
	res.Events = append(res.Events, Event{
		EventType:      "shop_purchase",
		EntityID:       idStr(shopEntity.id),
		CorrelationID:  in.CorrelationID,
		ShopID:         shopID,
		OfferID:        offer.OfferID,
		ItemInstanceID: idStr(item.instanceID),
		Price:          intPtr(offer.BuyPrice),
		TotalGold:      intPtr(s.gold),
		Item:           ptrItemView(s.itemView(item)),
		Offers:         offers,
		SellAppraisals: s.shopSellAppraisals(shopID),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleShopSell(in Input, res *TickResult) {
	if in.ShopSell == nil || in.ShopSell.ShopEntityID == "" || in.ShopSell.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	shopEntity, shopID, ok, reason := s.resolveShopIntentTarget(in.ShopSell.ShopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	item := s.findItem(in.ShopSell.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	if item.equipped {
		res.reject(in.MessageID, "item_equipped")
		return
	}
	price, ok := s.inventorySellPrice(shopID, item)
	if !ok {
		res.reject(in.MessageID, "unsellable_item")
		return
	}
	shop := s.rules.Shops[shopID]
	soldItem := cloneInvItem(item)

	removedID := idStr(item.instanceID)
	s.removeItemByID(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})
	s.clearHotbarReferences(item.instanceID, res)
	s.addShopBuyback(shopID, soldItem, shop.buybackPrice(price))

	s.gold += price
	s.progression.Gold = s.gold
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	offers, _ := s.shopCatalogWithChanges(shopID, res)
	res.Events = append(res.Events, Event{
		EventType:      "shop_sale",
		EntityID:       idStr(shopEntity.id),
		CorrelationID:  in.CorrelationID,
		ShopID:         shopID,
		ItemInstanceID: removedID,
		Price:          intPtr(price),
		TotalGold:      intPtr(s.gold),
		Offers:         offers,
		SellAppraisals: s.shopSellAppraisals(shopID),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleShopReroll(in Input, res *TickResult) {
	if in.ShopReroll == nil || in.ShopReroll.ShopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	shopEntity, shopID, ok, reason := s.resolveShopIntentTarget(in.ShopReroll.ShopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	shop, ok := s.rules.Shops[shopID]
	if !ok || !shop.MysteryOffers.Enabled {
		res.reject(in.MessageID, "reroll_unavailable")
		return
	}
	cost := shop.MysteryOffers.RerollCost
	if s.gold < cost {
		res.reject(in.MessageID, "insufficient_gold")
		return
	}
	state := s.rerollMysteryShopStock(shopID, shop, res)
	s.gold -= cost
	s.progression.Gold = s.gold
	res.Changes = append(res.Changes, Change{Op: OpGoldUpdate, Gold: intPtr(s.gold)})
	view := s.CharacterProgressionView()
	res.Changes = append(res.Changes, Change{Op: OpCharacterProgressionUpdate, Progression: &view})
	offers, _ := s.shopCatalogWithChanges(shopID, res)
	res.Events = append(res.Events, Event{
		EventType:     "shop_reroll",
		EntityID:      idStr(shopEntity.id),
		CorrelationID: in.CorrelationID,
		ShopID:        shopID,
		Price:         intPtr(cost),
		TotalGold:     intPtr(s.gold),
		RefreshKey:    state.RefreshKey,
		Offers:        offers,
	})
	res.ack(in.MessageID)
}

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
	cost := s.respecCostGold()
	if s.gold < cost {
		res.reject(in.MessageID, "not_enough_gold")
		return
	}
	s.gold -= cost
	s.progression.Gold = s.gold
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
		Price:              intPtr(cost),
		TotalGold:          intPtr(s.gold),
		UnspentStatPoints:  intPtr(s.progression.UnspentStatPoints),
		UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleStashDepositItem(in Input, res *TickResult) {
	if in.StashDepositItem == nil || in.StashDepositItem.StashEntityID == "" || in.StashDepositItem.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.StashDepositItem.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	item := s.findItem(in.StashDepositItem.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
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
	if len(s.stashItems) >= s.stashCapacity {
		res.reject(in.MessageID, "stash_full")
		return
	}

	stashItemID := s.alloc()
	deposited := &stashItem{
		stashItemID: stashItemID,
		itemDefID:   item.itemDefID,
		rollPayload: cloneRollPayload(item.rollPayload),
	}
	removedID := idStr(item.instanceID)
	transferID := "stash_deposit_item:" + idStr(stashItemID)
	wasEquipped := item.equipped
	if wasEquipped {
		for _, slot := range sortedStringKeys(s.equipped) {
			if s.equipped[slot] == item.instanceID {
				s.equipped[slot] = 0
				res.Changes = append(res.Changes, Change{
					Op:             OpEquippedUpdate,
					Slot:           slot,
					ItemInstanceID: nil,
					HotbarCapacity: intPtr(s.hotbarCapacity()),
					InventoryRows:  intPtr(s.inventoryRows()),
					InventoryCap:   intPtr(s.inventoryCapacity()),
				})
			}
		}
		s.appendEquipmentProgressionChanges(res)
	}
	s.removeItemByID(item.instanceID)
	s.stashItems = append(s.stashItems, deposited)

	res.Changes = append(res.Changes,
		Change{Op: OpInventoryRemove, ItemInstanceID: &removedID, StashTransferID: transferID},
		Change{Op: OpStashItemAdd, StashItem: ptrStashItemView(s.stashItemView(deposited)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:      "stash_item_deposited",
		EntityID:       idStr(stashEntity.id),
		CorrelationID:  in.CorrelationID,
		StashID:        stashID,
		ItemInstanceID: removedID,
		StashItemID:    idStr(stashItemID),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleStashWithdrawItem(in Input, res *TickResult) {
	if in.StashWithdrawItem == nil || in.StashWithdrawItem.StashEntityID == "" || in.StashWithdrawItem.StashItemID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.StashWithdrawItem.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	stored := s.findStashItem(in.StashWithdrawItem.StashItemID)
	if stored == nil {
		res.reject(in.MessageID, "stash_item_not_found")
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
	stashItemID := idStr(stored.stashItemID)
	transferID := "stash_withdraw_item:" + stashItemID
	s.removeStashItemByID(stored.stashItemID)
	s.inventory = append(s.inventory, item)

	res.Changes = append(res.Changes,
		Change{Op: OpStashItemRemove, StashItemID: stashItemID, StashTransferID: transferID},
		Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item)), StashTransferID: transferID},
	)
	res.Events = append(res.Events, Event{
		EventType:      "stash_item_withdrawn",
		EntityID:       idStr(stashEntity.id),
		CorrelationID:  in.CorrelationID,
		StashID:        stashID,
		ItemInstanceID: idStr(item.instanceID),
		StashItemID:    stashItemID,
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleStashDepositGold(in Input, res *TickResult) {
	if in.StashDepositGold == nil || in.StashDepositGold.StashEntityID == "" || in.StashDepositGold.Amount <= 0 {
		res.reject(in.MessageID, "invalid_amount")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.StashDepositGold.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	amount := in.StashDepositGold.Amount
	if s.gold < amount {
		res.reject(in.MessageID, "insufficient_gold")
		return
	}
	transferID := fmt.Sprintf("stash_deposit_gold:%d:%d", s.tick, amount)
	s.gold -= amount
	s.progression.Gold = s.gold
	s.stashGold += amount
	s.appendStashGoldChanges(res, transferID)
	res.Events = append(res.Events, Event{
		EventType:     "stash_gold_deposited",
		EntityID:      idStr(stashEntity.id),
		CorrelationID: in.CorrelationID,
		StashID:       stashID,
		Amount:        intPtr(amount),
		TotalGold:     intPtr(s.gold),
		StashGold:     intPtr(s.stashGold),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleStashWithdrawGold(in Input, res *TickResult) {
	if in.StashWithdrawGold == nil || in.StashWithdrawGold.StashEntityID == "" || in.StashWithdrawGold.Amount <= 0 {
		res.reject(in.MessageID, "invalid_amount")
		return
	}
	stashEntity, stashID, ok, reason := s.resolveStashIntentTarget(in.StashWithdrawGold.StashEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	amount := in.StashWithdrawGold.Amount
	if s.stashGold < amount {
		res.reject(in.MessageID, "insufficient_stash_gold")
		return
	}
	transferID := fmt.Sprintf("stash_withdraw_gold:%d:%d", s.tick, amount)
	s.stashGold -= amount
	s.gold += amount
	s.progression.Gold = s.gold
	s.appendStashGoldChanges(res, transferID)
	res.Events = append(res.Events, Event{
		EventType:     "stash_gold_withdrawn",
		EntityID:      idStr(stashEntity.id),
		CorrelationID: in.CorrelationID,
		StashID:       stashID,
		Amount:        intPtr(amount),
		TotalGold:     intPtr(s.gold),
		StashGold:     intPtr(s.stashGold),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleLevelTravel(in Input, res *TickResult) *TickResult {
	if in.Type == "teleport_intent" {
		return s.handleTeleport(in, res)
	}
	return s.handleTransition(in, res)
}

func (s *Sim) handleTransition(in Input, res *TickResult) *TickResult {
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
		return nil
	}
	if s.playerDead() {
		res.reject(in.MessageID, "player_dead")
		return nil
	}

	var (
		stairDefID string
		destLevel  int
		arrivalDef string
	)
	switch in.Type {
	case "descend_intent":
		if in.Descend == nil {
			res.reject(in.MessageID, "invalid_payload")
			return nil
		}
		stairDefID = stairsDownDefID
		destLevel = s.currentLevel - 1
		arrivalDef = stairsUpDefID
	case "ascend_intent":
		if in.Ascend == nil {
			res.reject(in.MessageID, "invalid_payload")
			return nil
		}
		if s.currentLevel >= townLevel {
			res.reject(in.MessageID, "already_at_entry")
			return nil
		}
		stairDefID = stairsUpDefID
		destLevel = s.currentLevel + 1
		arrivalDef = stairsDownDefID
	default:
		res.reject(in.MessageID, "invalid_payload")
		return nil
	}
	current := s.activeLevel()
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	stair := s.findReachableStair(current, stairDefID, player.pos)
	if stair == nil {
		res.reject(in.MessageID, "no_stair_in_range")
		return nil
	}
	if stair.state == interactableLocked || stair.state == interactableDisabled {
		reason := s.rules.DungeonGeneration.BossFloor.LockedExitReason
		if reason == "" {
			reason = "locked"
		}
		res.reject(in.MessageID, reason)
		eventType := "descend_blocked"
		if in.Type == "ascend_intent" {
			eventType = "ascend_blocked"
		}
		res.Events = append(res.Events, Event{EventType: eventType, EntityID: idStr(stair.id), CorrelationID: in.CorrelationID, Reason: reason})
		return nil
	}

	dest, err := s.ensureTravelLevel(destLevel)
	if err != nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	arrival := s.findStair(dest, arrivalDef)
	if arrival == nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	return s.movePlayerToLevel(in, res, current, dest, s.travelArrivalPosition(dest, arrival.pos, s.playerID))
}

func (s *Sim) handleTeleport(in Input, res *TickResult) *TickResult {
	if in.Teleport == nil {
		res.reject(in.MessageID, "invalid_payload")
		return nil
	}
	if !s.multiLevel {
		res.reject(in.MessageID, "not_dungeon_world")
		return nil
	}
	if s.playerDead() {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	targetLevel := in.Teleport.TargetLevel
	if targetLevel > townLevel {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	if !s.discoveredTeleporters[s.currentLevel] {
		res.reject(in.MessageID, "teleporter_not_discovered")
		return nil
	}
	if !s.discoveredTeleporters[targetLevel] {
		res.reject(in.MessageID, "target_level_not_discovered")
		return nil
	}
	current := s.activeLevel()
	player := current.entities[s.playerID]
	if player == nil {
		res.reject(in.MessageID, "player_dead")
		return nil
	}
	teleporter := s.findReachableTeleporter(current, player.pos)
	if teleporter == nil {
		res.reject(in.MessageID, "no_teleporter_in_range")
		return nil
	}
	if teleporter.state == interactableDisabled || teleporter.state == interactableLocked {
		reason := s.rules.DungeonGeneration.BossFloor.LockedExitReason
		if reason == "" {
			reason = "locked"
		}
		res.reject(in.MessageID, reason)
		res.Events = append(res.Events, Event{EventType: "teleport_blocked", EntityID: idStr(teleporter.id), CorrelationID: in.CorrelationID, Reason: reason})
		return nil
	}
	dest, err := s.ensureTravelLevel(targetLevel)
	if err != nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	arrival := s.findTeleporter(dest)
	if arrival == nil {
		res.reject(in.MessageID, "invalid_level")
		return nil
	}
	return s.movePlayerToLevel(in, res, current, dest, s.travelArrivalPosition(dest, arrival.pos, s.playerID))
}

func (s *Sim) handleEquip(in Input, res *TickResult) {
	if in.Equip == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	if !isEquipmentSlot(in.Equip.Slot) {
		res.reject(in.MessageID, "wrong_slot")
		return
	}
	item := s.findItem(in.Equip.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	itemSlot, ok := s.itemEquipSlot(item)
	if !ok {
		res.reject(in.MessageID, "not_equippable")
		return
	}
	if !s.slotAcceptsItem(in.Equip.Slot, item, itemSlot) {
		res.reject(in.MessageID, "wrong_slot")
		return
	}
	if s.slotBlockedByHands(in.Equip.Slot, item) {
		res.reject(in.MessageID, "hands_blocked")
		return
	}
	if !s.itemClassAllowed(item) {
		res.reject(in.MessageID, "class_requirement_not_met")
		return
	}
	if item.rollPayload != nil && !s.requirementsMet(item.rollPayload.Requirements) {
		res.reject(in.MessageID, "requirements_not_met")
		return
	}

	clearedSlots := s.slotsClearedByEquip(in.Equip.Slot, item)
	bagCountAfter := s.bagOccupancyCount()
	for _, slot := range clearedSlots {
		prevID := s.equipped[slot]
		if prevID == 0 || prevID == item.instanceID {
			continue
		}
		prev := s.findItemByID(prevID)
		if prev != nil && !s.hotbarHasItem(prev.instanceID) {
			bagCountAfter++
		}
	}
	if !item.equipped && !s.hotbarHasItem(item.instanceID) {
		bagCountAfter--
	}
	capacityAfter := inventoryCapacityForRows(s.inventoryRowsAfterEquip(in.Equip.Slot, item, clearedSlots))
	if bagCountAfter > capacityAfter {
		res.reject(in.MessageID, "capacity_would_overflow")
		return
	}
	for _, slot := range clearedSlots {
		prevID := s.equipped[slot]
		if prevID == 0 || prevID == item.instanceID {
			continue
		}
		if prev := s.findItemByID(prevID); prev != nil {
			prev.equipped = false
			res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(s.itemView(prev))})
		}
		s.equipped[slot] = 0
		res.Changes = append(res.Changes, Change{Op: OpEquippedUpdate, Slot: slot, ItemInstanceID: nil})
	}

	item.slot = in.Equip.Slot
	item.equipped = true
	s.equipped[in.Equip.Slot] = item.instanceID

	res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(s.itemView(item))})
	idCopy := idStr(item.instanceID)
	res.Changes = append(res.Changes, Change{
		Op:             OpEquippedUpdate,
		Slot:           in.Equip.Slot,
		ItemInstanceID: &idCopy,
		HotbarCapacity: intPtr(s.hotbarCapacity()),
		InventoryRows:  intPtr(s.inventoryRows()),
		InventoryCap:   intPtr(s.inventoryCapacity()),
	})
	s.appendEquipmentProgressionChanges(res)
	res.Events = append(res.Events, Event{EventType: "item_equipped", EntityID: idCopy, CorrelationID: in.CorrelationID})
	res.ack(in.MessageID)
}

func (s *Sim) handleUnequip(in Input, res *TickResult) {
	if in.Unequip == nil || !isEquipmentSlot(in.Unequip.Slot) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	instanceID := s.equipped[in.Unequip.Slot]
	if instanceID == 0 {
		res.reject(in.MessageID, "slot_empty")
		return
	}
	item := s.findItemByID(instanceID)
	if item == nil {
		res.reject(in.MessageID, "slot_empty")
		return
	}
	capacityAfter := s.inventoryCapacityWithItemUnequipped(item)
	bagCountAfter := s.bagOccupancyCount()
	if !s.hotbarHasItem(item.instanceID) {
		bagCountAfter++
	}
	if bagCountAfter > capacityAfter {
		res.reject(in.MessageID, "capacity_would_overflow")
		return
	}
	item.equipped = false
	s.equipped[in.Unequip.Slot] = 0
	res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(s.itemView(item))})
	res.Changes = append(res.Changes, Change{
		Op:             OpEquippedUpdate,
		Slot:           in.Unequip.Slot,
		ItemInstanceID: nil,
		HotbarCapacity: intPtr(s.hotbarCapacity()),
		InventoryRows:  intPtr(s.inventoryRows()),
		InventoryCap:   intPtr(s.inventoryCapacity()),
	})
	s.appendEquipmentProgressionChanges(res)
	idCopy := idStr(item.instanceID)
	res.Events = append(res.Events, Event{EventType: "item_unequipped", EntityID: idCopy, CorrelationID: in.CorrelationID})
	res.ack(in.MessageID)
}

func (s *Sim) handleDrop(in Input, res *TickResult) {
	if in.Drop == nil || in.Drop.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	item := s.findItem(in.Drop.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	dropPos, ok := s.findDropPosition()
	if !ok {
		res.reject(in.MessageID, "no_drop_space")
		return
	}

	wasEquipped := item.equipped
	if item.equipped {
		for _, slot := range sortedStringKeys(s.equipped) {
			if s.equipped[slot] == item.instanceID {
				s.equipped[slot] = 0
				res.Changes = append(res.Changes, Change{Op: OpEquippedUpdate, Slot: slot, ItemInstanceID: nil})
			}
		}
	}

	removedID := idStr(item.instanceID)
	itemDefID := item.itemDefID
	rollPayload := cloneRollPayload(item.rollPayload)
	s.removeItemByID(item.instanceID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &removedID})
	s.clearHotbarReferences(item.instanceID, res)

	loot := s.newLootEntity(itemDefID, dropPos, rollPayload, goldRollContext{levelNum: s.activeLevel().levelNum})
	loot.id = s.alloc()
	s.activeLevel().entities[loot.id] = loot
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(loot))})
	if wasEquipped {
		s.appendEquipmentProgressionChanges(res)
	}
	res.Events = append(res.Events, Event{
		EventType:      "item_dropped",
		EntityID:       idStr(loot.id),
		CorrelationID:  in.CorrelationID,
		ItemInstanceID: removedID,
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleAssignHotbar(in Input, res *TickResult) {
	if in.AssignHotbar == nil || !validHotbarSlot(in.AssignHotbar.SlotIndex) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	if in.AssignHotbar.ItemInstanceID == nil {
		assignedID := s.hotbar[in.AssignHotbar.SlotIndex]
		if assignedID != 0 && s.bagOccupancyCount()+1 > s.inventoryCapacity() {
			res.reject(in.MessageID, "inventory_full")
			return
		}
		s.hotbar[in.AssignHotbar.SlotIndex] = 0
		if assignedID != 0 && !s.hotbarHasItem(assignedID) {
			if item := s.findItemByID(assignedID); item != nil {
				res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
			}
		}
		res.Changes = append(res.Changes, Change{
			Op:             OpHotbarUpdate,
			SlotIndex:      in.AssignHotbar.SlotIndex,
			ItemInstanceID: nil,
			InventoryRows:  intPtr(s.inventoryRows()),
			InventoryCap:   intPtr(s.inventoryCapacity()),
		})
		res.ack(in.MessageID)
		return
	}
	item := s.findItem(*in.AssignHotbar.ItemInstanceID)
	if item == nil || item.equipped {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	if !s.itemIsConsumable(item) {
		res.reject(in.MessageID, "not_consumable")
		return
	}
	assignedID := s.hotbar[in.AssignHotbar.SlotIndex]
	if assignedID != 0 && assignedID != item.instanceID {
		bagCountAfter := s.bagOccupancyCount()
		if !s.hotbarHasItem(item.instanceID) {
			bagCountAfter--
		}
		if !s.hotbarHasItemExcept(assignedID, in.AssignHotbar.SlotIndex) {
			bagCountAfter++
		}
		if bagCountAfter > s.inventoryCapacity() {
			res.reject(in.MessageID, "inventory_full")
			return
		}
	}
	s.hotbar[in.AssignHotbar.SlotIndex] = item.instanceID
	if assignedID != 0 && assignedID != item.instanceID && !s.hotbarHasItem(assignedID) {
		if oldItem := s.findItemByID(assignedID); oldItem != nil {
			res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(oldItem))})
		}
	}
	idCopy := idStr(item.instanceID)
	if assignedID != item.instanceID && !s.hotbarHasItemExcept(item.instanceID, in.AssignHotbar.SlotIndex) {
		res.Changes = append(res.Changes, Change{Op: OpInventoryRemove, ItemInstanceID: &idCopy})
	}
	itemView := s.itemView(item)
	res.Changes = append(res.Changes, Change{
		Op:             OpHotbarUpdate,
		SlotIndex:      in.AssignHotbar.SlotIndex,
		ItemInstanceID: &idCopy,
		Item:           &itemView,
		InventoryRows:  intPtr(s.inventoryRows()),
		InventoryCap:   intPtr(s.inventoryCapacity()),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleUseHotbar(in Input, res *TickResult) {
	if in.UseHotbar == nil || !validHotbarSlot(in.UseHotbar.SlotIndex) {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	if in.UseHotbar.SlotIndex >= s.hotbarCapacity() {
		res.reject(in.MessageID, "hotbar_slot_disabled")
		return
	}
	itemID := s.hotbar[in.UseHotbar.SlotIndex]
	if itemID == 0 {
		res.reject(in.MessageID, "slot_empty")
		return
	}
	item := s.findItemByID(itemID)
	if item == nil || item.equipped {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	if !s.itemIsConsumable(item) {
		res.reject(in.MessageID, "not_consumable")
		return
	}
	if ok, reason := s.consumeItem(item, in.CorrelationID, res); !ok {
		res.reject(in.MessageID, reason)
		return
	}
	res.ack(in.MessageID)
}

func (s *Sim) handleUse(in Input, res *TickResult) {
	if in.Use == nil || in.Use.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	item := s.findItem(in.Use.ItemInstanceID)
	if item == nil {
		res.reject(in.MessageID, "not_in_inventory")
		return
	}
	if !s.itemIsConsumable(item) {
		res.reject(in.MessageID, "not_consumable")
		return
	}
	if ok, reason := s.consumeItem(item, in.CorrelationID, res); !ok {
		res.reject(in.MessageID, reason)
		return
	}
	res.ack(in.MessageID)
}

func (s *Sim) handleSetSkillBindings(in Input, res *TickResult) {
	if in.SetSkillBindings == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	keys := normalizeSkillFunctionKeys(in.SetSkillBindings.FunctionKeys)
	for _, skillID := range keys {
		if skillID == "" {
			continue
		}
		if _, ok := s.rules.Skills[skillID]; !ok {
			res.reject(in.MessageID, "unknown_skill")
			return
		}
	}
	if rightClick := in.SetSkillBindings.RightClickSkillID; rightClick != "" {
		if _, ok := s.rules.Skills[rightClick]; !ok {
			res.reject(in.MessageID, "unknown_skill")
			return
		}
	}
	s.skillFunctionKeys = keys
	s.rightClickSkillID = in.SetSkillBindings.RightClickSkillID
	view := s.SkillBindingsView()
	res.Changes = append(res.Changes, Change{Op: OpSkillBindingsUpdate, SkillBindings: &view})
	res.ack(in.MessageID)
}

func (s *Sim) handleAllocateStat(in Input, res *TickResult) {
	if in.AllocateStat == nil || !isBaseStat(in.AllocateStat.Stat) || in.AllocateStat.Points <= 0 {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	if in.AllocateStat.Points > s.progression.UnspentStatPoints {
		res.reject(in.MessageID, "not_enough_stat_points")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	beforeMaxHP := s.currentMaxHP()
	beforeMaxMana := s.currentMaxMana()
	switch in.AllocateStat.Stat {
	case "str":
		s.progression.BaseStats.Str += in.AllocateStat.Points
	case "dex":
		s.progression.BaseStats.Dex += in.AllocateStat.Points
	case "vit":
		s.progression.BaseStats.Vit += in.AllocateStat.Points
	case "magic":
		s.progression.BaseStats.Magic += in.AllocateStat.Points
	}
	s.progression.UnspentStatPoints -= in.AllocateStat.Points
	afterMaxHP := s.currentMaxHP()
	afterMaxMana := s.currentMaxMana()
	entityChanged := false
	if afterMaxHP != player.maxHP {
		player.maxHP = afterMaxHP
		if delta := afterMaxHP - beforeMaxHP; delta > 0 {
			player.hp += delta
			if player.hp > player.maxHP {
				player.hp = player.maxHP
			}
		}
		if player.hp > player.maxHP {
			player.hp = player.maxHP
		}
		entityChanged = true
	}
	if afterMaxMana != player.maxMana {
		player.maxMana = afterMaxMana
		if delta := afterMaxMana - beforeMaxMana; delta > 0 {
			player.mana += delta
			if player.mana > player.maxMana {
				player.mana = player.maxMana
			}
		}
		if player.mana > player.maxMana {
			player.mana = player.maxMana
		}
		entityChanged = true
	}
	if entityChanged {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	}

	s.appendProgressionAndSkillUpdates(res)
	res.Events = append(res.Events, Event{
		EventType:         "stat_allocated",
		CorrelationID:     in.CorrelationID,
		Stat:              in.AllocateStat.Stat,
		Amount:            intPtr(in.AllocateStat.Points),
		UnspentStatPoints: intPtr(s.progression.UnspentStatPoints),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleAllocateSkillPoint(in Input, res *TickResult) {
	if in.AllocateSkillPoint == nil || in.AllocateSkillPoint.SkillID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	skillID := in.AllocateSkillPoint.SkillID
	def, ok := s.rules.Skills[skillID]
	if !ok {
		res.reject(in.MessageID, "unknown_skill")
		return
	}
	if s.progression.UnspentSkillPoints <= 0 {
		res.reject(in.MessageID, "not_enough_skill_points")
		return
	}
	rank := s.progression.SkillRanks[skillID]
	if rank >= def.MaxRank {
		res.reject(in.MessageID, "skill_max_rank")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	nextRank := rank + 1
	if !s.skillClassAllowed(def) {
		res.reject(in.MessageID, "skill_class_not_allowed")
		return
	}
	if !s.skillRequirementsMet(def, nextRank) {
		res.reject(in.MessageID, "skill_requirements_not_met")
		return
	}

	rank = nextRank
	s.progression.SkillRanks[skillID] = rank
	s.progression.UnspentSkillPoints--
	s.appendProgressionAndSkillUpdates(res)
	res.Events = append(res.Events, Event{
		EventType:          "skill_rank_updated",
		EntityID:           idStr(player.id),
		CorrelationID:      in.CorrelationID,
		SkillID:            skillID,
		Rank:               intPtr(rank),
		MaxRank:            intPtr(def.MaxRank),
		UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
	})
	res.ack(in.MessageID)
}

func (s *Sim) handleCastSkill(in Input, res *TickResult) {
	if in.CastSkill == nil || in.CastSkill.SkillID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	skillID := in.CastSkill.SkillID
	def, ok := s.rules.Skills[skillID]
	if !ok {
		res.reject(in.MessageID, "unknown_skill")
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}
	rank := s.effectiveSkillRank(skillID)
	if rank <= 0 {
		res.reject(in.MessageID, "skill_not_learned")
		return
	}
	if !s.skillClassAllowed(def) {
		res.reject(in.MessageID, "skill_class_not_allowed")
		return
	}
	if !s.skillRequirementsMet(def, rank) {
		res.reject(in.MessageID, "skill_requirements_not_met")
		return
	}
	if remaining, onCooldown := s.skillCooldownRemaining(skillID); onCooldown {
		res.Events = append(res.Events, Event{
			EventType:      "skill_cooldown_rejected",
			EntityID:       idStr(player.id),
			CorrelationID:  in.CorrelationID,
			SkillID:        skillID,
			Reason:         "skill_on_cooldown",
			RemainingTicks: intPtr(remaining),
		})
		res.reject(in.MessageID, "skill_on_cooldown")
		return
	}
	manaCost := skillManaCost(def, rank)
	if player.mana < manaCost && !s.tryBloodPriceForSkill(player, manaCost, in.CorrelationID, res) {
		res.reject(in.MessageID, "not_enough_mana")
		return
	}
	switch def.Kind {
	case "projectile_attack", "cold_projectile_attack", "chain_projectile_attack":
		if def.Pierce.MaxHits > 0 || def.Root.DurationTicks > 0 || def.Volley.ArrowCount > 0 {
			s.handleRangerProjectileSkillCast(in, res, player, skillID, def, rank, manaCost)
			return
		}
		s.handleProjectileSkillCast(in, res, player, skillID, def, rank, manaCost)
	case "cone_attack":
		if def.Dash.RangeBase > 0 {
			s.handleDashSkillCast(in, res, player, skillID, def, rank, manaCost)
			return
		}
		s.handleConeSkillCast(in, res, player, skillID, def, rank, manaCost)
	case "self_buff":
		s.handleSelfBuffSkillCast(in, res, player, skillID, def, rank, manaCost)
	case "area_heal":
		s.handleAreaHealSkillCast(in, res, player, skillID, def, rank, manaCost)
	case "area_stat_buff":
		s.handleAreaStatBuffSkillCast(in, res, player, skillID, def, rank, manaCost)
	default:
		res.reject(in.MessageID, "unsupported_skill_kind")
	}
}

func (s *Sim) commitSkillSpend(player *entity, skillID string, def SkillDef, manaCost int) int {
	player.mana -= manaCost
	if player.mana < 0 {
		player.mana = 0
	}
	cooldownTicks := s.skillCooldownTicks(def)
	s.skillCooldowns[skillID] = skillCooldownState{EndsTick: s.tick + uint64(cooldownTicks), TotalTicks: cooldownTicks}
	return cooldownTicks
}

func (s *Sim) appendSkillCastEvent(res *TickResult, player *entity, skillID string, rank int, manaCost int, correlationID string, targetID uint64, projectileDefID string) {
	event := Event{
		EventType:      "skill_cast",
		EntityID:       idStr(player.id),
		SourceEntityID: idStr(player.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Rank:           intPtr(rank),
		Mana:           intPtr(manaCost),
	}
	if targetID != 0 {
		event.TargetEntityID = idStr(targetID)
	}
	if projectileDefID != "" {
		event.ProjectileDefID = projectileDefID
	}
	res.Events = append(res.Events, event)
}

func (s *Sim) appendConeSkillCastEvent(res *TickResult, player *entity, skillID string, rank int, manaCost int, correlationID string, targetID uint64, dir Vec2, cone SkillConeDef) {
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, correlationID, targetID, "")
	if len(res.Events) == 0 {
		return
	}
	event := &res.Events[len(res.Events)-1]
	event.Position = cloneVec2Ptr(&player.pos)
	event.Direction = cloneVec2Ptr(&dir)
	event.Range = floatPtr(cone.Range)
	event.AngleDegrees = floatPtr(cone.AngleDegrees)
}

func (s *Sim) appendSkillCooldownStartedEvent(res *TickResult, player *entity, skillID string, correlationID string, cooldownTicks int) {
	res.Events = append(res.Events, Event{
		EventType:      "skill_cooldown_started",
		EntityID:       idStr(player.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		RemainingTicks: intPtr(cooldownTicks),
		TotalTicks:     intPtr(cooldownTicks),
	})
}

func (s *Sim) handleProjectileSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	dir, targetID, rejectReason := s.skillCastDirection(def, in.CastSkill, player)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, def.Projectile.Range, true)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	projectile := s.spawnSkillProjectile(player, skillID, def, rank, dir, targetID, in)
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(projectile))})
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, skillID)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) handleConeSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	dir, targetID, rejectReason := s.skillCastDirectionWithRange(def, in.CastSkill, player, def.Cone.Range)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, def.Cone.Range, false)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}
	targets := s.coneSkillTargets(player, dir, def.Cone)
	if len(targets) == 0 {
		res.reject(in.MessageID, "no_valid_targets")
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendConeSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, dir, def.Cone)
	s.applyConeSkill(player, skillID, def, targets, in.CorrelationID, res)
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) handleSelfBuffSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, 0, "")
	s.applySkillBuff(player, skillID, def, rank, in.CorrelationID, res)
	s.syncActivePlayerMaxResources()
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) handleAreaHealSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	applications, rejectReason := s.areaHealApplications(player, def, rank, in.CastSkill)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, skillCastRange(def), false)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, 0, "")
	s.applyAreaHeal(player, skillID, rank, applications, in.CorrelationID, res)
	s.startAreaHealZones(player, skillID, def, rank, in.CastSkill, in.CorrelationID)
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) handleAreaStatBuffSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	applications, rejectReason := s.areaStatBuffApplications(player, def, rank, in.CastSkill)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, skillCastRange(def), false)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}
	if len(applications) == 0 {
		res.reject(in.MessageID, "no_valid_targets")
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, 0, "")
	s.applyAreaStatBuff(player, skillID, rank, applications, in.CorrelationID, res)
	s.syncActivePlayerMaxResources()
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) beginSkillAutoNav(in Input, res *TickResult, castRange float64, requireClearShot bool) {
	if in.CastSkill == nil || in.CastSkill.TargetID == "" {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	target := s.findEntity(in.CastSkill.TargetID)
	if target == nil || (target.kind != monsterEntity && target.kind != playerEntity) || target.hp <= 0 {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	_, steps, ok := s.findSkillCastApproachGoal(target, castRange, requireClearShot)
	if !ok {
		res.reject(in.MessageID, "no_path")
		return
	}
	if len(steps) > s.activeNav().MaxAutoSteps {
		res.reject(in.MessageID, "path_too_long")
		return
	}
	s.activeLevel().move = nil
	s.activeLevel().autoNav = &autoNavState{
		steps: steps,
		pendingSkill: &CastSkillIntent{
			SkillID:   in.CastSkill.SkillID,
			TargetID:  in.CastSkill.TargetID,
			Direction: cloneVec2Ptr(in.CastSkill.Direction),
		},
		sourceMsgID:  in.MessageID,
		sourceCorrID: in.CorrelationID,
	}
	res.ack(in.MessageID)
}
