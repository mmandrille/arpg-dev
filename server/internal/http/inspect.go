package httpapi

import (
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// registerInspectRoutes wires the debug-gated inspection + replay endpoints.
func (s *Server) registerInspectRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/sessions/{session_id}/state",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionState))))
	mux.Handle("GET /v0/sessions/{session_id}/replay",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionReplay))))
	mux.Handle("GET /v0/sessions/{session_id}/replay/timeline",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionReplayTimeline))))
}

// requireDebug enforces the X-Debug-Token header (ADR-0001 D8.4 / spec 4.1).
// It must wrap a handler already protected by requireAuth.
func (s *Server) requireDebug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Debug-Token") != s.cfg.DebugToken {
			writeError(w, http.StatusForbidden, "forbidden", "missing or invalid X-Debug-Token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loadOwnedSession loads a session and verifies it belongs to the caller.
func (s *Server) loadOwnedSession(w http.ResponseWriter, r *http.Request) (store.Session, bool) {
	accountID, _ := accountFromContext(r.Context())
	sessionID := r.PathValue("session_id")
	if !safePathID(sessionID) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return store.Session{}, false
	}
	sess, err := s.store.GetSession(r.Context(), sessionID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && sess.AccountID != accountID) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return store.Session{}, false
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
		return store.Session{}, false
	}
	return sess, true
}

func safePathID(id string) bool {
	if len(id) == 0 || len(id) > 80 {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

// handleSessionState returns the current authoritative state, reconstructed
// from the recorded seed + input stream.
func (s *Server) handleSessionState(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadOwnedSession(w, r)
	if !ok {
		return
	}
	recon, err := replay.Reconstruct(r.Context(), s.store, s.rules, sess.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not reconstruct state")
		return
	}
	writeJSON(w, http.StatusOK, recon.Snapshot)
}

// handleSessionReplay returns replay metadata plus the latest verification
// result (re-computed on request).
func (s *Server) handleSessionReplay(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadOwnedSession(w, r)
	if !ok {
		return
	}
	report, err := replay.Verify(r.Context(), s.store, s.rules, sess.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not verify replay")
		return
	}
	if !report.Match {
		s.metrics.ReplayFailures.Inc()
	}
	writeJSON(w, http.StatusOK, report)
}

// handleSessionReplayTimeline returns protocol-shaped replay envelopes for
// local visual tooling.
func (s *Server) handleSessionReplayTimeline(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.loadOwnedSession(w, r)
	if !ok {
		return
	}
	timeline, err := replay.BuildTimeline(r.Context(), s.store, s.rules, sess.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not build replay timeline")
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}
