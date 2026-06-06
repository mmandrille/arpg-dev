package inputdecode

import (
	"encoding/json"
	"testing"
)

func TestDecodeTeleportAllowsTownLevel(t *testing.T) {
	in, ok := Decode(TypeTeleport, "msg_tp_town", "", json.RawMessage(`{"target_level":0}`))
	if !ok {
		t.Fatal("Decode teleport target_level 0 rejected")
	}
	if in.Teleport == nil || in.Teleport.TargetLevel != 0 {
		t.Fatalf("decoded teleport = %+v, want target level 0", in.Teleport)
	}
}

func TestDecodeTeleportRejectsMissingOrPositiveTargetLevel(t *testing.T) {
	tests := []struct {
		name    string
		payload json.RawMessage
	}{
		{name: "missing", payload: json.RawMessage(`{}`)},
		{name: "positive", payload: json.RawMessage(`{"target_level":1}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := Decode(TypeTeleport, "msg_tp", "", tt.payload); ok {
				t.Fatalf("Decode(%s) accepted invalid teleport payload", tt.payload)
			}
		})
	}
}
