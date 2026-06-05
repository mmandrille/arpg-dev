package httpapi

import (
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerRealtimeRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/ws", s.requireAuth(http.HandlerFunc(s.handleWS)))
}

// handleWS authenticates (via requireAuth), validates session ownership, loads
// the character's persisted inventory, then upgrades to the realtime protocol.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "missing session_id query parameter")
		return
	}

	sess, err := s.store.GetSession(r.Context(), sessionID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && sess.AccountID != accountID) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
		return
	}

	inventory, err := s.store.ListInventory(r.Context(), sess.ID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load inventory")
		return
	}

	s.realtime.Run(w, r, sess, inventory)
}
