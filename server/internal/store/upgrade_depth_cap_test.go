package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestAccountStashItemUpgradeRejectsDepthCap(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_depth_"+suffix, "upgrade-depth+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_depth_"+suffix, acct.ID, "Depth Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	prog := store.CharacterProgression{
		AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1, Gold: 500,
		DeepestDungeonDepth: 25,
		Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{},
	}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, prog); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID: "depth_upgrade_item_"+suffix, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade",
		Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4,"item_level":2}`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, "depth_upgrade_item_"+suffix, "depth_upgrade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 300); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, err = s.UpgradeAccountStashItem(ctx, acct.ID, "depth_upgrade_stash_"+suffix, 100, 50, 3, 100, 1, 0, map[string]struct{}{"cave_blade": {}}, testUpgradeOptionsWithDepthCap(t, 25))
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("depth-capped upgrade err = %v, want ErrConflict", err)
	}
}
