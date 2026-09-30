package realtime

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type failingMemberStore struct{ store.Repository }

func (failingMemberStore) SetSessionMemberDisconnected(context.Context, string, string, string, int, int64) error {
	return errors.New("db down")
}

// A dropped member write leaves connected=true and rejects the next attach, so
// the failure must be metered instead of discarded.
func TestMemberDisconnectWriteFailureIsMetered(t *testing.T) {
	h := &Hub{store: failingMemberStore{}, log: slog.New(slog.NewTextHandler(io.Discard, nil)), metrics: metrics.New()}
	before := counterValue(t, h.metrics.PersistenceErrors)
	h.persistMemberDisconnected("sess", "acct", "char", 0, 7)
	if got := counterValue(t, h.metrics.PersistenceErrors) - before; got != 1 {
		t.Fatalf("persistence errors delta = %v, want 1", got)
	}
}

func counterValue(t *testing.T, c prometheus.Counter) float64 {
	t.Helper()
	var m dto.Metric
	if err := c.Write(&m); err != nil {
		t.Fatal(err)
	}
	return m.GetCounter().GetValue()
}
