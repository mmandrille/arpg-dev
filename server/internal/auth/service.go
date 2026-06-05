package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// Errors surfaced by the auth service.
var (
	ErrInvalidDevToken = errors.New("auth: invalid dev token")
	ErrInvalidEmail    = errors.New("auth: invalid email")
)

// defaultTTL is how long an issued dev access token remains valid.
const defaultTTL = 24 * time.Hour

// defaultCharacterName is the name given to the single v0 character.
const defaultCharacterName = "Hero"

// Service performs dev-token login and bearer-token authentication.
type Service struct {
	devToken   string
	tokens     *TokenStore
	accounts   store.AccountRepo
	characters store.CharacterRepo
	ttl        time.Duration
}

// NewService builds the auth service against the account/character repos.
func NewService(devToken string, accounts store.AccountRepo, characters store.CharacterRepo) *Service {
	return &Service{
		devToken:   devToken,
		tokens:     NewTokenStore(),
		accounts:   accounts,
		characters: characters,
		ttl:        defaultTTL,
	}
}

// LoginResult is the outcome of a successful dev-login.
type LoginResult struct {
	Account   store.Account
	Character store.Character
	Token     string
	ExpiresAt time.Time
}

// DevLogin validates the dev token, upserts the account by email, ensures the
// account's default character exists, and issues a bearer token. The result is
// treated as real account identity (no anonymous play).
func (s *Service) DevLogin(ctx context.Context, email, devToken string) (LoginResult, error) {
	if devToken != s.devToken {
		return LoginResult{}, ErrInvalidDevToken
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return LoginResult{}, ErrInvalidEmail
	}

	acct, err := s.accounts.UpsertAccountByEmail(ctx, ids.New("acct"), email)
	if err != nil {
		return LoginResult{}, err
	}
	char, err := s.characters.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, defaultCharacterName)
	if err != nil {
		return LoginResult{}, err
	}
	token, expires := s.tokens.Issue(acct.ID, s.ttl)
	return LoginResult{Account: acct, Character: char, Token: token, ExpiresAt: expires}, nil
}

// Authenticate returns the account id for a bearer token, if valid.
func (s *Service) Authenticate(token string) (string, bool) {
	return s.tokens.Lookup(token)
}
