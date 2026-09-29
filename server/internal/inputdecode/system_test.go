package inputdecode

import (
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestStoredLoadShedRoundTrip(t *testing.T) {
	want := game.LoadShedDirective{OverloadDegrade: true, CombatMovementThrottle: true}
	raw, err := EncodeStoredLoadShed("sys-1", want)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	in, ok := DecodeStored(raw)
	if !ok {
		t.Fatal("DecodeStored rejected system_load_shed")
	}
	if in.Type != TypeSystemLoadShed || in.MessageID != "sys-1" || in.LoadShed == nil || *in.LoadShed != want {
		t.Fatalf("decoded = %+v, want load shed %+v", in, want)
	}
}

func TestLoadShedIsNotAClientIntent(t *testing.T) {
	if IsClientIntent(TypeSystemLoadShed) {
		t.Fatal("clients must not be able to submit system_load_shed")
	}
	payload, _ := json.Marshal(map[string]bool{"overload_degrade": true})
	if _, ok := Decode(TypeSystemLoadShed, "msg-1", "", payload); ok {
		t.Fatal("client Decode must reject system_load_shed")
	}
}

func TestStoredMemberLifecycleRoundTrip(t *testing.T) {
	want := game.MemberLifecycle{PlayerID: 1737871872194362512, AccountID: "acct_g", CharacterID: "char_g", Respawn: true}
	for _, typ := range []string{TypeSystemMemberJoin, TypeSystemMemberLeave, TypeSystemMemberRejoin} {
		raw, err := EncodeStoredMemberLifecycle(typ, "sys-1", want)
		if err != nil {
			t.Fatalf("encode %s: %v", typ, err)
		}
		in, ok := DecodeStored(raw)
		if !ok {
			t.Fatalf("DecodeStored rejected %s", typ)
		}
		if in.Type != typ || in.MessageID != "sys-1" || in.Member == nil || *in.Member != want {
			t.Fatalf("decoded %s = %+v, want member %+v", typ, in, want)
		}
	}
}

func TestMemberLifecycleIsNotAClientIntent(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"player_entity_id": "7"})
	for _, typ := range []string{TypeSystemMemberJoin, TypeSystemMemberLeave, TypeSystemMemberRejoin} {
		if IsClientIntent(typ) {
			t.Fatalf("clients must not be able to submit %s", typ)
		}
		if _, ok := Decode(typ, "msg-1", "", payload); ok {
			t.Fatalf("client Decode must reject %s", typ)
		}
	}
}

func TestStoredMemberLifecycleRejectsMissingPlayer(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{"type": TypeSystemMemberLeave, "message_id": "sys-1", "payload": map[string]string{}})
	if _, ok := DecodeStored(raw); ok {
		t.Fatal("a membership row without a player entity must not decode")
	}
}
