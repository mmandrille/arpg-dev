// Package metrics owns the Prometheus registry and the metric set required by
// ADR-0001 D8.3. A dedicated (non-global) registry keeps metrics isolated and
// testable. /metrics exposition is served via the Handler.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds every collector emitted by the server.
type Metrics struct {
	reg *prometheus.Registry

	HTTPRequests   *prometheus.CounterVec // method,path,status
	ActiveSessions prometheus.Gauge
	WSConnections  prometheus.Gauge
	TickDuration   prometheus.Histogram // seconds
	MessageLatency prometheus.Histogram // seconds
	RejectedIntents   prometheus.Counter
	PersistenceErrors prometheus.Counter
	ReplayFailures    prometheus.Counter
	ReconciliationDeltas prometheus.Histogram // client-reported distance units
}

// New builds the registry, registers process/Go health collectors, and the
// application metric set.
func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := &Metrics{
		reg: reg,
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "arpg_http_requests_total",
			Help: "Total HTTP requests by method, route, and status.",
		}, []string{"method", "route", "status"}),
		ActiveSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "arpg_active_sessions",
			Help: "Number of active authoritative game sessions.",
		}),
		WSConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "arpg_websocket_connections",
			Help: "Number of open realtime WebSocket connections.",
		}),
		TickDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "arpg_tick_duration_seconds",
			Help:    "Authoritative simulation tick processing time.",
			Buckets: prometheus.ExponentialBuckets(0.0005, 2, 12),
		}),
		MessageLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "arpg_message_latency_seconds",
			Help:    "Server-side time from intent receipt to acknowledgement.",
			Buckets: prometheus.ExponentialBuckets(0.0005, 2, 12),
		}),
		RejectedIntents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "arpg_rejected_intents_total",
			Help: "Total realtime intents rejected by the authoritative server.",
		}),
		PersistenceErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "arpg_persistence_errors_total",
			Help: "Total persistence (Postgres) operation errors.",
		}),
		ReplayFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "arpg_replay_verification_failures_total",
			Help: "Total replay verifications that did not reproduce recorded output.",
		}),
		ReconciliationDeltas: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "arpg_reconciliation_delta",
			Help:    "Client-reported reconciliation distance between predicted and authoritative state.",
			Buckets: prometheus.LinearBuckets(0, 0.5, 12),
		}),
	}

	reg.MustRegister(
		m.HTTPRequests,
		m.ActiveSessions,
		m.WSConnections,
		m.TickDuration,
		m.MessageLatency,
		m.RejectedIntents,
		m.PersistenceErrors,
		m.ReplayFailures,
		m.ReconciliationDeltas,
	)
	return m
}

// Handler returns the Prometheus exposition handler for this registry.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}
