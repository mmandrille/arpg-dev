package store_test

import (
	"path/filepath"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func testRulesPath() string {
	return filepath.Join("..", "..", "..", "shared", "rules")
}

func testRules(t *testing.T) *game.Rules {
	t.Helper()
	rules, err := game.LoadRules(testRulesPath())
	if err != nil {
		t.Fatal(err)
	}

	return rules
}

func testUpgradeOptions(t *testing.T) game.ItemUpgradeOptions {
	t.Helper()
	rules := testRules(t)

	return game.ItemUpgradeOptions{
		Scaling: rules.DungeonGeneration.MonsterDepthScaling,
		Tiers:   rules.DungeonGeneration.ItemLevelTiers,
	}
}

func testUpgradeOptionsWithDepthCap(t *testing.T, deepestDepth int) game.ItemUpgradeOptions {
	t.Helper()
	opts := testUpgradeOptions(t)
	opts.MaxItemLevelDepth = game.MaxItemLevelForDepth(deepestDepth, opts.Tiers)

	return opts
}
