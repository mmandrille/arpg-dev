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
