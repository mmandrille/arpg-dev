package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpgradeFailureCurveMatchesGoldenCases(t *testing.T) {
	goldenPath := filepath.Join("..", "..", "..", "shared", "golden", "upgrade_success_chance.json")
	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Curve struct {
			SafeTargetLevelMax          int   `json:"safe_target_level_max"`
			LevelAnchors                []int `json:"level_anchors"`
			FailureChancePercentAnchors []int `json:"failure_chance_percent_anchors"`
		} `json:"curve"`
		ShardBonusPercentPerTier int `json:"shard_bonus_percent_per_tier"`
		Cases                    []struct {
			CurrentLevel   int `json:"current_level"`
			ShardLevel     int `json:"shard_level"`
			SuccessPercent int `json:"success_percent"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	rules := ItemUpgradeChanceRules{
		FailureCurve: ItemUpgradeFailureCurve{
			SafeTargetLevelMax:          golden.Curve.SafeTargetLevelMax,
			LevelAnchors:                golden.Curve.LevelAnchors,
			FailureChancePercentAnchors: golden.Curve.FailureChancePercentAnchors,
		},
		ShardSuccessBonusPercentPerTier: golden.ShardBonusPercentPerTier,
	}
	for _, tc := range golden.Cases {
		minShard := UpgradeShardMinLevel(tc.CurrentLevel)
		got := EffectiveUpgradeSuccessPercent(tc.CurrentLevel, tc.ShardLevel, minShard, rules)
		if got != tc.SuccessPercent {
			t.Fatalf("current=%d shard=%d success=%d, want %d", tc.CurrentLevel, tc.ShardLevel, got, tc.SuccessPercent)
		}
	}
}

func TestUpgradeFailureAtTargetLevel10(t *testing.T) {
	curve := ItemUpgradeFailureCurve{
		SafeTargetLevelMax:          1,
		LevelAnchors:                []int{2, 10},
		FailureChancePercentAnchors: []int{10, 75},
	}
	if got := UpgradeFailureChancePercent(10, curve); got != 75 {
		t.Fatalf("target 10 failure = %d, want 75", got)
	}
}

func TestEffectiveItemUpgradeMaxLevelDepthOnlyWhenConfigZero(t *testing.T) {
	if got := EffectiveItemUpgradeMaxLevel(0, 14); got != 14 {
		t.Fatalf("depth-only max = %d, want 14", got)
	}
	if got := EffectiveItemUpgradeMaxLevel(0, 0); got != 1_000_000 {
		t.Fatalf("uncapped depth-only max = %d, want large sentinel", got)
	}
	if got := EffectiveItemUpgradeMaxLevel(5, 0); got != 5 {
		t.Fatalf("config cap without depth = %d, want 5", got)
	}
	if got := EffectiveItemUpgradeMaxLevel(5, 14); got != 5 {
		t.Fatalf("config cap below depth = %d, want 5", got)
	}
}
