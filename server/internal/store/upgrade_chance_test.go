package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestAccountStashItemUpgradeFailureSpendsGoldWithoutStats(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_fail_"+suffix, "upgrade-fail+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_fail_"+suffix, acct.ID, "Failed Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	prog := store.CharacterProgression{AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1, Gold: 150, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, prog); err != nil {
		t.Fatal(err)
	}
	itemID := "fail_upgrade_item_" + suffix
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	stashID := "fail_upgrade_stash_" + suffix
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, itemID, stashID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 100); err != nil {
		t.Fatal(err)
	}
	item, gold, cost, success, err := s.UpgradeAccountStashItem(ctx, acct.ID, stashID, 100, 50, 2, 0, 100, 0, map[string]struct{}{"cave_blade": {}}, testUpgradeOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	if success {
		t.Fatal("forced failure upgrade returned success")
	}
	if cost != 100 || gold != 0 {
		t.Fatalf("failed upgrade cost/gold = %d/%d, want 100/0", cost, gold)
	}
	var stats struct {
		ItemLevel int `json:"item_level"`
		DamageMax int `json:"damage_max"`
		Pity      struct {
			Failures int `json:"failures"`
		} `json:"upgrade_pity"`
	}
	if err := json.Unmarshal(item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ItemLevel != 0 || stats.DamageMax != 4 || stats.Pity.Failures != 1 {
		t.Fatalf("failed upgrade stats changed: %+v", stats)
	}
}

func TestAccountStashItemUpgradePityGuaranteesSuccess(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_pity_"+suffix, "upgrade-pity+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_pity_"+suffix, acct.ID, "Pity Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	prog := store.CharacterProgression{AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1, Gold: 300, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, prog); err != nil {
		t.Fatal(err)
	}
	itemID := "pity_upgrade_item_" + suffix
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	stashID := "pity_upgrade_stash_" + suffix
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, itemID, stashID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 300); err != nil {
		t.Fatal(err)
	}
	var stats struct {
		ItemLevel int `json:"item_level"`
		DamageMax int `json:"damage_max"`
		Pity      struct {
			Failures int `json:"failures"`
		} `json:"upgrade_pity"`
	}
	for attempt := 1; attempt <= 2; attempt++ {
		item, gold, cost, success, err := s.UpgradeAccountStashItem(ctx, acct.ID, stashID, 100, 50, 2, 0, 100, 2, map[string]struct{}{"cave_blade": {}}, testUpgradeOptions(t))
		if err != nil {
			t.Fatal(err)
		}
		if success {
			t.Fatalf("attempt %d unexpectedly succeeded", attempt)
		}
		if cost != 100 || gold != 300-attempt*100 {
			t.Fatalf("attempt %d cost/gold = %d/%d", attempt, cost, gold)
		}
		if err := json.Unmarshal(item.RolledStats, &stats); err != nil {
			t.Fatal(err)
		}
		if stats.ItemLevel != 0 || stats.DamageMax != 4 || stats.Pity.Failures != attempt {
			t.Fatalf("attempt %d stats = %+v", attempt, stats)
		}
	}
	item, gold, cost, success, err := s.UpgradeAccountStashItem(ctx, acct.ID, stashID, 100, 50, 2, 0, 100, 2, map[string]struct{}{"cave_blade": {}}, testUpgradeOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("pity attempt should be guaranteed success")
	}
	if cost != 100 || gold != 0 {
		t.Fatalf("pity attempt cost/gold = %d/%d, want 100/0", cost, gold)
	}
	if err := json.Unmarshal(item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ItemLevel != 1 || stats.DamageMax != 5 || stats.Pity.Failures != 0 {
		t.Fatalf("pity success stats = %+v raw=%s", stats, string(item.RolledStats))
	}
}
