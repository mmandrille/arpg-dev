package game

import (
	"fmt"
	"math"
)

// ItemUpgradeFailureCurve configures logarithmic failure growth by target item level.
type ItemUpgradeFailureCurve struct {
	SafeTargetLevelMax          int `json:"safe_target_level_max"`
	LevelAnchors                []int `json:"level_anchors"`
	FailureChancePercentAnchors []int `json:"failure_chance_percent_anchors"`
}

// ItemUpgradeChanceRules combines failure curve and shard-tier success bonus.
type ItemUpgradeChanceRules struct {
	FailureCurve                ItemUpgradeFailureCurve
	ShardSuccessBonusPercentPerTier int
}

// UpgradeTargetLevel returns the item level after a successful upgrade attempt.
func UpgradeTargetLevel(currentLevel int) int {
	if currentLevel < 0 {
		return 1
	}

	return currentLevel + 1
}

// UpgradeFailureChancePercent returns failure percent for the post-upgrade target level.
func UpgradeFailureChancePercent(targetLevel int, curve ItemUpgradeFailureCurve) int {
	if targetLevel <= curve.SafeTargetLevelMax {
		return 0
	}
	if len(curve.LevelAnchors) == 0 || len(curve.FailureChancePercentAnchors) == 0 {
		return 0
	}
	if len(curve.LevelAnchors) != len(curve.FailureChancePercentAnchors) {
		return 0
	}

	anchors := curve.LevelAnchors
	failures := curve.FailureChancePercentAnchors
	if targetLevel <= anchors[0] {
		return clampPercent(failures[0])
	}

	lastIndex := len(anchors) - 1
	if targetLevel >= anchors[lastIndex] {
		if lastIndex == 0 {
			return clampPercent(failures[0])
		}

		return clampPercent(extrapolateLogFailure(targetLevel, anchors[lastIndex-1], failures[lastIndex-1], anchors[lastIndex], failures[lastIndex]))
	}

	for i := 1; i < len(anchors); i++ {
		if targetLevel > anchors[i] {
			continue
		}

		return clampPercent(interpolateLogFailure(targetLevel, anchors[i-1], failures[i-1], anchors[i], failures[i]))
	}

	return clampPercent(failures[lastIndex])
}

// EffectiveUpgradeSuccessPercent returns success chance for an upgrade attempt.
func EffectiveUpgradeSuccessPercent(currentLevel, shardLevel, minShardLevel int, rules ItemUpgradeChanceRules) int {
	targetLevel := UpgradeTargetLevel(currentLevel)
	failure := UpgradeFailureChancePercent(targetLevel, rules.FailureCurve)
	success := 100 - failure
	if rules.ShardSuccessBonusPercentPerTier > 0 && shardLevel > minShardLevel {
		success += rules.ShardSuccessBonusPercentPerTier * (shardLevel - minShardLevel)
	}

	return clampPercent(success)
}

func interpolateLogFailure(target, leftLevel, leftFailure, rightLevel, rightFailure int) int {
	leftLog := math.Log10(float64(leftLevel))
	rightLog := math.Log10(float64(rightLevel))
	targetLog := math.Log10(float64(target))
	if rightLog <= leftLog {
		return leftFailure
	}

	weight := (targetLog - leftLog) / (rightLog - leftLog)
	value := float64(leftFailure) + weight*float64(rightFailure-leftFailure)

	return int(math.Round(value))
}

func extrapolateLogFailure(target, leftLevel, leftFailure, rightLevel, rightFailure int) int {
	leftLog := math.Log10(float64(leftLevel))
	rightLog := math.Log10(float64(rightLevel))
	targetLog := math.Log10(float64(target))
	if rightLog <= leftLog {
		return rightFailure
	}

	slope := float64(rightFailure-leftFailure) / (rightLog - leftLog)
	value := float64(rightFailure) + slope*(targetLog-rightLog)

	return int(math.Round(value))
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}

	return value
}

func validateItemUpgradeFailureCurve(curve ItemUpgradeFailureCurve) error {
	if curve.SafeTargetLevelMax < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve.safe_target_level_max: must be non-negative")
	}
	if len(curve.LevelAnchors) == 0 || len(curve.FailureChancePercentAnchors) == 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve: anchors must not be empty")
	}
	if len(curve.LevelAnchors) != len(curve.FailureChancePercentAnchors) {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve: anchor arrays must match length")
	}
	for i, level := range curve.LevelAnchors {
		if level <= curve.SafeTargetLevelMax {
			return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve.level_anchors[%d]: must be above safe_target_level_max", i)
		}
		if i > 0 && level <= curve.LevelAnchors[i-1] {
			return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve.level_anchors[%d]: must increase", i)
		}
	}
	for i, failure := range curve.FailureChancePercentAnchors {
		if failure < 0 || failure > 100 {
			return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_failure_curve.failure_chance_percent_anchors[%d]: must be within [0,100]", i)
		}
	}

	return nil
}

func validateItemUpgradeChanceConfig(maxLevel int, curve ItemUpgradeFailureCurve, shardBonusPercentPerTier int) error {
	if maxLevel < 0 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_max_level: must be non-negative")
	}
	if shardBonusPercentPerTier < 0 || shardBonusPercentPerTier > 100 {
		return fmt.Errorf("game: invalid rules main_config.gameplay.item_upgrade_shard_success_bonus_percent_per_tier: must be 0-100")
	}

	return validateItemUpgradeFailureCurve(curve)
}
