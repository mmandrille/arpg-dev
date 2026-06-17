package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestSessionStartItemEquippedMutationPreservesWeaponSetAndLocation(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "session-item+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	char, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Hero", "barbarian")
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	sess := store.Session{ID: ids.New("sess"), AccountID: acct.ID, CharacterID: char.ID, Seed: "seed", WorldID: "gear_before_combat", Status: store.SessionActive}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	item := store.CharacterItemInstance{
		ID:          ids.New("item"),
		AccountID:   acct.ID,
		CharacterID: char.ID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{}`),
	}
	progression := store.CharacterProgression{
		AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1,
		Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{},
	}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, acct.ID, char.ID, []store.CharacterItemInstance{item}, nil, nil, store.CharacterSkillBindings{}, nil, nil, store.AccountStashGold{AccountID: acct.ID}, nil, progression); err != nil {
		t.Fatalf("create session snapshot: %v", err)
	}
	if err := s.SetSessionStartItemEquipped(ctx, sess.ID, acct.ID, char.ID, item.ID, "main_hand", true, 1); err != nil {
		t.Fatalf("set session start item equipped: %v", err)
	}
	snap, err := s.LoadSessionStartSnapshot(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load session snapshot: %v", err)
	}
	if len(snap.Items) != 1 || !snap.Items[0].Equipped || snap.Items[0].Slot != "main_hand" || snap.Items[0].WeaponSet != 1 || snap.Items[0].Location != store.ItemLocationEquipped {
		t.Fatalf("session start item weapon-set update not preserved: %+v", snap.Items)
	}
}
