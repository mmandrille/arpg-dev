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

func TestCharacterProgressionPersistEquipWaypointAndSnapshot(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "inv+"+ids.Token()[:12]+"@example.test")
	char, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), acct.ID, "Hero")
	defaultProgression := store.CharacterProgressionDefaults{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		Stats:             store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
	}
	progression, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, char.ID, defaultProgression)
	if err != nil {
		t.Fatalf("get or create progression: %v", err)
	}
	if progression.Level != 1 || progression.Experience != 0 || progression.UnspentStatPoints != 0 ||
		progression.Stats.Str != 5 || progression.Stats.Dex != 5 || progression.Stats.Vit != 5 || progression.Stats.Magic != 5 {
		t.Fatalf("default progression mismatch: %+v", progression)
	}
	progression.Level = 2
	progression.Experience = 25
	progression.UnspentStatPoints = 5
	progression.Stats.Vit = 6
	if err := s.UpsertCharacterProgression(ctx, acct.ID, progression); err != nil {
		t.Fatalf("upsert progression: %v", err)
	}
	loadedProgression, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, char.ID, store.CharacterProgressionDefaults{
		Level:             9,
		Experience:        999,
		UnspentStatPoints: 99,
		Stats:             store.CharacterBaseStats{Str: 1, Dex: 1, Vit: 1, Magic: 1},
	})
	if err != nil {
		t.Fatalf("reload progression: %v", err)
	}
	if loadedProgression.Level != 2 || loadedProgression.Experience != 25 || loadedProgression.UnspentStatPoints != 5 || loadedProgression.Stats.Vit != 6 {
		t.Fatalf("progression not persisted/stable: %+v", loadedProgression)
	}

	sess := store.Session{ID: ids.New("sess"), AccountID: acct.ID, CharacterID: char.ID, Seed: "ab", WorldID: "vertical_slice", Status: store.SessionActive}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	item := store.CharacterItemInstance{
		ID:          "1004",
		AccountID:   acct.ID,
		CharacterID: char.ID,
		ItemDefID:   "cave_blade",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"item_template_id":"cave_blade","display_name":"Rare Cave Blade","rarity":"rare","stats":{"damage_min":4,"damage_max":5,"max_hp":3},"requirements":{"level":1},"effect_ids":[]}`),
	}
	if err := s.AddCharacterItem(ctx, item); err != nil {
		t.Fatalf("add character item: %v", err)
	}
	if err := s.AddCharacterItem(ctx, item); err != nil {
		t.Fatalf("re-add character item: %v", err)
	}

	if err := s.SetCharacterItemEquipped(ctx, acct.ID, char.ID, item.ID, "weapon", true); err != nil {
		t.Fatalf("set equipped: %v", err)
	}
	if err := s.AddCharacterWaypoint(ctx, char.ID, -1); err != nil {
		t.Fatalf("add waypoint: %v", err)
	}
	if err := s.AddCharacterWaypoint(ctx, char.ID, -1); err != nil {
		t.Fatalf("re-add waypoint: %v", err)
	}

	items, err := s.ListCharacterItems(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list character items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("character item count = %d, want 1", len(items))
	}
	if !items[0].Equipped || items[0].Location != store.ItemLocationEquipped || items[0].Slot != "weapon" || items[0].ItemDefID != "cave_blade" {
		t.Fatalf("character item not persisted/equipped correctly: %+v", items[0])
	}
	var payload struct {
		ItemTemplateID string         `json:"item_template_id"`
		DisplayName    string         `json:"display_name"`
		Rarity         string         `json:"rarity"`
		Stats          map[string]int `json:"stats"`
		Requirements   map[string]int `json:"requirements"`
		EffectIDs      []string       `json:"effect_ids"`
	}
	if err := json.Unmarshal(items[0].RolledStats, &payload); err != nil {
		t.Fatalf("rolled stats payload invalid: %v", err)
	}
	if payload.ItemTemplateID != "cave_blade" || payload.DisplayName != "Rare Cave Blade" || payload.Rarity != "rare" ||
		payload.Stats["damage_min"] != 4 || payload.Stats["damage_max"] != 5 || payload.Stats["max_hp"] != 3 ||
		payload.Requirements["level"] != 1 || len(payload.EffectIDs) != 0 {
		t.Fatalf("rolled stats not preserved: %+v raw=%s", payload, string(items[0].RolledStats))
	}

	waypoints, err := s.ListCharacterWaypoints(ctx, char.ID)
	if err != nil {
		t.Fatalf("list waypoints: %v", err)
	}
	if len(waypoints) != 1 || waypoints[0].Level != -1 {
		t.Fatalf("waypoints = %+v, want level -1", waypoints)
	}

	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, acct.ID, char.ID, items, waypoints, loadedProgression); err != nil {
		t.Fatalf("create session snapshot: %v", err)
	}
	if err := s.SetCharacterItemEquipped(ctx, acct.ID, char.ID, item.ID, "", false); err != nil {
		t.Fatalf("mutate live item: %v", err)
	}
	if err := s.AddCharacterWaypoint(ctx, char.ID, -2); err != nil {
		t.Fatalf("mutate live waypoints: %v", err)
	}
	mutatedProgression := loadedProgression
	mutatedProgression.Level = 3
	mutatedProgression.Experience = 70
	mutatedProgression.UnspentStatPoints = 10
	mutatedProgression.Stats.Str = 7
	if err := s.UpsertCharacterProgression(ctx, acct.ID, mutatedProgression); err != nil {
		t.Fatalf("mutate live progression: %v", err)
	}
	snap, err := s.LoadSessionStartSnapshot(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load session snapshot: %v", err)
	}
	if len(snap.Items) != 1 || !snap.Items[0].Equipped || snap.Items[0].Slot != "weapon" {
		t.Fatalf("snapshot item mutated with live state: %+v", snap.Items)
	}
	if len(snap.Waypoints) != 1 || snap.Waypoints[0].Level != -1 {
		t.Fatalf("snapshot waypoints mutated with live state: %+v", snap.Waypoints)
	}
	if snap.Progression == nil {
		t.Fatalf("snapshot progression missing")
	}
	if snap.Progression.Level != 2 || snap.Progression.Experience != 25 || snap.Progression.UnspentStatPoints != 5 ||
		snap.Progression.Stats.Str != 5 || snap.Progression.Stats.Vit != 6 {
		t.Fatalf("snapshot progression mutated with live state: %+v", snap.Progression)
	}

	otherAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "other+"+ids.Token()[:12]+"@example.test")
	otherChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), otherAcct.ID, "Hero")
	if err := s.SetCharacterItemEquipped(ctx, otherAcct.ID, otherChar.ID, item.ID, "weapon", true); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("equip missing item: expected ErrNotFound, got %v", err)
	}
	if _, err := s.GetCharacterProgression(ctx, otherAcct.ID, char.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign get progression: expected ErrNotFound, got %v", err)
	}
	if err := s.UpsertCharacterProgression(ctx, otherAcct.ID, mutatedProgression); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign update progression: expected ErrNotFound, got %v", err)
	}

	if err := s.RemoveCharacterItem(ctx, acct.ID, char.ID, item.ID); err != nil {
		t.Fatalf("remove character item: %v", err)
	}
	items, err = s.ListCharacterItems(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list after remove: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("character item count after remove = %d, want 0", len(items))
	}
	if err := s.RemoveCharacterItem(ctx, acct.ID, char.ID, item.ID); !errors.Is(err, store.ErrNotFound) {
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
