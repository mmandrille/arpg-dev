package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// fullStack builds a real httptest server (Postgres-backed) including the
// realtime hub, or skips when Postgres/rules are unavailable.
func fullStack(t *testing.T) *httptest.Server {
	return fullStackWithRules(t, nil)
}

func fullStackWithRules(t *testing.T, tweak func(*game.Rules)) *httptest.Server {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := store.Connect(ctx, "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable")
	if err != nil {
		t.Skipf("skipping ws test: no Postgres: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(db.Close)

	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	if tweak != nil {
		tweak(rules)
	}
	m := metrics.New()
	authSvc := auth.NewService(testDevToken, db, db)
	h := New(Deps{
		Config:   config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, DebugToken: testDebugToken, MetricsEnabled: true},
		Logger:   logging.NewTo(io.Discard, "local"),
		Metrics:  m,
		Store:    db,
		Auth:     authSvc,
		Realtime: realtime.NewHub(db, rules, logging.NewTo(io.Discard, "local"), m),
		Rules:    rules,
		Ready:    db.Ping,
	}).Handler()

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// wire-decoding structs for the client side of the test.
type wEntity struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ItemDefID       string `json:"item_def_id"`
	ProjectileDefID string `json:"projectile_def_id"`
	HP              *int   `json:"hp"`
	MaxHP           *int   `json:"max_hp"`
	Mana            *int   `json:"mana"`
	MaxMana         *int   `json:"max_mana"`
}
type wItem struct {
	ItemInstanceID string `json:"item_instance_id"`
	ItemDefID      string `json:"item_def_id"`
	Slot           string `json:"slot"`
	Equipped       bool   `json:"equipped"`
}
type wChange struct {
	Op                   string                         `json:"op"`
	Entity               *wEntity                       `json:"entity"`
	EntityID             string                         `json:"entity_id"`
	Item                 *wItem                         `json:"item"`
	Slot                 string                         `json:"slot"`
	ItemInstanceID       *string                        `json:"item_instance_id"`
	CharacterProgression *game.CharacterProgressionView `json:"character_progression"`
	SkillProgression     *game.SkillProgressionView     `json:"skill_progression"`
	SkillCooldowns       []game.SkillCooldownView       `json:"skill_cooldowns"`
}
type wEvent struct {
	EventType string `json:"event_type"`
	EntityID  string `json:"entity_id"`
	SkillID   string `json:"skill_id"`
}
type wMsg struct {
	Type    string          `json:"type"`
	Tick    uint64          `json:"tick"`
	Payload json.RawMessage `json:"payload"`
}

type wireDelta struct {
	Tick    uint64
	Changes []wChange `json:"changes"`
	Events  []wEvent  `json:"events"`
}

func loginAndSession(t *testing.T, srv *httptest.Server) (token, sessionID string) {
	return loginAndSessionWithWorld(t, srv, "")
}

func loginAndSessionWithWorld(t *testing.T, srv *httptest.Server, worldID string) (token, sessionID string) {
	t.Helper()
	// dev-login
	rec := doHTTP(t, srv, "POST", "/v0/auth/dev-login", "", map[string]string{
		"email": "ws+" + ids.Token()[:12] + "@example.test", "dev_token": testDevToken,
	})
	var lr devLoginResponse
	mustJSON(t, rec, &lr)
	// create session
	body := map[string]any{"mode": "solo"}
	if worldID != "" {
		body["world_id"] = worldID
	}
	rec = doHTTP(t, srv, "POST", "/v0/sessions", lr.AccessToken, body)
	var sr createSessionResponse
	mustJSON(t, rec, &sr)
	return lr.AccessToken, sr.SessionID
}

func createSessionWithToken(t *testing.T, srv *httptest.Server, token, worldID string) string {
	t.Helper()
	body := map[string]any{"mode": "solo"}
	if worldID != "" {
		body["world_id"] = worldID
	}
	rec := doHTTP(t, srv, "POST", "/v0/sessions", token, body)
	var sr createSessionResponse
	mustJSON(t, rec, &sr)
	return sr.SessionID
}

func dialWS(t *testing.T, srv *httptest.Server, token, sessionID string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v0/ws?session_id=" + sessionID
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+token)
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
	if err != nil {
		body := ""
		if resp != nil {
			b, _ := io.ReadAll(resp.Body)
			body = string(b)
		}
		t.Fatalf("dial ws: %v (%s)", err, body)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestWebSocketRejectsUnauthenticated(t *testing.T) {
	srv := fullStack(t)
	_, sessionID := loginAndSession(t, srv)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v0/ws?session_id=" + sessionID
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected handshake failure without bearer token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %v, want 401", resp)
	}
}

func TestCoopWebSocketAllowsJoinedMemberAndRejectsNonMember(t *testing.T) {
	srv := fullStack(t)

	loginReq := func(email string) devLoginResponse {
		resp := doHTTP(t, srv, http.MethodPost, "/v0/auth/dev-login", "", map[string]string{
			"email": email, "dev_token": testDevToken,
		})
		var res devLoginResponse
		mustJSON(t, resp, &res)
		return res
	}
	createChar := func(token, name string) characterResponse {
		resp := doHTTP(t, srv, http.MethodPost, "/v0/characters", token, map[string]string{"name": name})
		var res characterResponse
		mustJSON(t, resp, &res)
		return res
	}

	host := loginReq("ws-coop-host+" + ids.Token()[:12] + "@example.test")
	guest := loginReq("ws-coop-guest+" + ids.Token()[:12] + "@example.test")
	outsider := loginReq("ws-coop-outsider+" + ids.Token()[:12] + "@example.test")
	guestChar := createChar(guest.AccessToken, "Guest")

	resp := doHTTP(t, srv, http.MethodPost, "/v0/sessions", host.AccessToken, map[string]any{"mode": "coop"})
	var created createSessionResponse
	mustJSON(t, resp, &created)
	resp = doHTTP(t, srv, http.MethodPost, "/v0/sessions/"+created.SessionID+"/join", guest.AccessToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	var joined createSessionResponse
	mustJSON(t, resp, &joined)

	guestConn := dialWS(t, srv, guest.AccessToken, created.SessionID)
	guestSnap := readSnapshot(t, guestConn)
	if guestSnap.SessionID != created.SessionID {
		t.Fatalf("guest snapshot session = %s, want %s", guestSnap.SessionID, created.SessionID)
	}
	_ = guestConn.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v0/ws?session_id=" + created.SessionID
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+outsider.AccessToken)
	_, rejectResp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
	if err == nil {
		t.Fatal("expected outsider websocket handshake failure")
	}
	if rejectResp == nil || rejectResp.StatusCode != http.StatusNotFound {
		t.Fatalf("outsider status = %v, want 404", rejectResp)
	}
}

func TestCoopWebSocketRejectsPersistedConnectedMember(t *testing.T) {
	srv := fullStack(t)
	host := loginWS(t, srv, "ws-connected-host")
	guest := loginWS(t, srv, "ws-connected-guest")
	guestChar := createCharacterWS(t, srv, guest.AccessToken, "Connected Guest")

	resp := doHTTP(t, srv, http.MethodPost, "/v0/sessions", host.AccessToken, map[string]any{"mode": "coop"})
	var created createSessionResponse
	mustJSON(t, resp, &created)
	resp = doHTTP(t, srv, http.MethodPost, "/v0/sessions/"+created.SessionID+"/join", guest.AccessToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	var joined createSessionResponse
	mustJSON(t, resp, &joined)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := store.Connect(ctx, "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable")
	if err != nil {
		t.Skipf("skipping persisted connected member test: no Postgres: %v", err)
	}
	if err := db.SetSessionMemberConnected(ctx, created.SessionID, guest.AccountID, guestChar.CharacterID, "1007", 0, 0); err != nil {
		t.Fatalf("mark guest connected: %v", err)
	}
	t.Cleanup(func() {
		_ = db.SetSessionMemberDisconnected(context.Background(), created.SessionID, guest.AccountID, guestChar.CharacterID, 0, 0)
		db.Close()
	})

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v0/ws?session_id=" + created.SessionID
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+guest.AccessToken)
	_, rejectResp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
	if err == nil {
		t.Fatal("expected persisted connected member websocket rejection")
	}
	if rejectResp == nil || rejectResp.StatusCode != http.StatusConflict {
		t.Fatalf("persisted connected status = %v, want 409", rejectResp)
	}
}

func TestCoopWebSocketSharedSessionLoopMovementDisconnectAndReconnect(t *testing.T) {
	srv := fullStack(t)
	host, guest := loginWS(t, srv, "ws-coop-loop-host"), loginWS(t, srv, "ws-coop-loop-guest")
	guestChar := createCharacterWS(t, srv, guest.AccessToken, "Guest")

	resp := doHTTP(t, srv, http.MethodPost, "/v0/sessions", host.AccessToken, map[string]any{"mode": "coop"})
	var created createSessionResponse
	mustJSON(t, resp, &created)
	resp = doHTTP(t, srv, http.MethodPost, "/v0/sessions/"+created.SessionID+"/join", guest.AccessToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	var joined createSessionResponse
	mustJSON(t, resp, &joined)

	hostConn := dialWS(t, srv, host.AccessToken, created.SessionID)
	hostSnap := readSnapshot(t, hostConn)
	guestConn := dialWS(t, srv, guest.AccessToken, created.SessionID)
	guestSnap := readSnapshotEventually(t, guestConn)
	hostSnap = readSnapshotEventually(t, hostConn)
	guestSnap = readSnapshotEventually(t, guestConn)
	if hostSnap.LocalPlayerID == "" || guestSnap.LocalPlayerID == "" || hostSnap.LocalPlayerID == guestSnap.LocalPlayerID {
		t.Fatalf("local player ids host=%q guest=%q", hostSnap.LocalPlayerID, guestSnap.LocalPlayerID)
	}
	if len(hostSnap.Party) != 2 || len(guestSnap.Party) != 2 {
		t.Fatalf("party metadata host=%+v guest=%+v", hostSnap.Party, guestSnap.Party)
	}

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v0/ws?session_id=" + created.SessionID
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+guest.AccessToken)
	_, dupResp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
	if err == nil {
		t.Fatal("expected duplicate guest websocket rejection")
	}
	if dupResp == nil || dupResp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate status = %v, want 409", dupResp)
	}

	hostBefore := entityPosition(t, hostSnap, hostSnap.LocalPlayerID)
	guestBefore := entityPosition(t, guestSnap, guestSnap.LocalPlayerID)
	sendIntent(t, guestConn, created.SessionID, guestSnap.ServerTick, "msg-guest-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	readAccepted(t, guestConn, "msg-guest-move")
	requestSnapshot(t, hostConn, created.SessionID, "msg-host-ready")
	requestSnapshot(t, guestConn, created.SessionID, "msg-guest-ready")
	hostAfterGuestMove := readSnapshotEventually(t, hostConn)
	guestAfterMove := readSnapshotEventually(t, guestConn)
	if got := entityPosition(t, hostAfterGuestMove, hostSnap.LocalPlayerID); got != hostBefore {
		t.Fatalf("host moved after guest input: before=%+v after=%+v", hostBefore, got)
	}
	if got := entityPosition(t, guestAfterMove, guestSnap.LocalPlayerID); got == guestBefore {
		t.Fatalf("guest did not move from %+v", guestBefore)
	}

	_ = guestConn.Close()
	readEntityRemove(t, hostConn, guestSnap.LocalPlayerID)
	sendIntent(t, hostConn, created.SessionID, hostAfterGuestMove.ServerTick, "msg-host-move-after-disconnect", "move_intent", map[string]any{"direction": map[string]any{"x": 0, "y": 1}, "duration_ticks": 1})
	readAccepted(t, hostConn, "msg-host-move-after-disconnect")

	reconnected := dialWS(t, srv, guest.AccessToken, created.SessionID)
	defer reconnected.Close()
	reconnectSnap := readSnapshotEventually(t, reconnected)
	if reconnectSnap.LocalPlayerID != guestSnap.LocalPlayerID || reconnectSnap.CurrentLevel != 0 {
		t.Fatalf("reconnect snapshot = %+v, want same player in town", reconnectSnap)
	}
	if findEntity(reconnectSnap, guestSnap.LocalPlayerID) == nil {
		t.Fatalf("reconnected guest entity missing from snapshot: %+v", reconnectSnap.Entities)
	}
}

func TestThreeClientListedCoopWebSocketVisibilityMovementAndDisconnect(t *testing.T) {
	srv := fullStack(t)
	host := loginWS(t, srv, "ws-three-host")
	guestA := loginWS(t, srv, "ws-three-guest-a")
	guestB := loginWS(t, srv, "ws-three-guest-b")
	lateGuest := loginWS(t, srv, "ws-three-late-guest")
	guestAChar := createCharacterWS(t, srv, guestA.AccessToken, "Guest A")
	guestBChar := createCharacterWS(t, srv, guestB.AccessToken, "Guest B")
	lateGuestChar := createCharacterWS(t, srv, lateGuest.AccessToken, "Late Guest")

	resp := doHTTP(t, srv, http.MethodPost, "/v0/sessions", host.AccessToken, map[string]any{
		"mode": "coop", "listed": true, "world_id": "dungeon_levels",
	})
	var created createSessionResponse
	mustJSON(t, resp, &created)
	for _, peer := range []struct {
		token string
		char  characterResponse
	}{
		{guestA.AccessToken, guestAChar},
		{guestB.AccessToken, guestBChar},
	} {
		resp = doHTTP(t, srv, http.MethodPost, "/v0/sessions/"+created.SessionID+"/join", peer.token, map[string]any{
			"character_id": peer.char.CharacterID,
		})
		var joined createSessionResponse
		mustJSON(t, resp, &joined)
		if joined.SessionID != created.SessionID || !joined.Listed || joined.JoinCode != "" {
			t.Fatalf("listed join response = %+v", joined)
		}
	}

	hostConn := dialWS(t, srv, host.AccessToken, created.SessionID)
	hostSnap := readSnapshotEventually(t, hostConn)
	guestAConn := dialWS(t, srv, guestA.AccessToken, created.SessionID)
	guestASnap := readSnapshotEventually(t, guestAConn)
	guestBConn := dialWS(t, srv, guestB.AccessToken, created.SessionID)
	guestBSnap := readSnapshotEventually(t, guestBConn)
	requestSnapshot(t, hostConn, created.SessionID, "msg-three-host-ready")
	requestSnapshot(t, guestAConn, created.SessionID, "msg-three-guest-a-ready")
	requestSnapshot(t, guestBConn, created.SessionID, "msg-three-guest-b-ready")
	hostSnap = readSnapshotEventually(t, hostConn)
	guestASnap = readSnapshotEventually(t, guestAConn)
	guestBSnap = readSnapshotEventually(t, guestBConn)

	ids := map[string]bool{hostSnap.LocalPlayerID: true, guestASnap.LocalPlayerID: true, guestBSnap.LocalPlayerID: true}
	if len(ids) != 3 {
		t.Fatalf("local player ids not distinct: host=%q guestA=%q guestB=%q", hostSnap.LocalPlayerID, guestASnap.LocalPlayerID, guestBSnap.LocalPlayerID)
	}
	for label, snap := range map[string]game.Snapshot{"host": hostSnap, "guestA": guestASnap, "guestB": guestBSnap} {
		if len(snap.Party) != 3 {
			t.Fatalf("%s party = %+v, want 3 members", label, snap.Party)
		}
		if countSnapshotPlayers(snap) != 3 {
			t.Fatalf("%s visible players = %+v, want 3", label, snap.Entities)
		}
	}
	assertActiveSessionVisible(t, srv, guestA.AccessToken, created.SessionID, 3, 3)

	hostBefore := entityPosition(t, hostSnap, hostSnap.LocalPlayerID)
	guestABefore := entityPosition(t, guestASnap, guestASnap.LocalPlayerID)
	guestBBefore := entityPosition(t, guestBSnap, guestBSnap.LocalPlayerID)
	sendIntent(t, guestBConn, created.SessionID, guestBSnap.ServerTick, "msg-three-guest-b-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	readAccepted(t, guestBConn, "msg-three-guest-b-move")
	requestSnapshot(t, hostConn, created.SessionID, "msg-three-host-after-move")
	requestSnapshot(t, guestAConn, created.SessionID, "msg-three-guest-a-after-move")
	requestSnapshot(t, guestBConn, created.SessionID, "msg-three-guest-b-after-move")
	hostAfter := readSnapshotEventually(t, hostConn)
	guestAAfter := readSnapshotEventually(t, guestAConn)
	guestBAfter := readSnapshotEventually(t, guestBConn)
	if got := entityPosition(t, hostAfter, hostSnap.LocalPlayerID); got != hostBefore {
		t.Fatalf("host moved from %+v to %+v after guest B input", hostBefore, got)
	}
	if got := entityPosition(t, guestAAfter, guestASnap.LocalPlayerID); got != guestABefore {
		t.Fatalf("guest A moved from %+v to %+v after guest B input", guestABefore, got)
	}
	if got := entityPosition(t, guestBAfter, guestBSnap.LocalPlayerID); got == guestBBefore {
		t.Fatalf("guest B did not move from %+v", guestBBefore)
	}

	_ = guestBConn.Close()
	readEntityRemove(t, hostConn, guestBSnap.LocalPlayerID)
	readEntityRemove(t, guestAConn, guestBSnap.LocalPlayerID)
	assertActiveSessionVisible(t, srv, guestA.AccessToken, created.SessionID, 3, 2)
	sendIntent(t, guestAConn, created.SessionID, guestAAfter.ServerTick, "msg-three-guest-a-move-after-disconnect", "move_intent", map[string]any{"direction": map[string]any{"x": 0, "y": 1}, "duration_ticks": 1})
	readAccepted(t, guestAConn, "msg-three-guest-a-move-after-disconnect")

	_ = guestAConn.Close()
	_ = hostConn.Close()
	time.Sleep(100 * time.Millisecond)
	assertActiveSessionHidden(t, srv, lateGuest.AccessToken, created.SessionID)
	resp = doHTTP(t, srv, http.MethodPost, "/v0/sessions/"+created.SessionID+"/join", lateGuest.AccessToken, map[string]any{
		"character_id": lateGuestChar.CharacterID,
	})
	if resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("late listed join after empty session status = %d, want 409, body = %s", resp.StatusCode, b)
	}
	_ = resp.Body.Close()
}

func TestWebSocketMalformedMessageGetsError(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	conn := dialWS(t, srv, token, sessionID)

	// First message is the initial snapshot.
	first := readMsg(t, conn)
	if first.Type != "session_snapshot" {
		t.Fatalf("first message = %q, want session_snapshot", first.Type)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{not json")); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Expect a structured error, not a dropped connection.
	for {
		m := readMsg(t, conn)
		if m.Type == "error" {
			var p struct {
				Code string `json:"code"`
			}
			_ = json.Unmarshal(m.Payload, &p)
			if p.Code != "bad_message" {
				t.Fatalf("error code = %q, want bad_message", p.Code)
			}
			return
		}
	}
}

func TestWebSocketCompletesSlice(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	driveSlice(t, srv, token, sessionID)
}

func TestResumeSnapshotMatchesStateEndpoint(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	itemID := driveSlice(t, srv, token, sessionID)

	conn := dialWS(t, srv, token, sessionID)
	defer conn.Close()
	resumed := readSnapshot(t, conn)
	assertResumeSliceSnapshot(t, resumed, itemID)

	state := fetchState(t, srv, token, sessionID)
	if !reflect.DeepEqual(resumed, state) {
		t.Fatalf("resume snapshot != state endpoint\nresume: %+v\nstate:  %+v", resumed, state)
	}

	sendIntent(t, conn, sessionID, resumed.ServerTick, "msg-1", "action_intent", map[string]any{"target_id": "1002"})
	rej := readRejected(t, conn, "msg-1")
	if rej.Reason != "duplicate" {
		t.Fatalf("duplicate rejection reason = %q, want duplicate", rej.Reason)
	}
}

func TestCharacterPersistenceLoadsInventoryAndEquipmentAcrossFreshSessions(t *testing.T) {
	srv := fullStack(t)
	token, firstSessionID := loginAndSession(t, srv)
	itemID := driveSlice(t, srv, token, firstSessionID)

	secondSessionID := createSessionWithToken(t, srv, token, "")
	conn := dialWS(t, srv, token, secondSessionID)
	defer conn.Close()
	snap := readSnapshot(t, conn)
	if len(snap.Inventory) != 1 {
		t.Fatalf("fresh persisted inventory count = %d, want 1: %+v", len(snap.Inventory), snap.Inventory)
	}
	if snap.Inventory[0].ItemInstanceID != itemID || snap.Inventory[0].ItemDefID != "rusty_sword" || !snap.Inventory[0].Equipped {
		t.Fatalf("fresh persisted inventory item = %+v, want equipped rusty_sword %s", snap.Inventory[0], itemID)
	}
	if snap.Equipped["main_hand"] == nil || *snap.Equipped["main_hand"] != itemID {
		t.Fatalf("fresh persisted equipped main_hand = %v, want %s", snap.Equipped["main_hand"], itemID)
	}
}

func TestCharacterProgressionPersistsAcrossStateResumeAndFreshSession(t *testing.T) {
	srv := fullStackWithRules(t, func(rules *game.Rules) {
		dummy := rules.Monsters["training_dummy"]
		dummy.MaxHP = 1
		dummy.XPReward = 20
		dummy.LootTable = "no_drop"
		dummy.RetaliationDamage = nil
		rules.Monsters["training_dummy"] = dummy
	})
	token, sessionID := loginAndSession(t, srv)

	conn := dialWS(t, srv, token, sessionID)
	first := readSnapshot(t, conn)
	if first.CharacterProgression.Level != 1 || first.CharacterProgression.UnspentStatPoints != 0 {
		t.Fatalf("initial progression = %+v, want level 1 no points", first.CharacterProgression)
	}

	sendIntent(t, conn, sessionID, first.ServerTick, "msg-prog-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	tick := waitStateDeltaTick(t, conn, first.ServerTick)
	sendIntent(t, conn, sessionID, tick, "msg-prog-kill", "action_intent", map[string]any{"target_id": "1002"})
	levelDelta := readProgressionDelta(t, conn, 2, 3, 5)
	if !hasWireEvent(levelDelta, "experience_gained") || !hasWireEvent(levelDelta, "character_leveled") {
		t.Fatalf("level delta missing XP events: %+v", levelDelta.Events)
	}

	sendIntent(t, conn, sessionID, levelDelta.Tick, "msg-prog-vit", "allocate_stat_intent", map[string]any{"stat": "vit", "points": 1})
	allocDelta := readProgressionDelta(t, conn, 2, 2, 6)
	if !hasWireEvent(allocDelta, "stat_allocated") || !hasWirePlayerMaxHP(allocDelta, 11) {
		t.Fatalf("allocation delta missing stat event/player max hp: changes=%+v events=%+v", allocDelta.Changes, allocDelta.Events)
	}
	_ = conn.Close()

	state := fetchState(t, srv, token, sessionID)
	assertProgressionSnapshot(t, state, 2, 2, 6)

	resume := dialWS(t, srv, token, sessionID)
	resumed := readSnapshot(t, resume)
	assertProgressionSnapshot(t, resumed, 2, 2, 6)
	_ = resume.Close()

	freshSessionID := createSessionWithToken(t, srv, token, "")
	fresh := dialWS(t, srv, token, freshSessionID)
	freshSnap := readSnapshot(t, fresh)
	assertProgressionSnapshot(t, freshSnap, 2, 2, 6)
	_ = fresh.Close()
}

func TestWebSocketSkillPointSpendAndMagicBoltCast(t *testing.T) {
	srv := fullStackWithRules(t, func(rules *game.Rules) {
		dummy := rules.Monsters["training_dummy"]
		dummy.MaxHP = 1
		dummy.XPReward = 50
		dummy.LootTable = "no_drop"
		dummy.RetaliationDamage = nil
		rules.Monsters["training_dummy"] = dummy
	})
	token, sessionID := loginAndSession(t, srv)

	conn := dialWS(t, srv, token, sessionID)
	first := readSnapshot(t, conn)
	if first.SkillProgression.UnspentSkillPoints != 0 || len(first.SkillProgression.Skills) == 0 || first.SkillProgression.Skills[0].Rank != 0 {
		t.Fatalf("initial skill progression = %+v, want rank 0 and no points", first.SkillProgression)
	}

	sendIntent(t, conn, sessionID, first.ServerTick, "msg-skill-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	tick := waitStateDeltaTick(t, conn, first.ServerTick)
	sendIntent(t, conn, sessionID, tick, "msg-skill-kill", "action_intent", map[string]any{"target_id": "1002"})
	levelDelta := readSkillProgressionDelta(t, conn, 1, 0)
	if !hasWireEvent(levelDelta, "skill_point_gained") {
		t.Fatalf("skill point delta missing event: %+v", levelDelta.Events)
	}

	sendIntent(t, conn, sessionID, levelDelta.Tick, "msg-skill-spend", "allocate_skill_point_intent", map[string]any{"skill_id": "magic_bolt"})
	spendDelta := readSkillProgressionDelta(t, conn, 0, 1)
	if !hasWireEvent(spendDelta, "skill_rank_updated") {
		t.Fatalf("skill spend delta missing event: %+v", spendDelta.Events)
	}

	sendIntent(t, conn, sessionID, spendDelta.Tick, "msg-skill-cast", "cast_skill_intent", map[string]any{"skill_id": "magic_bolt", "direction": map[string]any{"x": 1, "y": 0}})
	castDelta := readSkillCooldownDelta(t, conn, "magic_bolt", 40, 40)
	if !hasWireEvent(castDelta, "skill_cast") || !hasWireEvent(castDelta, "skill_cooldown_started") {
		t.Fatalf("cast delta missing skill events: %+v", castDelta.Events)
	}

	sendIntent(t, conn, sessionID, castDelta.Tick, "msg-skill-recast", "cast_skill_intent", map[string]any{"skill_id": "magic_bolt", "direction": map[string]any{"x": 1, "y": 0}})
	reject := readRejected(t, conn, "msg-skill-recast")
	if reject.Reason != "skill_on_cooldown" {
		t.Fatalf("recast reject = %+v, want skill_on_cooldown", reject)
	}
	_ = conn.Close()
}

func TestPostResumePickupAllocatesAfterHistoricalEntities(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)

	conn := dialWS(t, srv, token, sessionID)
	first := readMsg(t, conn)
	if first.Type != "session_snapshot" {
		t.Fatalf("first = %q, want session_snapshot", first.Type)
	}
	lootID := killUntilLoot(t, conn, sessionID, first.Tick)
	_ = conn.Close()

	resume := dialWS(t, srv, token, sessionID)
	defer resume.Close()
	snap := readSnapshot(t, resume)
	if findEntity(snap, lootID) == nil {
		t.Fatalf("resume snapshot missing historical loot entity %s: %+v", lootID, snap.Entities)
	}
	sendIntent(t, resume, sessionID, snap.ServerTick, "msg-pick-after-resume", "action_intent", map[string]any{"target_id": lootID})
	item := readInventoryAdd(t, resume)
	if item.ItemInstanceID == lootID {
		t.Fatalf("post-resume item id collided with loot entity id %s", lootID)
	}
	if item.ItemInstanceID != "1004" {
		t.Fatalf("post-resume item id = %s, want next historical id 1004", item.ItemInstanceID)
	}
}

func TestDeadPlayerResumeRejectsGameplayIntents(t *testing.T) {
	srv := fullStackWithRules(t, func(rules *game.Rules) {
		dmg := game.DamageRange{Min: 11, Max: 11}
		dummy := rules.Monsters["training_dummy"]
		dummy.MaxHP = 100
		dummy.RetaliationDamage = &dmg
		rules.Monsters["training_dummy"] = dummy
		rules.Combat.PlayerDamage = game.DamageRange{Min: 1, Max: 1}
	})
	token, sessionID := loginAndSession(t, srv)

	conn := dialWS(t, srv, token, sessionID)
	first := readMsg(t, conn)
	if first.Type != "session_snapshot" {
		t.Fatalf("first = %q, want session_snapshot", first.Type)
	}
	sendIntent(t, conn, sessionID, first.Tick, "msg-move-lethal", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	moveTick := waitStateDeltaTick(t, conn, first.Tick)
	sendIntent(t, conn, sessionID, moveTick, "msg-lethal", "action_intent", map[string]any{"target_id": "1002"})
	readEvent(t, conn, "player_killed")
	_ = conn.Close()

	charactersResp := doHTTP(t, srv, http.MethodGet, "/v0/characters", token, nil)
	var characters listCharactersResponse
	mustJSON(t, charactersResp, &characters)
	if len(characters.Characters) == 0 || !characters.Characters[0].Dead {
		t.Fatalf("dead character not listed as dead: %+v", characters.Characters)
	}
	createResp := doHTTP(t, srv, http.MethodPost, "/v0/sessions", token, map[string]any{"mode": "solo"})
	if createResp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(createResp.Body)
		_ = createResp.Body.Close()
		t.Fatalf("fresh session after death status = %d, want 409, body = %s", createResp.StatusCode, body)
	}
	_ = createResp.Body.Close()

	resume := dialWS(t, srv, token, sessionID)
	defer resume.Close()
	snap := readSnapshot(t, resume)
	player := findEntity(snap, "1001")
	if player == nil || player.HP == nil || *player.HP != 0 {
		t.Fatalf("resumed player = %+v, want hp 0", player)
	}

	intents := []struct {
		id      string
		typ     string
		payload any
	}{
		{"msg-dead-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1}},
		{"msg-dead-attack", "action_intent", map[string]any{"target_id": "1002"}},
		{"msg-dead-pickup", "action_intent", map[string]any{"target_id": "1003"}},
		{"msg-dead-equip", "equip_intent", map[string]any{"item_instance_id": "1004", "slot": "main_hand"}},
	}
	for _, in := range intents {
		sendIntent(t, resume, sessionID, snap.ServerTick, in.id, in.typ, in.payload)
	}
	for _, in := range intents {
		rej := readRejected(t, resume, in.id)
		if rej.Reason != "player_dead" {
			t.Fatalf("%s rejection reason = %q, want player_dead", in.typ, rej.Reason)
		}
	}
}

// driveSlice plays the full vertical slice over the WebSocket protocol and
// returns the equipped item's instance id. It fatals on any failure.
func driveSlice(t *testing.T, srv *httptest.Server, token, sessionID string) string {
	t.Helper()
	conn := dialWS(t, srv, token, sessionID)

	// Background reader: pushes every received message to a channel so the
	// driver can act on a timer even when the (change-only) server is quiet.
	recv := make(chan wMsg, 256)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				close(recv)
				return
			}
			var m wMsg
			if json.Unmarshal(data, &m) == nil {
				recv <- m
			}
		}
	}()

	first := <-recv
	if first.Type != "session_snapshot" {
		t.Fatalf("first = %q", first.Type)
	}

	var lastTick uint64
	var lootID, itemID string
	movedIntoRange, killed, pickedUp, equipSent, equipped := false, false, false, false, false
	seq := 0
	send := func(typ string, payload any) {
		seq++
		env := map[string]any{
			"type": typ, "message_id": "msg-" + strconv.Itoa(seq),
			"session_id": sessionID, "tick": lastTick, "payload": payload,
		}
		if err := conn.WriteJSON(env); err != nil {
			t.Fatalf("send %s: %v", typ, err)
		}
	}
	lastTick = first.Tick
	send("move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})

	attackTicker := time.NewTicker(120 * time.Millisecond)
	defer attackTicker.Stop()
	overall := time.After(10 * time.Second)

	for !equipped {
		select {
		case <-overall:
			t.Fatalf("slice stalled: killed=%v pickedUp=%v equipped=%v", killed, pickedUp, equipped)
		case <-attackTicker.C:
			if movedIntoRange && !killed {
				send("action_intent", map[string]any{"target_id": "1002"})
			}
		case m, ok := <-recv:
			if !ok {
				t.Fatal("connection closed mid-slice")
			}
			if m.Tick > lastTick {
				lastTick = m.Tick
			}
			if m.Type != "state_delta" {
				continue
			}
			var d struct {
				Changes []wChange `json:"changes"`
				Events  []wEvent  `json:"events"`
			}
			_ = json.Unmarshal(m.Payload, &d)
			for _, ev := range d.Events {
				if ev.EventType == "monster_killed" {
					killed = true
				}
			}
			for _, c := range d.Changes {
				if c.Op == "entity_update" && c.Entity != nil && c.Entity.Type == "player" {
					movedIntoRange = true
				}
				if c.Op == "entity_spawn" && c.Entity != nil && c.Entity.Type == "loot" {
					lootID = c.Entity.ID
				}
				if c.Op == "inventory_add" && c.Item != nil {
					itemID = c.Item.ItemInstanceID
				}
				// Equip is confirmed only when the authoritative delta reports it.
				if c.Op == "equipped_update" && c.Slot == "main_hand" && c.ItemInstanceID != nil && *c.ItemInstanceID == itemID {
					equipped = true
				}
			}
			if killed && !pickedUp && lootID != "" {
				send("action_intent", map[string]any{"target_id": lootID})
				pickedUp = true
			}
			if pickedUp && itemID != "" && !equipSent {
				send("equip_intent", map[string]any{"item_instance_id": itemID, "slot": "main_hand"})
				equipSent = true
			}
		}
	}

	// Confirm via a fresh snapshot triggered by client_ready.
	send("client_ready", map[string]any{"client_version": "test", "last_seen_tick": int(lastTick)})
	confirm := time.After(3 * time.Second)
	for {
		select {
		case <-confirm:
			t.Fatal("no confirming snapshot after equip")
		case m, ok := <-recv:
			if !ok {
				t.Fatal("connection closed before confirm")
			}
			if m.Type != "session_snapshot" {
				continue
			}
			var s struct {
				Inventory []wItem            `json:"inventory"`
				Equipped  map[string]*string `json:"equipped"`
			}
			_ = json.Unmarshal(m.Payload, &s)
			if len(s.Inventory) == 1 && s.Inventory[0].ItemDefID == "rusty_sword" && s.Inventory[0].Equipped &&
				s.Equipped["main_hand"] != nil && *s.Equipped["main_hand"] == itemID {
				_ = conn.Close()
				return itemID // success
			}
		}
	}
}

// --- small helpers ----------------------------------------------------------

func loginWS(t *testing.T, srv *httptest.Server, label string) devLoginResponse {
	t.Helper()
	resp := doHTTP(t, srv, http.MethodPost, "/v0/auth/dev-login", "", map[string]string{
		"email": label + "+" + ids.Token()[:12] + "@example.test", "dev_token": testDevToken,
	})
	var res devLoginResponse
	mustJSON(t, resp, &res)
	return res
}

func createCharacterWS(t *testing.T, srv *httptest.Server, token, name string) characterResponse {
	t.Helper()
	resp := doHTTP(t, srv, http.MethodPost, "/v0/characters", token, map[string]string{"name": name})
	var res characterResponse
	mustJSON(t, resp, &res)
	return res
}

func readSnapshot(t *testing.T, conn *websocket.Conn) game.Snapshot {
	t.Helper()
	m := readMsg(t, conn)
	if m.Type != "session_snapshot" {
		t.Fatalf("message = %q, want session_snapshot", m.Type)
	}
	var snap game.Snapshot
	if err := json.Unmarshal(m.Payload, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	return snap
}

func readSnapshotEventually(t *testing.T, conn *websocket.Conn) game.Snapshot {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("no session_snapshot before timeout")
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "session_snapshot" {
			continue
		}
		var snap game.Snapshot
		if err := json.Unmarshal(m.Payload, &snap); err != nil {
			t.Fatalf("decode snapshot: %v", err)
		}
		return snap
	}
}

func readMsg(t *testing.T, conn *websocket.Conn) wMsg {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read ws: %v", err)
	}
	_ = conn.SetReadDeadline(time.Time{})
	var m wMsg
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode ws: %v", err)
	}
	return m
}

func requestSnapshot(t *testing.T, conn *websocket.Conn, sessionID, messageID string) {
	t.Helper()
	sendIntent(t, conn, sessionID, 0, messageID, "client_ready", map[string]any{"client_version": "test", "last_seen_tick": 0})
}

func sendIntent(t *testing.T, conn *websocket.Conn, sessionID string, tick uint64, messageID, typ string, payload any) {
	t.Helper()
	env := map[string]any{
		"type":       typ,
		"message_id": messageID,
		"session_id": sessionID,
		"tick":       tick,
		"payload":    payload,
	}
	if err := conn.WriteJSON(env); err != nil {
		t.Fatalf("send %s: %v", typ, err)
	}
}

func readAccepted(t *testing.T, conn *websocket.Conn, messageID string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no intent_accepted for %s", messageID)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "intent_accepted" {
			continue
		}
		var p struct {
			AcceptedMessageID string `json:"accepted_message_id"`
		}
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			t.Fatalf("decode accepted: %v", err)
		}
		if p.AcceptedMessageID == messageID {
			return
		}
	}
}

type rejectPayload struct {
	RejectedMessageID string `json:"rejected_message_id"`
	Reason            string `json:"reason"`
}

func readRejected(t *testing.T, conn *websocket.Conn, messageID string) rejectPayload {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no intent_rejected for %s", messageID)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "intent_rejected" {
			continue
		}
		var p rejectPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			t.Fatalf("decode reject: %v", err)
		}
		if p.RejectedMessageID == messageID {
			return p
		}
	}
}

func readEntityRemove(t *testing.T, conn *websocket.Conn, entityID string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no entity_remove for %s", entityID)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d wireDelta
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, change := range d.Changes {
			if change.Op == "entity_remove" && change.EntityID == entityID {
				return
			}
		}
	}
}

func readEvent(t *testing.T, conn *websocket.Conn, eventType string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no state_delta event %s", eventType)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d struct {
			Events []wEvent `json:"events"`
		}
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, ev := range d.Events {
			if ev.EventType == eventType {
				return
			}
		}
	}
}

func readProgressionDelta(t *testing.T, conn *websocket.Conn, level, unspent, vit int) wireDelta {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no progression delta level=%d unspent=%d vit=%d", level, unspent, vit)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d wireDelta
		d.Tick = m.Tick
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, change := range d.Changes {
			if change.Op != "character_progression_update" || change.CharacterProgression == nil {
				continue
			}
			p := change.CharacterProgression
			if p.Level == level && p.UnspentStatPoints == unspent && p.BaseStats.Vit == vit {
				return d
			}
		}
	}
}

func readSkillProgressionDelta(t *testing.T, conn *websocket.Conn, unspent, rank int) wireDelta {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no skill progression delta unspent=%d rank=%d", unspent, rank)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d wireDelta
		d.Tick = m.Tick
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, change := range d.Changes {
			if change.Op != "skill_progression_update" || change.SkillProgression == nil {
				continue
			}
			p := change.SkillProgression
			if p.UnspentSkillPoints != unspent {
				continue
			}
			for _, skill := range p.Skills {
				if skill.SkillID == "magic_bolt" && skill.Rank == rank {
					return d
				}
			}
		}
	}
}

func readSkillCooldownDelta(t *testing.T, conn *websocket.Conn, skillID string, remaining, total int) wireDelta {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("no skill cooldown delta %s %d/%d", skillID, remaining, total)
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d wireDelta
		d.Tick = m.Tick
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, change := range d.Changes {
			if change.Op != "skill_cooldown_update" {
				continue
			}
			for _, cooldown := range change.SkillCooldowns {
				if cooldown.SkillID == skillID && cooldown.RemainingTicks == remaining && cooldown.TotalTicks == total {
					return d
				}
			}
		}
	}
}

func hasWireEvent(delta wireDelta, eventType string) bool {
	for _, ev := range delta.Events {
		if ev.EventType == eventType {
			return true
		}
	}
	return false
}

func hasWirePlayerMaxHP(delta wireDelta, maxHP int) bool {
	for _, change := range delta.Changes {
		if change.Op == "entity_update" && change.Entity != nil && change.Entity.Type == "player" && change.Entity.MaxHP != nil && *change.Entity.MaxHP == maxHP {
			return true
		}
	}
	return false
}

func assertProgressionSnapshot(t *testing.T, snap game.Snapshot, level, unspent, vit int) {
	t.Helper()
	p := snap.CharacterProgression
	if p.Level != level || p.UnspentStatPoints != unspent || p.BaseStats.Vit != vit || int(p.DerivedStats.MaxHP) != 11 {
		t.Fatalf("snapshot progression = %+v, want level=%d unspent=%d vit=%d max_hp=11", p, level, unspent, vit)
	}
	player := findEntity(snap, "1001")
	if player == nil || player.MaxHP == nil || *player.MaxHP != 11 {
		t.Fatalf("snapshot player = %+v, want max_hp 11", player)
	}
}

func waitStateDeltaTick(t *testing.T, conn *websocket.Conn, fallback uint64) uint64 {
	t.Helper()
	deadline := time.After(5 * time.Second)
	lastTick := fallback
	for {
		select {
		case <-deadline:
			t.Fatal("no state_delta before timeout")
		default:
		}
		m := readMsg(t, conn)
		if m.Tick > lastTick {
			lastTick = m.Tick
		}
		if m.Type == "state_delta" {
			return lastTick
		}
	}
}

func killUntilLoot(t *testing.T, conn *websocket.Conn, sessionID string, startTick uint64) string {
	t.Helper()
	lastTick := startTick
	seq := 0
	deadline := time.After(5 * time.Second)
	sendIntent(t, conn, sessionID, lastTick, "msg-pre-pick-move", "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	lastTick = waitStateDeltaTick(t, conn, lastTick)
	for {
		seq++
		sendIntent(t, conn, sessionID, lastTick, "msg-pre-pick-attack-"+strconv.Itoa(seq), "action_intent", map[string]any{"target_id": "1002"})
		select {
		case <-deadline:
			t.Fatal("no loot before timeout")
		default:
		}
		m := readMsg(t, conn)
		if m.Tick > lastTick {
			lastTick = m.Tick
		}
		if m.Type != "state_delta" {
			continue
		}
		var d struct {
			Changes []wChange `json:"changes"`
		}
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, c := range d.Changes {
			if c.Op == "entity_spawn" && c.Entity != nil && c.Entity.Type == "loot" {
				return c.Entity.ID
			}
		}
	}
}

func readInventoryAdd(t *testing.T, conn *websocket.Conn) wItem {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("no inventory_add before timeout")
		default:
		}
		m := readMsg(t, conn)
		if m.Type != "state_delta" {
			continue
		}
		var d struct {
			Changes []wChange `json:"changes"`
		}
		if err := json.Unmarshal(m.Payload, &d); err != nil {
			t.Fatalf("decode delta: %v", err)
		}
		for _, c := range d.Changes {
			if c.Op == "inventory_add" && c.Item != nil {
				return *c.Item
			}
		}
	}
}

func fetchState(t *testing.T, srv *httptest.Server, token, sessionID string) game.Snapshot {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/v0/sessions/"+sessionID+"/state", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Debug-Token", testDebugToken)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("fetch state: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("fetch state: status %d body %s", resp.StatusCode, b)
	}
	var snap game.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	return snap
}

func assertResumeSliceSnapshot(t *testing.T, snap game.Snapshot, itemID string) {
	t.Helper()
	player := findEntity(snap, "1001")
	if player == nil || player.HP == nil || *player.HP >= 10 {
		t.Fatalf("resumed player = %+v, want reduced hp", player)
	}
	monster := findEntity(snap, "1002")
	if monster == nil || monster.HP == nil || *monster.HP != 0 {
		t.Fatalf("resumed monster = %+v, want hp 0", monster)
	}
	if len(snap.Inventory) != 1 || snap.Inventory[0].ItemDefID != "rusty_sword" || !snap.Inventory[0].Equipped {
		t.Fatalf("resumed inventory = %+v, want equipped rusty_sword", snap.Inventory)
	}
	if snap.Equipped["main_hand"] == nil || *snap.Equipped["main_hand"] != itemID {
		t.Fatalf("resumed equipped main_hand = %v, want %s", snap.Equipped["main_hand"], itemID)
	}
}

func findEntity(snap game.Snapshot, id string) *game.EntityView {
	for i := range snap.Entities {
		if snap.Entities[i].ID == id {
			return &snap.Entities[i]
		}
	}
	return nil
}

func entityPosition(t *testing.T, snap game.Snapshot, id string) game.Vec2 {
	t.Helper()
	entity := findEntity(snap, id)
	if entity == nil {
		t.Fatalf("snapshot missing entity %s: %+v", id, snap.Entities)
	}
	return entity.Position
}

func countSnapshotPlayers(snap game.Snapshot) int {
	count := 0
	for _, ent := range snap.Entities {
		if ent.Type == "player" {
			count++
		}
	}
	return count
}

func assertActiveSessionVisible(t *testing.T, srv *httptest.Server, token, sessionID string, wantMembers, wantConnected int) {
	t.Helper()
	resp := doHTTP(t, srv, http.MethodGet, "/v0/sessions/active", token, nil)
	var active activeSessionsResponse
	mustJSON(t, resp, &active)
	for _, sess := range active.Sessions {
		if sess.SessionID == sessionID {
			if sess.MemberCount != wantMembers || sess.ConnectedCount != wantConnected {
				t.Fatalf("active session counts = members %d connected %d, want %d/%d: %+v", sess.MemberCount, sess.ConnectedCount, wantMembers, wantConnected, sess)
			}
			return
		}
	}
	t.Fatalf("active session %s not found in %+v", sessionID, active.Sessions)
}

func assertActiveSessionHidden(t *testing.T, srv *httptest.Server, token, sessionID string) {
	t.Helper()
	resp := doHTTP(t, srv, http.MethodGet, "/v0/sessions/active", token, nil)
	var active activeSessionsResponse
	mustJSON(t, resp, &active)
	for _, sess := range active.Sessions {
		if sess.SessionID == sessionID {
			t.Fatalf("session %s should be hidden from active list: %+v", sessionID, active.Sessions)
		}
	}
}

func doHTTP(t *testing.T, srv *httptest.Server, method, path, bearer string, body any) *http.Response {
	t.Helper()
	buf, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(string(buf)))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func mustJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, b)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
