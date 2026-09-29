package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v478: corpses are frozen in the member's start snapshot. Looting the body in
// the live tables afterwards must not change what the snapshot returns.
func TestSessionStartSnapshotFreezesRecoverableCorpses(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "session-corpse+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	hero, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Hero", "barbarian")
	if err != nil {
		t.Fatalf("create hero: %v", err)
	}
	dead, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Fallen", "sorcerer")
	if err != nil {
		t.Fatalf("create dead character: %v", err)
	}
	item := store.CharacterItemInstance{
		ID:          ids.New("item"),
		AccountID:   acct.ID,
		CharacterID: dead.ID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationEquipped,
		Slot:        "main_hand",
		Equipped:    true,
		WeaponSet:   1,
		RolledStats: json.RawMessage(`{"item_level":3}`),
	}
	if err := s.AddCharacterItem(ctx, item); err != nil {
		t.Fatalf("add corpse item: %v", err)
	}
	if err := s.MarkCharacterDead(ctx, acct.ID, dead.ID, -2); err != nil {
		t.Fatalf("mark dead: %v", err)
	}
	corpses, err := s.ListRecoverableCharacterCorpses(ctx, acct.ID, hero.ID)
	if err != nil || len(corpses) != 1 {
		t.Fatalf("recoverable corpses = %+v err=%v, want 1", corpses, err)
	}
	sess := store.Session{ID: ids.New("sess"), AccountID: acct.ID, CharacterID: hero.ID, Seed: "seed", WorldID: "dungeon_levels", Status: store.SessionActive}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	progression := store.CharacterProgression{AccountID: acct.ID, CharacterID: hero.ID, CharacterClass: "barbarian", Level: 1, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}}
	if err := s.CreateSessionStartSnapshot(ctx, store.SessionStartSnapshot{SessionID: sess.ID, AccountID: acct.ID, CharacterID: hero.ID, StashGold: store.AccountStashGold{AccountID: acct.ID}, Corpses: corpses, Progression: &progression}); err != nil {
		t.Fatalf("create session snapshot: %v", err)
	}

	newItemID := ids.New("item")
	if _, err := s.TransferCorpseItemToCharacter(ctx, acct.ID, dead.ID, hero.ID, item.ID, newItemID); err != nil {
		t.Fatalf("loot corpse: %v", err)
	}
	if live, _ := s.ListRecoverableCharacterCorpses(ctx, acct.ID, hero.ID); len(live) != 0 {
		t.Fatalf("live corpses after loot = %+v, want none", live)
	}

	snap, err := s.LoadSessionStartSnapshotForMember(ctx, sess.ID, acct.ID, hero.ID)
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if len(snap.Corpses) != 1 {
		t.Fatalf("snapshot corpses = %+v, want the frozen body", snap.Corpses)
	}
	got := snap.Corpses[0]
	if got.CharacterID != dead.ID || got.Name != "Fallen" || got.DeathLevel != -2 || len(got.Items) != 1 {
		t.Fatalf("snapshot corpse = %+v", got)
	}
	gotItem := got.Items[0]
	if gotItem.ID != item.ID || gotItem.ItemDefID != item.ItemDefID || gotItem.Slot != "main_hand" || !gotItem.Equipped ||
		gotItem.WeaponSet != 1 || gotItem.Location != store.ItemLocationEquipped || !jsonSame(gotItem.RolledStats, item.RolledStats) {
		t.Fatalf("snapshot corpse item = %+v, want %+v", gotItem, item)
	}
}

func jsonSame(a, b json.RawMessage) bool {
	var va, vb any
	return json.Unmarshal(a, &va) == nil && json.Unmarshal(b, &vb) == nil && string(mustJSON(va)) == string(mustJSON(vb))
}

func mustJSON(v any) []byte {
	out, _ := json.Marshal(v)
	return out
}
