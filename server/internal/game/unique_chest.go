package game

import (
	"fmt"
	"strconv"
	"strings"
)

func (r *Rules) validateUniqueItemRules(uniqueItems map[string]UniqueItemDef) error {
	for uniqueID, unique := range uniqueItems {
		if unique.ID != uniqueID {
			return fmt.Errorf("game: invalid rules unique_items.%s.id: must match key", uniqueID)
		}
		template, ok := r.ItemTemplates[unique.BaseTemplateID]
		if !ok {
			return fmt.Errorf("game: invalid rules unique_items.%s.base_template_id: unknown template %s", uniqueID, unique.BaseTemplateID)
		}
		if unique.Enabled && unique.Status != "ready" {
			return fmt.Errorf("game: invalid rules unique_items.%s.status: enabled entries must be ready", uniqueID)
		}
		if !unique.Enabled && unique.Status != "disabled_seed" {
			return fmt.Errorf("game: invalid rules unique_items.%s.status: disabled entries must remain disabled_seed", uniqueID)
		}
		if err := r.validateNamedUniqueEffects(uniqueID, unique, template.ItemType); err != nil {
			return err
		}
	}
	return nil
}

func (r *Rules) validateNamedUniqueEffects(uniqueID string, unique UniqueItemDef, itemType string) error {
	seenEffects := map[string]bool{}
	for _, effectID := range unique.FixedEffectIDs {
		if seenEffects[effectID] {
			return fmt.Errorf("game: invalid rules unique_items.%s.fixed_effect_ids.%s: duplicate effect", uniqueID, effectID)
		}
		seenEffects[effectID] = true
		effect, ok := r.UniqueEffects[effectID]
		if !ok || !effect.Enabled || effect.Status != "ready" {
			return fmt.Errorf("game: invalid rules unique_items.%s.fixed_effect_ids.%s: unknown or inactive effect", uniqueID, effectID)
		}
		if !uniqueChestEffectCompatible(effect, itemType) {
			return fmt.Errorf("game: invalid rules unique_items.%s.fixed_effect_ids.%s: incompatible with template type %s", uniqueID, effectID, itemType)
		}
	}
	return nil
}

func (s *Sim) openUniqueTestChest(e *entity, in Input, res *TickResult, ack bool) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if e.state != interactableReady && e.state != interactableOpen {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	state := s.uniqueChests[e.id]
	if state == nil {
		items, ok := s.uniqueTestChestItems()
		if !ok {
			res.reject(in.MessageID, "invalid_target")
			return
		}
		state = &uniqueChestState{items: make([]*stashItem, 0, len(items))}
		for _, item := range items {
			state.items = append(state.items, &stashItem{
				stashItemID: s.alloc(),
				itemDefID:   item.itemDefID,
				rollPayload: cloneRollPayload(item.rollPayload),
			})
		}
		s.uniqueChests[e.id] = state
	} else if changed, ok := s.backfillUniqueTestChestState(state); !ok {
		res.reject(in.MessageID, "invalid_target")
		return
	} else if changed {
		s.savePlayer(s.defaultPlayer())
	}
	if e.state != interactableOpen {
		e.state = interactableOpen
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
	}
	res.Events = append(res.Events, Event{
		EventType:     "unique_chest_opened",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		Service:       uniqueTestChestService,
		Amount:        intPtr(len(state.items)),
		StashID:       uniqueTestChestService,
		StashItems:    s.uniqueChestItemViews(state),
		StashCapacity: intPtr(len(state.items)),
		Inventory:     s.inventoryView(),
		Equipped:      newSnapshotEquippedMap(s.equipped),
		Gold:          intPtr(s.gold),
		Hotbar:        s.hotbarView(),
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) handleUniqueChestTakeItem(in Input, res *TickResult) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if in.UniqueChestTakeItem == nil || in.UniqueChestTakeItem.ChestEntityID == "" || in.UniqueChestTakeItem.ChestItemID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	entity, _, ok := s.findEntityAnyLevel(in.UniqueChestTakeItem.ChestEntityID)
	if !ok || entity == nil || entity.kind != interactableEntity || s.serviceForInteractable(entity) != uniqueTestChestService {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if !s.inDispatchRange(entity) {
		res.reject(in.MessageID, "out_of_range")
		return
	}
	state := s.uniqueChests[entity.id]
	if state == nil {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	chestItemID, err := strconv.ParseUint(in.UniqueChestTakeItem.ChestItemID, 10, 64)
	if err != nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	idx := -1
	var stored *stashItem
	for i, candidate := range state.items {
		if candidate.stashItemID == chestItemID {
			idx = i
			stored = candidate
			break
		}
	}
	if stored == nil {
		res.reject(in.MessageID, "not_found")
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
	item.slot = s.itemSlot(item.itemDefID, item.rollPayload)
	state.items = append(state.items[:idx], state.items[idx+1:]...)
	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	res.Events = append(res.Events, Event{
		EventType:      "unique_chest_item_taken",
		EntityID:       idStr(entity.id),
		CorrelationID:  in.CorrelationID,
		Service:        uniqueTestChestService,
		StashID:        uniqueTestChestService,
		ItemInstanceID: idStr(item.instanceID),
		StashItemID:    idStr(chestItemID),
		StashItems:     s.uniqueChestItemViews(state),
		Inventory:      s.inventoryView(),
		Equipped:       newSnapshotEquippedMap(s.equipped),
		Gold:           intPtr(s.gold),
		Hotbar:         s.hotbarView(),
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) uniqueChestItemViews(state *uniqueChestState) []StashItemView {
	if state == nil {
		return []StashItemView{}
	}
	out := make([]StashItemView, 0, len(state.items))
	for _, item := range state.items {
		out = append(out, s.stashItemView(item))
	}
	return out
}

func cloneUniqueChestItems(chests map[uint64]*uniqueChestState) map[uint64][]*stashItem {
	if chests == nil {
		return nil
	}
	out := make(map[uint64][]*stashItem, len(chests))
	for chestID, state := range chests {
		if state == nil {
			continue
		}
		items := make([]*stashItem, 0, len(state.items))
		for _, item := range state.items {
			if item == nil {
				continue
			}
			items = append(items, &stashItem{
				stashItemID: item.stashItemID,
				itemDefID:   item.itemDefID,
				rollPayload: cloneRollPayload(item.rollPayload),
			})
		}
		out[chestID] = items
	}
	return out
}

func restoreUniqueChestItems(items map[uint64][]*stashItem) map[uint64]*uniqueChestState {
	out := make(map[uint64]*uniqueChestState, len(items))
	for chestID, chestItems := range items {
		state := &uniqueChestState{items: make([]*stashItem, 0, len(chestItems))}
		for _, item := range chestItems {
			if item == nil {
				continue
			}
			state.items = append(state.items, &stashItem{
				stashItemID: item.stashItemID,
				itemDefID:   item.itemDefID,
				rollPayload: cloneRollPayload(item.rollPayload),
			})
		}
		out[chestID] = state
	}
	return out
}

func (s *Sim) seedUniqueTestChests() {
	if !s.gameplayDebug {
		return
	}

	items, ok := s.uniqueTestChestItems()
	if !ok {
		return
	}

	for _, level := range s.levels {
		for _, e := range level.entities {
			if e == nil || e.kind != interactableEntity {
				continue
			}

			if s.serviceForInteractable(e) != uniqueTestChestService {
				continue
			}

			state := s.uniqueChests[e.id]
			if state != nil && len(state.items) > 0 {
				continue
			}

			state = &uniqueChestState{items: make([]*stashItem, 0, len(items))}
			for _, item := range items {
				state.items = append(state.items, &stashItem{
					stashItemID: s.alloc(),
					itemDefID:   item.itemDefID,
					rollPayload: cloneRollPayload(item.rollPayload),
				})
			}

			s.uniqueChests[e.id] = state
		}
	}
}

func (r *Rules) namedUniqueCoveredEffectIDs() map[string]bool {
	covered := map[string]bool{}
	for _, uniqueID := range sortedStringKeys(r.UniqueItems) {
		unique := r.UniqueItems[uniqueID]
		if !unique.Enabled || unique.Status != "ready" {
			continue
		}
		for _, effectID := range unique.FixedEffectIDs {
			covered[effectID] = true
		}
	}

	return covered
}

func (r *Rules) uniqueTestChestCatalogSize() int {
	covered := r.namedUniqueCoveredEffectIDs()
	count := 0
	for _, effectID := range sortedStringKeys(r.UniqueEffects) {
		effect := r.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" || covered[effectID] {
			continue
		}
		count++
	}
	for _, unique := range r.UniqueItems {
		if unique.Enabled && unique.Status == "ready" {
			count++
		}
	}
	for range r.SetItems {
		count++
	}

	return count
}

func (s *Sim) uniqueTestChestItems() ([]*invItem, bool) {
	items := []*invItem{}
	covered := s.rules.namedUniqueCoveredEffectIDs()
	for _, effectID := range sortedStringKeys(s.rules.UniqueEffects) {
		effect := s.rules.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" || covered[effectID] {
			continue
		}
		templateID, ok := s.rules.rollUniqueChestTemplateForEffect(s.seed, effectID, effect)
		if !ok {
			return nil, false
		}
		payload, ok := s.rules.uniqueChestPayload(templateID, effectID)
		if !ok {
			return nil, false
		}
		items = append(items, &invItem{
			itemDefID:   payload.ItemTemplateID,
			rollPayload: cloneRollPayload(&payload),
		})
	}
	namedItems, ok := s.rules.namedUniqueChestItems()
	if !ok {
		return nil, false
	}
	items = append(items, namedItems...)
	setItems, ok := s.rules.setChestItems()
	if !ok {
		return nil, false
	}
	items = append(items, setItems...)
	return items, true
}

func (s *Sim) backfillUniqueTestChestState(state *uniqueChestState) (bool, bool) {
	if state == nil {
		return false, true
	}
	catalog, ok := s.uniqueTestChestItems()
	if !ok {
		return false, false
	}
	seen := map[string]bool{}
	for _, item := range state.items {
		if item == nil {
			continue
		}
		seen[uniqueChestCatalogKey(item.itemDefID, item.rollPayload)] = true
	}
	changed := false
	for _, item := range catalog {
		if item == nil {
			continue
		}
		key := uniqueChestCatalogKey(item.itemDefID, item.rollPayload)
		if seen[key] {
			continue
		}
		state.items = append(state.items, &stashItem{
			stashItemID: s.alloc(),
			itemDefID:   item.itemDefID,
			rollPayload: cloneRollPayload(item.rollPayload),
		})
		seen[key] = true
		changed = true
	}
	return changed, true
}

func uniqueChestCatalogKey(itemDefID string, payload *ItemRollPayload) string {
	if payload == nil {
		return itemDefID + "|base"
	}
	effects := strings.Join(payload.EffectIDs, ",")
	return itemDefID + "|" + payload.ItemTemplateID + "|" + payload.Rarity + "|" + payload.DisplayName + "|" + effects
}

func (r *Rules) namedUniqueChestItems() ([]*invItem, bool) {
	items := []*invItem{}
	for _, uniqueID := range sortedStringKeys(r.UniqueItems) {
		unique := r.UniqueItems[uniqueID]
		if !unique.Enabled || unique.Status != "ready" {
			continue
		}
		payload, ok := r.namedUniquePayload(uniqueID)
		if !ok {
			return nil, false
		}
		items = append(items, &invItem{
			itemDefID:   payload.ItemTemplateID,
			rollPayload: cloneRollPayload(&payload),
		})
	}
	return items, true
}

func (r *Rules) compatibleUniqueChestTemplateIDs(effect UniqueEffectDef) []string {
	out := []string{}
	for _, templateID := range sortedStringKeys(r.ItemTemplates) {
		template := r.ItemTemplates[templateID]
		if uniqueChestEffectCompatible(effect, template.ItemType) {
			out = append(out, templateID)
		}
	}

	return out
}

func (r *Rules) rollUniqueChestTemplateForEffect(sessionSeed, effectID string, effect UniqueEffectDef) (string, bool) {
	compatible := r.compatibleUniqueChestTemplateIDs(effect)
	if len(compatible) == 0 {
		return "", false
	}

	rng := NewRNG(SeedToUint64(sessionSeed + "|unique_chest|" + effectID))

	return compatible[rng.IntN(len(compatible))], true
}

func (r *Rules) uniqueChestPayload(templateID string, effectID string) (ItemRollPayload, bool) {
	template, ok := r.ItemTemplates[templateID]
	if !ok {
		return ItemRollPayload{}, false
	}
	effect, ok := r.UniqueEffects[effectID]
	if !ok || !effect.Enabled || effect.Status != "ready" || !uniqueChestEffectCompatible(effect, template.ItemType) {
		return ItemRollPayload{}, false
	}
	itemLevel := MaxItemLevelForDepth(maxInt(1, template.Requirements["level"]), r.DungeonGeneration.ItemLevelTiers)
	payload := ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    r.uniqueItemDisplayName(template, cloneIntMap(template.BaseStats), effect),
		Rarity:         "unique",
		ItemLevel:      1,
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      []string{effectID},
	}

	return FinalizeItemRollPayload(payload, itemLevel, r.DungeonGeneration.MonsterDepthScaling, r.DungeonGeneration.ItemLevelTiers), true
}

func (r *Rules) namedUniquePayload(uniqueID string) (ItemRollPayload, bool) {
	unique, ok := r.UniqueItems[uniqueID]
	if !ok || !unique.Enabled || unique.Status != "ready" {
		return ItemRollPayload{}, false
	}
	template, ok := r.ItemTemplates[unique.BaseTemplateID]
	if !ok {
		return ItemRollPayload{}, false
	}
	if len(unique.FixedEffectIDs) == 0 {
		return ItemRollPayload{}, false
	}
	for _, effectID := range unique.FixedEffectIDs {
		effect, ok := r.UniqueEffects[effectID]
		if !ok || !effect.Enabled || effect.Status != "ready" || !uniqueChestEffectCompatible(effect, template.ItemType) {
			return ItemRollPayload{}, false
		}
	}
	stats := cloneIntMap(template.BaseStats)
	for stat, value := range unique.FixedStats {
		stats[stat] = value
	}
	requirements := cloneIntMap(template.Requirements)
	if unique.MinimumLevel > requirements["level"] {
		requirements["level"] = unique.MinimumLevel
	}
	itemLevel := MaxItemLevelForDepth(maxInt(1, requirements["level"]), r.DungeonGeneration.ItemLevelTiers)
	primaryEffect := r.UniqueEffects[unique.FixedEffectIDs[0]]
	payload := ItemRollPayload{
		ItemTemplateID: unique.BaseTemplateID,
		DisplayName:    r.uniqueItemDisplayName(template, stats, primaryEffect),
		Rarity:         "unique",
		ItemLevel:      1,
		Stats:          stats,
		Requirements:   requirements,
		EffectIDs:      cloneStringSlice(unique.FixedEffectIDs),
		NamedUniqueID:  uniqueID,
	}

	return FinalizeItemRollPayload(payload, itemLevel, r.DungeonGeneration.MonsterDepthScaling, r.DungeonGeneration.ItemLevelTiers), true
}

func uniqueChestEffectCompatible(effect UniqueEffectDef, itemType string) bool {
	if itemType == "" {
		return false
	}
	for _, compatible := range effect.CompatibleItemTypes {
		if compatible == itemType {
			return true
		}
	}
	return false
}

func (r *Rules) uniqueItemDisplayName(template ItemTemplateDef, stats map[string]int, effect UniqueEffectDef) string {
	return r.rolledEquipmentDisplayName(template, "unique", stats, "of "+effect.DisplayName)
}
