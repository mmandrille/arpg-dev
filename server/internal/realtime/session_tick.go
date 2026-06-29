package realtime

import (
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (l *sessionLoop) doTick() {
	start := time.Now()
	l.mu.Lock()
	l.flushDeferredPersist()
	tick := l.sim.CurrentTick()
	inputs := l.buffer[tick]
	inputTypes := make(map[string]string, len(inputs))
	for _, in := range inputs {
		inputTypes[in.MessageID] = in.Type
	}
	delete(l.buffer, tick)
	sortInputs(inputs)
	simStart := time.Now()
	profiler := newBackendTickProfiler()
	results := l.sim.TickResultsProfiled(inputs, profiler)
	simDuration := time.Since(simStart)
	snapshot := l.sim.PerfSnapshot()
	counters := l.sim.PerfCounters()
	nav := l.sim.ActiveNavigationRules()
	latencies := []time.Duration{}
	for _, res := range results {
		for _, ack := range res.Acks {
			if recv, ok := l.received[ack.MessageID]; ok {
				latencies = append(latencies, time.Since(recv))
				delete(l.received, ack.MessageID)
			}
		}
	}
	clients := make([]*loopClient, 0, len(l.clients))
	membersByPlayerID := make(map[uint64]store.SessionMember, len(l.clients))
	levelsByPlayerID := make(map[uint64]int, len(l.clients))
	for _, client := range l.clients {
		clients = append(clients, client)
		membersByPlayerID[client.playerID] = client.member
		if level, ok := l.sim.PlayerCurrentLevel(client.playerID); ok {
			levelsByPlayerID[client.playerID] = level
		}
	}
	l.mu.Unlock()
	l.hub.metrics.TickDuration.Observe(time.Since(start).Seconds())
	for _, latency := range latencies {
		l.hub.metrics.MessageLatency.Observe(latency.Seconds())
	}
	eventSequence := int64(0)
	persistDuration := time.Duration(0)
	broadcastDuration := time.Duration(0)
	simGuardrail := evaluateTickGuardrail(simDuration)
	deferNonCritical := simGuardrail.OverBudget
	for _, res := range results {
		persistStart := time.Now()
		eventSequence = l.persistTick(res, membersByPlayerID, eventSequence, deferNonCritical)
		persistDuration += time.Since(persistStart)
		broadcastStart := time.Now()
		l.fanoutResult(res, clients, inputTypes, levelsByPlayerID)
		broadcastDuration += time.Since(broadcastStart)
	}
	totalDuration := time.Since(start)
	guardrail := evaluateTickGuardrail(totalDuration)
	combatBudget := game.CombatPhaseBudgetForTick()
	degradationApplied := false
	if guardrail.OverBudget {
		l.mu.Lock()
		if l.sim != nil && shouldApplyOverloadDegradation(counters, snapshot, nav) {
			degradationApplied = l.sim.ApplyOverloadDegradation()
		}
		if l.sim != nil {
			l.sim.SetCombatMovementThrottle(degradationApplied || combatPhaseOverBudget(profiler, combatBudget))
		}
		l.mu.Unlock()
		logTickBudgetWarning(l.log, tick, totalDuration, guardrail, simDuration, persistDuration, broadcastDuration, len(inputs), results, len(clients), snapshot, counters, degradationApplied)
	} else {
		l.mu.Lock()
		if l.sim != nil {
			l.sim.SetCombatMovementThrottle(combatPhaseOverBudget(profiler, combatBudget))
		}
		l.mu.Unlock()
	}
	if l.perfDebug && time.Since(l.lastPerfLog) >= defaultPerfDebugInterval {
		l.lastPerfLog = time.Now()
		logBackendPerf(l.log, tick, start, simDuration, persistDuration, broadcastDuration, len(inputs), results, len(clients), snapshot, counters, profiler)
	}
	if time.Since(l.lastPerfStatus) >= defaultPerfDebugInterval {
		l.lastPerfStatus = time.Now()
		perf := buildPerformanceStatus(tick, totalDuration, simDuration, persistDuration, broadcastDuration, len(inputs), results, len(clients), snapshot, counters, profiler, degradationApplied)
		l.fanoutPerformanceStatus(perf, clients, levelsByPlayerID)
	}
}
