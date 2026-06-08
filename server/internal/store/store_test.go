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

func TestDeleteCharacterRemovesProgressionAndSessions(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "delete+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	keep, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Keep")
	if err != nil {
		t.Fatalf("create keep character: %v", err)
	}
	remove, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Remove")
	if err != nil {
		t.Fatalf("create remove character: %v", err)
	}
	defaultProgression := store.CharacterProgressionDefaults{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		Stats:             store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
	}
	if _, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, remove.ID, defaultProgression); err != nil {
		t.Fatalf("create progression: %v", err)
	}
	sess := store.Session{
		ID:          ids.New("sess"),
		AccountID:   acct.ID,
		CharacterID: remove.ID,
		Seed:        "deadbeef",
		WorldID:     "dungeon_levels",
		Status:      store.SessionActive,
	}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := s.DeleteCharacter(ctx, acct.ID, remove.ID); err != nil {
		t.Fatalf("delete character: %v", err)
	}
	if _, err := s.GetCharacter(ctx, remove.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("deleted character still exists: %v", err)
	}
	if _, err := s.GetSession(ctx, sess.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("deleted character session still exists: %v", err)
	}

	chars, err := s.ListCharacters(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list characters: %v", err)
	}
	if len(chars) != 1 || chars[0].ID != keep.ID {
		t.Fatalf("remaining characters = %+v, want only %s", chars, keep.ID)
	}
	if err := s.DeleteCharacter(ctx, acct.ID, "char_missing"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("delete missing character = %v, want ErrNotFound", err)
	}
}

func TestCoopSessionMembersActorInputsAndSnapshots(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	hostAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "host+"+ids.Token()[:12]+"@example.test")
	hostChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), hostAcct.ID, "Host")
	guestAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "guest+"+ids.Token()[:12]+"@example.test")
	guestChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), guestAcct.ID, "Guest")
	thirdAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "third+"+ids.Token()[:12]+"@example.test")
	thirdChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), thirdAcct.ID, "Third")

	sess := store.Session{
		ID:           ids.New("sess"),
		AccountID:    hostAcct.ID,
		CharacterID:  hostChar.ID,
		Seed:         "c001",
		WorldID:      "dungeon_levels",
		Mode:         store.SessionModeCoop,
		JoinCodeHash: "join_hash",
		Status:       store.SessionActive,
	}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create coop session: %v", err)
	}
	loaded, err := s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load coop session: %v", err)
	}
	if loaded.Mode != store.SessionModeCoop || loaded.JoinCodeHash != "join_hash" {
		t.Fatalf("coop session metadata mismatch: %+v", loaded)
	}

	if err := s.CreateSessionHostMember(ctx, store.SessionMember{
		SessionID:      sess.ID,
		AccountID:      hostAcct.ID,
		CharacterID:    hostChar.ID,
		PlayerEntityID: "1001",
		Role:           store.SessionMemberHost,
		CurrentLevel:   -1,
	}); err != nil {
		t.Fatalf("create host member: %v", err)
	}
	if err := s.CreateSessionGuestMember(ctx, store.SessionMember{
		SessionID:      sess.ID,
		AccountID:      guestAcct.ID,
		CharacterID:    guestChar.ID,
		PlayerEntityID: "1007",
		CurrentLevel:   0,
	}); err != nil {
		t.Fatalf("create guest member: %v", err)
	}
	if err := s.CreateSessionGuestMember(ctx, store.SessionMember{
		SessionID:   sess.ID,
		AccountID:   guestAcct.ID,
		CharacterID: guestChar.ID,
	}); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("duplicate guest = %v, want ErrConflict", err)
	}
	if err := s.CreateSessionGuestMember(ctx, store.SessionMember{
		SessionID:   sess.ID,
		AccountID:   thirdAcct.ID,
		CharacterID: thirdChar.ID,
	}); !errors.Is(err, store.ErrPartyFull) {
		t.Fatalf("third guest = %v, want ErrPartyFull", err)
	}

	members, err := s.ListSessionMembers(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 2 || members[0].Role != store.SessionMemberHost || members[1].Role != store.SessionMemberGuest {
		t.Fatalf("members order = %+v", members)
	}
	if err := s.SetSessionMemberConnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1007", 0, 9); err != nil {
		t.Fatalf("connect guest: %v", err)
	}
	member, err := s.GetSessionMemberByAccount(ctx, sess.ID, guestAcct.ID)
	if err != nil {
		t.Fatalf("get member by account: %v", err)
	}
	if !member.Connected || member.PlayerEntityID != "1007" || member.CurrentLevel != 0 {
		t.Fatalf("connected member mismatch: %+v", member)
	}
	if err := s.SetSessionMemberDisconnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, -1, 12); err != nil {
		t.Fatalf("disconnect guest: %v", err)
	}
	member, _ = s.GetSessionMember(ctx, sess.ID, guestAcct.ID, guestChar.ID)
	if member.Connected || member.LeftTick == nil || *member.LeftTick != 12 || member.CurrentLevel != -1 {
		t.Fatalf("disconnected member mismatch: %+v", member)
	}

	if err := s.AppendInput(ctx, store.SessionInput{
		ID:                  ids.New("inp"),
		SessionID:           sess.ID,
		Tick:                3,
		Sequence:            1,
		MessageID:           ids.New("msg"),
		ActorAccountID:      guestAcct.ID,
		ActorCharacterID:    guestChar.ID,
		ActorPlayerEntityID: "1007",
		Payload:             json.RawMessage(`{"type":"move_intent"}`),
	}); err != nil {
		t.Fatalf("append actor input: %v", err)
	}
	inputs, err := s.ListInputs(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list inputs: %v", err)
	}
	if len(inputs) != 1 || inputs[0].ActorAccountID != guestAcct.ID || inputs[0].ActorCharacterID != guestChar.ID || inputs[0].ActorPlayerEntityID != "1007" {
		t.Fatalf("actor input mismatch: %+v", inputs)
	}

	defaultProgression := store.CharacterProgressionDefaults{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		Stats:             store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
	}
	hostProgression, _ := s.GetOrCreateCharacterProgression(ctx, hostAcct.ID, hostChar.ID, defaultProgression)
	guestProgression, _ := s.GetOrCreateCharacterProgression(ctx, guestAcct.ID, guestChar.ID, defaultProgression)
	hostItem := store.CharacterItemInstance{ID: "2001", AccountID: hostAcct.ID, CharacterID: hostChar.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory}
	guestItem := store.CharacterItemInstance{ID: "2001", AccountID: guestAcct.ID, CharacterID: guestChar.ID, ItemDefID: "cave_bow", Location: store.ItemLocationInventory}
	hostHotbar := []store.CharacterHotbarSlot{{AccountID: hostAcct.ID, CharacterID: hostChar.ID, SlotIndex: 0, ItemInstanceID: &hostItem.ID}}
	guestHotbar := []store.CharacterHotbarSlot{{AccountID: guestAcct.ID, CharacterID: guestChar.ID, SlotIndex: 0, ItemInstanceID: &guestItem.ID}}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, hostAcct.ID, hostChar.ID, []store.CharacterItemInstance{hostItem}, nil, hostHotbar, hostProgression); err != nil {
		t.Fatalf("host start snapshot: %v", err)
	}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, guestAcct.ID, guestChar.ID, []store.CharacterItemInstance{guestItem}, nil, guestHotbar, guestProgression); err != nil {
		t.Fatalf("guest start snapshot: %v", err)
	}
	snaps, err := s.LoadSessionStartSnapshots(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load start snapshots: %v", err)
	}
	if len(snaps) != 2 || len(snaps[0].Items) != 1 || len(snaps[1].Items) != 1 {
		t.Fatalf("snapshot count mismatch: %+v", snaps)
	}
	if snaps[0].Items[0].ItemDefID != "cave_blade" || snaps[1].Items[0].ItemDefID != "cave_bow" {
		t.Fatalf("member snapshots collided: %+v", snaps)
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

	if err := s.SetCharacterItemEquipped(ctx, acct.ID, char.ID, item.ID, "main_hand", true); err != nil {
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
	if !items[0].Equipped || items[0].Location != store.ItemLocationEquipped || items[0].Slot != "main_hand" || items[0].ItemDefID != "cave_blade" {
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

	hotbar, err := s.ListCharacterHotbar(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list hotbar: %v", err)
	}
	if len(hotbar) != 10 {
		t.Fatalf("hotbar slots = %d, want 10", len(hotbar))
	}
	if err := s.SetCharacterHotbarSlot(ctx, acct.ID, char.ID, 2, &item.ID); err != nil {
		t.Fatalf("set hotbar slot: %v", err)
	}
	hotbar, err = s.ListCharacterHotbar(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("reload hotbar: %v", err)
	}
	if hotbar[2].ItemInstanceID == nil || *hotbar[2].ItemInstanceID != item.ID {
		t.Fatalf("hotbar slot 2 = %+v, want item %s", hotbar[2], item.ID)
	}

	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, acct.ID, char.ID, items, waypoints, hotbar, loadedProgression); err != nil {
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
	if err := s.SetCharacterHotbarSlot(ctx, acct.ID, char.ID, 2, nil); err != nil {
		t.Fatalf("mutate live hotbar: %v", err)
	}
	snap, err := s.LoadSessionStartSnapshot(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load session snapshot: %v", err)
	}
	if len(snap.Items) != 1 || !snap.Items[0].Equipped || snap.Items[0].Slot != "main_hand" {
		t.Fatalf("snapshot item mutated with live state: %+v", snap.Items)
	}
	if len(snap.Hotbar) != 10 || snap.Hotbar[2].ItemInstanceID == nil || *snap.Hotbar[2].ItemInstanceID != item.ID {
		t.Fatalf("snapshot hotbar mutated with live state: %+v", snap.Hotbar)
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
	if err := s.SetCharacterItemEquipped(ctx, otherAcct.ID, otherChar.ID, item.ID, "main_hand", true); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("equip missing item: expected ErrNotFound, got %v", err)
	}
	if err := s.SetCharacterHotbarSlot(ctx, otherAcct.ID, otherChar.ID, 2, &item.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign hotbar assign: expected ErrNotFound, got %v", err)
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
	hotbar, err = s.ListCharacterHotbar(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list hotbar after remove: %v", err)
	}
	if hotbar[2].ItemInstanceID != nil {
		t.Fatalf("removed item still assigned in hotbar: %+v", hotbar[2])
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
