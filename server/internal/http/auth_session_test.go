package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const (
	testDevToken   = "test-dev-token"
	testDebugToken = "test-debug-token"
)

// fullServer builds a server backed by real Postgres, or skips if unreachable.
func fullServer(t *testing.T) http.Handler {
	t.Helper()
	url := "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := store.Connect(ctx, url)
	if err != nil {
		t.Skipf("skipping auth/session test: no Postgres: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(db.Close)
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("rules dir: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return New(Deps{
		Config:  config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, MetricsEnabled: true},
		Logger:  logging.NewTo(io.Discard, "local"),
		Metrics: metrics.New(),
		Store:   db,
		Auth:    auth.NewService(testDevToken, db, db),
		Rules:   rules,
		Ready:   db.Ping,
	}).Handler()
}

func postJSON(h http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func login(t *testing.T, h http.Handler) (accountID, token string) {
	t.Helper()
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "dev@example.test", "dev_token": testDevToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res devLoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if res.AccessToken == "" || res.AccountID == "" {
		t.Fatalf("login missing token/account: %+v", res)
	}
	return res.AccountID, res.AccessToken
}

func TestDevLoginInvalidToken(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "dev@example.test", "dev_token": "wrong",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestDevLoginInvalidEmail(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "not-an-email", "dev_token": testDevToken,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCreateSessionRequiresAuth(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/sessions", "", map[string]any{"mode": "solo"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestCreateSessionInvalidToken(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/sessions", "garbage-token", map[string]any{"mode": "solo"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestCreateAndResumeSession(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.SessionID == "" || created.Seed == "" || created.CharacterID == "" || created.WorldID != game.DefaultWorldID {
		t.Fatalf("incomplete session response: %+v", created)
	}
	if created.WSURL != "/v0/ws?session_id="+created.SessionID {
		t.Fatalf("ws_url = %q", created.WSURL)
	}

	// Resume the same session.
	resumeID := created.SessionID
	rec = postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "resume_session_id": resumeID})
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resumed createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resumed)
	if resumed.SessionID != resumeID || resumed.Seed != created.Seed || resumed.WorldID != created.WorldID {
		t.Fatalf("resume mismatch: %+v vs %+v", resumed, created)
	}
}

func TestCreateSessionWorldID(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "gear_before_combat"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.WorldID != "gear_before_combat" {
		t.Fatalf("world_id = %q, want gear_before_combat", created.WorldID)
	}

	resumeID := created.SessionID
	rec = postJSON(h, "/v0/sessions", token, map[string]any{
		"mode": "solo", "resume_session_id": resumeID, "world_id": game.DefaultWorldID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resumed createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resumed)
	if resumed.WorldID != "gear_before_combat" {
		t.Fatalf("resume world_id = %q, want persisted gear_before_combat", resumed.WorldID)
	}
}

func TestCreateSessionRejectsUnknownWorldID(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "missing"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "invalid_world_id" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestResumeUnknownSession(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)
	missing := "sess_00000000000000000000000000"
	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "resume_session_id": missing})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
