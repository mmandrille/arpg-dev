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
	store         store.Repository
	rules         *game.Rules
	log           *slog.Logger
	metrics       *metrics.Metrics
	gameplayDebug bool
	upgrader      websocket.Upgrader
	mu            sync.Mutex
	loops         map[string]*sessionLoop
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

// SetGameplayDebug enables local development-only gameplay conveniences for
// sessions built by this hub.
func (h *Hub) SetGameplayDebug(enabled bool) {
	h.gameplayDebug = enabled
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
			h.persistMemberDisconnected(sess.ID, member.AccountID, member.CharacterID, member.CurrentLevel, 0)
		}
		return
	}
	for attempt := 0; !loop.attach(r.Context(), conn, member); attempt++ {
		// The loop stopped between lookup and attach: resume on its successor.
		if loop, err = h.loopForSession(r.Context(), sess); err != nil || attempt >= 2 {
			h.log.Error("attach to session loop", "session_id", sess.ID, "attempt", attempt, "error", err)
			if claimed {
				h.persistMemberDisconnected(sess.ID, member.AccountID, member.CharacterID, member.CurrentLevel, 0)
			}
			_ = conn.Close()
			return
		}
	}
}

func (h *Hub) loopForSession(ctx context.Context, sess store.Session) (*sessionLoop, error) {
	h.mu.Lock()
	loop := h.loops[sess.ID]
	h.mu.Unlock()
	if loop != nil {
		if !loop.stopping() {
			return loop, nil
		}
		// Its last client just left. Wait for it to record its final tick
		// (v481), or the successor would resume from storage behind it.
		<-loop.stopDone
		h.removeLoop(sess.ID, loop)
	}

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

func storeShopStock(accountID, characterID string, items []game.PersistedShopStockItem) []store.CharacterShopStockItem {
	out := make([]store.CharacterShopStockItem, 0, len(items))
	for _, item := range items {
		out = append(out, store.CharacterShopStockItem{
			AccountID:      accountID,
			CharacterID:    characterID,
			ShopID:         item.ShopID,
			RefreshKey:     item.RefreshKey,
			OfferIndex:     item.OfferIndex,
			OfferID:        item.OfferID,
			SourceDepth:    item.SourceDepth,
			ItemTemplateID: item.ItemTemplateID,
			RolledPayload:  item.RolledPayload,
			BuyPrice:       item.BuyPrice,
			Available:      item.Available,
		})
	}
	return out
}
