package realtime

import (
	"log/slog"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const sessionTickBudgetOverrunMessage = "session_tick_budget_overrun"

type tickGuardrailDecision struct {
	OverBudget bool
	Budget     time.Duration
	Overrun    time.Duration
}

func evaluateTickGuardrail(total time.Duration) tickGuardrailDecision {
	budget := time.Second / tickHz
	overrun := total - budget
	if overrun <= 0 {
		return tickGuardrailDecision{Budget: budget}
	}
	return tickGuardrailDecision{OverBudget: true, Budget: budget, Overrun: overrun}
}

func shouldApplyOverloadDegradation(counters game.PerfCounters, snapshot game.PerfSnapshot, nav game.NavigationRules) bool {
	if counters.PathNodesVisited >= nav.MonsterPathNodesPerTick {
		return true
	}
	if snapshot.LiveMonsters >= nav.MonsterOverloadLiveMonsterThreshold && counters.PathRequests > 0 {
		return true
	}

	return counters.PathRequests > 0 ||
		counters.PathCacheHits > 0 ||
		counters.PathNodesVisited > 0 ||
		counters.MonstersMoved > 0
}

func combatPhaseOverBudget(profiler *backendTickProfiler, budget time.Duration) bool {
	if profiler == nil {
		return false
	}
	return profiler.phaseDuration(game.TickPhaseCombat) > budget
}

func logTickBudgetWarning(log *slog.Logger, tick uint64, total time.Duration, decision tickGuardrailDecision, simDuration, persistDuration, broadcastDuration time.Duration, inputs int, results []game.TickResult, clients int, snapshot game.PerfSnapshot, counters game.PerfCounters, degradationApplied bool) {
	if log == nil || !decision.OverBudget {
		return
	}
	changes := 0
	events := 0
	for _, res := range results {
		changes += len(res.Changes)
		events += len(res.Events)
	}
	log.Warn(sessionTickBudgetOverrunMessage,
		"tick", tick,
		"total_ms", durationMS(total),
		"sim_ms", durationMS(simDuration),
		"persist_ms", durationMS(persistDuration),
		"broadcast_ms", durationMS(broadcastDuration),
		"tick_budget_ms", durationMS(decision.Budget),
		"tick_overrun_ms", durationMS(decision.Overrun),
		"inputs", inputs,
		"results", len(results),
		"changes", changes,
		"events", events,
		"clients", clients,
		"live_monsters", snapshot.LiveMonsters,
		"walls", snapshot.Walls,
		"path_requests", counters.PathRequests,
		"path_cache_hits", counters.PathCacheHits,
		"path_nodes_visited", counters.PathNodesVisited,
		"monsters_moved", counters.MonstersMoved,
		"degradation_applied", degradationApplied,
	)
}
