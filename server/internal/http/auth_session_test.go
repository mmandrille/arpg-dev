package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
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
	return fullServerWithConfig(t, config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, MetricsEnabled: true})
}

func fullServerWithConfig(t *testing.T, cfg config.Config) http.Handler {
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
		Config:  cfg,
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

func getJSON(h http.Handler, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func deleteJSON(h http.Handler, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func login(t *testing.T, h http.Handler) (accountID, token string) {
	return loginEmail(t, h, "dev@example.test")
}

func loginEmail(t *testing.T, h http.Handler, email string) (accountID, token string) {
	t.Helper()
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": email, "dev_token": testDevToken,
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

func createCharacter(t *testing.T, h http.Handler, token, name string) characterResponse {
	t.Helper()
	rec := postJSON(h, "/v0/characters", token, map[string]string{"name": name})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create character status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res characterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode character: %v", err)
	}
	if res.CharacterID == "" || res.Name == "" || res.CreatedAt == "" {
		t.Fatalf("incomplete character response: %+v", res)
	}
	return res
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

func TestCharacterAPIRequiresAuth(t *testing.T) {
	h := fullServer(t)

	rec := getJSON(h, "/v0/characters", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("list status = %d, want 401", rec.Code)
	}
	rec = postJSON(h, "/v0/characters", "", map[string]string{"name": "Mara"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("create status = %d, want 401", rec.Code)
	}
	rec = deleteJSON(h, "/v0/characters/char_missing", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("delete status = %d, want 401", rec.Code)
	}
}

func TestCreateCharacterValidationAndList(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, "characters-validation@example.test")

	rec := postJSON(h, "/v0/characters", token, map[string]string{"name": "   "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty name status = %d, want 400", rec.Code)
	}
	rec = postJSON(h, "/v0/characters", token, map[string]string{"name": strings.Repeat("x", 25)})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("long name status = %d, want 400", rec.Code)
	}

	first := createCharacter(t, h, token, "  Mara  ")
	if first.Name != "Mara" {
		t.Fatalf("trimmed name = %q, want Mara", first.Name)
	}
	second := createCharacter(t, h, token, "Mara")
	if second.Name != "Mara" || second.CharacterID == first.CharacterID {
		t.Fatalf("duplicate character not created independently: first=%+v second=%+v", first, second)
	}

	rec = getJSON(h, "/v0/characters", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	var found int
	for _, c := range listed.Characters {
		if c.CharacterID == first.CharacterID || c.CharacterID == second.CharacterID {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("listed characters missing created rows: %+v", listed.Characters)
	}
}

func TestCharactersAreAccountScoped(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, "characters-account-a@example.test")
	_, tokenB := loginEmail(t, h, "characters-account-b@example.test")

	charA := createCharacter(t, h, tokenA, "Account A Hero")
	rec := getJSON(h, "/v0/characters", tokenB)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == charA.CharacterID {
			t.Fatalf("account B saw account A character: %+v", listed.Characters)
		}
	}
}

func TestDeleteCharacterRemovesOwnedCharacter(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, "characters-delete+"+ids.Token()[:12]+"@example.test")
	keep := createCharacter(t, h, token, "Keep")
	remove := createCharacter(t, h, token, "Remove")

	rec := deleteJSON(h, "/v0/characters/"+remove.CharacterID, token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = getJSON(h, "/v0/characters", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Characters) != 2 {
		t.Fatalf("listed characters after delete = %+v, want 2 remaining", listed.Characters)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == remove.CharacterID {
			t.Fatalf("deleted character still listed: %+v", listed.Characters)
		}
	}
	var foundKeep bool
	for _, c := range listed.Characters {
		if c.CharacterID == keep.CharacterID {
			foundKeep = true
			break
		}
	}
	if !foundKeep {
		t.Fatalf("keep character missing after delete: %+v", listed.Characters)
	}

	rec = deleteJSON(h, "/v0/characters/"+remove.CharacterID, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", rec.Code)
	}
}

func TestDeleteCharacterIsAccountScoped(t *testing.T) {
	h := fullServer(t)
	suffix := ids.Token()[:12]
	_, tokenA := loginEmail(t, h, "characters-delete-a+"+suffix+"@example.test")
	_, tokenB := loginEmail(t, h, "characters-delete-b+"+suffix+"@example.test")
	charA := createCharacter(t, h, tokenA, "Account A Hero")

	rec := deleteJSON(h, "/v0/characters/"+charA.CharacterID, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete other account character status = %d, want 404", rec.Code)
	}

	rec = getJSON(h, "/v0/characters", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Characters) != 2 {
		t.Fatalf("account A characters deleted by account B: %+v", listed.Characters)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == charA.CharacterID {
			return
		}
	}
	t.Fatalf("account A character missing after cross-account delete attempt: %+v", listed.Characters)
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

func TestCreateCoopSessionAndJoinLifecycle(t *testing.T) {
	h := fullServer(t)
	_, hostToken := loginEmail(t, h, "coop-host+"+ids.Token()[:12]+"@example.test")
	_, guestToken := loginEmail(t, h, "coop-guest+"+ids.Token()[:12]+"@example.test")
	_, thirdToken := loginEmail(t, h, "coop-third+"+ids.Token()[:12]+"@example.test")
	_, fourthToken := loginEmail(t, h, "coop-fourth+"+ids.Token()[:12]+"@example.test")
	guestChar := createCharacter(t, h, guestToken, "Guest Hero")
	thirdChar := createCharacter(t, h, thirdToken, "Third Hero")
	fourthChar := createCharacter(t, h, fourthToken, "Fourth Hero")

	rec := postJSON(h, "/v0/sessions", hostToken, map[string]any{"mode": "coop", "world_id": "dungeon_levels"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode coop create: %v", err)
	}
	if created.Mode != store.SessionModeCoop || created.JoinCode == "" || created.WorldID != "dungeon_levels" {
		t.Fatalf("coop create response mismatch: %+v", created)
	}
	if created.Listed {
		t.Fatalf("private coop unexpectedly listed: %+v", created)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", guestToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("join status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var joined createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &joined); err != nil {
		t.Fatalf("decode join: %v", err)
	}
	if joined.SessionID != created.SessionID || joined.CharacterID != guestChar.CharacterID || joined.Mode != store.SessionModeCoop || joined.JoinCode != "" {
		t.Fatalf("join response mismatch: %+v", joined)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", guestToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate join status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "duplicate_member" {
		t.Fatalf("duplicate code = %q", body.Error.Code)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": created.JoinCode, "character_id": thirdChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("third join status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", fourthToken, map[string]any{
		"join_code": created.JoinCode, "character_id": fourthChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("fourth join status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": "join_wrong", "character_id": thirdChar.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("wrong code status = %d, want 404", rec.Code)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": created.JoinCode, "character_id": created.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign character status = %d, want 404", rec.Code)
	}

	rec = postJSON(h, "/v0/sessions", hostToken, map[string]any{"mode": "coop"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create second coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var ended createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &ended)
	rec = postJSON(h, "/v0/sessions/"+ended.SessionID+"/end", hostToken, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("end coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+ended.SessionID+"/join", guestToken, map[string]any{
		"join_code": ended.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("join ended status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "session_ended" {
		t.Fatalf("ended join code = %q", body.Error.Code)
	}
}

func TestListedSessionListAndJoinWithoutCode(t *testing.T) {
	h := fullServer(t)
	_, hostToken := loginEmail(t, h, "listed-host+"+ids.Token()[:12]+"@example.test")
	_, guestToken := loginEmail(t, h, "listed-guest+"+ids.Token()[:12]+"@example.test")
	_, outsiderToken := loginEmail(t, h, "listed-outsider+"+ids.Token()[:12]+"@example.test")
	hostChar := createCharacter(t, h, hostToken, "Listed Host")
	guestChar := createCharacter(t, h, guestToken, "Listed Guest")
	outsiderChar := createCharacter(t, h, outsiderToken, "Listed Outsider")

	privateRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "coop", "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if privateRec.Code != http.StatusCreated {
		t.Fatalf("private create status = %d, body = %s", privateRec.Code, privateRec.Body.String())
	}
	soloRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if soloRec.Code != http.StatusCreated {
		t.Fatalf("solo create status = %d, body = %s", soloRec.Code, soloRec.Body.String())
	}
	listedRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "coop", "listed": true, "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if listedRec.Code != http.StatusCreated {
		t.Fatalf("listed create status = %d, body = %s", listedRec.Code, listedRec.Body.String())
	}
	var listed createSessionResponse
	if err := json.Unmarshal(listedRec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode listed create: %v", err)
	}
	if !listed.Listed || listed.JoinCode == "" {
		t.Fatalf("listed create response mismatch: %+v", listed)
	}

	rec := getJSON(h, "/v0/sessions/active", guestToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("active list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var active activeSessionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &active); err != nil {
		t.Fatalf("decode active sessions: %v", err)
	}
	var found activeSessionSummaryResponse
	for _, sess := range active.Sessions {
		if sess.SessionID == listed.SessionID {
			found = sess
		}
		if strings.Contains(rec.Body.String(), "join_") {
			t.Fatalf("active list exposed join code: %s", rec.Body.String())
		}
	}
	if found.SessionID != "" {
		t.Fatalf("unconnected listed session should not be active-browser visible: %+v", found)
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", guestToken, map[string]any{
		"character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("listed join status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var joined createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &joined); err != nil {
		t.Fatalf("decode listed join: %v", err)
	}
	if joined.SessionID != listed.SessionID || joined.CharacterID != guestChar.CharacterID || !joined.Listed || joined.JoinCode != "" {
		t.Fatalf("listed join response mismatch: %+v", joined)
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", outsiderToken, map[string]any{
		"character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account listed join status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", outsiderToken, map[string]any{
		"character_id": outsiderChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("second listed join status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSessionCustomSeedLocalOnly(t *testing.T) {
	local := fullServer(t)
	_, localToken := login(t, local)
	rec := postJSON(local, "/v0/sessions", localToken, map[string]any{"mode": "solo", "seed": "pinned-test-seed"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("local custom seed status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.Seed != "pinned-test-seed" {
		t.Fatalf("local seed = %q, want pinned-test-seed", created.Seed)
	}

	remote := fullServerWithConfig(t, config.Config{Addr: ":0", Env: "remote", DevToken: testDevToken, MetricsEnabled: true})
	_, remoteToken := loginEmail(t, remote, "remote-seed@example.test")
	rec = postJSON(remote, "/v0/sessions", remoteToken, map[string]any{"mode": "solo", "seed": "pinned-test-seed"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("remote custom seed status = %d, body = %s", rec.Code, rec.Body.String())
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

func TestCreateSessionWithSelectedCharacter(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, "selected-character-a@example.test")
	_, tokenB := loginEmail(t, h, "selected-character-b@example.test")
	charA := createCharacter(t, h, tokenA, "Selected Hero")

	rec := postJSON(h, "/v0/sessions", tokenA, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": charA.CharacterID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create selected status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if created.CharacterID != charA.CharacterID || created.WorldID != "dungeon_levels" {
		t.Fatalf("selected session mismatch: %+v, character=%+v", created, charA)
	}

	rec = postJSON(h, "/v0/sessions", tokenB, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": charA.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account selected status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "character_not_found" {
		t.Fatalf("error code = %q, want character_not_found", body.Error.Code)
	}
}

func TestCreateSessionOmittedCharacterUsesDefault(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, "default-character-session@example.test")

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "dungeon_levels"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if created.CharacterID == "" || created.WorldID != "dungeon_levels" {
		t.Fatalf("default-character session incomplete: %+v", created)
	}
}

func TestEndSessionOwnerOnlyAndIdempotent(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, "end-session-a@example.test")
	_, tokenB := loginEmail(t, h, "end-session-b@example.test")

	rec := postJSON(h, "/v0/sessions", tokenA, map[string]any{"mode": "solo"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenB, map[string]any{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account end status = %d, want 404", rec.Code)
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenA, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("end status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenA, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("second end status = %d, body = %s", rec.Code, rec.Body.String())
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
