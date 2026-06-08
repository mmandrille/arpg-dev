package httpapi

import (
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerRealtimeRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/ws", s.requireAuth(http.HandlerFunc(s.handleWS)))
}

// handleWS authenticates (via requireAuth), validates session ownership, then
// upgrades to the realtime protocol. The hub restores authoritative state from
// recorded inputs when this is a same-session resume.
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
	member, memberErr := s.store.GetSessionMemberByAccount(r.Context(), sessionID, accountID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && sess.AccountID != accountID && errors.Is(memberErr, store.ErrNotFound)) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if err != nil || (memberErr != nil && !errors.Is(memberErr, store.ErrNotFound)) {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
		return
	}
	if sess.Mode == store.SessionModeCoop && member.CharacterID == "" {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if member.CharacterID == "" {
		member = store.SessionMember{
			SessionID:   sess.ID,
			AccountID:   accountID,
			CharacterID: sess.CharacterID,
			Role:        store.SessionMemberHost,
			Status:      store.SessionMemberActive,
		}
	}

	s.realtime.Run(w, r, sess, member)
}
