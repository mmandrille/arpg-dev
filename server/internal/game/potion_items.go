package game

import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	RejuvPotionItemDefID = "rejuv_potion"
)

// PotionRulesConfig holds data-driven leveled potion tuning.
type PotionRulesConfig struct {
	RestoreMultiplierPerLevel int `json:"restore_multiplier_per_level"`
	RejuvMinRestorePercent    int `json:"rejuv_min_restore_percent"`
	RejuvDropWeightPercent    int `json:"rejuv_drop_weight_percent"`
	ShopBaseBuyPrice          map[string]int `json:"shop_base_buy_price"`
	ShopBuyPricePerLevel      int            `json:"shop_buy_price_per_level"`
}

func (r *Rules) potionRules() PotionRulesConfig {
	if r == nil {
		return PotionRulesConfig{
			RestoreMultiplierPerLevel: 3,
			RejuvMinRestorePercent:    33,
			RejuvDropWeightPercent:    20,
			ShopBuyPricePerLevel:      40,
		}
	}

	return r.MainConfig.Gameplay.PotionRules
}

// IsLeveledPotion reports whether an item def uses floor-scaled potion levels.
func IsLeveledPotion(itemDefID string) bool {
	switch itemDefID {
	case "red_potion", "blue_potion", RejuvPotionItemDefID:
		return true
	default:
		return false
	}
}

// IsRejuvPotion reports whether an item def is the dual-resource rejuv potion.
func IsRejuvPotion(itemDefID string) bool {
	return itemDefID == RejuvPotionItemDefID
}

// PotionGroundDisplayName is the generic floor label for a potion type.
func PotionGroundDisplayName(itemDefID string) string {
	switch itemDefID {
	case "red_potion":
		return "Health Potion"
	case "blue_potion":
		return "Mana Potion"
	case RejuvPotionItemDefID:
		return "Rejuv Potion"
	default:
		return ""
	}
}

// PotionInventoryDisplayName is the bag/tooltip name for a leveled potion.
func PotionInventoryDisplayName(itemDefID string) string {
	name := PotionGroundDisplayName(itemDefID)
	if name != "" {
		return name
	}
	return itemDefID
}

// NewPotionRollPayload builds authoritative loot/inventory payload for a leveled potion.
func NewPotionRollPayload(itemDefID string, level int) *ItemRollPayload {
	if level < 1 {
		level = 1
	}

	return &ItemRollPayload{
		ItemTemplateID: itemDefID,
		DisplayName:    PotionInventoryDisplayName(itemDefID),
		ItemLevel:      level,
		Stats:          map[string]int{"item_level": level},
	}
}

// MarshalPotionRolledStats returns persisted JSON for a leveled potion tier.
func MarshalPotionRolledStats(level int) (json.RawMessage, error) {
	if level < 1 {
		level = 1
	}
	raw, err := json.Marshal(map[string]int{"item_level": level})
	if err != nil {
		return nil, fmt.Errorf("game: marshal potion stats: %w", err)
	}

	return raw, nil
}

// PotionLevelFromItem reads potion level from inventory payload, defaulting to 1.
func PotionLevelFromItem(item *invItem) int {
	if item == nil {
		return 1
	}
	if item.rollPayload != nil {
		if item.rollPayload.ItemLevel > 0 {
			return item.rollPayload.ItemLevel
		}
		if level := item.rollPayload.Stats["item_level"]; level > 0 {
			return level
		}
	}

	return 1
}

// PotionLevelFromRaw reads potion level from rolled_stats JSON.
func PotionLevelFromRaw(itemDefID string, raw json.RawMessage) (int, error) {
	if !IsLeveledPotion(itemDefID) {
		return 0, fmt.Errorf("game: not a leveled potion %q", itemDefID)
	}
	if len(raw) == 0 || string(raw) == "{}" {
		return 1, nil
	}

	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err == nil && payload.ItemTemplateID != "" {
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
		return 0, fmt.Errorf("game: decode potion level: %w", err)
	}
	if flat.ItemLevel < 1 {
		return 1, nil
	}

	return flat.ItemLevel, nil
}

// PotionShopTierLevel maps character progression to vendor potion tier.
func PotionShopTierLevel(deepestDepth int) int {
	if deepestDepth < 1 {
		return 1
	}

	return deepestDepth
}

// PotionShopBuyPrice returns the buy price for a leveled potion at a tier.
func (r *Rules) PotionShopBuyPrice(itemDefID string, level int, roundTo int) int {
	cfg := r.potionRules()
	if level < 1 {
		level = 1
	}
	base := cfg.ShopBaseBuyPrice[itemDefID]
	if base < 1 {
		base = cfg.ShopBuyPricePerLevel
	}
	if base < 1 {
		base = 40
	}
	price := base * level
	if roundTo < 1 {
		roundTo = 5
	}

	return int(math.Round(float64(price)/float64(roundTo))) * roundTo
}

// ResolvePotionDropKind remaps a red/blue treasure-class roll into the potion band.
func (r *Rules) ResolvePotionDropKind(rng *RNG, rolledItemDefID string) string {
	return r.resolvePotionDropKindFromRoll(rolledItemDefID, rng.IntN(100))
}

// resolvePotionDropKindFromRoll remaps potion drops without advancing unrelated RNG streams.
func (r *Rules) resolvePotionDropKindFromRoll(rolledItemDefID string, roll int) string {
	if rolledItemDefID != "red_potion" && rolledItemDefID != "blue_potion" {
		return rolledItemDefID
	}
	cfg := r.potionRules()
	rejuvWeight := cfg.RejuvDropWeightPercent
	if rejuvWeight < 0 {
		rejuvWeight = 0
	}
	if rejuvWeight > 100 {
		rejuvWeight = 100
	}
	remaining := 100 - rejuvWeight
	healthWeight := remaining / 2
	bandRoll := roll % 100
	if bandRoll < 0 {
		bandRoll += 100
	}
	if bandRoll < rejuvWeight {
		return RejuvPotionItemDefID
	}
	if bandRoll < rejuvWeight+healthWeight {
		return "red_potion"
	}

	return "blue_potion"
}

// PotionRestoreAmount returns flat HP or mana restore for health/mana potions.
func (r *Rules) PotionRestoreAmount(level int) int {
	cfg := r.potionRules()
	multiplier := cfg.RestoreMultiplierPerLevel
	if multiplier < 1 {
		multiplier = 3
	}
	if level < 1 {
		level = 1
	}

	return multiplier * level
}

// RejuvRestorePercent returns the percent restore for rejuv potions at a tier.
func (r *Rules) RejuvRestorePercent(level int) int {
	cfg := r.potionRules()
	minPercent := cfg.RejuvMinRestorePercent
	if minPercent < 1 {
		minPercent = 33
	}
	if level < 1 {
		level = 1
	}
	if level > minPercent {
		return level
	}

	return minPercent
}

// PotionSummaryLines returns inventory/shop summary text for a leveled potion.
func (r *Rules) PotionSummaryLines(itemDefID string, level int) []string {
	if !IsLeveledPotion(itemDefID) {
		return nil
	}
	if level < 1 {
		level = 1
	}
	lines := []string{fmt.Sprintf("Level %d", level)}
	switch itemDefID {
	case "red_potion":
		amount := r.PotionRestoreAmount(level)
		lines = append(lines, fmt.Sprintf("Restores %d HP", amount))
	case "blue_potion":
		amount := r.PotionRestoreAmount(level)
		lines = append(lines, fmt.Sprintf("Restores %d mana", amount))
	case RejuvPotionItemDefID:
		percent := r.RejuvRestorePercent(level)
		lines = append(lines, fmt.Sprintf("Restores %d%% HP and mana", percent))
	}

	return lines
}

// PotionFloorLevelFromGoldContext maps loot source depth to potion item level.
func PotionFloorLevelFromGoldContext(goldCtx goldRollContext) int {
	depth := absInt(goldCtx.levelNum)
	if depth < 1 {
		return 1
	}

	return depth
}
