package game

import (
	"encoding/json"
	"fmt"
)

const UpgradeShardItemDefID = "upgrade_shard"

const defaultAppraisalShopID = "town_vendor"

// UpgradeShardMinLevel returns the minimum leveled upgrade-shard tier required to
// upgrade an item currently at currentLevel. The shard must be the same tier or higher;
// consumable tiers floor at 1.
func UpgradeShardMinLevel(currentLevel int) int {
	if currentLevel < 1 {
		return 1
	}

	return currentLevel
}

// NewUpgradeShardRollPayload builds the durable loot/inventory payload for a leveled shard.
func NewUpgradeShardRollPayload(level int) *ItemRollPayload {
	if level < 1 {
		level = 1
	}

	return &ItemRollPayload{
		ItemTemplateID: UpgradeShardItemDefID,
		DisplayName:    "Upgrade Shard",
		ItemLevel:      level,
		Stats:          map[string]int{"item_level": level},
	}
}

// UpgradeShardLevelFromRaw reads shard level from rolled_stats JSON.
func UpgradeShardLevelFromRaw(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "{}" {
		return 1, nil
	}

	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err == nil && payload.ItemTemplateID == UpgradeShardItemDefID {
		if payload.ItemLevel > 0 {
			return payload.ItemLevel, nil
		}
		if level := payload.Stats["item_level"]; level > 0 {
			return level, nil
		}
	}

	var flat struct {
		ItemLevel int `json:"item_level"`
	}
	if err := json.Unmarshal(raw, &flat); err != nil {
		return 0, fmt.Errorf("game: decode upgrade shard level: %w", err)
	}
	if flat.ItemLevel < 1 {
		return 1, nil
	}

	return flat.ItemLevel, nil
}

// MarshalUpgradeShardRolledStats returns persisted stash/inventory JSON for a shard level.
func MarshalUpgradeShardRolledStats(level int) (json.RawMessage, error) {
	if level < 1 {
		level = 1
	}
	raw, err := json.Marshal(map[string]int{"item_level": level})
	if err != nil {
		return nil, fmt.Errorf("game: marshal upgrade shard stats: %w", err)
	}

	return raw, nil
}

// ItemSellPrice returns shop appraisal sell price for an item row.
func ItemSellPrice(rules *Rules, shopID, itemDefID string, rolled json.RawMessage) (int, bool) {
	if rules == nil || itemDefID == "" {
		return 0, false
	}
	shop, ok := rules.Shops[shopID]
	if !ok {
		return 0, false
	}

	payload := parseRollPayload(rolled)
	if payload != nil {
		buyPrice, ok := shop.generatedBuyPrice(payload.ItemTemplateID, payload.Rarity, payload.Stats, rules)
		if !ok {
			return 0, false
		}

		return shop.sellPrice(buyPrice), true
	}

	buyPrice, ok := shop.fixedBuyPrice(itemDefID)
	if !ok {
		return 0, false
	}

	return shop.sellPrice(buyPrice), true
}

// DefaultItemSellPrice uses the town vendor shop for blacksmith appraisal.
func DefaultItemSellPrice(rules *Rules, itemDefID string, rolled json.RawMessage) (int, bool) {
	payload := parseRollPayload(rolled)
	if payload == nil {
		payload = inferRollPayloadFromFlatStats(rules, itemDefID, rolled)
	}
	if payload != nil {
		return itemSellPriceFromPayload(rules, defaultAppraisalShopID, payload)
	}

	return ItemSellPrice(rules, defaultAppraisalShopID, itemDefID, rolled)
}

func itemSellPriceFromPayload(rules *Rules, shopID string, payload *ItemRollPayload) (int, bool) {
	if rules == nil || payload == nil {
		return 0, false
	}
	shop, ok := rules.Shops[shopID]
	if !ok {
		return 0, false
	}
	buyPrice, ok := shop.generatedBuyPrice(payload.ItemTemplateID, payload.Rarity, payload.Stats, rules)
	if !ok {
		return 0, false
	}

	return shop.sellPrice(buyPrice), true
}

func inferRollPayloadFromFlatStats(rules *Rules, itemDefID string, raw json.RawMessage) *ItemRollPayload {
	if rules == nil || itemDefID == "" || len(raw) == 0 {
		return nil
	}
	template, ok := rules.ItemTemplates[itemDefID]
	if !ok {
		return nil
	}

	var flat map[string]json.RawMessage
	if err := json.Unmarshal(raw, &flat); err != nil || len(flat) == 0 {
		return nil
	}
	if _, hasTemplate := flat["item_template_id"]; hasTemplate {
		return nil
	}

	stats := make(map[string]int)
	itemLevel := 1
	for key, value := range flat {
		if key == "upgrade_pity" {
			continue
		}
		var n int
		if err := json.Unmarshal(value, &n); err != nil {
			continue
		}
		if key == "item_level" {
			if n > 0 {
				itemLevel = n
			}
			continue
		}
		stats[key] = n
	}
	if len(stats) == 0 {
		return nil
	}

	rarity := "common"
	if r, ok := flat["rarity"]; ok {
		_ = json.Unmarshal(r, &rarity)
	}

	return &ItemRollPayload{
		ItemTemplateID: itemDefID,
		DisplayName:    template.Name,
		Rarity:         rarity,
		ItemLevel:      itemLevel,
		Stats:          stats,
	}
}
