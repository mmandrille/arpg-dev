package game

import "math"

// DepthIndex converts a 1-based dungeon depth into the zero-based index used by
// depth scaling curves.
func DepthIndex(depth int) int {
	if depth < 1 {
		return 0
	}

	return depth - 1
}

// DepthFactor returns 1 + perDepth*depthIndex, the shared multiplier used by
// monsters and item tier scaling.
func DepthFactor(perDepth float64, depthIndex int) float64 {
	if depthIndex < 0 {
		depthIndex = 0
	}

	return 1 + perDepth*float64(depthIndex)
}

// ScaleIntByDepthFactor multiplies an integer stat by the shared depth factor.
func ScaleIntByDepthFactor(value int, perDepth float64, depthIndex int) int {
	if value == 0 || depthIndex <= 0 {
		return value
	}

	return roundPositive(float64(value) * DepthFactor(perDepth, depthIndex))
}

// UnscaleIntByDepthFactor reverses ScaleIntByDepthFactor for proportional upgrades.
func UnscaleIntByDepthFactor(value int, perDepth float64, depthIndex int) int {
	if value == 0 || depthIndex <= 0 {
		return value
	}

	factor := DepthFactor(perDepth, depthIndex)
	if factor <= 0 {
		return value
	}

	return roundPositive(float64(value) / factor)
}

// ScaleAdditiveDepthStat adds per-depth growth to a base integer stat.
func ScaleAdditiveDepthStat(value int, perDepth float64, depthIndex int) int {
	if depthIndex <= 0 {
		return value
	}

	return int(math.Round(float64(value) + perDepth*float64(depthIndex)))
}

// UnscaleAdditiveDepthStat reverses ScaleAdditiveDepthStat.
func UnscaleAdditiveDepthStat(value int, perDepth float64, depthIndex int) int {
	if depthIndex <= 0 {
		return value
	}

	return int(math.Round(float64(value) - perDepth*float64(depthIndex)))
}

// ScaleItemStatValue applies monster-parity scaling to one rolled item stat key.
func ScaleItemStatValue(statKey string, value int, depthIndex int, scaling MonsterDepthScalingRules) int {
	if depthIndex <= 0 || value == 0 {
		return value
	}

	switch statKey {
	case "damage_min", "damage_max":
		return ScaleIntByDepthFactor(value, scaling.DamagePerDepth, depthIndex)
	case "max_hp", "max_mana", "str", "dex", "vit", "magic",
		"all_skills", "attack_speed_percent", "health_regen_per_10_seconds", "mana_regen_per_10_seconds",
		"skill_damage_percent", "hotbar_slots", "inventory_rows", "level":
		return ScaleIntByDepthFactor(value, scaling.HPPerDepth, depthIndex)
	case "armor":
		return ScaleAdditiveDepthStat(value, scaling.ArmorPerDepth, depthIndex)
	case "block_percent":
		return ScaleAdditiveDepthStat(value, scaling.BlockPercentPerDepth, depthIndex)
	default:
		return ScaleIntByDepthFactor(value, scaling.HPPerDepth, depthIndex)
	}
}

// UnscaleItemStatValue reverses ScaleItemStatValue for one stat key.
func UnscaleItemStatValue(statKey string, value int, depthIndex int, scaling MonsterDepthScalingRules) int {
	if depthIndex <= 0 || value == 0 {
		return value
	}

	switch statKey {
	case "damage_min", "damage_max":
		return UnscaleIntByDepthFactor(value, scaling.DamagePerDepth, depthIndex)
	case "max_hp", "max_mana", "str", "dex", "vit", "magic",
		"all_skills", "attack_speed_percent", "health_regen_per_10_seconds", "mana_regen_per_10_seconds",
		"skill_damage_percent", "hotbar_slots", "inventory_rows", "level":
		return UnscaleIntByDepthFactor(value, scaling.HPPerDepth, depthIndex)
	case "armor":
		return UnscaleAdditiveDepthStat(value, scaling.ArmorPerDepth, depthIndex)
	case "block_percent":
		return UnscaleAdditiveDepthStat(value, scaling.BlockPercentPerDepth, depthIndex)
	default:
		return UnscaleIntByDepthFactor(value, scaling.HPPerDepth, depthIndex)
	}
}

func scaleItemStatMap(stats map[string]int, depthIndex int, scaling MonsterDepthScalingRules) map[string]int {
	if depthIndex <= 0 {
		return cloneIntMap(stats)
	}

	out := make(map[string]int, len(stats))
	for key, value := range stats {
		out[key] = ScaleItemStatValue(key, value, depthIndex, scaling)
	}

	return out
}

func unscaleItemStatMap(stats map[string]int, depthIndex int, scaling MonsterDepthScalingRules) map[string]int {
	if depthIndex <= 0 {
		return cloneIntMap(stats)
	}

	out := make(map[string]int, len(stats))
	for key, value := range stats {
		out[key] = UnscaleItemStatValue(key, value, depthIndex, scaling)
	}

	return out
}
