package game

import "strconv"

type corpseItem struct {
	instanceID  uint64
	itemDefID   string
	slot        string
	equipped    bool
	rollPayload *ItemRollPayload
}

type corpseState struct {
	entityID    uint64
	accountID   string
	characterID string
	name        string
	level       int
	deathLevel  int
	items       []*corpseItem
}

// PersistedCorpse is a same-account dead character body reloaded at session start.
type PersistedCorpse struct {
	AccountID   string
	CharacterID string
	Name        string
	Level       int
	DeathLevel  int
	Items       []PersistedItem
}

// LoadCharacterCorpses restores recoverable same-account dead-character bodies.
func (s *Sim) LoadCharacterCorpses(corpses []PersistedCorpse) {
	if s.corpses == nil {
		s.corpses = map[string]*corpseState{}
	}
	for _, p := range corpses {
		if p.CharacterID == "" || p.DeathLevel >= 0 || len(p.Items) == 0 {
			continue
		}
		state := &corpseState{
			accountID:   p.AccountID,
			characterID: p.CharacterID,
			name:        p.Name,
			level:       maxInt(1, p.Level),
			deathLevel:  p.DeathLevel,
			items:       make([]*corpseItem, 0, len(p.Items)),
		}
		for _, item := range p.Items {
			id, err := strconv.ParseUint(item.InstanceID, 10, 64)
			if err != nil || item.ItemDefID == "" {
				continue
			}
			state.items = append(state.items, &corpseItem{
				instanceID:  id,
				itemDefID:   item.ItemDefID,
				slot:        item.Slot,
				equipped:    item.Equipped,
				rollPayload: parseRollPayload(item.RolledStats),
			})
			if id >= s.nextID {
				s.nextID = id + 1
			}
		}
		if len(state.items) == 0 {
			continue
		}
		s.corpses[state.characterID] = state
		if level := s.levels[state.deathLevel]; level != nil {
			s.spawnCorpseOnLevel(level, state)
		}
	}
}

func (s *Sim) spawnCorpsesOnLevel(level *LevelState) {
	if level == nil || len(s.corpses) == 0 {
		return
	}
	for _, characterID := range sortedStringKeys(s.corpses) {
		corpse := s.corpses[characterID]
		if corpse == nil || corpse.deathLevel != level.levelNum || corpse.entityID != 0 || len(corpse.items) == 0 {
			continue
		}
		s.spawnCorpseOnLevel(level, corpse)
	}
}

func (s *Sim) spawnCorpseOnLevel(level *LevelState, corpse *corpseState) {
	if level == nil || corpse == nil || corpse.entityID != 0 || len(corpse.items) == 0 {
		return
	}
	e := &entity{
		kind:              interactableEntity,
		pos:               s.corpsePosition(level, corpse),
		interactableDefID: heroCorpseDefID,
		state:             interactableReady,
		corpseCharacterID: corpse.characterID,
		corpseName:        corpse.name,
		corpseLevel:       corpse.level,
		corpseItemCount:   len(corpse.items),
	}
	e.id = s.alloc()
	corpse.entityID = e.id
	level.entities[e.id] = e
}

func (s *Sim) corpsePosition(level *LevelState, corpse *corpseState) Vec2 {
	nav := level.nav
	if nav == nil || nav.GridBounds.MaxX <= nav.GridBounds.MinX || nav.GridBounds.MaxY <= nav.GridBounds.MinY {
		return Vec2{X: 2, Y: 2}
	}
	spanX := nav.GridBounds.MaxX - nav.GridBounds.MinX
	spanY := nav.GridBounds.MaxY - nav.GridBounds.MinY
	seed := 0
	for _, r := range corpse.characterID {
		seed += int(r)
	}
	seed += absInt(corpse.deathLevel) * 17
	total := spanX * spanY
	for i := 0; i < total; i++ {
		n := (seed + i*7) % total
		pos := Vec2{X: float64(nav.GridBounds.MinX + n%spanX), Y: float64(nav.GridBounds.MinY + n/spanX)}
		if !s.positionBlockedOnLevel(level, pos, interactableInteractionRadius) {
			return pos
		}
	}
	return Vec2{X: 2, Y: 2}
}

func (s *Sim) positionBlockedOnLevel(level *LevelState, pos Vec2, radius float64) bool {
	for _, wall := range level.walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, radius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && distance(pos, e.pos) < radius+s.targetInteractionRadius(e) {
			return true
		}
	}
	return false
}

func (s *Sim) openCorpse(e *entity, in Input, res *TickResult, ack bool) {
	if e.state != interactableReady || e.corpseCharacterID == "" {
		res.reject(in.MessageID, "not_actionable")
		return
	}
	corpse := s.corpses[e.corpseCharacterID]
	if corpse == nil || len(corpse.items) == 0 {
		res.reject(in.MessageID, "empty_corpse")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:          "corpse_opened",
		EntityID:           idStr(e.id),
		CorrelationID:      in.CorrelationID,
		CorpseCharacterID:  corpse.characterID,
		CorpseName:         corpse.name,
		CorpseItems:        s.corpseItemViews(corpse),
		Inventory:          s.inventoryView(),
		Equipped:           newSnapshotEquippedMap(s.equipped),
		Gold:               intPtr(s.gold),
		Hotbar:             s.hotbarView(),
		InventoryRows:      intPtr(s.inventoryRows()),
		InventoryCapacity:  intPtr(s.inventoryCapacity()),
		HotbarCapacity:     intPtr(s.hotbarCapacity()),
		CharacterClass:     s.progression.CharacterClass,
		CharacterLevel:     intPtr(s.progression.Level),
		CharacterXP:        intPtr(s.progression.Experience),
		UnspentStatPoints:  intPtr(s.progression.UnspentStatPoints),
		UnspentSkillPoints: intPtr(s.progression.UnspentSkillPoints),
	})
	if ack {
		res.ack(in.MessageID)
	}
}

func (s *Sim) handleCorpseWithdrawItem(in Input, res *TickResult) {
	if in.CorpseWithdrawItem == nil || in.CorpseWithdrawItem.CorpseEntityID == "" || in.CorpseWithdrawItem.ItemInstanceID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	entity, _, ok := s.findEntityAnyLevel(in.CorpseWithdrawItem.CorpseEntityID)
	if !ok || entity == nil || entity.kind != interactableEntity || entity.interactableDefID != heroCorpseDefID {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	if !s.inDispatchRange(entity) {
		res.reject(in.MessageID, "out_of_range")
		return
	}
	corpse := s.corpses[entity.corpseCharacterID]
	if corpse == nil {
		res.reject(in.MessageID, "invalid_target")
		return
	}
	itemID, err := strconv.ParseUint(in.CorpseWithdrawItem.ItemInstanceID, 10, 64)
	if err != nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	idx := -1
	var item *corpseItem
	for i, candidate := range corpse.items {
		if candidate.instanceID == itemID {
			idx = i
			item = candidate
			break
		}
	}
	if item == nil {
		res.reject(in.MessageID, "not_found")
		return
	}
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		res.reject(in.MessageID, "inventory_full")
		return
	}
	newID := s.alloc()
	owned := &invItem{instanceID: newID, itemDefID: item.itemDefID, rollPayload: item.rollPayload}
	s.inventory = append(s.inventory, owned)
	corpse.items = append(corpse.items[:idx], corpse.items[idx+1:]...)
	entity.corpseItemCount = len(corpse.items)
	transferID := "corpse_recover_item:" + idStr(itemID)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(owned)), StashTransferID: transferID})
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(entity))})
	res.Events = append(res.Events, Event{
		EventType:         "corpse_item_recovered",
		EntityID:          idStr(entity.id),
		CorrelationID:     in.CorrelationID,
		CorpseCharacterID: corpse.characterID,
		CorpseName:        corpse.name,
		ItemInstanceID:    idStr(newID),
		StashItemID:       idStr(itemID),
		Item:              ptrItemView(s.itemView(owned)),
		CorpseItems:       s.corpseItemViews(corpse),
		Inventory:         s.inventoryView(),
		Equipped:          newSnapshotEquippedMap(s.equipped),
		Gold:              intPtr(s.gold),
		Hotbar:            s.hotbarView(),
	})
	if len(corpse.items) == 0 {
		delete(s.corpses, corpse.characterID)
		delete(s.activeLevel().entities, entity.id)
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(entity.id)})
	}
	res.ack(in.MessageID)
}

func (s *Sim) inventoryView() []ItemView {
	out := make([]ItemView, 0, len(s.inventory))
	for _, item := range s.inventory {
		if item == nil || s.hotbarHasItem(item.instanceID) {
			continue
		}
		out = append(out, s.itemView(item))
	}
	return out
}

func (s *Sim) corpseItemViews(corpse *corpseState) []ItemView {
	if corpse == nil {
		return []ItemView{}
	}
	out := make([]ItemView, 0, len(corpse.items))
	for _, item := range corpse.items {
		out = append(out, s.itemView(&invItem{
			instanceID:  item.instanceID,
			itemDefID:   item.itemDefID,
			slot:        item.slot,
			equipped:    item.equipped,
			rollPayload: item.rollPayload,
		}))
	}
	return out
}
