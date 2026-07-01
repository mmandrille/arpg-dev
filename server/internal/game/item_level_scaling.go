package game

// ItemLevelTierRules controls how dungeon depth maps to droppable item tiers.
type ItemLevelTierRules struct {
	LevelsPerTier int `json:"levels_per_tier"`
}

func (t ItemLevelTierRules) levelsPerTier() int {
	if t.LevelsPerTier < 1 {
		return 10
	}

	return t.LevelsPerTier
}

// MaxItemLevelForDepth returns the highest item level allowed at a dungeon depth.
// Level 1 is the base tier; depth 11 unlocks level 2 because tiers start above base.
func MaxItemLevelForDepth(depth int, tiers ItemLevelTierRules) int {
	if depth < 1 {
		return 1
	}

	return 1 + (depth-1)/tiers.levelsPerTier()
}

// RepresentativeDepthForItemLevel returns the anchor dungeon depth for an item level tier.
func RepresentativeDepthForItemLevel(itemLevel int, tiers ItemLevelTierRules) int {
	if itemLevel < 1 {
		itemLevel = 1
	}

	return (itemLevel-1)*tiers.levelsPerTier() + 1
}

// DepthIndexForItemLevel returns the depth index used by scaling curves for an item level.
func DepthIndexForItemLevel(itemLevel int, tiers ItemLevelTierRules) int {
	return DepthIndex(RepresentativeDepthForItemLevel(itemLevel, tiers))
}

// RollItemLevel picks a uniform item level from 1..max for the given source depth.
func RollItemLevel(rng *RNG, sourceDepth int, tiers ItemLevelTierRules) int {
	maxLevel := MaxItemLevelForDepth(sourceDepth, tiers)
	if maxLevel <= 1 {
		return 1
	}

	return 1 + rng.IntN(maxLevel)
}

// ApplyItemLevelScaling scales ilvl-1 stats and requirements to the target item level.
func ApplyItemLevelScaling(stats map[string]int, requirements map[string]int, itemLevel int, scaling MonsterDepthScalingRules, tiers ItemLevelTierRules) (map[string]int, map[string]int) {
	depthIndex := DepthIndexForItemLevel(itemLevel, tiers)

	return scaleItemStatMap(stats, depthIndex, scaling), scaleItemStatMap(requirements, depthIndex, scaling)
}

// UpgradeItemLevelPayload increments item level and rescales stats/requirements proportionally.
func UpgradeItemLevelPayload(stats map[string]int, requirements map[string]int, currentLevel int, scaling MonsterDepthScalingRules, tiers ItemLevelTierRules) (map[string]int, map[string]int, int, error) {
	if currentLevel < 0 {
		currentLevel = 0
	}

	nextLevel := currentLevel + 1
	currentIndex := DepthIndexForItemLevel(currentLevel, tiers)
	nextIndex := DepthIndexForItemLevel(nextLevel, tiers)
	baselineStats := unscaleItemStatMap(stats, currentIndex, scaling)
	baselineRequirements := unscaleItemStatMap(requirements, currentIndex, scaling)
	nextStats := scaleItemStatMap(baselineStats, nextIndex, scaling)
	nextRequirements := scaleItemStatMap(baselineRequirements, nextIndex, scaling)

	return nextStats, nextRequirements, nextLevel, nil
}

// FinalizeItemRollPayload assigns item level and applies tier scaling to a level-1 roll payload.
func FinalizeItemRollPayload(payload ItemRollPayload, itemLevel int, scaling MonsterDepthScalingRules, tiers ItemLevelTierRules) ItemRollPayload {
	if itemLevel < 1 {
		itemLevel = 1
	}

	stats, requirements := ApplyItemLevelScaling(payload.Stats, payload.Requirements, itemLevel, scaling, tiers)
	payload.ItemLevel = itemLevel
	payload.Stats = stats
	payload.Requirements = requirements

	return payload
}
