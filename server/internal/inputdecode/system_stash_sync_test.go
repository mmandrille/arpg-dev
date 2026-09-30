package inputdecode

import (
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestStoredAccountStashSyncRoundTrip(t *testing.T) {
	want := game.AccountStashSync{
		PlayerID: 1001,
		Gold:     42,
		Items: []game.PersistedStashItem{
			{StashItemID: "5010", ItemDefID: "red_potion", RolledStats: json.RawMessage(`{}`)},
			{StashItemID: "5020", ItemDefID: "long_sword", RolledStats: json.RawMessage(`{"rarity":"magic"}`)},
		},
	}
	raw, err := EncodeStoredAccountStashSync("sys-stash-1", want)
	if err != nil {
		t.Fatal(err)
	}
	in, ok := DecodeStored(raw)
	if !ok || in.Type != TypeSystemAccountStashSync || in.MessageID != "sys-stash-1" || in.StashSync == nil {
		t.Fatalf("DecodeStored = %+v ok=%v", in, ok)
	}
	got := *in.StashSync
	if got.PlayerID != want.PlayerID || got.Gold != want.Gold || len(got.Items) != len(want.Items) {
		t.Fatalf("round trip = %+v, want %+v", got, want)
	}
	for i := range want.Items {
		if got.Items[i].StashItemID != want.Items[i].StashItemID || got.Items[i].ItemDefID != want.Items[i].ItemDefID || string(got.Items[i].RolledStats) != string(want.Items[i].RolledStats) {
			t.Fatalf("item %d = %+v, want %+v", i, got.Items[i], want.Items[i])
		}
	}
}

func TestAccountStashSyncIsNotAClientIntent(t *testing.T) {
	if IsClientIntent(TypeSystemAccountStashSync) {
		t.Fatal("clients must not be able to submit system_account_stash_sync")
	}
	if _, ok := Decode(TypeSystemAccountStashSync, "m1", "", json.RawMessage(`{"player_entity_id":"1001","gold":1,"items":[]}`)); ok {
		t.Fatal("Decode accepted a client-submitted system_account_stash_sync")
	}
}

func TestStoredAccountStashSyncRejectsBadPlayer(t *testing.T) {
	raw := json.RawMessage(`{"type":"system_account_stash_sync","message_id":"m","payload":{"player_entity_id":"x","gold":0,"items":[]}}`)
	if _, ok := DecodeStored(raw); ok {
		t.Fatal("DecodeStored accepted a non-numeric player_entity_id")
	}
}
