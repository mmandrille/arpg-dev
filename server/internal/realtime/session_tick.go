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
	delete(l.buffer, tick)
	inputs = append(inputs, l.stashSyncInputsLocked(tick)...)
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
	elapsedAfterSim := time.Since(start)
	deferNonCritical := shouldDeferNonCritical(simDuration, elapsedAfterSim, 0) || shouldPrearmPersistDefer(results, simDuration)
	for _, res := range results {
		if shouldDeferNonCritical(simDuration, time.Since(start), persistDuration) {
			deferNonCritical = true
		}
		persistStart := time.Now()
		eventSequence = l.persistTick(res, membersByPlayerID, eventSequence, deferNonCritical, start)
		persistDuration += time.Since(persistStart)
	}
	broadcastStart := time.Now()
	l.fanoutTickResults(results, clients, levelsByPlayerID)
	broadcastDuration = time.Since(broadcastStart)
	totalDuration := time.Since(start)
	guardrail := evaluateTickGuardrail(totalDuration)
	// Wall-clock load shedding mutates the sim, so it is recorded as a
	// server-authored input for this tick (v476) to keep replay exact.
	l.mu.Lock()
	degradationApplied, loadShedInput := l.applyLoadShed(tick, loadShedSample{
		guardrail:   guardrail,
		simDuration: simDuration,
		counters:    counters,
		snapshot:    snapshot,
		nav:         nav,
		profiler:    profiler,
	})
	if resultsHaveEvents(results) {
		l.noteDurableLocked(int64(tick))
	}
	// Quiet ticks leave no row, so replay would stop short of them (v481).
	checkpoint := l.quietCheckpointLocked(tick)
	l.mu.Unlock()
	l.persistSystemInput(loadShedInput)
	l.persistSystemInput(checkpoint)
	if guardrail.OverBudget {
		logTickBudgetWarning(l.log, tick, totalDuration, guardrail, simDuration, persistDuration, broadcastDuration, len(inputs), results, len(clients), snapshot, counters, degradationApplied)
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

func resultsHaveEvents(results []game.TickResult) bool {
	for _, res := range results {
		if len(res.Events) > 0 {
			return true
		}
	}
	return false
}
