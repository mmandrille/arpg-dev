package realtime

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func TestLogBackendPerfIncludesCrowdedCombatFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	profiler := newBackendTickProfiler()
	profiler.phases[game.TickPhaseAI] = 5 * time.Millisecond
	profiler.phases[game.TickPhasePathfind] = 2 * time.Millisecond
	profiler.phases[game.TickPhaseCombat] = 7 * time.Millisecond

	logBackendPerf(
		logger,
		7,
		time.Now().Add(-150*time.Millisecond),
		20*time.Millisecond,
		3*time.Millisecond,
		4*time.Millisecond,
		2,
		[]game.TickResult{{
			Tick:    7,
			Changes: []game.Change{{Op: game.OpEntityUpdate}},
			Events:  []game.Event{{EventType: "monster_aggro"}},
			Acks:    []game.Ack{{MessageID: "accepted"}},
			Rejects: []game.Reject{{MessageID: "rejected", Reason: "invalid"}},
		}},
		1,
		game.PerfSnapshot{Level: 0, Entities: 40, LiveMonsters: 36, Walls: 9},
		game.PerfCounters{PathRequests: 12, PathNodesVisited: 345, MonstersMoved: 9},
		profiler,
	)

	payload := map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal perf log: %v\n%s", err, buf.String())
	}
	assertJSONValue(t, payload, "msg", "backend_perf")
	assertJSONValue(t, payload, "tick", float64(7))
	assertJSONValue(t, payload, "ai_ms", float64(5))
	assertJSONValue(t, payload, "pathfind_ms", float64(2))
	assertJSONValue(t, payload, "combat_ms", float64(7))
	assertJSONValue(t, payload, "broadcast_ms", float64(4))
	assertJSONValue(t, payload, "persist_ms", float64(3))
	assertJSONValue(t, payload, "path_requests", float64(12))
	assertJSONValue(t, payload, "path_cache_hits", float64(0))
	assertJSONValue(t, payload, "path_nodes_visited", float64(345))
	assertJSONValue(t, payload, "monsters_moved", float64(9))
	assertJSONValue(t, payload, "tick_budget_ms", durationMS(time.Second/tickHz))
	assertJSONValue(t, payload, "tick_over_budget", true)
	assertJSONValue(t, payload, "live_monsters", float64(36))
	assertJSONValue(t, payload, "walls", float64(9))
	if got, ok := payload["tick_overrun_ms"].(float64); !ok || got <= 0 {
		t.Fatalf("tick_overrun_ms = %v, want positive float64", payload["tick_overrun_ms"])
	}
}

func TestBuildPerformanceStatusIncludesCrowdedCombatFields(t *testing.T) {
	profiler := newBackendTickProfiler()
	profiler.phases[game.TickPhaseAI] = 5 * time.Millisecond
	profiler.phases[game.TickPhasePathfind] = 2 * time.Millisecond
	profiler.phases[game.TickPhaseCombat] = 7 * time.Millisecond

	perf := buildPerformanceStatus(
		9,
		140*time.Millisecond,
		20*time.Millisecond,
		3*time.Millisecond,
		4*time.Millisecond,
		2,
		[]game.TickResult{{
			Tick:    9,
			Changes: []game.Change{{Op: game.OpEntityUpdate}, {Op: game.OpEntityRemove}},
			Events:  []game.Event{{EventType: "monster_aggro"}},
			Acks:    []game.Ack{{MessageID: "accepted"}},
			Rejects: []game.Reject{{MessageID: "rejected", Reason: "invalid"}},
		}},
		3,
		game.PerfSnapshot{Level: -3, Entities: 41, Players: 2, Monsters: 36, LiveMonsters: 34, Companions: 1, Projectiles: 1, Loot: 2, Interactables: 2, Walls: 9},
		game.PerfCounters{PathRequests: 12, PathCacheHits: 5, PathNodesVisited: 345, MonstersMoved: 9},
		profiler,
		true,
	)

	if perf.Tick != 9 || !perf.TickOverBudget || perf.TickOverrunMS <= 0 {
		t.Fatalf("tick budget fields = tick %d over %v overrun %.1f", perf.Tick, perf.TickOverBudget, perf.TickOverrunMS)
	}
	if perf.AIMS != 5 || perf.PathfindMS != 2 || perf.CombatMS != 7 || perf.PersistMS != 3 || perf.BroadcastMS != 4 {
		t.Fatalf("phase timings = %+v", perf)
	}
	if perf.PathRequests != 12 || perf.PathCacheHits != 5 || perf.PathNodesVisited != 345 || perf.MonstersMoved != 9 {
		t.Fatalf("path counters = %+v", perf)
	}
	if perf.Changes != 2 || perf.Events != 1 || perf.Acks != 1 || perf.Rejects != 1 || perf.Clients != 3 {
		t.Fatalf("loop counters = %+v", perf)
	}
	if perf.GameLevel != -3 || perf.Entities != 41 || perf.Players != 2 || perf.LiveMonsters != 34 || perf.Walls != 9 {
		t.Fatalf("room shape = %+v", perf)
	}
	if !perf.DegradationApplied {
		t.Fatalf("degradation_applied = false, want true")
	}
}

func assertJSONValue(t *testing.T, payload map[string]any, key string, want any) {
	t.Helper()
	got, ok := payload[key]
	if !ok {
		t.Fatalf("missing JSON key %s in %+v", key, payload)
	}
	if got != want {
		t.Fatalf("%s = %#v, want %#v", key, got, want)
	}
}
