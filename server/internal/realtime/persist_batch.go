package realtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func shouldDeferNonCritical(simDuration, elapsed time.Duration, persistDuration time.Duration) bool {
	if evaluateTickGuardrail(simDuration).OverBudget {
		return true
	}
	if evaluateTickGuardrail(elapsed).OverBudget {
		return true
	}
	if evaluateTickGuardrail(persistDuration).OverBudget {
		return true
	}

	return false
}

func (l *sessionLoop) appendSessionEvents(ctx context.Context, res game.TickResult, events []game.Event, eventSequence int64) int64 {
	if len(events) == 0 {
		return eventSequence
	}
	records := make([]store.SessionEvent, 0, len(events))
	for _, ev := range events {
		payload, _ := json.Marshal(ev)
		records = append(records, store.SessionEvent{
			ID:            ids.New("evt"),
			SessionID:     l.sess.ID,
			Tick:          int64(res.Tick),
			Sequence:      eventSequence,
			EventType:     ev.EventType,
			CorrelationID: ev.CorrelationID,
			Payload:       payload,
		})
		eventSequence++
	}
	if err := l.hub.store.AppendEvents(ctx, records); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist events batch", "count", len(records), "error", err)
	}

	return eventSequence
}

func (l *sessionLoop) persistChangeOrDefer(c game.Change, member store.SessionMember, deferNonCritical bool) {
	if deferNonCritical && persistChangeDeferrable(c.Op) {
		l.deferredPersistChanges = append(l.deferredPersistChanges, deferredPersistChange{change: c, member: member})
		return
	}
	l.persistChange(context.Background(), c, member)
}
