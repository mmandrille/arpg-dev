package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func testDatabaseURL() string {
	if v := os.Getenv("ARPG_TEST_DATABASE_URL"); v != "" {
		return v
	}
	if v := os.Getenv("ARPG_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"
}

// newStore connects + migrates, or skips when no Postgres is reachable.
func newStore(t *testing.T) *store.Store {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := store.Connect(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("skipping store integration test: cannot connect to Postgres: %v", err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func TestMigrateIdempotent(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	// A second migrate must be a no-op, not an error.
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if err := s.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestAccountCharacterSessionFlow(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	email := "dev+" + ids.Token()[:12] + "@example.test"
	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), email)
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}

	// Upsert with the same email must return the same account id.
	acct2, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), email)
	if err != nil {
		t.Fatalf("upsert account again: %v", err)
	}
	if acct2.ID != acct.ID {
		t.Fatalf("re-upsert changed account id: %s != %s", acct2.ID, acct.ID)
	}

	char, err := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, "Hero")
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	// Second call returns the same character.
	char2, err := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, "Hero")
	if err != nil {
		t.Fatalf("get character: %v", err)
	}
	if char2.ID != char.ID {
		t.Fatalf("default character not stable: %s != %s", char2.ID, char.ID)
	}

	sess := store.Session{
		ID:          ids.New("sess"),
		AccountID:   acct.ID,
		CharacterID: char.ID,
		Seed:        "deadbeef",
		WorldID:     "gear_before_combat",
		Status:      store.SessionActive,
	}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	got, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.Seed != "deadbeef" || got.WorldID != "gear_before_combat" || got.Status != store.SessionActive {
		t.Fatalf("session round-trip mismatch: %+v", got)
	}

	if _, err := s.GetSession(ctx, "sess_does_not_exist"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInventoryPersistAndEquip(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "inv+"+ids.Token()[:12]+"@example.test")
	char, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, "Hero")
	sess := store.Session{ID: ids.New("sess"), AccountID: acct.ID, CharacterID: char.ID, Seed: "ab", WorldID: "vertical_slice", Status: store.SessionActive}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	item := store.InventoryItem{
		ID:          "1004",
		SessionID:   sess.ID,
		AccountID:   acct.ID,
		CharacterID: char.ID,
		ItemDefID:   "rusty_sword",
		Slot:        "",
		Equipped:    false,
	}
	if err := s.AddInventoryItem(ctx, item); err != nil {
		t.Fatalf("add inventory: %v", err)
	}
	// Idempotent: adding the same (session, id) again is a no-op.
	if err := s.AddInventoryItem(ctx, item); err != nil {
		t.Fatalf("re-add inventory: %v", err)
	}

	if err := s.SetEquipped(ctx, sess.ID, item.ID, "weapon", true); err != nil {
		t.Fatalf("set equipped: %v", err)
	}

	items, err := s.ListInventory(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list inventory: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("inventory count = %d, want 1", len(items))
	}
	if !items[0].Equipped || items[0].Slot != "weapon" || items[0].ItemDefID != "rusty_sword" {
		t.Fatalf("inventory item not persisted/equipped correctly: %+v", items[0])
	}

	if err := s.SetEquipped(ctx, sess.ID, "nope", "weapon", true); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("equip missing item: expected ErrNotFound, got %v", err)
	}

	if err := s.RemoveInventoryItem(ctx, sess.ID, item.ID); err != nil {
		t.Fatalf("remove inventory: %v", err)
	}
	items, err = s.ListInventory(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list after remove: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("inventory count after remove = %d, want 0", len(items))
	}
	if err := s.RemoveInventoryItem(ctx, sess.ID, item.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("remove missing item: expected ErrNotFound, got %v", err)
	}
}

func TestInputsAndEventsOrdering(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "ev+"+ids.Token()[:12]+"@example.test")
	char, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, "Hero")
	sess := store.Session{ID: ids.New("sess"), AccountID: acct.ID, CharacterID: char.ID, Seed: "ab", WorldID: "vertical_slice", Status: store.SessionActive}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Insert inputs out of order; expect ordered read by (tick, sequence).
	type in struct{ tick, seq int64 }
	for _, x := range []in{{2, 0}, {1, 1}, {1, 0}} {
		err := s.AppendInput(ctx, store.SessionInput{
			ID:        ids.New("inp"),
			SessionID: sess.ID,
			Tick:      x.tick,
			Sequence:  x.seq,
			MessageID: ids.New("msg"),
			Payload:   json.RawMessage(`{"k":"v"}`),
		})
		if err != nil {
			t.Fatalf("append input: %v", err)
		}
	}
	inputs, err := s.ListInputs(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list inputs: %v", err)
	}
	if len(inputs) != 3 {
		t.Fatalf("inputs = %d, want 3", len(inputs))
	}
	want := []in{{1, 0}, {1, 1}, {2, 0}}
	for i, w := range want {
		if inputs[i].Tick != w.tick || inputs[i].Sequence != w.seq {
			t.Fatalf("input[%d] = (%d,%d), want (%d,%d)", i, inputs[i].Tick, inputs[i].Sequence, w.tick, w.seq)
		}
	}

	// Duplicate message_id within the session is ignored (no error).
	dupMsg := ids.New("msg")
	base := store.SessionInput{ID: ids.New("inp"), SessionID: sess.ID, Tick: 5, Sequence: 0, MessageID: dupMsg, Payload: json.RawMessage(`{}`)}
	if err := s.AppendInput(ctx, base); err != nil {
		t.Fatalf("append dup base: %v", err)
	}
	base.ID = ids.New("inp")
	if err := s.AppendInput(ctx, base); err != nil {
		t.Fatalf("append dup: %v", err)
	}
	inputs, _ = s.ListInputs(ctx, sess.ID)
	if len(inputs) != 4 {
		t.Fatalf("after dup, inputs = %d, want 4", len(inputs))
	}

	if err := s.AppendEvent(ctx, store.SessionEvent{
		ID: ids.New("evt"), SessionID: sess.ID, Tick: 3, Sequence: 0,
		EventType: "monster_killed", CorrelationID: "corr_x", Payload: json.RawMessage(`{"entity_id":"1002"}`),
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	events, err := s.ListEvents(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "monster_killed" {
		t.Fatalf("events round-trip mismatch: %+v", events)
	}
}
