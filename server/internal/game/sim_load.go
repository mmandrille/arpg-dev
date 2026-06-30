package game

import (
	"encoding/json"
	"sort"
	"strconv"
)

// PersistedItem is a durable inventory item reloaded on session resume.
type PersistedItem struct {
	InstanceID  string
	ItemDefID   string
	Slot        string
	Equipped    bool
	WeaponSet   int
	RolledStats json.RawMessage
}

// PersistedHotbarSlot is a durable hotbar assignment reloaded on session resume.
type PersistedHotbarSlot struct {
	SlotIndex      int
	ItemInstanceID *string
}

// PersistedSkillBindings is the durable skill control layout reloaded on resume.
type PersistedSkillBindings struct {
	FunctionKeys      []string
	RightClickSkillID string
}

// PersistedStashItem is an account-stash item reloaded at session start.
type PersistedStashItem struct {
	StashItemID string
	ItemDefID   string
	RolledStats json.RawMessage
}

// PersistedResourceAmount is an account resource balance reloaded at session start.
type PersistedResourceAmount struct {
	ResourceID string
	Amount     int
}

// LoadInventory restores persisted inventory into a fresh sim (used on resume).
// The entity counter is advanced past any reloaded instance id so newly
// allocated ids never collide with reloaded ones.
func (s *Sim) LoadInventory(items []PersistedItem) {
	s.ensureWeaponSets()
	for _, p := range items {
		id, err := strconv.ParseUint(p.InstanceID, 10, 64)
		if err != nil {
			continue
		}
		it := &invItem{instanceID: id, itemDefID: p.ItemDefID, slot: p.Slot, equipped: p.Equipped, rollPayload: parseRollPayload(p.RolledStats)}
		s.inventory = append(s.inventory, it)
		if p.Equipped && p.Slot != "" {
			s.setEquippedSlot(p.Slot, id, normalizeWeaponSetIndex(p.WeaponSet))
		}
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
	s.syncActiveWeaponSetToEquipped()
	s.syncActivePlayerResourceCaps(nil)
	s.savePlayer(s.defaultPlayer())
}

// LoadHotbar restores fixed hotbar assignments into a fresh sim.
func (s *Sim) LoadHotbar(slots []PersistedHotbarSlot) {
	if len(s.hotbar) != 10 {
		s.hotbar = make([]uint64, 10)
	}
	for _, slot := range slots {
		if slot.SlotIndex < 0 || slot.SlotIndex >= len(s.hotbar) {
			continue
		}
		if slot.ItemInstanceID == nil || *slot.ItemInstanceID == "" {
			s.hotbar[slot.SlotIndex] = 0
			continue
		}
		id, err := strconv.ParseUint(*slot.ItemInstanceID, 10, 64)
		if err != nil {
			continue
		}
		s.hotbar[slot.SlotIndex] = id
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) LoadSkillBindings(bindings PersistedSkillBindings) {
	s.skillFunctionKeys = normalizeSkillFunctionKeys(bindings.FunctionKeys)
	s.rightClickSkillID = bindings.RightClickSkillID
	s.savePlayer(s.defaultPlayer())
}

// LoadShopStock restores durable generated shop stock into a fresh sim. Buyback
// rows are session-local and are intentionally not loaded here.
func (s *Sim) LoadShopStock(items []PersistedShopStockItem) {
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	for _, p := range items {
		if p.ShopID == "" || p.OfferID == "" || p.ItemTemplateID == "" {
			continue
		}
		payload := parseRollPayload(p.RolledPayload)
		if payload == nil {
			continue
		}
		state := s.shopStock[p.ShopID]
		if state == nil {
			state = &shopStockState{RefreshKey: p.RefreshKey}
			s.shopStock[p.ShopID] = state
		}
		if state.RefreshKey == "" {
			state.RefreshKey = p.RefreshKey
		}
		state.Generated = append(state.Generated, &shopStockItem{
			OfferIndex:     p.OfferIndex,
			OfferID:        p.OfferID,
			SourceDepth:    p.SourceDepth,
			ItemTemplateID: p.ItemTemplateID,
			Payload:        *payload,
			BuyPrice:       p.BuyPrice,
			Available:      p.Available,
		})
	}
	for _, shopID := range sortedStringKeys(s.shopStock) {
		state := s.shopStock[shopID]
		sort.Slice(state.Generated, func(i, j int) bool {
			if state.Generated[i].OfferIndex != state.Generated[j].OfferIndex {
				return state.Generated[i].OfferIndex < state.Generated[j].OfferIndex
			}
			return state.Generated[i].OfferID < state.Generated[j].OfferID
		})
	}
	s.savePlayer(s.defaultPlayer())
}

// LoadAccountStash restores account-owned stash contents into the active
// player's private state.
func (s *Sim) LoadAccountStash(items []PersistedStashItem, gold int, capacity int) {
	if capacity <= 0 {
		capacity = defaultStashCapacity
	}
	s.stashItems = []*stashItem{}
	for _, p := range items {
		id, err := strconv.ParseUint(p.StashItemID, 10, 64)
		if err != nil || p.ItemDefID == "" {
			continue
		}
		s.stashItems = append(s.stashItems, &stashItem{
			stashItemID: id,
			itemDefID:   p.ItemDefID,
			rollPayload: parseRollPayload(p.RolledStats),
		})
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
	sort.Slice(s.stashItems, func(i, j int) bool {
		return s.stashItems[i].stashItemID < s.stashItems[j].stashItemID
	})
	if gold < 0 {
		gold = 0
	}
	s.stashGold = gold
	s.stashCapacity = capacity
	s.savePlayer(s.defaultPlayer())
}

// LoadResourceWallet restores account-owned resource balances into the active
// player's private state.
func (s *Sim) LoadResourceWallet(resources []PersistedResourceAmount) {
	s.resourceWallet = make(map[string]int)
	for _, resource := range resources {
		if resource.ResourceID == "" || resource.Amount <= 0 {
			continue
		}
		s.resourceWallet[resource.ResourceID] += resource.Amount
	}
	s.savePlayer(s.defaultPlayer())
}

func parseRollPayload(raw json.RawMessage) *ItemRollPayload {
	if len(raw) == 0 || string(raw) == "{}" {
		return nil
	}
	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.ItemTemplateID == "" {
		return nil
	}
	if payload.ItemLevel < 1 {
		payload.ItemLevel = 1
	}
	payload.Stats = cloneIntMap(payload.Stats)
	payload.Requirements = cloneIntMap(payload.Requirements)
	payload.EffectIDs = cloneStringSlice(payload.EffectIDs)
	return &payload
}

func cloneRollPayload(in *ItemRollPayload) *ItemRollPayload {
	if in == nil {
		return nil
	}
	out := &ItemRollPayload{
		ItemTemplateID: in.ItemTemplateID,
		DisplayName:    in.DisplayName,
		Rarity:         in.Rarity,
		ItemLevel:      in.ItemLevel,
		Stats:          cloneIntMap(in.Stats),
		Requirements:   cloneIntMap(in.Requirements),
		EffectIDs:      cloneStringSlice(in.EffectIDs),
	}
	if len(in.ClassAffinities) > 0 {
		out.ClassAffinities = make([]ClassAffinityRoll, len(in.ClassAffinities))
		copy(out.ClassAffinities, in.ClassAffinities)
	}

	return out
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for key, value := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = value
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneVec2Ptr(in *Vec2) *Vec2 {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func sortedUniqueStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	for _, value := range in {
		if value != "" {
			seen[value] = true
		}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func removeStringValue(in []string, value string) []string {
	if value == "" || len(in) == 0 {
		return cloneStringSlice(in)
	}
	out := []string{}
	for _, current := range in {
		if current != value {
			out = append(out, current)
		}
	}
	return out
}

func containsStringValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func normalizeSkillFunctionKeys(in []string) []string {
	out := make([]string, skillFunctionKeyCount)
	copy(out, in)
	return out
}

// LoadDiscoveredTeleporters restores durable character waypoint unlocks into a
// fresh session. Town remains discovered even if callers omit it.
func (s *Sim) LoadDiscoveredTeleporters(levels []int) {
	if !s.multiLevel {
		return
	}
	s.discoveredTeleporters[townLevel] = true
	for _, level := range levels {
		if s.levelHasTeleporter(level) {
			s.discoveredTeleporters[level] = true
		}
	}
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) LoadInventoryForPlayer(playerID uint64, items []PersistedItem) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadInventory(items)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadHotbarForPlayer(playerID uint64, slots []PersistedHotbarSlot) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadHotbar(slots)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadSkillBindingsForPlayer(playerID uint64, bindings PersistedSkillBindings) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadSkillBindings(bindings)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadShopStockForPlayer(playerID uint64, items []PersistedShopStockItem) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadShopStock(items)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadAccountStashForPlayer(playerID uint64, items []PersistedStashItem, gold int, capacity int) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadAccountStash(items, gold, capacity)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadResourceWalletForPlayer(playerID uint64, resources []PersistedResourceAmount) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadResourceWallet(resources)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}

func (s *Sim) LoadDiscoveredTeleportersForPlayer(playerID uint64, levels []int) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	s.usePlayer(ps)
	s.LoadDiscoveredTeleporters(levels)
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
}
