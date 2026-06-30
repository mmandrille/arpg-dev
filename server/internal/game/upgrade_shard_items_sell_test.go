package game_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestDefaultItemSellPriceFlatBlade(t *testing.T) {
	rules, err := game.LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatal(err)
	}
	price, ok := game.DefaultItemSellPrice(rules, "cave_blade", json.RawMessage(`{"damage_min":2,"damage_max":4,"item_level":2}`))
	if !ok || price <= 0 {
		t.Fatalf("sell price = %d ok=%v", price, ok)
	}
	t.Logf("sell price = %d", price)
}
