package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// registerInspectRoutes wires the debug-gated inspection + replay endpoints.
func (s *Server) registerInspectRoutes(mux *http.ServeMux) {
	mux.Handle("PUT /v0/debug/characters/{character_id}/progression",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleDebugCharacterProgression))))
	mux.Handle("GET /v0/sessions/{session_id}/state",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionState))))
	mux.Handle("GET /v0/sessions/{session_id}/replay",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionReplay))))
	mux.Handle("GET /v0/sessions/{session_id}/replay/timeline",
		s.requireAuth(s.requireDebug(http.HandlerFunc(s.handleSessionReplayTimeline))))
}

type debugCharacterProgressionRequest struct {
	Level              int                      `json:"level"`
	Experience         int                      `json:"experience"`
	UnspentStatPoints  int                      `json:"unspent_stat_points"`
	UnspentSkillPoints int                      `json:"unspent_skill_points"`
	Stats              store.CharacterBaseStats `json:"stats"`
	SkillRanks         map[string]int           `json:"skill_ranks"`
}

func (s *Server) handleDebugCharacterProgression(w http.ResponseWriter, r *http.Request) {
	accountID, _ := accountFromContext(r.Context())
	characterID := r.PathValue("character_id")
	if !safePathID(characterID) {
		writeError(w, http.StatusNotFound, "character_not_found", "character not found")
		return
	}
	character, err := s.store.GetCharacter(r.Context(), characterID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && character.AccountID != accountID) {
		writeError(w, http.StatusNotFound, "character_not_found", "character not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character")
		return
	}

	var req debugCharacterProgressionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if req.Level < 1 || req.Experience < 0 || req.UnspentStatPoints < 0 || req.UnspentSkillPoints < 0 {
		writeError(w, http.StatusBadRequest, "invalid_progression", "progression values are out of range")
		return
	}
	for skillID, rank := range req.SkillRanks {
		if skillID == "" || rank < 0 {
			writeError(w, http.StatusBadRequest, "invalid_progression", "skill ranks are out of range")
			return
		}
	}

	progression := store.CharacterProgression{
		AccountID:           accountID,
		CharacterID:         characterID,
		Level:               req.Level,
		Experience:          req.Experience,
		UnspentStatPoints:   req.UnspentStatPoints,
		UnspentSkillPoints:  req.UnspentSkillPoints,
		Stats:               req.Stats,
		Gold:                0,
		DeepestDungeonDepth: 0,
		SkillRanks:          req.SkillRanks,
	}
	if err := s.store.UpsertCharacterProgression(r.Context(), accountID, progression); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not seed character progression")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	throughTick := int64(-1)
	if raw := r.URL.Query().Get("through_tick"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid_request", "through_tick must be a non-negative integer")
			return
		}
		throughTick = parsed
	}
	timeline, err := replay.BuildTimeline(r.Context(), s.store, s.rules, sess.ID, throughTick)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not build replay timeline")
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}
