package realtime

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

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
	mu       sync.Mutex
	loops    map[string]*sessionLoop
}

// NewHub constructs a realtime hub.
func NewHub(st store.Repository, rules *game.Rules, log *slog.Logger, m *metrics.Metrics) *Hub {
	return &Hub{
		store:   st,
		rules:   rules,
		log:     log,
		metrics: m,
		loops:   make(map[string]*sessionLoop),
		upgrader: websocket.Upgrader{
			// v0 dev default: accept any origin. Remote deployments must
			// restrict this (deferred to the wire-protocol / auth ADRs).
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

// Run upgrades the request to a WebSocket and attaches it to the authoritative
// session loop. The caller must have already validated session membership.
func (h *Hub) Run(w http.ResponseWriter, r *http.Request, sess store.Session, member store.SessionMember) {
	loop, err := h.loopForSession(r.Context(), sess)
	if err != nil {
		h.metrics.PersistenceErrors.Inc()
		h.log.Error("load session loop", "session_id", sess.ID, "error", err)
		http.Error(w, "could not load session", http.StatusInternalServerError)
		return
	}
	if loop.hasConnectedMember(memberKey(member)) {
		http.Error(w, "member_already_connected", http.StatusConflict)
		return
	}
	claimed := false
	if isCoopSession(sess) && member.CharacterID != "" {
		ok, err := h.store.ClaimSessionMemberConnection(r.Context(), sess.ID, member.AccountID, member.CharacterID)
		if err != nil {
			h.metrics.PersistenceErrors.Inc()
			h.log.Error("claim session member connection", "session_id", sess.ID, "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
			http.Error(w, "could not claim session member connection", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "member_already_connected", http.StatusConflict)
			return
		}
		claimed = true
		member.Connected = true
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own HTTP error response on failure.
		if claimed {
			_ = h.store.SetSessionMemberDisconnected(context.Background(), sess.ID, member.AccountID, member.CharacterID, member.CurrentLevel, 0)
		}
		return
	}
	loop.attach(r.Context(), conn, member)
}

func (h *Hub) loopForSession(ctx context.Context, sess store.Session) (*sessionLoop, error) {
	h.mu.Lock()
	if loop := h.loops[sess.ID]; loop != nil {
		h.mu.Unlock()
		return loop, nil
	}
	h.mu.Unlock()

	loop, err := newSessionLoop(ctx, h, sess)
	if err != nil {
		return nil, err
	}
	h.mu.Lock()
	if existing := h.loops[sess.ID]; existing != nil {
		h.mu.Unlock()
		loop.stop()
		return existing, nil
	}
	h.loops[sess.ID] = loop
	h.mu.Unlock()
	loop.start()
	return loop, nil
}

func (h *Hub) removeLoop(sessionID string, loop *sessionLoop) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.loops[sessionID] == loop {
		delete(h.loops, sessionID)
	}
}

func progressionStateFromStore(rules *game.Rules, progression *store.CharacterProgression) game.CharacterProgressionState {
	if progression == nil {
		return rules.DefaultCharacterProgressionState()
	}
	return game.CharacterProgressionState{
		Level:               progression.Level,
		Experience:          progression.Experience,
		UnspentStatPoints:   progression.UnspentStatPoints,
		Gold:                progression.Gold,
		DeepestDungeonDepth: progression.DeepestDungeonDepth,
		BaseStats: game.BaseStatsView{
			Str:   progression.Stats.Str,
			Dex:   progression.Stats.Dex,
			Vit:   progression.Stats.Vit,
			Magic: progression.Stats.Magic,
		},
	}
}

func persistedItems(items []store.CharacterItemInstance) []game.PersistedItem {
	out := make([]game.PersistedItem, 0, len(items))
	for _, item := range items {
		if item.Location != store.ItemLocationInventory && item.Location != store.ItemLocationEquipped {
			continue
		}
		out = append(out, game.PersistedItem{
			InstanceID:  item.ID,
			ItemDefID:   item.ItemDefID,
			Slot:        item.Slot,
			Equipped:    item.Equipped,
			RolledStats: item.RolledStats,
		})
	}
	return out
}

func persistedHotbar(slots []store.CharacterHotbarSlot) []game.PersistedHotbarSlot {
	out := make([]game.PersistedHotbarSlot, 0, len(slots))
	for _, slot := range slots {
		out = append(out, game.PersistedHotbarSlot{
			SlotIndex:      slot.SlotIndex,
			ItemInstanceID: slot.ItemInstanceID,
		})
	}
	return out
}

func waypointLevels(waypoints []store.CharacterWaypoint) []int {
	out := make([]int, 0, len(waypoints))
	for _, wp := range waypoints {
		out = append(out, wp.Level)
	}
	return out
}
