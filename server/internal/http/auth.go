package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
)

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v0/auth/dev-login", s.handleDevLogin)
}

type devLoginRequest struct {
	Email    string `json:"email"`
	DevToken string `json:"dev_token"`
}

type devLoginResponse struct {
	AccountID   string `json:"account_id"`
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}

func (s *Server) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	var req devLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	res, err := s.auth.DevLogin(r.Context(), req.Email, req.DevToken)
	switch {
	case errors.Is(err, auth.ErrInvalidDevToken):
		writeError(w, http.StatusUnauthorized, "invalid_dev_token", "invalid dev token")
		return
	case errors.Is(err, auth.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "invalid_email", "email must be a non-empty address")
		return
	case err != nil:
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not complete login")
		return
	}

	writeJSON(w, http.StatusOK, devLoginResponse{
		AccountID:   res.Account.ID,
		AccessToken: res.Token,
		ExpiresAt:   res.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// decodeJSON strictly decodes a JSON request body into v.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errors.New("malformed JSON body")
	}
	return nil
}
