package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerSessionRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v0/sessions", s.requireAuth(http.HandlerFunc(s.handleCreateSession)))
	mux.Handle("GET /v0/sessions/active", s.requireAuth(http.HandlerFunc(s.handleListActiveSessions)))
	mux.Handle("POST /v0/sessions/{session_id}/join", s.requireAuth(http.HandlerFunc(s.handleJoinSession)))
	mux.Handle("POST /v0/sessions/{session_id}/end", s.requireAuth(http.HandlerFunc(s.handleEndSession)))
}

type createSessionRequest struct {
	Mode            string  `json:"mode"`
	ResumeSessionID *string `json:"resume_session_id"`
	WorldID         string  `json:"world_id"`
	CharacterID     string  `json:"character_id"`
	Seed            string  `json:"seed"`
	Listed          bool    `json:"listed"`
}

type createSessionResponse struct {
	SessionID   string `json:"session_id"`
	CharacterID string `json:"character_id"`
	Seed        string `json:"seed"`
	WorldID     string `json:"world_id"`
	Mode        string `json:"mode"`
	Listed      bool   `json:"listed"`
	JoinCode    string `json:"join_code,omitempty"`
	WSURL       string `json:"ws_url"`
}

type activeSessionSummaryResponse struct {
	SessionID       string `json:"session_id"`
	WorldID         string `json:"world_id"`
	Mode            string `json:"mode"`
	Listed          bool   `json:"listed"`
	HostCharacterID string `json:"host_character_id"`
	HostDisplayName string `json:"host_display_name"`
	MemberCount     int    `json:"member_count"`
	ConnectedCount  int    `json:"connected_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type activeSessionsResponse struct {
	Sessions []activeSessionSummaryResponse `json:"sessions"`
}

type joinSessionRequest struct {
	JoinCode    string `json:"join_code"`
	CharacterID string `json:"character_id"`
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
	mode := req.Mode
	if mode == "" {
		mode = store.SessionModeSolo
	}
	if mode != store.SessionModeSolo && mode != store.SessionModeCoop {
		writeError(w, http.StatusBadRequest, "invalid_mode", "mode must be \"solo\" or \"coop\"")
		return
	}

	ctx := r.Context()

	// Resume path: the session must exist and belong to the caller.
	if req.ResumeSessionID != nil && *req.ResumeSessionID != "" {
		sess, err := s.store.GetSession(ctx, *req.ResumeSessionID)
		member, memberErr := s.store.GetSessionMemberByAccount(ctx, *req.ResumeSessionID, accountID)
		if errors.Is(err, store.ErrNotFound) || (err == nil && sess.AccountID != accountID && errors.Is(memberErr, store.ErrNotFound)) {
			writeError(w, http.StatusNotFound, "session_not_found", "session not found")
			return
		}
		if err != nil || (memberErr != nil && !errors.Is(memberErr, store.ErrNotFound)) {
			s.metrics.PersistenceErrors.Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
			return
		}
		if err := s.store.TouchSession(ctx, sess.ID); err != nil {
			s.metrics.PersistenceErrors.Inc()
		}
		characterID := sess.CharacterID
		if member.CharacterID != "" {
			characterID = member.CharacterID
		}
		writeJSON(w, http.StatusOK, sessionResponse(sess, characterID, ""))
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

	var joinCode string
	var joinHash string
	if mode == store.SessionModeCoop {
		var err error
		joinCode, err = newJoinCode()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not generate join code")
			return
		}
		joinHash = hashJoinCode(joinCode)
	}

	sess := store.Session{
		ID:           ids.New("sess"),
		AccountID:    accountID,
		CharacterID:  char.ID,
		Seed:         seed,
		WorldID:      worldID,
		Mode:         mode,
		Listed:       mode == store.SessionModeCoop && req.Listed,
		JoinCodeHash: joinHash,
		Status:       store.SessionActive,
	}
	if err := s.store.CreateSession(ctx, sess); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}
	if err := s.store.CreateSessionHostMember(ctx, store.SessionMember{
		SessionID:    sess.ID,
		AccountID:    accountID,
		CharacterID:  char.ID,
		Role:         store.SessionMemberHost,
		Status:       store.SessionMemberActive,
		CurrentLevel: 0,
	}); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session member")
		return
	}
	if !s.createSessionStartSnapshot(w, ctx, sess.ID, accountID, char.ID) {
		return
	}

	writeJSON(w, http.StatusCreated, sessionResponse(sess, char.ID, joinCode))
}

func (s *Server) handleListActiveSessions(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.ListActiveListedSessions(r.Context())
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list sessions")
		return
	}
	res := activeSessionsResponse{Sessions: make([]activeSessionSummaryResponse, 0, len(summaries))}
	for _, summary := range summaries {
		res.Sessions = append(res.Sessions, activeSessionSummaryResponse{
			SessionID:       summary.SessionID,
			WorldID:         summary.WorldID,
			Mode:            summary.Mode,
			Listed:          summary.Listed,
			HostCharacterID: summary.HostCharacterID,
			HostDisplayName: summary.HostDisplayName,
			MemberCount:     summary.MemberCount,
			ConnectedCount:  summary.ConnectedCount,
			CreatedAt:       summary.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt:       summary.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	writeJSON(w, http.StatusOK, res)
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

func (s *Server) handleJoinSession(w http.ResponseWriter, r *http.Request) {
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
	var req joinSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if req.CharacterID == "" {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}

	ctx := r.Context()
	sess, err := s.store.GetSession(ctx, sessionID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load session")
		return
	}
	if sess.Mode != store.SessionModeCoop {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if sess.Status == store.SessionEnded {
		writeError(w, http.StatusConflict, "session_ended", "session has ended")
		return
	}
	if !sess.Listed {
		if req.JoinCode == "" || sess.JoinCodeHash == "" || !joinCodeMatches(sess.JoinCodeHash, req.JoinCode) {
			writeError(w, http.StatusNotFound, "session_not_found", "session not found")
			return
		}
	}
	char, err := s.store.GetCharacter(ctx, req.CharacterID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && char.AccountID != accountID) {
		writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character")
		return
	}
	if err := s.store.CreateSessionGuestMember(ctx, store.SessionMember{
		SessionID:    sess.ID,
		AccountID:    accountID,
		CharacterID:  char.ID,
		Role:         store.SessionMemberGuest,
		Status:       store.SessionMemberActive,
		CurrentLevel: 0,
	}); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			writeError(w, http.StatusConflict, "duplicate_member", "account or character already joined")
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "session_not_found", "session not found")
		default:
			s.metrics.PersistenceErrors.Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "could not join session")
		}
		return
	}
	if !s.createSessionStartSnapshot(w, ctx, sess.ID, accountID, char.ID) {
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse(sess, char.ID, ""))
}

func (s *Server) createSessionStartSnapshot(w http.ResponseWriter, ctx context.Context, sessionID, accountID, characterID string) bool {
	items, err := s.store.ListCharacterItems(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character items")
		return false
	}
	progression, err := s.store.GetOrCreateCharacterProgression(ctx, accountID, characterID, progressionDefaultsFromRules(s.rules))
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character progression")
		return false
	}
	waypoints, err := s.store.ListCharacterWaypoints(ctx, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character waypoints")
		return false
	}
	hotbar, err := s.store.ListCharacterHotbar(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character hotbar")
		return false
	}
	if err := s.store.CreateSessionStartSnapshot(ctx, sessionID, accountID, characterID, items, waypoints, hotbar, progression); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session start snapshot")
		return false
	}
	return true
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
	if sess.Mode == store.SessionModeCoop {
		if member.CharacterID == "" {
			member = store.SessionMember{SessionID: sess.ID, AccountID: sess.AccountID, CharacterID: sess.CharacterID}
		}
		if err := s.store.SetSessionMemberDisconnected(r.Context(), sess.ID, member.AccountID, member.CharacterID, member.CurrentLevel, 0); err != nil && !errors.Is(err, store.ErrNotFound) {
			s.metrics.PersistenceErrors.Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "could not leave session")
			return
		}
		members, err := s.store.ListSessionMembers(r.Context(), sess.ID)
		if err != nil {
			s.metrics.PersistenceErrors.Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "could not load session members")
			return
		}
		anyConnected := false
		for _, m := range members {
			if m.Connected {
				anyConnected = true
				break
			}
		}
		if !anyConnected {
			if err := s.store.SetSessionStatus(r.Context(), sessionID, store.SessionEnded); err != nil {
				s.metrics.PersistenceErrors.Inc()
				writeError(w, http.StatusInternalServerError, "internal_error", "could not end session")
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": store.SessionEnded})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
		return
	}
	if err := s.store.SetSessionStatus(r.Context(), sessionID, store.SessionEnded); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not end session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": store.SessionEnded})
}

func sessionResponse(sess store.Session, characterID, joinCode string) createSessionResponse {
	mode := sess.Mode
	if mode == "" {
		mode = store.SessionModeSolo
	}
	return createSessionResponse{
		SessionID:   sess.ID,
		CharacterID: characterID,
		Seed:        sess.Seed,
		WorldID:     sess.WorldID,
		Mode:        mode,
		Listed:      sess.Listed,
		JoinCode:    joinCode,
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

func newJoinCode() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "join_" + hex.EncodeToString(b[:]), nil
}

func hashJoinCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func joinCodeMatches(hash, code string) bool {
	got := hashJoinCode(code)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(got)) == 1
}
