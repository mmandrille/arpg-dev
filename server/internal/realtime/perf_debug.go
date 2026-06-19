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

type tickResultSummary struct {
	Changes int
	Events  int
	Acks    int
	Rejects int
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

func summarizeTickResults(results []game.TickResult) tickResultSummary {
	summary := tickResultSummary{}
	for _, res := range results {
		summary.Changes += len(res.Changes)
		summary.Events += len(res.Events)
		summary.Acks += len(res.Acks)
		summary.Rejects += len(res.Rejects)
	}
	return summary
}

func buildPerformanceStatus(tick uint64, totalDuration, simDuration, persistDuration, broadcastDuration time.Duration, inputs int, results []game.TickResult, clients int, snapshot game.PerfSnapshot, counters game.PerfCounters, profiler *backendTickProfiler, degradationApplied bool) performanceStatusPayload {
	summary := summarizeTickResults(results)
	guardrail := evaluateTickGuardrail(totalDuration)
	return performanceStatusPayload{
		Tick:               tick,
		TotalMS:            durationMS(totalDuration),
		SimMS:              durationMS(simDuration),
		AIMS:               durationMS(profiler.phaseDuration(game.TickPhaseAI)),
		PathfindMS:         durationMS(profiler.phaseDuration(game.TickPhasePathfind)),
		CombatMS:           durationMS(profiler.phaseDuration(game.TickPhaseCombat)),
		BroadcastMS:        durationMS(broadcastDuration),
		PersistMS:          durationMS(persistDuration),
		PathRequests:       counters.PathRequests,
		PathCacheHits:      counters.PathCacheHits,
		PathNodesVisited:   counters.PathNodesVisited,
		MonstersMoved:      counters.MonstersMoved,
		TickBudgetMS:       durationMS(guardrail.Budget),
		TickOverBudget:     guardrail.OverBudget,
		TickOverrunMS:      durationMS(guardrail.Overrun),
		DegradationApplied: degradationApplied,
		Inputs:             inputs,
		Results:            len(results),
		Changes:            summary.Changes,
		Events:             summary.Events,
		Acks:               summary.Acks,
		Rejects:            summary.Rejects,
		Clients:            clients,
		GameLevel:          snapshot.Level,
		Entities:           snapshot.Entities,
		Players:            snapshot.Players,
		Monsters:           snapshot.Monsters,
		LiveMonsters:       snapshot.LiveMonsters,
		Companions:         snapshot.Companions,
		Projectiles:        snapshot.Projectiles,
		Loot:               snapshot.Loot,
		Interactables:      snapshot.Interactables,
		Walls:              snapshot.Walls,
	}
}

func logBackendPerf(log *slog.Logger, tick uint64, started time.Time, simDuration, persistDuration, broadcastDuration time.Duration, inputs int, results []game.TickResult, clients int, snapshot game.PerfSnapshot, counters game.PerfCounters, profiler *backendTickProfiler) {
	totalDuration := time.Since(started)
	perf := buildPerformanceStatus(tick, totalDuration, simDuration, persistDuration, broadcastDuration, inputs, results, clients, snapshot, counters, profiler, false)
	log.Info("backend_perf",
		"tick", perf.Tick,
		"total_ms", perf.TotalMS,
		"sim_ms", perf.SimMS,
		"ai_ms", perf.AIMS,
		"pathfind_ms", perf.PathfindMS,
		"combat_ms", perf.CombatMS,
		"broadcast_ms", perf.BroadcastMS,
		"persist_ms", perf.PersistMS,
		"path_requests", perf.PathRequests,
		"path_cache_hits", perf.PathCacheHits,
		"path_nodes_visited", perf.PathNodesVisited,
		"monsters_moved", perf.MonstersMoved,
		"tick_budget_ms", perf.TickBudgetMS,
		"tick_over_budget", perf.TickOverBudget,
		"tick_overrun_ms", perf.TickOverrunMS,
		"inputs", perf.Inputs,
		"results", perf.Results,
		"changes", perf.Changes,
		"events", perf.Events,
		"acks", perf.Acks,
		"rejects", perf.Rejects,
		"clients", perf.Clients,
		"game_level", perf.GameLevel,
		"entities", perf.Entities,
		"players", perf.Players,
		"monsters", perf.Monsters,
		"live_monsters", perf.LiveMonsters,
		"companions", perf.Companions,
		"projectiles", perf.Projectiles,
		"loot", perf.Loot,
		"interactables", perf.Interactables,
		"walls", perf.Walls,
	)
}
