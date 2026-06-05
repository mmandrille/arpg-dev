// Package auth implements the dev-token authentication baseline (ADR-0001
// auth/session default): email/dev-token login that issues opaque bearer
// access tokens carrying real account identity. The production auth provider
// is deferred to ADR-0005; everything here sits behind interfaces.
package auth

import (
	"sync"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
)

// TokenStore is an in-memory bearer-token registry. Tokens are opaque and map
// to an account id with an expiry. In-memory is acceptable for the dev-token
// baseline; a server restart invalidates tokens (clients simply re-login).
type TokenStore struct {
	mu      sync.Mutex
	entries map[string]tokenEntry
}

type tokenEntry struct {
	accountID string
	expires   time.Time
}

// NewTokenStore returns an empty token store.
func NewTokenStore() *TokenStore {
	return &TokenStore{entries: make(map[string]tokenEntry)}
}

// Issue mints a new opaque token for the account, valid for ttl.
func (t *TokenStore) Issue(accountID string, ttl time.Duration) (string, time.Time) {
	token := ids.Token()
	expires := time.Now().Add(ttl)
	t.mu.Lock()
	t.entries[token] = tokenEntry{accountID: accountID, expires: expires}
	t.mu.Unlock()
	return token, expires
}

// Lookup returns the account id for a token if it exists and has not expired.
func (t *TokenStore) Lookup(token string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[token]
	if !ok {
		return "", false
	}
	if time.Now().After(e.expires) {
		delete(t.entries, token)
		return "", false
	}
	return e.accountID, true
}
