package realtime

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// Hub builds session runners for authenticated WebSocket connections.
type Hub struct {
	store    store.Repository
	rules    *game.Rules
	log      *slog.Logger
	metrics  *metrics.Metrics
	upgrader websocket.Upgrader
}

// NewHub constructs a realtime hub.
func NewHub(st store.Repository, rules *game.Rules, log *slog.Logger, m *metrics.Metrics) *Hub {
	return &Hub{
		store:   st,
		rules:   rules,
		log:     log,
		metrics: m,
		upgrader: websocket.Upgrader{
			// v0 dev default: accept any origin. Remote deployments must
			// restrict this (deferred to the wire-protocol / auth ADRs).
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

// Run upgrades the request to a WebSocket and runs the authoritative session
// loop. The caller must have already validated session ownership.
func (h *Hub) Run(w http.ResponseWriter, r *http.Request, sess store.Session) {
	storedInputs, err := h.store.ListInputs(r.Context(), sess.ID)
	if err != nil {
		h.metrics.PersistenceErrors.Inc()
		http.Error(w, "could not load session inputs", http.StatusInternalServerError)
		return
	}

	sim := game.NewSim(sess.ID, sess.Seed, h.rules)
	var meta *replay.ResumeMetadata
	if len(storedInputs) > 0 {
		recon, err := replay.Reconstruct(r.Context(), h.store, h.rules, sess.ID)
		if err != nil {
			h.metrics.PersistenceErrors.Inc()
			h.log.Error("reconstruct session for websocket resume", "session_id", sess.ID, "error", err)
			http.Error(w, "could not reconstruct session", http.StatusInternalServerError)
			return
		}
		sim = recon.Sim
		meta = &recon.Metadata
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own HTTP error response on failure.
		return
	}

	newRunner(conn, sim, sess, h.store, h.log, h.metrics, meta).run(r.Context())
}
