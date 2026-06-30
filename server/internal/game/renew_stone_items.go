package game

import (
	"encoding/json"
	"fmt"
)

const RenewStoneItemDefID = "renew_stone"

// NewRenewStoneRollPayload builds the durable loot/inventory payload for a leveled renew stone.
func NewRenewStoneRollPayload(level int) *ItemRollPayload {
	if level < 1 {
		level = 1
	}

	return &ItemRollPayload{
		ItemTemplateID: RenewStoneItemDefID,
		DisplayName:    "Renew Stone",
		ItemLevel:      level,
		Stats:          map[string]int{"item_level": level},
	}
}

// RenewStoneLevelFromRaw reads renew stone level from rolled_stats JSON.
func RenewStoneLevelFromRaw(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "{}" {
		return 1, nil
	}

	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err == nil && payload.ItemTemplateID == RenewStoneItemDefID {
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
		return 0, fmt.Errorf("game: decode renew stone level: %w", err)
	}
	if flat.ItemLevel < 1 {
		return 1, nil
	}

	return flat.ItemLevel, nil
}

// MarshalRenewStoneRolledStats returns persisted stash/inventory JSON for a renew stone level.
func MarshalRenewStoneRolledStats(level int) (json.RawMessage, error) {
	if level < 1 {
		level = 1
	}
	raw, err := json.Marshal(map[string]int{"item_level": level})
	if err != nil {
		return nil, fmt.Errorf("game: marshal renew stone stats: %w", err)
	}

	return raw, nil
}

// LeveledConsumableLevelFromRaw reads level for upgrade shards or renew stones.
func LeveledConsumableLevelFromRaw(itemDefID string, raw json.RawMessage) (int, error) {
	switch itemDefID {
	case UpgradeShardItemDefID:
		return UpgradeShardLevelFromRaw(raw)
	case RenewStoneItemDefID:
		return RenewStoneLevelFromRaw(raw)
	default:
		return 0, fmt.Errorf("game: unknown leveled consumable %q", itemDefID)
	}
}
