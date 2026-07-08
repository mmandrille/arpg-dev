package realtime

import (
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestCoalesceOutboundStateDelta(t *testing.T) {
	first := outEnvelope{
		Type: typeStateDelta,
		Tick: 5,
		Payload: stateDeltaPayload{
			ServerTick: 5,
			Level:      -1,
			Changes:    []game.Change{{Op: game.OpEntityUpdate}},
			Events:     []game.Event{{EventType: "monster_aggro"}},
		},
	}
	second := outEnvelope{
		Type: typeStateDelta,
		Tick: 6,
		Payload: stateDeltaPayload{
			ServerTick: 6,
			Level:      -1,
			Changes:    []game.Change{{Op: game.OpGoldUpdate, Gold: intPtr(3)}},
			Events:     []game.Event{{EventType: "monster_damaged"}},
			Performance: &performanceStatusPayload{Tick: 6, LiveMonsters: 26},
		},
	}
	merged, ok := coalesceOutbound(first, second)
	if !ok {
		t.Fatal("expected state_delta coalesce")
	}
	payload, ok := merged.Payload.(stateDeltaPayload)
	if !ok {
		t.Fatalf("payload type = %T", merged.Payload)
	}
	if merged.Tick != 6 || payload.ServerTick != 6 {
		t.Fatalf("merged tick = %d server_tick = %d", merged.Tick, payload.ServerTick)
	}
	if len(payload.Changes) != 2 || len(payload.Events) != 2 {
		t.Fatalf("merged changes/events = %d/%d, want 2/2", len(payload.Changes), len(payload.Events))
	}
	if payload.Performance == nil || payload.Performance.LiveMonsters != 26 {
		t.Fatalf("merged performance = %+v", payload.Performance)
	}
}

func TestCoalesceOutboundRejectsMixedTypes(t *testing.T) {
	_, ok := coalesceOutbound(
		outEnvelope{Type: typeStateDelta, Payload: stateDeltaPayload{}},
		outEnvelope{Type: typeIntentAccepted, Payload: intentAcceptedPayload{}},
	)
	if ok {
		t.Fatal("mixed envelope types should not coalesce")
	}
}
