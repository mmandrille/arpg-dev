package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const maxCharacterNameRunes = 24
const defaultCharacterClass = "barbarian"

func (s *Server) registerCharacterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/characters", s.requireAuth(http.HandlerFunc(s.handleListCharacters)))
	mux.Handle("POST /v0/characters", s.requireAuth(http.HandlerFunc(s.handleCreateCharacter)))
	mux.Handle("PATCH /v0/characters/{character_id}", s.requireAuth(http.HandlerFunc(s.handleRenameCharacter)))
	mux.Handle("DELETE /v0/characters/{character_id}", s.requireAuth(http.HandlerFunc(s.handleDeleteCharacter)))
}

type characterResponse struct {
	CharacterID         string `json:"character_id"`
	Name                string `json:"name"`
	CharacterClass      string `json:"character_class"`
	Dead                bool   `json:"dead"`
	Level               int    `json:"level"`
	Gold                int    `json:"gold"`
	DeepestDungeonDepth int    `json:"deepest_dungeon_depth"`
	CreatedAt           string `json:"created_at"`
}

type listCharactersResponse struct {
	Characters []characterResponse `json:"characters"`
}

type createCharacterRequest struct {
	Name           string `json:"name"`
	CharacterClass string `json:"character_class"`
}

type renameCharacterRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}

	chars, err := s.store.ListCharacters(r.Context(), accountID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list characters")
		return
	}

	res := listCharactersResponse{Characters: make([]characterResponse, 0, len(chars))}
	for _, c := range chars {
		res.Characters = append(res.Characters, characterSummaryToResponse(c))
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}

	var req createCharacterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "invalid_character_name", "character name is required")
		return
	}
	if utf8.RuneCountInString(name) > maxCharacterNameRunes {
		writeError(w, http.StatusBadRequest, "invalid_character_name", "character name must be 24 characters or fewer")
		return
	}

	characterClass := strings.TrimSpace(req.CharacterClass)
	if characterClass == "" {
		characterClass = defaultCharacterClass
	}
	if _, ok := s.rules.CharacterProgression.Classes[characterClass]; !ok {
		writeError(w, http.StatusBadRequest, "invalid_character_class", "character class is not available")
		return
	}

	char, err := s.store.CreateCharacter(r.Context(), ids.New("char"), accountID, name, characterClass)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create character")
		return
	}
	writeJSON(w, http.StatusCreated, characterToResponse(char))
}

func (s *Server) handleRenameCharacter(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}

	characterID := strings.TrimSpace(r.PathValue("character_id"))
	if characterID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "character_id is required")
		return
	}
	var req renameCharacterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "invalid_character_name", "character name is required")
		return
	}
	if utf8.RuneCountInString(name) > maxCharacterNameRunes {
		writeError(w, http.StatusBadRequest, "invalid_character_name", "character name must be 24 characters or fewer")
		return
	}

	char, err := s.store.RenameCharacter(r.Context(), accountID, characterID, name)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "character not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not rename character")
		return
	}
	writeJSON(w, http.StatusOK, characterToResponse(char))
}

func (s *Server) handleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account context")
		return
	}

	characterID := strings.TrimSpace(r.PathValue("character_id"))
	if characterID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "character_id is required")
		return
	}

	err := s.store.DeleteCharacter(r.Context(), accountID, characterID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "character not found")
		return
	}
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not delete character")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func characterToResponse(c store.Character) characterResponse {
	return characterResponse{
		CharacterID:         c.ID,
		Name:                c.Name,
		CharacterClass:      c.CharacterClass,
		Dead:                c.Dead,
		Level:               1,
		Gold:                0,
		DeepestDungeonDepth: 0,
		CreatedAt:           c.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func characterSummaryToResponse(c store.CharacterSummary) characterResponse {
	return characterResponse{
		CharacterID:         c.ID,
		Name:                c.Name,
		CharacterClass:      c.CharacterClass,
		Dead:                c.Dead,
		Level:               c.Level,
		Gold:                c.Gold,
		DeepestDungeonDepth: c.DeepestDungeonDepth,
		CreatedAt:           c.CreatedAt.UTC().Format(time.RFC3339),
	}
}
