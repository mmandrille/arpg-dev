package realtime

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestEvaluateTickGuardrail(t *testing.T) {
	within := evaluateTickGuardrail(50 * time.Millisecond)
	if within.OverBudget || within.Overrun != 0 {
		t.Fatalf("within budget decision = %+v, want no overrun", within)
	}
	over := evaluateTickGuardrail(150 * time.Millisecond)
	if !over.OverBudget || over.Budget != time.Second/tickHz || over.Overrun != 50*time.Millisecond {
		t.Fatalf("over budget decision = %+v", over)
	}
}

func TestShouldApplyOverloadDegradationRequiresRoomPressure(t *testing.T) {
	if shouldApplyOverloadDegradation(game.PerfCounters{}) {
		t.Fatalf("empty counters should not apply overload degradation")
	}
	cases := []struct {
		name     string
		counters game.PerfCounters
	}{
		{name: "path requests", counters: game.PerfCounters{PathRequests: 1}},
		{name: "path cache hits", counters: game.PerfCounters{PathCacheHits: 1}},
		{name: "path nodes", counters: game.PerfCounters{PathNodesVisited: 1}},
		{name: "monsters moved", counters: game.PerfCounters{MonstersMoved: 1}},
	}
	for _, tc := range cases {
		if !shouldApplyOverloadDegradation(tc.counters) {
			t.Fatalf("%s should apply overload degradation", tc.name)
		}
	}
}

func TestLogTickBudgetWarningPayload(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	decision := evaluateTickGuardrail(150 * time.Millisecond)
	logTickBudgetWarning(
		logger,
		42,
		150*time.Millisecond,
		decision,
		110*time.Millisecond,
		20*time.Millisecond,
		5*time.Millisecond,
		3,
		[]game.TickResult{{
			Tick:    42,
			Changes: []game.Change{{Op: game.OpEntityUpdate}},
			Events:  []game.Event{{EventType: "monster_aggro"}},
		}},
		2,
		game.PerfSnapshot{LiveMonsters: 36, Walls: 9},
		game.PerfCounters{PathRequests: 10, PathCacheHits: 4, PathNodesVisited: 1201, MonstersMoved: 1},
		true,
	)

	payload := map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal guardrail log: %v\n%s", err, buf.String())
	}
	assertJSONValue(t, payload, "msg", sessionTickBudgetOverrunMessage)
	assertJSONValue(t, payload, "tick", float64(42))
	assertJSONValue(t, payload, "tick_budget_ms", durationMS(time.Second/tickHz))
	assertJSONValue(t, payload, "tick_overrun_ms", float64(50))
	assertJSONValue(t, payload, "live_monsters", float64(36))
	assertJSONValue(t, payload, "path_requests", float64(10))
	assertJSONValue(t, payload, "path_cache_hits", float64(4))
	assertJSONValue(t, payload, "path_nodes_visited", float64(1201))
	assertJSONValue(t, payload, "monsters_moved", float64(1))
	assertJSONValue(t, payload, "degradation_applied", true)
}
