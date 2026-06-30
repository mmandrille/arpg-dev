package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func testItemSellPrice(t *testing.T, itemDefID string, rolled json.RawMessage) int {
	t.Helper()
	price, ok := game.DefaultItemSellPrice(testRules(t), itemDefID, rolled)
	if !ok || price <= 0 {
		t.Fatalf("sell price unavailable for %s rolled=%s", itemDefID, string(rolled))
	}

	return price
}

func addUpgradeShardStash(t *testing.T, s *store.Store, ctx context.Context, accountID, characterID, stashItemID string, level int) {
	t.Helper()
	stats, err := game.MarshalUpgradeShardRolledStats(level)
	if err != nil {
		t.Fatal(err)
	}
	itemID := "shard_src_" + ids.Token()[:10]
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          itemID,
		AccountID:   accountID,
		CharacterID: characterID,
		ItemDefID:   game.UpgradeShardItemDefID,
		Location:    store.ItemLocationInventory,
		RolledStats: stats,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, accountID, characterID, itemID, stashItemID); err != nil {
		t.Fatal(err)
	}
}

func addUpgradeShardInventory(t *testing.T, s *store.Store, ctx context.Context, accountID, characterID, itemID string, level int) {
	t.Helper()
	stats, err := game.MarshalUpgradeShardRolledStats(level)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          itemID,
		AccountID:   accountID,
		CharacterID: characterID,
		ItemDefID:   game.UpgradeShardItemDefID,
		Location:    store.ItemLocationInventory,
		RolledStats: stats,
	}); err != nil {
		t.Fatal(err)
	}
}
