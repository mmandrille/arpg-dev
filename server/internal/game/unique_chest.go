package game

import "fmt"

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
		DisplayName:    fmt.Sprintf("Unique %s", template.Name),
		Rarity:         "unique",
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      []string{effectID},
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
