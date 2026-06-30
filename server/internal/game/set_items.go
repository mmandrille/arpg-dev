package game

import "fmt"

type SetItemCatalogDef struct {
	ID           string            `json:"id"`
	Enabled      bool              `json:"enabled"`
	DisplayName  string            `json:"display_name"`
	Rarity       string            `json:"rarity"`
	Items        []SetItemPieceDef `json:"items"`
	PieceBonuses []SetItemBonusDef `json:"piece_bonuses"`
	FullSetBonus SetItemBonusDef   `json:"full_set_bonus"`
	Status       string            `json:"status"`
}

type SetItemPieceDef struct {
	ID             string         `json:"id"`
	BaseTemplateID string         `json:"base_template_id"`
	DisplayName    string         `json:"display_name"`
	MinimumLevel   int            `json:"minimum_level"`
	FixedStats     map[string]int `json:"fixed_stats"`
}

type SetItemBonusDef struct {
	RequiredPieces int            `json:"required_pieces"`
	Stats          map[string]int `json:"stats"`
}

type SetItemDef struct {
	SetID          string
	SetDisplayName string
	Piece          SetItemPieceDef
}

func (r *Rules) validateSetItemRules(sets map[string]SetItemCatalogDef) error {
	r.SetItems = map[string]SetItemDef{}
	for setID, set := range sets {
		if set.ID != setID {
			return fmt.Errorf("game: invalid rules set_items.%s.id: must match key", setID)
		}
		if set.Rarity != "set" {
			return fmt.Errorf("game: invalid rules set_items.%s.rarity: must be set", setID)
		}
		if set.Enabled && set.Status != "ready" {
			return fmt.Errorf("game: invalid rules set_items.%s.status: enabled entries must be ready", setID)
		}
		if !set.Enabled && set.Status != "disabled_seed" {
			return fmt.Errorf("game: invalid rules set_items.%s.status: disabled entries must remain disabled_seed", setID)
		}
		if !set.Enabled {
			continue
		}
		seenPieces := map[string]bool{}
		seenTemplates := map[string]bool{}
		for _, piece := range set.Items {
			if piece.ID == "" {
				return fmt.Errorf("game: invalid rules set_items.%s.items: id required", setID)
			}
			if seenPieces[piece.ID] {
				return fmt.Errorf("game: invalid rules set_items.%s.items.%s: duplicate piece", setID, piece.ID)
			}
			seenPieces[piece.ID] = true
			template, ok := r.ItemTemplates[piece.BaseTemplateID]
			if !ok {
				return fmt.Errorf("game: invalid rules set_items.%s.items.%s.base_template_id: unknown template %s", setID, piece.ID, piece.BaseTemplateID)
			}
			if seenTemplates[template.Slot] {
				return fmt.Errorf("game: invalid rules set_items.%s.items.%s: duplicate slot %s", setID, piece.ID, template.Slot)
			}
			seenTemplates[template.Slot] = true
			r.SetItems[piece.ID] = SetItemDef{SetID: setID, SetDisplayName: set.DisplayName, Piece: piece}
		}
		if set.FullSetBonus.RequiredPieces != len(set.Items) {
			return fmt.Errorf("game: invalid rules set_items.%s.full_set_bonus.required_pieces: must match item count", setID)
		}
		for _, bonus := range set.PieceBonuses {
			if bonus.RequiredPieces < 2 || bonus.RequiredPieces >= len(set.Items) {
				return fmt.Errorf("game: invalid rules set_items.%s.piece_bonuses.%d: invalid required pieces", setID, bonus.RequiredPieces)
			}
		}
	}
	return nil
}

func (r *Rules) setItemPayload(setItemID string) (ItemRollPayload, bool) {
	setItem, ok := r.SetItems[setItemID]
	if !ok {
		return ItemRollPayload{}, false
	}
	template, ok := r.ItemTemplates[setItem.Piece.BaseTemplateID]
	if !ok {
		return ItemRollPayload{}, false
	}
	stats := cloneIntMap(template.BaseStats)
	for stat, value := range setItem.Piece.FixedStats {
		stats[stat] = value
	}
	requirements := cloneIntMap(template.Requirements)
	if setItem.Piece.MinimumLevel > requirements["level"] {
		requirements["level"] = setItem.Piece.MinimumLevel
	}
	itemLevel := MaxItemLevelForDepth(maxInt(1, requirements["level"]), r.DungeonGeneration.ItemLevelTiers)
	payload := ItemRollPayload{
		ItemTemplateID: setItem.Piece.BaseTemplateID,
		DisplayName:    setItem.Piece.DisplayName,
		Rarity:         "set",
		ItemLevel:      1,
		Stats:          stats,
		Requirements:   requirements,
		EffectIDs:      []string{},
	}

	return FinalizeItemRollPayload(payload, itemLevel, r.DungeonGeneration.MonsterDepthScaling, r.DungeonGeneration.ItemLevelTiers), true
}

func (r *Rules) setChestItems() ([]*invItem, bool) {
	items := []*invItem{}
	for _, setItemID := range sortedStringKeys(r.SetItems) {
		payload, ok := r.setItemPayload(setItemID)
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

func (s *Sim) equippedSetBonusStats() map[string]int {
	counts := map[string]int{}
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		setItem, ok := s.setItemForEquippedItem(item)
		if !ok {
			continue
		}
		counts[setItem.SetID]++
	}
	stats := map[string]int{}
	for _, setID := range sortedStringKeys(counts) {
		set, ok := s.rules.SetCatalogs[setID]
		if !ok || !set.Enabled || set.Status != "ready" {
			continue
		}
		pieces := counts[setID]
		for _, bonus := range set.PieceBonuses {
			if pieces >= bonus.RequiredPieces {
				addStats(stats, bonus.Stats)
			}
		}
		if pieces >= set.FullSetBonus.RequiredPieces {
			addStats(stats, set.FullSetBonus.Stats)
		}
	}
	return stats
}

func (s *Sim) equippedSetPieceCounts() map[string]int {
	counts := map[string]int{}
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		setItem, ok := s.setItemForEquippedItem(item)
		if !ok {
			continue
		}
		counts[setItem.SetID]++
	}
	return counts
}

func (s *Sim) setItemSummaryLines(item *invItem) []string {
	setItem, ok := s.setItemForEquippedItem(item)
	if !ok {
		return nil
	}
	set, ok := s.rules.SetCatalogs[setItem.SetID]
	if !ok || !set.Enabled || set.Status != "ready" {
		return nil
	}
	equippedCount := s.equippedSetPieceCounts()[setItem.SetID]
	totalCount := len(set.Items)
	lines := []string{
		fmt.Sprintf("Set: %s (%d/%d equipped)", set.DisplayName, equippedCount, totalCount),
	}
	for _, bonus := range set.PieceBonuses {
		lines = append(lines, setBonusSummaryLine(bonus, equippedCount))
	}
	lines = append(lines, setBonusSummaryLine(set.FullSetBonus, equippedCount))
	return lines
}

func setBonusSummaryLine(bonus SetItemBonusDef, equippedCount int) string {
	state := "inactive"
	if equippedCount >= bonus.RequiredPieces {
		state = "active"
	}
	return fmt.Sprintf("%d-piece set bonus: %s (%s)", bonus.RequiredPieces, setBonusStatsSummary(bonus.Stats), state)
}

func setBonusStatsSummary(stats map[string]int) string {
	lines := statSummaryLines(stats)
	if len(lines) == 0 {
		return "None"
	}
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += ", "
		}
		out += line
	}
	return out
}

func (s *Sim) setItemForEquippedItem(item *invItem) (SetItemDef, bool) {
	if item == nil || item.rollPayload == nil || item.rollPayload.Rarity != "set" {
		return SetItemDef{}, false
	}
	for _, setItemID := range sortedStringKeys(s.rules.SetItems) {
		setItem := s.rules.SetItems[setItemID]
		if setItem.Piece.BaseTemplateID == item.rollPayload.ItemTemplateID && setItem.Piece.DisplayName == item.rollPayload.DisplayName {
			return setItem, true
		}
	}
	return SetItemDef{}, false
}

func (s *Sim) setItemForInventoryItem(item *invItem) (SetItemDef, bool) {
	if item == nil || item.rollPayload == nil || item.rollPayload.Rarity != "set" {
		return SetItemDef{}, false
	}
	for _, setItemID := range sortedStringKeys(s.rules.SetItems) {
		setItem := s.rules.SetItems[setItemID]
		if setItem.Piece.BaseTemplateID == item.rollPayload.ItemTemplateID && setItem.Piece.DisplayName == item.rollPayload.DisplayName {
			return setItem, true
		}
	}
	return SetItemDef{}, false
}

func (s *Sim) appendSetItemInventoryUpdates(res *TickResult) {
	if res == nil {
		return
	}
	for _, item := range s.inventory {
		if item == nil || s.hotbarHasItem(item.instanceID) {
			continue
		}
		if _, ok := s.setItemForInventoryItem(item); !ok {
			continue
		}
		res.Changes = append(res.Changes, Change{Op: OpInventoryUpdate, Item: ptrItemView(s.itemView(item))})
	}
}

func addStats(out map[string]int, in map[string]int) {
	for stat, value := range in {
		out[stat] += value
	}
}

func applySetCombatStats(
	setStats map[string]int,
	damageMin *float64,
	damageMax *float64,
	armor *float64,
	maxHP *float64,
	maxMana *float64,
	healthRegen *float64,
	manaRegen *float64,
	blockPercent *float64,
	itemSpeedPercent *float64,
	hitChancePercent *float64,
	critChancePercent *float64,
	evadeChancePercent *float64,
	magicFindPercent *float64,
	damageMinSources *[]StatBreakdownSourceView,
	damageMaxSources *[]StatBreakdownSourceView,
	armorSources *[]StatBreakdownSourceView,
	maxHPSources *[]StatBreakdownSourceView,
	maxManaSources *[]StatBreakdownSourceView,
	healthRegenSources *[]StatBreakdownSourceView,
	manaRegenSources *[]StatBreakdownSourceView,
	blockSources *[]StatBreakdownSourceView,
	attackSpeedSources *[]StatBreakdownSourceView,
	hitChanceSources *[]StatBreakdownSourceView,
	critChanceSources *[]StatBreakdownSourceView,
	evadeChanceSources *[]StatBreakdownSourceView,
	magicFindSources *[]StatBreakdownSourceView,
) {
	addSetFloat(setStats, "damage_min", damageMin, damageMinSources)
	addSetFloat(setStats, "damage_max", damageMax, damageMaxSources)
	addSetFloat(setStats, "armor", armor, armorSources)
	addSetFloat(setStats, "max_hp", maxHP, maxHPSources)
	addSetFloat(setStats, "max_mana", maxMana, maxManaSources)
	addSetRegen(setStats, "health_regen_per_10_seconds", healthRegen, healthRegenSources)
	addSetRegen(setStats, "mana_regen_per_10_seconds", manaRegen, manaRegenSources)
	addSetFloat(setStats, "block_percent", blockPercent, blockSources)
	addSetFloat(setStats, "attack_speed_percent", itemSpeedPercent, attackSpeedSources)
	addSetPercentFloat(setStats, "hit_chance", hitChancePercent, hitChanceSources)
	addSetPercentFloat(setStats, "crit_chance", critChancePercent, critChanceSources)
	addSetPercentFloat(setStats, "evade_chance", evadeChancePercent, evadeChanceSources)
	addSetFloat(setStats, "magic_find_percent", magicFindPercent, magicFindSources)
}

func addSetFloat(setStats map[string]int, stat string, target *float64, sources *[]StatBreakdownSourceView) {
	if value := setStats[stat]; value != 0 {
		*target += float64(value)
		*sources = append(*sources, StatBreakdownSourceView{Label: "Set bonus", Value: float64(value), Kind: "set_bonus"})
	}
}

func addSetRegen(setStats map[string]int, stat string, target *float64, sources *[]StatBreakdownSourceView) {
	if value := setStats[stat]; value != 0 {
		perSecond := float64(value) / 10.0
		*target += perSecond
		*sources = append(*sources, StatBreakdownSourceView{Label: "Set bonus", Value: perSecond, Kind: "set_bonus"})
	}
}

func addSetPercentFloat(setStats map[string]int, stat string, target *float64, sources *[]StatBreakdownSourceView) {
	if value := setStats[stat]; value != 0 {
		*target += float64(value)
		*sources = append(*sources, StatBreakdownSourceView{Label: "Set bonus", Value: float64(value) / 100.0, Kind: "set_bonus"})
	}
}
