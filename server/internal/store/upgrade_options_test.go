package store_test

import (
	"path/filepath"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func testUpgradeOptions(t *testing.T) game.ItemUpgradeOptions {
	t.Helper()
	rules, err := game.LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatal(err)
	}

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
