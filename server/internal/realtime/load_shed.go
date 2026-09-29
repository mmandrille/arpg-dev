package realtime

import (
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// loadShedSample is the wall-clock and sim evidence gathered after a tick.
type loadShedSample struct {
	guardrail   tickGuardrailDecision
	simDuration time.Duration
	counters    game.PerfCounters
	snapshot    game.PerfSnapshot
	nav         game.NavigationRules
	profiler    *backendTickProfiler
}

// loadShedPolicy turns a post-tick sample into a directive. Production uses
// wallClockLoadShed; tests inject deterministic policies.
type loadShedPolicy func(loadShedSample) game.LoadShedDirective

func wallClockLoadShed(s loadShedSample) game.LoadShedDirective {
	degrade := s.guardrail.OverBudget &&
		(shouldApplyOverloadDegradation(s.counters, s.snapshot, s.nav) ||
			shouldApplySimPressureOverloadDegradation(s.simDuration, s.guardrail.Budget, s.snapshot, s.nav))
	return game.LoadShedDirective{
		OverloadDegrade:        degrade,
		CombatMovementThrottle: combatPhaseOverBudget(s.profiler, game.CombatPhaseBudgetForTick()),
	}
}

// applyLoadShed applies the directive for the tick that just finished and, when
// it changes sim state, returns the input row that makes it replayable. The
// caller must hold l.mu and persist the row after unlocking.
func (l *sessionLoop) applyLoadShed(tick uint64, sample loadShedSample) (bool, *store.SessionInput) {
	policy := l.loadShed
	if policy == nil {
		policy = wallClockLoadShed
	}
	d := policy(sample)
	if l.sim == nil || !l.sim.LoadShedChanges(d) {
		return false, nil
	}
	messageID := ids.New("sys")
	payload, err := inputdecode.EncodeStoredLoadShed(messageID, d)
	if err != nil {
		// Unrecordable state changes would break replay; skip shedding instead.
		l.log.Error("encode load shed input", "tick", tick, "error", err)
		return false, nil
	}
	applied := l.sim.ApplyLoadShed(d)
	return applied, l.systemInputRowLocked(int64(tick), messageID, payload)
}
