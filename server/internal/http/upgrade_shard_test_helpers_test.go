package httpapi

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func httpTestRules(t *testing.T) *game.Rules {
	t.Helper()
	rules, err := game.LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatal(err)
	}

	return rules
}

func httpTestItemSellPrice(t *testing.T, itemDefID string, rolled json.RawMessage) int {
	t.Helper()
	price, ok := game.DefaultItemSellPrice(httpTestRules(t), itemDefID, rolled)
	if !ok || price <= 0 {
		t.Fatalf("sell price unavailable for %s", itemDefID)
	}

	return price
}

func addHTTPUpgradeShardStash(t *testing.T, db *store.Store, ctx context.Context, accountID, characterID, stashItemID string, level int) {
	t.Helper()
	stats, err := game.MarshalUpgradeShardRolledStats(level)
	if err != nil {
		t.Fatal(err)
	}
	itemID := "http_shard_src_" + ids.Token()[:10]
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID: itemID, AccountID: accountID, CharacterID: characterID,
		ItemDefID: game.UpgradeShardItemDefID, Location: store.ItemLocationInventory, RolledStats: stats,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, characterID, itemID, stashItemID); err != nil {
		t.Fatal(err)
	}
}

func addHTTPUpgradeShardInventory(t *testing.T, db *store.Store, ctx context.Context, accountID, characterID, itemID string, level int) {
	t.Helper()
	stats, err := game.MarshalUpgradeShardRolledStats(level)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID: itemID, AccountID: accountID, CharacterID: characterID,
		ItemDefID: game.UpgradeShardItemDefID, Location: store.ItemLocationInventory, RolledStats: stats,
	}); err != nil {
		t.Fatal(err)
	}
}
