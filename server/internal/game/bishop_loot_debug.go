package game

import (
	"fmt"
	"sort"
)

type BishopLootDepthOptionView struct {
	Depth         int    `json:"depth"`
	Label         string `json:"label"`
	MaxItemLevel  int    `json:"max_item_level"`
}

type BishopLootDepthCatalogView struct {
	MaxReachableDepth int                         `json:"max_reachable_depth"`
	Depths            []BishopLootDepthOptionView `json:"depths"`
}

type BishopLootEntryView struct {
	EntryIndex        int    `json:"entry_index"`
	Label             string `json:"label"`
	ItemDefID         string `json:"item_def_id,omitempty"`
	ItemTemplateID    string `json:"item_template_id,omitempty"`
	UniqueItemID      string `json:"unique_item_id,omitempty"`
	SetItemID         string `json:"set_item_id,omitempty"`
	Weight            int    `json:"weight"`
	SupportsItemLevel bool   `json:"supports_item_level"`
}

type BishopLootAttemptView struct {
	AttemptID     string                `json:"attempt_id"`
	SuccessWeight int                   `json:"success_weight"`
	NoDropWeight  int                   `json:"no_drop_weight"`
	Entries       []BishopLootEntryView `json:"entries"`
}

type BishopLootResourcePoolEntryView struct {
	ItemDefID         string `json:"item_def_id"`
	Label             string `json:"label"`
	Weight            int    `json:"weight"`
	SupportsItemLevel bool   `json:"supports_item_level"`
}

type BishopLootResourceBranchView struct {
	ChancePercent int                               `json:"chance_percent"`
	Pool          []BishopLootResourcePoolEntryView `json:"pool"`
}

type BishopLootWalletItemView struct {
	ItemDefID string `json:"item_def_id"`
	Label     string `json:"label"`
}

type BishopLootSourceCatalogView struct {
	Depth         int                           `json:"depth"`
	SourceType    string                        `json:"source_type"`
	MaxItemLevel  int                           `json:"max_item_level"`
	LootTableID   string                        `json:"loot_table_id"`
	Attempts      []BishopLootAttemptView       `json:"attempts"`
	ResourceLoot  *BishopLootResourceBranchView `json:"resource_loot,omitempty"`
	WalletItems   []BishopLootWalletItemView    `json:"wallet_items"`
}

type BishopDebugLootCatalogIntent struct {
	BishopEntityID string
}

type BishopDebugLootSourceCatalogIntent struct {
	BishopEntityID string
	Depth          int
	SourceType     string
}

type BishopDebugForceLootIntent struct {
	BishopEntityID string
	Depth          int
	SourceType     string
	DropKind       string
	AttemptID      string
	EntryIndex     int
	ItemDefID      string
	ItemLevel      int
}

func (s *Sim) bishopMaxReachableDepth() int {
	return maxInt(1, s.progression.DeepestDungeonDepth)
}

func (s *Sim) validateBishopLootDepth(depth int) (int, bool) {
	if depth < 1 {
		return 0, false
	}
	maxDepth := s.bishopMaxReachableDepth()
	if depth > maxDepth {
		return 0, false
	}

	return depth, true
}

func (s *Sim) handleBishopDebugLootCatalog(in Input, res *TickResult) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if in.BishopDebugLootCatalog == nil || in.BishopDebugLootCatalog.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(in.BishopDebugLootCatalog.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	catalog := s.buildBishopLootDepthCatalog()
	healed, restored := s.restorePlayerResources(player, res)
	res.Events = append(res.Events, Event{
		EventType:           "bishop_debug_loot_catalog",
		EntityID:            idStr(bishopEntity.id),
		CorrelationID:       in.CorrelationID,
		Service:             "bishop",
		Heal:                intPtr(healed),
		Mana:                intPtr(restored),
		BishopLootDepthCatalog: &catalog,
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) handleBishopDebugLootSourceCatalog(in Input, res *TickResult) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if in.BishopDebugLootSourceCatalog == nil || in.BishopDebugLootSourceCatalog.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	intent := in.BishopDebugLootSourceCatalog
	depth, ok := s.validateBishopLootDepth(intent.Depth)
	if !ok || intent.SourceType == "" {
		res.reject(in.MessageID, "invalid_depth")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(intent.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	catalog, ok := s.buildBishopLootSourceCatalog(depth, intent.SourceType)
	if !ok {
		res.reject(in.MessageID, "invalid_source")
		return
	}
	healed, restored := s.restorePlayerResources(player, res)
	res.Events = append(res.Events, Event{
		EventType:            "bishop_debug_loot_source_catalog",
		EntityID:             idStr(bishopEntity.id),
		CorrelationID:        in.CorrelationID,
		Service:              "bishop",
		Heal:                 intPtr(healed),
		Mana:                 intPtr(restored),
		BishopLootSourceCatalog: &catalog,
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) handleBishopDebugForceLoot(in Input, res *TickResult) {
	if !s.gameplayDebug {
		res.reject(in.MessageID, "debug_disabled")
		return
	}
	if in.BishopDebugForceLoot == nil || in.BishopDebugForceLoot.BishopEntityID == "" {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	intent := in.BishopDebugForceLoot
	depth, ok := s.validateBishopLootDepth(intent.Depth)
	if !ok || intent.SourceType == "" || intent.DropKind == "" {
		res.reject(in.MessageID, "invalid_depth")
		return
	}
	bishopEntity, ok, reason := s.resolveBishopIntentTarget(intent.BishopEntityID)
	if !ok {
		res.reject(in.MessageID, reason)
		return
	}
	player := s.activeLevel().entities[s.playerID]
	if player == nil || player.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	lootID, itemLevel, ok := s.forceBishopLootDrop(intent, depth, player.pos, s.targetInteractionRadius(player), in.CorrelationID, res)
	if !ok {
		res.reject(in.MessageID, "drop_failed")
		return
	}

	healed, restored := s.restorePlayerResources(player, res)
	res.Events = append(res.Events, Event{
		EventType:      "bishop_debug_loot_dropped",
		EntityID:       idStr(bishopEntity.id),
		TargetEntityID: idStr(lootID),
		CorrelationID:  in.CorrelationID,
		Service:        "bishop",
		Amount:         intPtr(itemLevel),
		Heal:           intPtr(healed),
		Mana:           intPtr(restored),
	})
	res.ack(in.MessageID)
	s.savePlayer(s.defaultPlayer())
}

func (s *Sim) buildBishopLootDepthCatalog() BishopLootDepthCatalogView {
	maxDepth := s.bishopMaxReachableDepth()
	tiers := s.rules.DungeonGeneration.ItemLevelTiers
	seen := map[int]bool{}
	out := BishopLootDepthCatalogView{MaxReachableDepth: maxDepth}
	for _, band := range s.rules.DungeonGeneration.LootBands {
		depth := band.MinDepth
		if depth < 1 || depth > maxDepth || seen[depth] {
			continue
		}
		seen[depth] = true
		out.Depths = append(out.Depths, BishopLootDepthOptionView{
			Depth:        depth,
			Label:        fmt.Sprintf("Depth %d", depth),
			MaxItemLevel: MaxItemLevelForDepth(depth, tiers),
		})
	}
	sort.Slice(out.Depths, func(i, j int) bool { return out.Depths[i].Depth < out.Depths[j].Depth })

	return out
}

func (s *Sim) buildBishopLootSourceCatalog(depth int, sourceType string) (BishopLootSourceCatalogView, bool) {
	lootTableID, hook, ok := s.bishopLootTableForSource(depth, sourceType)
	if !ok {
		return BishopLootSourceCatalogView{}, false
	}
	tiers := s.rules.DungeonGeneration.ItemLevelTiers
	catalog := BishopLootSourceCatalogView{
		Depth:        depth,
		SourceType:   sourceType,
		MaxItemLevel: MaxItemLevelForDepth(depth, tiers),
		LootTableID:  lootTableID,
		WalletItems:  s.bishopLootWalletItems(),
	}
	table, ok := s.rules.LootTables[lootTableID]
	if !ok || table.TreasureClassID == "" {
		return BishopLootSourceCatalogView{}, false
	}
	classDef, ok := s.rules.TreasureClasses[table.TreasureClassID]
	if !ok {
		return BishopLootSourceCatalogView{}, false
	}
	for _, attempt := range classDef.Attempts {
		view := BishopLootAttemptView{
			AttemptID:     attempt.AttemptID,
			SuccessWeight: attempt.SuccessWeight,
			NoDropWeight:  attempt.NoDropWeight,
		}
		for idx, entry := range attempt.Entries {
			view.Entries = append(view.Entries, BishopLootEntryView{
				EntryIndex:        idx,
				Label:             s.bishopLootEntryLabel(entry),
				ItemDefID:         entry.ItemDefID,
				ItemTemplateID:    entry.ItemTemplateID,
				UniqueItemID:      entry.UniqueItemID,
				SetItemID:         entry.SetItemID,
				Weight:            entry.Weight,
				SupportsItemLevel: s.bishopLootEntrySupportsItemLevel(entry),
			})
		}
		catalog.Attempts = append(catalog.Attempts, view)
	}
	if chance := s.rules.resourceLootDropChancePercent(hook); chance > 0 {
		branch := &BishopLootResourceBranchView{ChancePercent: chance}
		for _, entry := range s.rules.MainConfig.Gameplay.ResourceLootDrops.Pool {
			if entry.ItemDefID == "" || entry.Weight <= 0 {
				continue
			}
			branch.Pool = append(branch.Pool, BishopLootResourcePoolEntryView{
				ItemDefID:         entry.ItemDefID,
				Label:             s.bishopLootItemLabel(entry.ItemDefID),
				Weight:            entry.Weight,
				SupportsItemLevel: entry.ItemDefID == UpgradeShardItemDefID || entry.ItemDefID == RenewStoneItemDefID,
			})
		}
		catalog.ResourceLoot = branch
	}

	return catalog, true
}

func (s *Sim) bishopLootTableForSource(depth int, sourceType string) (string, resourceLootDropHook, bool) {
	switch sourceType {
	case "monster":
		band, ok := s.rules.DungeonGeneration.LootBandForDepth(depth)
		if !ok || band.MonsterLootTable == "" {
			return "", 0, false
		}

		return band.MonsterLootTable, resourceLootMonsterCommonRare, true
	case "chest":
		band, ok := s.rules.DungeonGeneration.LootBandForDepth(depth)
		if !ok || band.ChestLootTable == "" {
			return "", 0, false
		}

		return band.ChestLootTable, resourceLootChestRegular, true
	case "boss":
		if len(s.rules.DungeonGeneration.BossFloor.BossTemplatePool) == 0 {
			return "", 0, false
		}
		templateID := s.rules.DungeonGeneration.BossFloor.BossTemplatePool[0]
		template, ok := s.rules.BossTemplates[templateID]
		if !ok || template.LootTable == "" {
			return "", 0, false
		}

		return template.LootTable, resourceLootBossKill, true
	case "boss_chest":
		tableID := s.rules.DungeonGeneration.BossFloor.ChestLootTable
		if tableID == "" {
			return "", 0, false
		}

		return tableID, resourceLootChestBoss, true
	default:
		return "", 0, false
	}
}

func (s *Sim) bishopLootWalletItems() []BishopLootWalletItemView {
	seen := map[string]bool{}
	out := []BishopLootWalletItemView{}
	for _, rule := range s.rules.MainConfig.Gameplay.BadgeRewardRules {
		if rule.ResourceItemDefID == "" || seen[rule.ResourceItemDefID] {
			continue
		}
		if !s.rules.isBadgeRewardResourceItem(rule.ResourceItemDefID) {
			continue
		}
		seen[rule.ResourceItemDefID] = true
		out = append(out, BishopLootWalletItemView{
			ItemDefID: rule.ResourceItemDefID,
			Label:     s.bishopLootItemLabel(rule.ResourceItemDefID),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ItemDefID < out[j].ItemDefID })

	return out
}

func (s *Sim) bishopLootEntryLabel(entry TreasureClassEntry) string {
	switch {
	case entry.ItemDefID != "":
		return s.bishopLootItemLabel(entry.ItemDefID)
	case entry.ItemTemplateID != "":
		if template, ok := s.rules.ItemTemplates[entry.ItemTemplateID]; ok && template.Name != "" {
			return template.Name
		}

		return entry.ItemTemplateID
	case entry.UniqueItemID != "":
		return entry.UniqueItemID
	case entry.SetItemID != "":
		return entry.SetItemID
	default:
		return "unknown"
	}
}

func (s *Sim) bishopLootItemLabel(itemDefID string) string {
	if def, ok := s.rules.Items[itemDefID]; ok && def.Name != "" {
		return def.Name
	}

	return itemDefID
}

func (s *Sim) bishopLootEntrySupportsItemLevel(entry TreasureClassEntry) bool {
	if entry.ItemTemplateID != "" {
		return true
	}

	return false
}

func (s *Sim) forceBishopLootDrop(intent *BishopDebugForceLootIntent, depth int, sourcePos Vec2, sourceRadius float64, corr string, res *TickResult) (uint64, int, bool) {
	itemLevel := s.clampBishopForcedItemLevel(intent.ItemLevel, depth)
	switch intent.DropKind {
	case "wallet_item":
		if !s.rules.isBadgeRewardResourceItem(intent.ItemDefID) {
			return 0, 0, false
		}
		lootID, ok := s.spawnWalletBadgeLoot(intent.ItemDefID, sourcePos, sourceRadius, corr, res)
		if !ok {
			return 0, 0, false
		}

		return lootID, 1, true
	case "resource_pool":
		if intent.ItemDefID != UpgradeShardItemDefID && intent.ItemDefID != RenewStoneItemDefID {
			return 0, 0, false
		}
		lootID, level, ok := s.spawnResourceLootAtLevel(intent.ItemDefID, sourcePos, sourceRadius, depth, itemLevel, corr, res)
		if !ok {
			return 0, 0, false
		}

		return lootID, level, true
	case "treasure_entry":
		catalog, ok := s.buildBishopLootSourceCatalog(depth, intent.SourceType)
		if !ok {
			return 0, 0, false
		}
		drop, ok := s.bishopLootDropFromCatalog(catalog, intent.AttemptID, intent.EntryIndex)
		if !ok {
			return 0, 0, false
		}
		goldCtx := s.bishopGoldRollContext(depth, intent.SourceType, itemLevel)
		s.spawnLootDrops([]LootDrop{drop}, sourcePos, sourceRadius, corr, res, goldCtx)
		lootID := s.latestSpawnedLootID(res)
		if lootID == 0 {
			return 0, 0, false
		}

		return lootID, itemLevel, true
	default:
		return 0, 0, false
	}
}

func (s *Sim) bishopLootDropFromCatalog(catalog BishopLootSourceCatalogView, attemptID string, entryIndex int) (LootDrop, bool) {
	for _, attempt := range catalog.Attempts {
		if attempt.AttemptID != attemptID {
			continue
		}
		for _, entry := range attempt.Entries {
			if entry.EntryIndex != entryIndex {
				continue
			}

			return LootDrop{
				ItemDefID:      entry.ItemDefID,
				ItemTemplateID: entry.ItemTemplateID,
				UniqueItemID:   entry.UniqueItemID,
				SetItemID:      entry.SetItemID,
			}, true
		}
	}

	return LootDrop{}, false
}

func (s *Sim) bishopGoldRollContext(depth int, sourceType string, forcedItemLevel int) goldRollContext {
	ctx := goldRollContext{
		levelNum:          -depth,
		forcedItemLevel:   forcedItemLevel,
	}
	if sourceType == "boss" {
		ctx.magicFind = true
		ctx.magicFindBonusPercent = s.rules.MainConfig.Gameplay.BossLootMagicFindBonusPercent
	}

	return ctx
}

func (s *Sim) clampBishopForcedItemLevel(itemLevel, depth int) int {
	maxLevel := MaxItemLevelForDepth(depth, s.rules.DungeonGeneration.ItemLevelTiers)
	if itemLevel < 1 {
		itemLevel = 1
	}
	if itemLevel > maxLevel {
		itemLevel = maxLevel
	}

	return itemLevel
}

func (s *Sim) latestSpawnedLootID(res *TickResult) uint64 {
	for i := len(res.Events) - 1; i >= 0; i-- {
		if res.Events[i].EventType != "loot_dropped" || res.Events[i].EntityID == "" {
			continue
		}
		if id, ok := ParseEntityID(res.Events[i].EntityID); ok {
			return id
		}
	}

	return 0
}
