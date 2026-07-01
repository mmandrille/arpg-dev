package game

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestUpgradeShardMinLevelMatchesItemTier(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		want         int
	}{
		{name: "level zero item uses tier one shard", currentLevel: 0, want: 1},
		{name: "level one item uses tier one shard", currentLevel: 1, want: 1},
		{name: "level two item uses tier two shard", currentLevel: 2, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := UpgradeShardMinLevel(test.currentLevel); got != test.want {
				t.Fatalf("UpgradeShardMinLevel(%d) = %d, want %d", test.currentLevel, got, test.want)
			}
		})
	}
}

func TestInferRollPayloadFromFlatStatsCaveBlade(t *testing.T) {
	rules, err := LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatal(err)
	}
	payload := inferRollPayloadFromFlatStats(rules, "long_sword", json.RawMessage(`{"damage_min":2,"damage_max":4,"item_level":2}`))
	if payload == nil {
		t.Fatal("payload is nil")
	}
	price, ok := itemSellPriceFromPayload(rules, defaultAppraisalShopID, payload)
	if !ok || price <= 0 {
		t.Fatalf("sell price = %d ok=%v payload=%+v", price, ok, payload)
	}
}
