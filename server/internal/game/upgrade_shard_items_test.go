package game

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestInferRollPayloadFromFlatStatsCaveBlade(t *testing.T) {
	rules, err := LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatal(err)
	}
	payload := inferRollPayloadFromFlatStats(rules, "cave_blade", json.RawMessage(`{"damage_min":2,"damage_max":4,"item_level":2}`))
	if payload == nil {
		t.Fatal("payload is nil")
	}
	price, ok := itemSellPriceFromPayload(rules, defaultAppraisalShopID, payload)
	if !ok || price <= 0 {
		t.Fatalf("sell price = %d ok=%v payload=%+v", price, ok, payload)
	}
}
