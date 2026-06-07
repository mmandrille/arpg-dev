package httpapi

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const maxCharacterNameRunes = 24

func (s *Server) registerCharacterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/characters", s.requireAuth(http.HandlerFunc(s.handleListCharacters)))
	mux.Handle("POST /v0/characters", s.requireAuth(http.HandlerFunc(s.handleCreateCharacter)))
}

type characterResponse struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at"`
}

type listCharactersResponse struct {
	Characters []characterResponse `json:"characters"`
}

type createCharacterRequest struct {
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
		res.Characters = append(res.Characters, characterToResponse(c))
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

	char, err := s.store.CreateCharacter(r.Context(), ids.New("char"), accountID, name)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create character")
		return
	}
	writeJSON(w, http.StatusCreated, characterToResponse(char))
}

func characterToResponse(c store.Character) characterResponse {
	return characterResponse{
		CharacterID: c.ID,
		Name:        c.Name,
		CreatedAt:   c.CreatedAt.UTC().Format(time.RFC3339),
	}
}
