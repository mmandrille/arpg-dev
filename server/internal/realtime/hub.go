package realtime

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
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
// loop. The caller must have already validated session ownership and supplies
// the session record and its character's persisted inventory.
func (h *Hub) Run(w http.ResponseWriter, r *http.Request, sess store.Session, inventory []store.InventoryItem) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own HTTP error response on failure.
		return
	}

	sim := game.NewSim(sess.ID, sess.Seed, h.rules)
	items := make([]game.PersistedItem, 0, len(inventory))
	for _, it := range inventory {
		items = append(items, game.PersistedItem{
			InstanceID: it.ID,
			ItemDefID:  it.ItemDefID,
			Slot:       it.Slot,
			Equipped:   it.Equipped,
		})
	}
	sim.LoadInventory(items)

	newRunner(conn, sim, sess, h.store, h.log, h.metrics).run(r.Context())
}
