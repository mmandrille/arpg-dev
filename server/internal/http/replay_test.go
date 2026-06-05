package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// testStoreAndRules opens a second store handle + loads rules for tests that
// inspect/verify recorded sessions directly.
func testStoreAndRules(t *testing.T) (*store.Store, *game.Rules) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := store.Connect(ctx, "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable")
	if err != nil {
		t.Skipf("skipping replay test: no Postgres: %v", err)
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
	return db, rules
}

func TestReplayVerifiesRecordedSlice(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	itemID := driveSlice(t, srv, token, sessionID)

	db, rules := testStoreAndRules(t)
	rep, err := replay.Verify(context.Background(), db, rules, sessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !rep.Match {
		t.Fatalf("replay mismatch on a clean recording: %s", rep.Mismatch)
	}
	if rep.RecordedEventCount == 0 || rep.InputCount == 0 {
		t.Fatalf("expected recorded inputs+events, got %+v", rep)
	}
	// Reconstructed snapshot must show the equipped sword.
	if rep.Snapshot.Equipped["weapon"] == nil || *rep.Snapshot.Equipped["weapon"] != itemID {
		t.Fatalf("reconstructed equipped weapon = %v, want %s", rep.Snapshot.Equipped["weapon"], itemID)
	}
}

func TestReplayDetectsMismatch(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	driveSlice(t, srv, token, sessionID)

	db, rules := testStoreAndRules(t)
	// Corrupt the recorded output: inject a bogus extra event.
	err := db.AppendEvent(context.Background(), store.SessionEvent{
		ID:        ids.New("evt"),
		SessionID: sessionID,
		Tick:      999999,
		Sequence:  0,
		EventType: "bogus_event",
		Payload:   json.RawMessage(`{"event_type":"bogus_event"}`),
	})
	if err != nil {
		t.Fatalf("inject event: %v", err)
	}

	rep, err := replay.Verify(context.Background(), db, rules, sessionID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if rep.Match {
		t.Fatal("expected replay mismatch after corrupting recorded events")
	}
	if rep.Mismatch == "" {
		t.Fatal("expected a mismatch reason")
	}
}

func TestInspectionDebugAuth(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	driveSlice(t, srv, token, sessionID)

	base := srv.URL
	statePath := "/v0/sessions/" + sessionID + "/state"

	// Missing debug token -> 403.
	req, _ := http.NewRequest(http.MethodGet, base+statePath, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("state no-debug: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("state without debug token: status = %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()

	// With debug token -> 200 and equipped sword.
	req, _ = http.NewRequest(http.MethodGet, base+statePath, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Debug-Token", testDebugToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("state with debug: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("state with debug: status = %d, body %s", resp.StatusCode, b)
	}
	var snap game.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Equipped["weapon"] == nil {
		t.Fatal("state endpoint: no weapon equipped in reconstructed snapshot")
	}
}

func TestReplayTimelineEndpoint(t *testing.T) {
	srv := fullStack(t)
	token, sessionID := loginAndSession(t, srv)
	driveSlice(t, srv, token, sessionID)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/v0/sessions/"+sessionID+"/replay/timeline", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Debug-Token", testDebugToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("timeline: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("timeline: status = %d, body %s", resp.StatusCode, b)
	}

	var body struct {
		SessionID string `json:"session_id"`
		Seed      string `json:"seed"`
		Envelopes []struct {
			Type      string          `json:"type"`
			MessageID string          `json:"message_id"`
			SessionID string          `json:"session_id"`
			Tick      uint64          `json:"tick"`
			Payload   json.RawMessage `json:"payload"`
		} `json:"envelopes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode timeline: %v", err)
	}
	if body.SessionID != sessionID {
		t.Fatalf("timeline session_id = %q, want %q", body.SessionID, sessionID)
	}
	if len(body.Envelopes) < 2 {
		t.Fatalf("expected snapshot plus deltas, got %d envelopes", len(body.Envelopes))
	}
	if body.Envelopes[0].Type != "session_snapshot" {
		t.Fatalf("first envelope type = %q, want session_snapshot", body.Envelopes[0].Type)
	}
	sawDelta := false
	sawMonsterKilled := false
	for _, env := range body.Envelopes {
		if env.Type != "state_delta" {
			continue
		}
		sawDelta = true
		var payload struct {
			ServerTick uint64       `json:"server_tick"`
			Events     []game.Event `json:"events"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("decode delta payload: %v", err)
		}
		for _, ev := range payload.Events {
			if ev.EventType == "monster_killed" {
				sawMonsterKilled = true
			}
		}
	}
	if !sawDelta {
		t.Fatal("expected at least one state_delta")
	}
	if !sawMonsterKilled {
		t.Fatal("expected replay timeline to include monster_killed event")
	}
}
