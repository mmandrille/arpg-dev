package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerSessionRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v0/sessions", s.requireAuth(http.HandlerFunc(s.handleCreateSession)))
	mux.Handle("POST /v0/sessions/{session_id}/end", s.requireAuth(http.HandlerFunc(s.handleEndSession)))
}

type createSessionRequest struct {
	Mode            string  `json:"mode"`
	ResumeSessionID *string `json:"resume_session_id"`
	WorldID         string  `json:"world_id"`
	CharacterID     string  `json:"character_id"`
	Seed            string  `json:"seed"`
}

type createSessionResponse struct {
	SessionID   string `json:"session_id"`
	CharacterID string `json:"character_id"`
	Seed        string `json:"seed"`
	WorldID     string `json:"world_id"`
	WSURL       string `json:"ws_url"`
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}

	var req createSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if req.Mode != "" && req.Mode != "solo" {
		writeError(w, http.StatusBadRequest, "invalid_mode", "only mode \"solo\" is supported in v0")
		return
	}

	ctx := r.Context()

	// Resume path: the session must exist and belong to the caller.
	if req.ResumeSessionID != nil && *req.ResumeSessionID != "" {
		sess, err := s.store.GetSession(ctx, *req.ResumeSessionID)
		if errors.Is(err, store.ErrNotFound) || (err == nil && sess.AccountID != accountID) {
			writeError(w, http.StatusNotFound, "session_not_found", "session not found")
			return
		}
		if err != nil {
			s.metrics.PersistenceErrors.Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
			return
		}
		if err := s.store.TouchSession(ctx, sess.ID); err != nil {
			s.metrics.PersistenceErrors.Inc()
		}
		writeJSON(w, http.StatusOK, sessionResponse(sess))
		return
	}

	// Create path: use a selected character when provided, otherwise preserve
	// the default-character compatibility path for bots, smoke, and dev flows.
	worldID := req.WorldID
	if worldID == "" {
		worldID = game.DefaultWorldID
	}
	if s.rules == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "rules not loaded")
		return
	}
	if _, ok := s.rules.Worlds[worldID]; !ok {
		writeError(w, http.StatusBadRequest, "invalid_world_id", "unknown world_id")
		return
	}

	char, err := s.characterForSessionCreate(ctx, accountID, req.CharacterID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "character_not_found", "character not found")
		return
	case err != nil:
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character")
		return
	}

	seed := req.Seed
	if seed != "" && !s.cfg.IsLocal() {
		writeError(w, http.StatusBadRequest, "invalid_seed", "custom session seeds are only available in local development")
		return
	}
	if seed == "" {
		var err error
		seed, err = newSeed()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not generate seed")
			return
		}
	}

	sess := store.Session{
		ID:          ids.New("sess"),
		AccountID:   accountID,
		CharacterID: char.ID,
		Seed:        seed,
		WorldID:     worldID,
		Status:      store.SessionActive,
	}
	if err := s.store.CreateSession(ctx, sess); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}
	items, err := s.store.ListCharacterItems(ctx, accountID, char.ID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character items")
		return
	}
	progression, err := s.store.GetOrCreateCharacterProgression(ctx, accountID, char.ID, progressionDefaultsFromRules(s.rules))
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character progression")
		return
	}
	waypoints, err := s.store.ListCharacterWaypoints(ctx, char.ID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character waypoints")
		return
	}
	if err := s.store.CreateSessionStartSnapshot(ctx, sess.ID, accountID, char.ID, items, waypoints, progression); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session start snapshot")
		return
	}

	writeJSON(w, http.StatusCreated, sessionResponse(sess))
}

func (s *Server) characterForSessionCreate(ctx context.Context, accountID, requestedCharacterID string) (store.Character, error) {
	if requestedCharacterID == "" {
		return s.store.GetOrCreateDefaultCharacter(ctx, ids.New("char"), accountID, "Hero")
	}
	char, err := s.store.GetCharacter(ctx, requestedCharacterID)
	if err != nil {
		return store.Character{}, err
	}
	if char.AccountID != accountID {
		return store.Character{}, store.ErrNotFound
	}
	return char, nil
}

func (s *Server) handleEndSession(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
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
	if err := s.store.SetSessionStatus(r.Context(), sessionID, store.SessionEnded); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not end session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": store.SessionEnded})
}

func sessionResponse(sess store.Session) createSessionResponse {
	return createSessionResponse{
		SessionID:   sess.ID,
		CharacterID: sess.CharacterID,
		Seed:        sess.Seed,
		WorldID:     sess.WorldID,
		WSURL:       "/v0/ws?session_id=" + sess.ID,
	}
}

func progressionDefaultsFromRules(rules *game.Rules) store.CharacterProgressionDefaults {
	state := rules.DefaultCharacterProgressionState()
	return store.CharacterProgressionDefaults{
		Level:             state.Level,
		Experience:        state.Experience,
		UnspentStatPoints: state.UnspentStatPoints,
		Stats: store.CharacterBaseStats{
			Str:   state.BaseStats.Str,
			Dex:   state.BaseStats.Dex,
			Vit:   state.BaseStats.Vit,
			Magic: state.BaseStats.Magic,
		},
	}
}

// newSeed returns a fresh 128-bit hex-encoded server seed from OS entropy.
func newSeed() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
