package game

import "fmt"

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
	if e.state == interactableOpen {
		res.reject(in.MessageID, "already_open")
		return
	}
	if e.state != interactableReady {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	items, ok := s.uniqueTestChestItems()
	if !ok {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	e.state = interactableOpen
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(e))})
	for _, item := range items {
		item.instanceID = s.alloc()
		item.slot = s.itemSlot(item.itemDefID, item.rollPayload)
		s.inventory = append(s.inventory, item)
		res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	}
	res.Events = append(res.Events, Event{
		EventType:     "interactable_activated",
		EntityID:      idStr(e.id),
		CorrelationID: in.CorrelationID,
		Service:       uniqueTestChestService,
		Amount:        intPtr(len(items)),
	})
	if ack {
		res.ack(in.MessageID)
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) uniqueTestChestItems() ([]*invItem, bool) {
	items := []*invItem{}
	for _, effectID := range sortedStringKeys(s.rules.UniqueEffects) {
		effect := s.rules.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" {
			continue
		}
		templateID, ok := s.rules.firstUniqueChestTemplate(effect)
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
	return items, true
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

func (r *Rules) firstUniqueChestTemplate(effect UniqueEffectDef) (string, bool) {
	for _, templateID := range sortedStringKeys(r.ItemTemplates) {
		template := r.ItemTemplates[templateID]
		if uniqueChestEffectCompatible(effect, template.ItemType) {
			return templateID, true
		}
	}
	return "", false
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
	return ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    "Unique " + template.Name,
		Rarity:         "unique",
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      []string{effectID},
	}, true
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
	return ItemRollPayload{
		ItemTemplateID: unique.BaseTemplateID,
		DisplayName:    unique.DisplayName,
		Rarity:         "unique",
		Stats:          stats,
		Requirements:   requirements,
		EffectIDs:      cloneStringSlice(unique.FixedEffectIDs),
	}, true
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
