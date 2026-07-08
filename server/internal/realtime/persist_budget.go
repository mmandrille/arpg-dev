package realtime

import (
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const (
	persistPrearmSimThreshold = 25 * time.Millisecond
	persistPrearmChangeCount  = 16
)

func persistTickBudgetExceeded(tickStart time.Time) bool {
	return evaluateTickGuardrail(time.Since(tickStart)).OverBudget
}

func shouldPrearmPersistDefer(results []game.TickResult, simDuration time.Duration) bool {
	if simDuration >= persistPrearmSimThreshold {
		return true
	}
	changes := 0
	for _, res := range results {
		changes += len(res.Changes)
	}

	return changes >= persistPrearmChangeCount
}
