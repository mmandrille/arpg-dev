package realtime

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const defaultPerfDebugInterval = time.Second

type backendTickProfiler struct {
	phases map[string]time.Duration
}

func perfDebugEnabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("ARPG_PERF_DEBUG"))
	return enabled
}

func newBackendTickProfiler() *backendTickProfiler {
	return &backendTickProfiler{phases: map[string]time.Duration{}}
}

func (p *backendTickProfiler) MeasureTickPhase(name string, fn func()) {
	if p == nil {
		fn()
		return
	}
	started := time.Now()
	fn()
	p.phases[name] += time.Since(started)
}

func (p *backendTickProfiler) phaseDuration(name string) time.Duration {
	if p == nil {
		return 0
	}
	return p.phases[name]
}

func durationMS(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}

func logBackendPerf(log *slog.Logger, tick uint64, started time.Time, simDuration, persistDuration, broadcastDuration time.Duration, inputs int, results []game.TickResult, clients int, snapshot game.PerfSnapshot, counters game.PerfCounters, profiler *backendTickProfiler) {
	changes := 0
	events := 0
	acks := 0
	rejects := 0
	for _, res := range results {
		changes += len(res.Changes)
		events += len(res.Events)
		acks += len(res.Acks)
		rejects += len(res.Rejects)
	}
	totalDuration := time.Since(started)
	tickBudget := time.Second / tickHz
	overrun := totalDuration - tickBudget
	tickOverBudget := overrun > 0
	if overrun < 0 {
		overrun = 0
	}
	log.Info("backend_perf",
		"tick", tick,
		"total_ms", durationMS(totalDuration),
		"sim_ms", durationMS(simDuration),
		"ai_ms", durationMS(profiler.phaseDuration(game.TickPhaseAI)),
		"pathfind_ms", durationMS(profiler.phaseDuration(game.TickPhasePathfind)),
		"combat_ms", durationMS(profiler.phaseDuration(game.TickPhaseCombat)),
		"broadcast_ms", durationMS(broadcastDuration),
		"persist_ms", durationMS(persistDuration),
		"path_requests", counters.PathRequests,
		"path_cache_hits", counters.PathCacheHits,
		"path_nodes_visited", counters.PathNodesVisited,
		"monsters_moved", counters.MonstersMoved,
		"tick_budget_ms", durationMS(tickBudget),
		"tick_over_budget", tickOverBudget,
		"tick_overrun_ms", durationMS(overrun),
		"inputs", inputs,
		"results", len(results),
		"changes", changes,
		"events", events,
		"acks", acks,
		"rejects", rejects,
		"clients", clients,
		"game_level", snapshot.Level,
		"entities", snapshot.Entities,
		"players", snapshot.Players,
		"monsters", snapshot.Monsters,
		"live_monsters", snapshot.LiveMonsters,
		"companions", snapshot.Companions,
		"projectiles", snapshot.Projectiles,
		"loot", snapshot.Loot,
		"interactables", snapshot.Interactables,
		"walls", snapshot.Walls,
	)
}
