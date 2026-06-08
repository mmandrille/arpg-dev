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

func TestDecodeAllocateStatIntent(t *testing.T) {
	in, ok := Decode(TypeAllocateStat, "msg_stat", "", json.RawMessage(`{"stat":"vit","points":1}`))
	if !ok {
		t.Fatal("Decode allocate_stat_intent rejected valid payload")
	}
	if in.AllocateStat == nil || in.AllocateStat.Stat != "vit" || in.AllocateStat.Points != 1 {
		t.Fatalf("decoded allocate stat = %+v", in.AllocateStat)
	}
}

func TestDecodeAllocateStatIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{"stat":"luck","points":1}`),
		json.RawMessage(`{"stat":"str","points":0}`),
		json.RawMessage(`{"stat":"str","points":-1}`),
		json.RawMessage(`{"points":1}`),
		json.RawMessage(`{"stat":"str"}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeAllocateStat, "msg_stat", "", payload); ok {
			t.Fatalf("Decode accepted invalid allocate stat payload %s", payload)
		}
	}
}

func TestDecodeDirectionalAttackIntent(t *testing.T) {
	in, ok := Decode(TypeDirectional, "msg_dir", "corr_dir", json.RawMessage(`{"direction":{"x":1,"y":0}}`))
	if !ok {
		t.Fatal("Decode directional_attack_intent rejected valid payload")
	}
	if in.DirectionalAttack == nil || in.DirectionalAttack.Direction.X != 1 || in.DirectionalAttack.Direction.Y != 0 {
		t.Fatalf("decoded directional attack = %+v", in.DirectionalAttack)
	}
	if !IsClientIntent(TypeDirectional) {
		t.Fatal("directional_attack_intent not marked as client intent")
	}
}

func TestDecodeDirectionalAttackIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"direction":null}`),
		json.RawMessage(`{"direction":{"x":"bad","y":0}}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeDirectional, "msg_dir", "", payload); ok {
			t.Fatalf("Decode accepted invalid directional attack payload %s", payload)
		}
	}
}
