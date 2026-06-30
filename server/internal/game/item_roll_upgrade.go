package game

import (
	"encoding/json"
	"fmt"
)

// ItemUpgradeOptions configures stash/inventory item level upgrades.
type ItemUpgradeOptions struct {
	Scaling           MonsterDepthScalingRules
	Tiers             ItemLevelTierRules
	MaxItemLevelDepth int
}

// EffectiveItemUpgradeMaxLevel returns the lowest allowed cap from config and depth progression.
func EffectiveItemUpgradeMaxLevel(configMaxLevel, depthMaxLevel int) int {
	maxLevel := configMaxLevel
	if depthMaxLevel > 0 && (maxLevel <= 0 || depthMaxLevel < maxLevel) {
		maxLevel = depthMaxLevel
	}

	if maxLevel < 1 {
		return 1
	}

	return maxLevel
}

// UpgradeRolledStatsJSON increments item level and rescales durable rolled stats JSON.
func UpgradeRolledStatsJSON(raw json.RawMessage, configMaxLevel int, opts ItemUpgradeOptions) ([]byte, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("game: decode rolled stats for upgrade: %w", err)
		}
	}

	statsMap := rollPayloadStatsMap(payload)
	requirementsMap := rollPayloadRequirementsMap(payload)
	currentLevel := intStatValue(statsMap["item_level"])
	effectiveMax := EffectiveItemUpgradeMaxLevel(configMaxLevel, opts.MaxItemLevelDepth)
	if currentLevel >= effectiveMax {
		return nil, fmt.Errorf("game: item upgrade at max level")
	}

	nextStats, nextRequirements, nextLevel, err := UpgradeItemLevelPayload(
		intStatMapFromAny(statsMap),
		intStatMapFromAny(requirementsMap),
		currentLevel,
		opts.Scaling,
		opts.Tiers,
	)
	if err != nil {
		return nil, err
	}

	nextStats["item_level"] = nextLevel
	writeRollPayloadStats(payload, nextStats)
	writeRollPayloadRequirements(payload, nextRequirements)

	out, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("game: encode upgraded rolled stats: %w", err)
	}

	return out, nil
}

func rollPayloadStatsMap(payload map[string]any) map[string]any {
	if nested, ok := payload["stats"].(map[string]any); ok {
		return nested
	}

	return payload
}

func rollPayloadRequirementsMap(payload map[string]any) map[string]any {
	if nested, ok := payload["requirements"].(map[string]any); ok {
		return nested
	}

	return map[string]any{}
}

func writeRollPayloadStats(payload map[string]any, stats map[string]int) {
	encoded := map[string]any{}
	for key, value := range stats {
		encoded[key] = value
	}

	if _, ok := payload["stats"].(map[string]any); ok {
		payload["stats"] = encoded

		return
	}

	for key, value := range encoded {
		payload[key] = value
	}
}

func writeRollPayloadRequirements(payload map[string]any, requirements map[string]int) {
	if len(requirements) == 0 {
		return
	}

	encoded := map[string]any{}
	for key, value := range requirements {
		encoded[key] = value
	}
	payload["requirements"] = encoded
}

func intStatMapFromAny(values map[string]any) map[string]int {
	out := make(map[string]int, len(values))
	for key, value := range values {
		if key == "upgrade_pity" {
			continue
		}

		if n, ok := intStatValueOK(value); ok {
			out[key] = n
		}
	}

	return out
}

func intStatValue(value any) int {
	n, _ := intStatValueOK(value)

	return n
}

func intStatValueOK(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case json.Number:
		n, err := v.Int64()

		return int(n), err == nil
	default:
		return 0, false
	}
}
