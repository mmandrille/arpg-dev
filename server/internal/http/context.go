package httpapi

import (
	"context"
	"net/http"
	"strings"
)

type accountCtxKey int

const accountKey accountCtxKey = iota

// withAccount stores the authenticated account id on the context.
func withAccount(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountKey, accountID)
}

// accountFromContext retrieves the authenticated account id, if any.
func accountFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountKey).(string)
	return id, ok && id != ""
}

// bearerToken extracts the token from an "Authorization: Bearer <token>"
// header. As a fallback it accepts an "access_token" query parameter, because
// WebSocket clients (e.g. Godot's WebSocketPeer, browsers) cannot reliably set
// the Authorization header on the handshake request.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return r.URL.Query().Get("access_token")
}

// requireAuth wraps a handler with bearer-token authentication. On success the
// account id is placed on the request context; otherwise it responds 401 with a
// structured error.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		accountID, ok := s.auth.Authenticate(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}
		next.ServeHTTP(w, r.WithContext(withAccount(r.Context(), accountID)))
	})
}
