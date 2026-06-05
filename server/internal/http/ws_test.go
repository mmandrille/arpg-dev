package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// fullStack builds a real httptest server (Postgres-backed) including the
// realtime hub, or skips when Postgres/rules are unavailable.
func fullStack(t *testing.T) *httptest.Server {
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
	ID        string `json:"id"`
	Type      string `json:"type"`
	ItemDefID string `json:"item_def_id"`
	HP        *int   `json:"hp"`
}
type wItem struct {
	ItemInstanceID string `json:"item_instance_id"`
	ItemDefID      string `json:"item_def_id"`
	Slot           string `json:"slot"`
	Equipped       bool   `json:"equipped"`
}
type wChange struct {
	Op             string   `json:"op"`
	Entity         *wEntity `json:"entity"`
	EntityID       string   `json:"entity_id"`
	Item           *wItem   `json:"item"`
	Slot           string   `json:"slot"`
	ItemInstanceID *string  `json:"item_instance_id"`
}
type wEvent struct {
	EventType string `json:"event_type"`
	EntityID  string `json:"entity_id"`
}
type wMsg struct {
	Type    string          `json:"type"`
	Tick    uint64          `json:"tick"`
	Payload json.RawMessage `json:"payload"`
}

func loginAndSession(t *testing.T, srv *httptest.Server) (token, sessionID string) {
	t.Helper()
	// dev-login
	rec := doHTTP(t, srv, "POST", "/v0/auth/dev-login", "", map[string]string{
		"email": "ws@example.test", "dev_token": testDevToken,
	})
	var lr devLoginResponse
	mustJSON(t, rec, &lr)
	// create session
	rec = doHTTP(t, srv, "POST", "/v0/sessions", lr.AccessToken, map[string]any{"mode": "solo"})
	var sr createSessionResponse
	mustJSON(t, rec, &sr)
	return lr.AccessToken, sr.SessionID
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
			var p struct{ Code string `json:"code"` }
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
	killed, pickedUp, equipSent, equipped := false, false, false, false
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

	attackTicker := time.NewTicker(120 * time.Millisecond)
	defer attackTicker.Stop()
	overall := time.After(10 * time.Second)

	for !equipped {
		select {
		case <-overall:
			t.Fatalf("slice stalled: killed=%v pickedUp=%v equipped=%v", killed, pickedUp, equipped)
		case <-attackTicker.C:
			if !killed {
				send("attack_intent", map[string]any{"target_id": "1002"})
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
				if c.Op == "entity_spawn" && c.Entity != nil && c.Entity.Type == "loot" {
					lootID = c.Entity.ID
				}
				if c.Op == "inventory_add" && c.Item != nil {
					itemID = c.Item.ItemInstanceID
				}
				// Equip is confirmed only when the authoritative delta reports it.
				if c.Op == "equipped_update" && c.Slot == "weapon" && c.ItemInstanceID != nil && *c.ItemInstanceID == itemID {
					equipped = true
				}
			}
			if killed && !pickedUp && lootID != "" {
				send("pick_up_intent", map[string]any{"entity_id": lootID})
				pickedUp = true
			}
			if pickedUp && itemID != "" && !equipSent {
				send("equip_intent", map[string]any{"item_instance_id": itemID, "slot": "weapon"})
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
				s.Equipped["weapon"] != nil && *s.Equipped["weapon"] == itemID {
				return itemID // success
			}
		}
	}
}

// --- small helpers ----------------------------------------------------------

func readMsg(t *testing.T, conn *websocket.Conn) wMsg {
	t.Helper()
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read ws: %v", err)
	}
	var m wMsg
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode ws: %v", err)
	}
	return m
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
