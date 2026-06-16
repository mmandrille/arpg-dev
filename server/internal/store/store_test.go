package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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
	cleanupMarketRowsForTestAccounts(t)
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
	if char.CharacterClass != "barbarian" {
		t.Fatalf("default character class = %q, want barbarian", char.CharacterClass)
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

	chars, err := s.ListCharacters(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list characters: %v", err)
	}
	if len(chars) != 1 || chars[0].ID != char.ID || chars[0].CharacterClass != "barbarian" || chars[0].Level != 1 || chars[0].Gold != 0 || chars[0].DeepestDungeonDepth != 0 {
		t.Fatalf("default character summary = %+v, want level 1 gold 0 depth 0", chars)
	}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, store.CharacterProgression{
		CharacterID:         char.ID,
		Level:               6,
		Experience:          88,
		Stats:               store.CharacterBaseStats{Str: 6, Dex: 7, Vit: 8, Magic: 9},
		Gold:                123,
		DeepestDungeonDepth: 3,
		SkillRanks:          map[string]int{"magic_bolt": 1},
	}); err != nil {
		t.Fatalf("upsert progression for summary: %v", err)
	}
	chars, err = s.ListCharacters(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list characters after progression: %v", err)
	}
	if len(chars) != 1 || chars[0].Level != 6 || chars[0].Gold != 123 || chars[0].DeepestDungeonDepth != 3 {
		t.Fatalf("progression character summary = %+v, want level 6 gold 123 depth 3", chars)
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
	keep, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Keep", "barbarian")
	if err != nil {
		t.Fatalf("create keep character: %v", err)
	}
	remove, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Remove", "barbarian")
	if err != nil {
		t.Fatalf("create remove character: %v", err)
	}
	defaultProgression := store.CharacterProgressionDefaults{
		Level:              1,
		Experience:         0,
		UnspentStatPoints:  0,
		UnspentSkillPoints: 0,
		Stats:              store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
		SkillRanks:         map[string]int{"magic_bolt": 0},
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
		Listed:       true,
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
	if loaded.Mode != store.SessionModeCoop || !loaded.Listed || loaded.JoinCodeHash != "join_hash" {
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
		JoinedTick:     1,
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
		SessionID:      sess.ID,
		AccountID:      thirdAcct.ID,
		CharacterID:    thirdChar.ID,
		PlayerEntityID: "1011",
		CurrentLevel:   0,
		JoinedTick:     2,
	}); err != nil {
		t.Fatalf("third guest: %v", err)
	}

	members, err := s.ListSessionMembers(ctx, sess.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(members) != 3 || members[0].Role != store.SessionMemberHost || members[1].Role != store.SessionMemberGuest || members[2].Role != store.SessionMemberGuest {
		t.Fatalf("members order = %+v", members)
	}
	summaries, err := s.ListActiveListedSessions(ctx)
	if err != nil {
		t.Fatalf("list active sessions: %v", err)
	}
	if summaryByID(summaries, sess.ID).SessionID != "" {
		t.Fatalf("inactive listed session with no connected members was listed: %+v", summaries)
	}
	claimed, err := s.ClaimSessionMemberConnection(ctx, sess.ID, hostAcct.ID, hostChar.ID)
	if err != nil {
		t.Fatalf("claim host connection: %v", err)
	}
	if !claimed {
		t.Fatalf("host connection was not claimed")
	}
	claimed, err = s.ClaimSessionMemberConnection(ctx, sess.ID, hostAcct.ID, hostChar.ID)
	if err != nil {
		t.Fatalf("claim duplicate host connection: %v", err)
	}
	if claimed {
		t.Fatalf("duplicate host connection was claimed")
	}
	if err := s.SetSessionMemberConnected(ctx, sess.ID, hostAcct.ID, hostChar.ID, "1001", 0, 0); err != nil {
		t.Fatalf("connect host: %v", err)
	}
	summaries, err = s.ListActiveListedSessions(ctx)
	if err != nil {
		t.Fatalf("list active sessions after connect: %v", err)
	}
	found := summaryByID(summaries, sess.ID)
	if found.SessionID == "" || found.HostCharacterID != hostChar.ID || found.HostDisplayName != hostChar.Name || found.MemberCount != 3 || found.ConnectedCount != 1 || !found.Listed {
		t.Fatalf("active listed summary mismatch: %+v", found)
	}
	if err := s.SetSessionMemberDisconnected(ctx, sess.ID, hostAcct.ID, hostChar.ID, 0, 10); err != nil {
		t.Fatalf("disconnect host: %v", err)
	}
	claimed, err = s.ClaimSessionMemberConnection(ctx, sess.ID, hostAcct.ID, hostChar.ID)
	if err != nil {
		t.Fatalf("claim host connection after disconnect: %v", err)
	}
	if !claimed {
		t.Fatalf("host connection was not claimable after disconnect")
	}
	if err := s.SetSessionMemberDisconnected(ctx, sess.ID, hostAcct.ID, hostChar.ID, 0, 11); err != nil {
		t.Fatalf("disconnect host after claim: %v", err)
	}
	summaries, err = s.ListActiveListedSessions(ctx)
	if err != nil {
		t.Fatalf("list active sessions after disconnect: %v", err)
	}
	if summaryByID(summaries, sess.ID).SessionID != "" {
		t.Fatalf("empty listed session was still listed: %+v", summaries)
	}
	ended, err := s.EndListedSessionIfNoConnected(ctx, sess.ID)
	if err != nil {
		t.Fatalf("end empty listed session: %v", err)
	}
	if !ended {
		t.Fatalf("empty listed session was not ended")
	}
	loaded, err = s.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get ended session: %v", err)
	}
	if loaded.Status != store.SessionEnded {
		t.Fatalf("session status = %q, want ended", loaded.Status)
	}
	if ended, err := s.EndListedSessionIfNoConnected(ctx, sess.ID); err != nil || ended {
		t.Fatalf("second empty listed end = %v, %v; want false, nil", ended, err)
	}
	if err := s.SetSessionMemberConnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1007", 0, 9); err != nil {
		t.Fatalf("connect guest: %v", err)
	}
	summaries, err = s.ListActiveListedSessions(ctx)
	if err != nil {
		t.Fatalf("list active sessions after ended reconnect: %v", err)
	}
	for _, summary := range summaries {
		if summary.SessionID == sess.ID {
			t.Fatalf("ended listed session was listed: %+v", summaries)
		}
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
	thirdProgression, _ := s.GetOrCreateCharacterProgression(ctx, thirdAcct.ID, thirdChar.ID, defaultProgression)
	hostItem := store.CharacterItemInstance{ID: "2001", AccountID: hostAcct.ID, CharacterID: hostChar.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory}
	guestItem := store.CharacterItemInstance{ID: "2001", AccountID: guestAcct.ID, CharacterID: guestChar.ID, ItemDefID: "cave_bow", Location: store.ItemLocationInventory}
	thirdItem := store.CharacterItemInstance{ID: "2001", AccountID: thirdAcct.ID, CharacterID: thirdChar.ID, ItemDefID: "cave_helm", Location: store.ItemLocationInventory}
	hostHotbar := []store.CharacterHotbarSlot{{AccountID: hostAcct.ID, CharacterID: hostChar.ID, SlotIndex: 0, ItemInstanceID: &hostItem.ID}}
	guestHotbar := []store.CharacterHotbarSlot{{AccountID: guestAcct.ID, CharacterID: guestChar.ID, SlotIndex: 0, ItemInstanceID: &guestItem.ID}}
	thirdHotbar := []store.CharacterHotbarSlot{{AccountID: thirdAcct.ID, CharacterID: thirdChar.ID, SlotIndex: 0, ItemInstanceID: &thirdItem.ID}}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, hostAcct.ID, hostChar.ID, []store.CharacterItemInstance{hostItem}, nil, hostHotbar, store.CharacterSkillBindings{}, nil, nil, store.AccountStashGold{AccountID: hostAcct.ID}, nil, hostProgression); err != nil {
		t.Fatalf("host start snapshot: %v", err)
	}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, guestAcct.ID, guestChar.ID, []store.CharacterItemInstance{guestItem}, nil, guestHotbar, store.CharacterSkillBindings{}, nil, nil, store.AccountStashGold{AccountID: guestAcct.ID}, nil, guestProgression); err != nil {
		t.Fatalf("guest start snapshot: %v", err)
	}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, thirdAcct.ID, thirdChar.ID, []store.CharacterItemInstance{thirdItem}, nil, thirdHotbar, store.CharacterSkillBindings{}, nil, nil, store.AccountStashGold{AccountID: thirdAcct.ID}, nil, thirdProgression); err != nil {
		t.Fatalf("third start snapshot: %v", err)
	}
	snaps, err := s.LoadSessionStartSnapshots(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load start snapshots: %v", err)
	}
	if len(snaps) != 3 || len(snaps[0].Items) != 1 || len(snaps[1].Items) != 1 || len(snaps[2].Items) != 1 {
		t.Fatalf("snapshot count mismatch: %+v", snaps)
	}
	if snaps[0].Items[0].ItemDefID != "cave_blade" || snaps[1].Items[0].ItemDefID != "cave_bow" || snaps[2].Items[0].ItemDefID != "cave_helm" {
		t.Fatalf("member snapshots collided: %+v", snaps)
	}
}

func TestSessionMemberConnectedPreservesTickZeroJoinOnReconnect(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	hostAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "host-zero+"+ids.Token()[:12]+"@example.test")
	hostChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), hostAcct.ID, "Host")
	guestAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "guest-zero+"+ids.Token()[:12]+"@example.test")
	guestChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), guestAcct.ID, "Guest")
	sess := store.Session{
		ID:           ids.New("sess"),
		AccountID:    hostAcct.ID,
		CharacterID:  hostChar.ID,
		Seed:         "joined-zero",
		WorldID:      "dungeon_levels",
		Mode:         store.SessionModeCoop,
		JoinCodeHash: "join_hash",
		Status:       store.SessionActive,
	}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.CreateSessionHostMember(ctx, store.SessionMember{
		SessionID: sess.ID, AccountID: hostAcct.ID, CharacterID: hostChar.ID, Role: store.SessionMemberHost,
	}); err != nil {
		t.Fatalf("create host member: %v", err)
	}
	if err := s.CreateSessionGuestMember(ctx, store.SessionMember{
		SessionID: sess.ID, AccountID: guestAcct.ID, CharacterID: guestChar.ID, Role: store.SessionMemberGuest, JoinedTick: -1,
	}); err != nil {
		t.Fatalf("create guest member: %v", err)
	}
	if err := s.SetSessionMemberConnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1005", 0, 0); err != nil {
		t.Fatalf("first connect: %v", err)
	}
	member, err := s.GetSessionMember(ctx, sess.ID, guestAcct.ID, guestChar.ID)
	if err != nil {
		t.Fatalf("get first connected member: %v", err)
	}
	if member.JoinedTick != 0 {
		t.Fatalf("joined_tick after tick-zero connect = %d, want 0", member.JoinedTick)
	}
	if err := s.SetSessionMemberDisconnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, 0, 10); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if claimed, err := s.ClaimSessionMemberConnection(ctx, sess.ID, guestAcct.ID, guestChar.ID); err != nil || !claimed {
		t.Fatalf("claim reconnect = %v, %v; want true, nil", claimed, err)
	}
	if err := s.SetSessionMemberConnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1005", 0, 12); err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	member, err = s.GetSessionMember(ctx, sess.ID, guestAcct.ID, guestChar.ID)
	if err != nil {
		t.Fatalf("get reconnected member: %v", err)
	}
	if member.JoinedTick != 0 {
		t.Fatalf("joined_tick after reconnect = %d, want original 0", member.JoinedTick)
	}
}

func summaryByID(summaries []store.SessionSummary, sessionID string) store.SessionSummary {
	for _, summary := range summaries {
		if summary.SessionID == sessionID {
			return summary
		}
	}
	return store.SessionSummary{}
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
	if progression.Level != 1 || progression.Experience != 0 || progression.UnspentStatPoints != 0 || progression.UnspentSkillPoints != 0 ||
		progression.DeepestDungeonDepth != 0 ||
		progression.Stats.Str != 5 || progression.Stats.Dex != 5 || progression.Stats.Vit != 5 || progression.Stats.Magic != 5 ||
		progression.SkillRanks["magic_bolt"] != 0 {
		t.Fatalf("default progression mismatch: %+v", progression)
	}
	progression.Level = 2
	progression.Experience = 25
	progression.UnspentStatPoints = 5
	progression.UnspentSkillPoints = 1
	progression.Stats.Vit = 6
	progression.DeepestDungeonDepth = 2
	progression.SkillRanks["magic_bolt"] = 1
	if err := s.UpsertCharacterProgression(ctx, acct.ID, progression); err != nil {
		t.Fatalf("upsert progression: %v", err)
	}
	loadedProgression, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, char.ID, store.CharacterProgressionDefaults{
		Level:              9,
		Experience:         999,
		UnspentStatPoints:  99,
		UnspentSkillPoints: 99,
		Stats:              store.CharacterBaseStats{Str: 1, Dex: 1, Vit: 1, Magic: 1},
		SkillRanks:         map[string]int{"magic_bolt": 5},
	})
	if err != nil {
		t.Fatalf("reload progression: %v", err)
	}
	if loadedProgression.Level != 2 || loadedProgression.Experience != 25 || loadedProgression.UnspentStatPoints != 5 ||
		loadedProgression.UnspentSkillPoints != 1 || loadedProgression.SkillRanks["magic_bolt"] != 1 ||
		loadedProgression.DeepestDungeonDepth != 2 || loadedProgression.Stats.Vit != 6 {
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

	if err := s.SetCharacterItemEquipped(ctx, acct.ID, char.ID, item.ID, "main_hand", true, 1); err != nil {
		t.Fatalf("set equipped: %v", err)
	}
	insertedWaypoint, err := s.AddAccountWaypoint(ctx, acct.ID, -1)
	if err != nil {
		t.Fatalf("add waypoint: %v", err)
	}
	if !insertedWaypoint {
		t.Fatalf("first waypoint insert returned false")
	}
	insertedWaypoint, err = s.AddAccountWaypoint(ctx, acct.ID, -1)
	if err != nil {
		t.Fatalf("re-add waypoint: %v", err)
	}
	if insertedWaypoint {
		t.Fatalf("duplicate waypoint insert returned true")
	}

	items, err := s.ListCharacterItems(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list character items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("character item count = %d, want 1", len(items))
	}
	if !items[0].Equipped || items[0].Location != store.ItemLocationEquipped || items[0].Slot != "main_hand" || items[0].WeaponSet != 1 || items[0].ItemDefID != "cave_blade" {
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

	waypoints, err := s.ListAccountWaypoints(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list waypoints: %v", err)
	}
	if len(waypoints) != 1 || waypoints[0].Level != -1 {
		t.Fatalf("waypoints = %+v, want level -1", waypoints)
	}
	altChar, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Secondary", "sorcerer")
	if err != nil {
		t.Fatalf("create secondary character: %v", err)
	}
	altWaypoints, err := s.ListAccountWaypoints(ctx, acct.ID, altChar.ID)
	if err != nil {
		t.Fatalf("list secondary waypoints: %v", err)
	}
	if len(altWaypoints) != 1 || altWaypoints[0].CharacterID != altChar.ID || altWaypoints[0].Level != -1 {
		t.Fatalf("secondary account waypoints = %+v, want level -1 for %s", altWaypoints, altChar.ID)
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
	skillBinds, err := s.GetOrCreateCharacterSkillBindings(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("get skill bindings: %v", err)
	}
	skillBinds.FunctionKeys[0] = "magic_bolt"
	skillBinds.FunctionKeys[1] = "heal"
	skillBinds.FunctionKeys[8] = "cleave"
	skillBinds.RightClickSkillID = "heal"
	if err := s.SetCharacterSkillBindings(ctx, skillBinds); err != nil {
		t.Fatalf("set skill bindings: %v", err)
	}
	skillBinds, err = s.GetOrCreateCharacterSkillBindings(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("reload skill bindings: %v", err)
	}
	if len(skillBinds.FunctionKeys) != 16 || skillBinds.FunctionKeys[0] != "magic_bolt" || skillBinds.FunctionKeys[1] != "heal" || skillBinds.FunctionKeys[8] != "cleave" || skillBinds.RightClickSkillID != "heal" {
		t.Fatalf("skill bindings mismatch: %+v", skillBinds)
	}

	initialStock := []store.CharacterShopStockItem{
		{
			AccountID:      acct.ID,
			CharacterID:    char.ID,
			ShopID:         "town_vendor",
			RefreshKey:     "wp:-1",
			OfferIndex:     0,
			OfferID:        "generated:depth1:000",
			SourceDepth:    1,
			ItemTemplateID: "cave_blade",
			RolledPayload:  json.RawMessage(`{"item_template_id":"cave_blade","display_name":"Common Cave Blade","rarity":"common","stats":{"damage_min":2,"damage_max":4},"requirements":{"level":1},"effect_ids":[]}`),
			BuyPrice:       100,
			Available:      true,
		},
		{
			AccountID:      acct.ID,
			CharacterID:    char.ID,
			ShopID:         "town_vendor",
			RefreshKey:     "wp:-1",
			OfferIndex:     1,
			OfferID:        "generated:depth2:001",
			SourceDepth:    2,
			ItemTemplateID: "cave_bow",
			RolledPayload:  json.RawMessage(`{"item_template_id":"cave_bow","display_name":"Rare Cave Bow","rarity":"rare","stats":{"damage_min":2,"damage_max":3},"requirements":{"level":1},"effect_ids":[]}`),
			BuyPrice:       210,
			Available:      true,
		},
	}
	if err := s.ReplaceCharacterShopStock(ctx, acct.ID, char.ID, "town_vendor", "wp:-1", initialStock); err != nil {
		t.Fatalf("replace shop stock: %v", err)
	}
	shopStock, err := s.ListCharacterShopStock(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("list shop stock: %v", err)
	}
	if len(shopStock) != 2 || shopStock[0].OfferID != "generated:depth1:000" || shopStock[1].SourceDepth != 2 {
		t.Fatalf("shop stock mismatch: %+v", shopStock)
	}
	if err := s.SetCharacterShopStockAvailable(ctx, acct.ID, char.ID, "town_vendor", "generated:depth1:000", false); err != nil {
		t.Fatalf("consume shop stock: %v", err)
	}
	shopStock, err = s.ListCharacterShopStock(ctx, acct.ID, char.ID)
	if err != nil {
		t.Fatalf("reload shop stock: %v", err)
	}
	if len(shopStock) != 2 || shopStock[0].Available {
		t.Fatalf("consumed shop stock did not persist: %+v", shopStock)
	}
	if err := s.SetCharacterShopStockAvailable(ctx, acct.ID, char.ID, "town_vendor", "missing", true); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("restore missing stock: expected ErrNotFound, got %v", err)
	}

	resources := []store.AccountResourceAmount{{AccountID: acct.ID, ResourceID: "upgrade_shard", Amount: 2}}
	if err := s.CreateSessionStartSnapshot(ctx, sess.ID, acct.ID, char.ID, items, waypoints, hotbar, skillBinds, shopStock, nil, store.AccountStashGold{AccountID: acct.ID}, resources, loadedProgression); err != nil {
		t.Fatalf("create session snapshot: %v", err)
	}
	if err := s.SetCharacterItemEquipped(ctx, acct.ID, char.ID, item.ID, "", false, 0); err != nil {
		t.Fatalf("mutate live item: %v", err)
	}
	if _, err := s.AddAccountWaypoint(ctx, acct.ID, -2); err != nil {
		t.Fatalf("mutate live waypoints: %v", err)
	}
	mutatedProgression := loadedProgression
	mutatedProgression.Level = 3
	mutatedProgression.Experience = 70
	mutatedProgression.UnspentStatPoints = 10
	mutatedProgression.UnspentSkillPoints = 0
	mutatedProgression.Stats.Str = 7
	mutatedProgression.DeepestDungeonDepth = 5
	mutatedProgression.SkillRanks["magic_bolt"] = 2
	if err := s.UpsertCharacterProgression(ctx, acct.ID, mutatedProgression); err != nil {
		t.Fatalf("mutate live progression: %v", err)
	}
	if err := s.SetCharacterHotbarSlot(ctx, acct.ID, char.ID, 2, nil); err != nil {
		t.Fatalf("mutate live hotbar: %v", err)
	}
	skillBinds.FunctionKeys[0] = "rage"
	skillBinds.RightClickSkillID = "rage"
	if err := s.SetCharacterSkillBindings(ctx, skillBinds); err != nil {
		t.Fatalf("mutate live skill bindings: %v", err)
	}
	replacementStock := []store.CharacterShopStockItem{{
		AccountID:      acct.ID,
		CharacterID:    char.ID,
		ShopID:         "town_vendor",
		RefreshKey:     "wp:-2",
		OfferIndex:     0,
		OfferID:        "generated:depth3:000",
		SourceDepth:    3,
		ItemTemplateID: "cave_helm",
		RolledPayload:  json.RawMessage(`{"item_template_id":"cave_helm","display_name":"Magic Cave Helm","rarity":"magic","stats":{"armor":5},"requirements":{"level":1},"effect_ids":[]}`),
		BuyPrice:       85,
		Available:      true,
	}}
	if err := s.ReplaceCharacterShopStock(ctx, acct.ID, char.ID, "town_vendor", "wp:-2", replacementStock); err != nil {
		t.Fatalf("mutate live shop stock: %v", err)
	}
	if _, err := s.AddAccountResource(ctx, acct.ID, "upgrade_shard", 3); err != nil {
		t.Fatalf("mutate live resource wallet: %v", err)
	}
	snap, err := s.LoadSessionStartSnapshot(ctx, sess.ID)
	if err != nil {
		t.Fatalf("load session snapshot: %v", err)
	}
	if len(snap.Items) != 1 || !snap.Items[0].Equipped || snap.Items[0].Slot != "main_hand" || snap.Items[0].WeaponSet != 1 {
		t.Fatalf("snapshot item mutated with live state: %+v", snap.Items)
	}
	if len(snap.Hotbar) != 10 || snap.Hotbar[2].ItemInstanceID == nil || *snap.Hotbar[2].ItemInstanceID != item.ID {
		t.Fatalf("snapshot hotbar mutated with live state: %+v", snap.Hotbar)
	}
	if len(snap.SkillBinds.FunctionKeys) != 16 || snap.SkillBinds.FunctionKeys[0] != "magic_bolt" || snap.SkillBinds.FunctionKeys[1] != "heal" || snap.SkillBinds.FunctionKeys[8] != "cleave" || snap.SkillBinds.RightClickSkillID != "heal" {
		t.Fatalf("snapshot skill bindings mutated with live state: %+v", snap.SkillBinds)
	}
	if len(snap.Waypoints) != 1 || snap.Waypoints[0].Level != -1 {
		t.Fatalf("snapshot waypoints mutated with live state: %+v", snap.Waypoints)
	}
	if len(snap.ShopStock) != 2 || snap.ShopStock[0].OfferID != "generated:depth1:000" || snap.ShopStock[0].Available || snap.ShopStock[1].OfferID != "generated:depth2:001" {
		t.Fatalf("snapshot shop stock mutated with live state: %+v", snap.ShopStock)
	}
	if len(snap.Resources) != 1 || snap.Resources[0].ResourceID != "upgrade_shard" || snap.Resources[0].Amount != 2 {
		t.Fatalf("snapshot resources mutated with live state: %+v", snap.Resources)
	}
	if snap.Progression == nil {
		t.Fatalf("snapshot progression missing")
	}
	if snap.Progression.Level != 2 || snap.Progression.Experience != 25 || snap.Progression.UnspentStatPoints != 5 ||
		snap.Progression.UnspentSkillPoints != 1 || snap.Progression.SkillRanks["magic_bolt"] != 1 ||
		snap.Progression.DeepestDungeonDepth != 2 || snap.Progression.Stats.Str != 5 || snap.Progression.Stats.Vit != 6 {
		t.Fatalf("snapshot progression mutated with live state: %+v", snap.Progression)
	}

	otherAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "other+"+ids.Token()[:12]+"@example.test")
	otherChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), otherAcct.ID, "Hero")
	if err := s.SetCharacterItemEquipped(ctx, otherAcct.ID, otherChar.ID, item.ID, "main_hand", true, 0); !errors.Is(err, store.ErrNotFound) {
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

func TestAccountStashTransfersAreAccountScopedAndAtomic(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "stash+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	charA, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Stasher", "barbarian")
	if err != nil {
		t.Fatalf("create character A: %v", err)
	}
	charB, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Alt", "barbarian")
	if err != nil {
		t.Fatalf("create character B: %v", err)
	}
	otherAcct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "stash-other+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert other account: %v", err)
	}
	otherChar, err := s.CreateCharacter(ctx, ids.New("char"), otherAcct.ID, "Other", "barbarian")
	if err != nil {
		t.Fatalf("create other character: %v", err)
	}

	defaults := store.CharacterProgressionDefaults{
		Level:              1,
		Experience:         0,
		UnspentStatPoints:  0,
		UnspentSkillPoints: 0,
		Stats:              store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
		Gold:               10,
		SkillRanks:         map[string]int{"magic_bolt": 0},
	}
	if _, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, charA.ID, defaults); err != nil {
		t.Fatalf("create progression A: %v", err)
	}
	otherDefaults := defaults
	otherDefaults.Gold = 5
	if _, err := s.GetOrCreateCharacterProgression(ctx, acct.ID, charB.ID, store.CharacterProgressionDefaults{Level: 1, Stats: defaults.Stats, Gold: 0, SkillRanks: defaults.SkillRanks}); err != nil {
		t.Fatalf("create progression B: %v", err)
	}
	if _, err := s.GetOrCreateCharacterProgression(ctx, otherAcct.ID, otherChar.ID, otherDefaults); err != nil {
		t.Fatalf("create progression other: %v", err)
	}

	itemA := store.CharacterItemInstance{
		ID:          "item_a",
		AccountID:   acct.ID,
		CharacterID: charA.ID,
		ItemDefID:   "cave_blade",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"item_template_id":"cave_blade","display_name":"Rare Cave Blade","rarity":"rare","stats":{"damage_min":4,"damage_max":6},"requirements":{"level":1},"effect_ids":[]}`),
	}
	if err := s.AddCharacterItem(ctx, itemA); err != nil {
		t.Fatalf("add character item A: %v", err)
	}
	stashed, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, charA.ID, itemA.ID, "stash_1")
	if err != nil {
		t.Fatalf("deposit item: %v", err)
	}
	if stashed.AccountID != acct.ID || stashed.StashItemID != "stash_1" || stashed.SourceCharacterID != charA.ID || stashed.ItemDefID != itemA.ItemDefID {
		t.Fatalf("stashed item mismatch: %+v", stashed)
	}
	charAItems, err := s.ListCharacterItems(ctx, acct.ID, charA.ID)
	if err != nil {
		t.Fatalf("list char A items: %v", err)
	}
	if len(charAItems) != 0 {
		t.Fatalf("deposited item remained on character A: %+v", charAItems)
	}
	stashItems, err := s.ListAccountStashItems(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list account stash: %v", err)
	}
	if len(stashItems) != 1 || stashItems[0].StashItemID != "stash_1" {
		t.Fatalf("account stash items = %+v", stashItems)
	}
	otherStashItems, err := s.ListAccountStashItems(ctx, otherAcct.ID)
	if err != nil {
		t.Fatalf("list other stash: %v", err)
	}
	if len(otherStashItems) != 0 {
		t.Fatalf("foreign account saw stash rows: %+v", otherStashItems)
	}
	if _, err := s.TransferAccountStashItemToCharacter(ctx, otherAcct.ID, otherChar.ID, "stash_1", "foreign_item"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign withdraw: expected ErrNotFound, got %v", err)
	}

	withdrawn, err := s.TransferAccountStashItemToCharacter(ctx, acct.ID, charB.ID, "stash_1", "item_b")
	if err != nil {
		t.Fatalf("withdraw item to char B: %v", err)
	}
	if withdrawn.AccountID != acct.ID || withdrawn.CharacterID != charB.ID || withdrawn.ID != "item_b" || withdrawn.ItemDefID != itemA.ItemDefID {
		t.Fatalf("withdrawn item mismatch: %+v", withdrawn)
	}
	stashItems, err = s.ListAccountStashItems(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list after withdraw: %v", err)
	}
	if len(stashItems) != 0 {
		t.Fatalf("stash item remained after withdraw: %+v", stashItems)
	}

	conflictA := store.CharacterItemInstance{
		ID:          "item_conflict_a",
		AccountID:   acct.ID,
		CharacterID: charA.ID,
		ItemDefID:   "red_potion",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{}`),
	}
	conflictB := conflictA
	conflictB.ID = "item_conflict_b"
	if err := s.AddCharacterItem(ctx, conflictA); err != nil {
		t.Fatalf("add conflict A: %v", err)
	}
	if err := s.AddCharacterItem(ctx, conflictB); err != nil {
		t.Fatalf("add conflict B: %v", err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, charA.ID, conflictA.ID, "stash_conflict"); err != nil {
		t.Fatalf("deposit conflict A: %v", err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, charA.ID, conflictB.ID, "stash_conflict"); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("deposit conflict B: expected ErrConflict, got %v", err)
	}
	charAItems, err = s.ListCharacterItems(ctx, acct.ID, charA.ID)
	if err != nil {
		t.Fatalf("list char A after conflict: %v", err)
	}
	if len(charAItems) != 1 || charAItems[0].ID != conflictB.ID {
		t.Fatalf("conflict deposit did not roll back character item: %+v", charAItems)
	}

	if gold, err := s.GetOrCreateAccountStashGold(ctx, acct.ID); err != nil || gold.Gold != 0 {
		t.Fatalf("initial stash gold = %+v err=%v", gold, err)
	}
	charGold, stashGold, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, charA.ID, 4)
	if err != nil {
		t.Fatalf("deposit gold: %v", err)
	}
	if charGold != 6 || stashGold != 4 {
		t.Fatalf("deposit gold balances character=%d stash=%d", charGold, stashGold)
	}
	charGold, stashGold, err = s.TransferAccountStashGoldToCharacter(ctx, acct.ID, charB.ID, 2)
	if err != nil {
		t.Fatalf("withdraw gold: %v", err)
	}
	if charGold != 2 || stashGold != 2 {
		t.Fatalf("withdraw gold balances character=%d stash=%d", charGold, stashGold)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, charA.ID, 0); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("zero gold transfer: expected ErrConflict, got %v", err)
	}
	if _, _, err := s.TransferAccountStashGoldToCharacter(ctx, otherAcct.ID, otherChar.ID, 1); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("foreign empty stash withdraw: expected ErrConflict, got %v", err)
	}
}

func TestMarketListingMovesStashItemAndCancelReturnsIt(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	acct, err := s.UpsertAccountByEmail(ctx, "acct_market_"+suffix, "market+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_market_"+suffix, acct.ID, "Market Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	item := store.CharacterItemInstance{
		ID:          "market_item_" + suffix,
		AccountID:   acct.ID,
		CharacterID: char.ID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":1}`),
	}
	if err := s.AddCharacterItem(ctx, item); err != nil {
		t.Fatal(err)
	}
	stashID := "stash_market_" + suffix
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, item.ID, stashID); err != nil {
		t.Fatal(err)
	}

	listingID := "listing_market_" + suffix
	listing, err := s.CreateMarketListingFromStash(ctx, acct.ID, stashID, listingID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if listing.Status != store.MarketListingActive || listing.ItemDefID != "rusty_sword" {
		t.Fatalf("listing = %+v", listing)
	}
	stashItems, err := s.ListAccountStashItems(ctx, acct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashItems) != 0 {
		t.Fatalf("stash after listing = %+v, want empty", stashItems)
	}
	active, err := s.ListActiveMarketListings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) == 0 || active[0].ID != listing.ID {
		t.Fatalf("active listings = %+v, want %s", active, listing.ID)
	}

	other, err := s.UpsertAccountByEmail(ctx, "acct_market_other_"+suffix, "market-other+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CancelMarketListing(ctx, other.ID, listing.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign cancel err = %v, want ErrNotFound", err)
	}
	canceled, err := s.CancelMarketListing(ctx, acct.ID, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if canceled.Status != store.MarketListingCanceled || canceled.CanceledAt == nil {
		t.Fatalf("canceled listing = %+v", canceled)
	}
	stashItems, err = s.ListAccountStashItems(ctx, acct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashItems) != 1 || stashItems[0].StashItemID != stashID {
		t.Fatalf("stash after cancel = %+v", stashItems)
	}
}

func TestAccountStashItemUpgradeSpendsGoldAndPersistsStats(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_"+suffix, "upgrade+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_"+suffix, acct.ID, "Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	prog := store.CharacterProgression{AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1, Gold: 300, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, prog); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "upgrade_item_" + suffix, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, "upgrade_item_"+suffix, "upgrade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 250); err != nil {
		t.Fatal(err)
	}
	item, gold, cost, success, err := s.UpgradeAccountStashItem(ctx, acct.ID, "upgrade_stash_"+suffix, 100, 50, 2, 100, 1, 0, map[string]struct{}{"cave_blade": {}})
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("first upgrade should succeed")
	}
	if cost != 100 || gold != 150 {
		t.Fatalf("first upgrade cost/gold = %d/%d, want 100/150", cost, gold)
	}
	var stats struct {
		ItemLevel int `json:"item_level"`
		DamageMax int `json:"damage_max"`
		DamageMin int `json:"damage_min"`
	}
	if err := json.Unmarshal(item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ItemLevel != 1 || stats.DamageMax != 5 || stats.DamageMin != 2 {
		t.Fatalf("upgraded stats = %+v", stats)
	}
	item, gold, cost, success, err = s.UpgradeAccountStashItem(ctx, acct.ID, "upgrade_stash_"+suffix, 100, 50, 2, 100, 1, 0, map[string]struct{}{"cave_blade": {}})
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("second upgrade should succeed")
	}
	if cost != 150 || gold != 0 {
		t.Fatalf("second upgrade cost/gold = %d/%d, want 150/0", cost, gold)
	}
	if err := json.Unmarshal(item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ItemLevel != 2 || stats.DamageMax != 6 || stats.DamageMin != 2 {
		t.Fatalf("second upgraded stats = %+v", stats)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 25); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := s.UpgradeAccountStashItem(ctx, acct.ID, "upgrade_stash_"+suffix, 1, 1, 2, 100, 1, 0, map[string]struct{}{"cave_blade": {}}); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("max level upgrade err = %v, want ErrConflict", err)
	}
}

func TestAccountStashItemUpgradeRejectsInsufficientGold(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_poor_"+suffix, "upgrade-poor+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_poor_"+suffix, acct.ID, "Poor Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "poor_upgrade_item_" + suffix, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, "poor_upgrade_item_"+suffix, "poor_upgrade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := s.UpgradeAccountStashItem(ctx, acct.ID, "poor_upgrade_stash_"+suffix, 100, 50, 2, 100, 1, 0, map[string]struct{}{"cave_blade": {}}); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("insufficient gold upgrade err = %v, want ErrConflict", err)
	}
}

func TestAccountStashItemUpgradeHandlesRolledPayloadStats(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, "acct_upgrade_payload_"+suffix, "upgrade-payload+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	char, err := s.CreateCharacter(ctx, "char_upgrade_payload_"+suffix, acct.ID, "Payload Upgrade Hero", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	prog := store.CharacterProgression{AccountID: acct.ID, CharacterID: char.ID, CharacterClass: "barbarian", Level: 1, Gold: 150, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := s.UpsertCharacterProgression(ctx, acct.ID, prog); err != nil {
		t.Fatal(err)
	}
	payload := `{"item_template_id":"cave_blade","display_name":"Rare Cave Blade","rarity":"rare","stats":{"damage_min":4,"damage_max":5},"requirements":{"level":1,"str":5},"effect_ids":[]}`
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "payload_upgrade_item_" + suffix, AccountID: acct.ID, CharacterID: char.ID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(payload)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, acct.ID, char.ID, "payload_upgrade_item_"+suffix, "payload_upgrade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, acct.ID, char.ID, 100); err != nil {
		t.Fatal(err)
	}
	item, gold, cost, success, err := s.UpgradeAccountStashItem(ctx, acct.ID, "payload_upgrade_stash_"+suffix, 100, 50, 2, 100, 1, 0, map[string]struct{}{"cave_blade": {}})
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("payload upgrade should succeed")
	}
	if cost != 100 || gold != 0 {
		t.Fatalf("payload upgrade cost/gold = %d/%d, want 100/0", cost, gold)
	}
	var upgraded struct {
		ItemTemplateID string         `json:"item_template_id"`
		DisplayName    string         `json:"display_name"`
		Rarity         string         `json:"rarity"`
		Stats          map[string]int `json:"stats"`
		Requirements   map[string]int `json:"requirements"`
	}
	if err := json.Unmarshal(item.RolledStats, &upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.ItemTemplateID != "cave_blade" || upgraded.DisplayName != "Rare Cave Blade" || upgraded.Rarity != "rare" {
		t.Fatalf("payload metadata after upgrade = %+v raw=%s", upgraded, string(item.RolledStats))
	}
	if upgraded.Stats["item_level"] != 1 || upgraded.Stats["damage_max"] != 6 || upgraded.Stats["damage_min"] != 4 {
		t.Fatalf("payload upgraded stats = %+v raw=%s", upgraded.Stats, string(item.RolledStats))
	}
	if upgraded.Requirements["str"] != 5 {
		t.Fatalf("payload requirements after upgrade = %+v raw=%s", upgraded.Requirements, string(item.RolledStats))
	}
}

func TestMarketOfferAcceptMovesItemsAndRefundsCompetingOffers(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_seller_"+suffix, "market-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_seller_"+suffix, seller.ID, "Market Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_bidder_"+suffix, "market-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_bidder_"+suffix, bidder.ID, "Market Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	otherBidder, err := s.UpsertAccountByEmail(ctx, "acct_market_other_bidder_"+suffix, "market-other-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	otherChar, err := s.CreateCharacter(ctx, "char_market_other_bidder_"+suffix, otherBidder.ID, "Market Other Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":3}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "seller_item_"+suffix, "seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	for _, itemID := range []string{"bidder_item_a_" + suffix, "bidder_item_b_" + suffix} {
		if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, itemID, "stash_"+itemID); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "other_item_" + suffix, AccountID: otherBidder.ID, CharacterID: otherChar.ID, ItemDefID: "blue_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, otherBidder.ID, otherChar.ID, "other_item_"+suffix, "stash_other_"+suffix); err != nil {
		t.Fatal(err)
	}

	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "seller_stash_"+suffix, "listing_offer_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateMarketOffer(ctx, seller.ID, listing.ID, "self_offer_"+suffix, []string{"seller_stash_" + suffix}); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("self offer err = %v, want ErrConflict", err)
	}
	offer, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "offer_accept_"+suffix, []string{"stash_bidder_item_a_" + suffix, "stash_bidder_item_b_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	if len(offer.Items) != 2 || offer.Status != store.MarketOfferActive {
		t.Fatalf("offer = %+v", offer)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "bidder_item_c_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "blue_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "bidder_item_c_"+suffix, "stash_bidder_item_c_"+suffix); err != nil {
		t.Fatal(err)
	}
	secondBidderOffer, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "offer_second_"+suffix, []string{"stash_bidder_item_c_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	if secondBidderOffer.Status != store.MarketOfferActive || len(secondBidderOffer.Items) != 1 {
		t.Fatalf("second bidder offer = %+v", secondBidderOffer)
	}
	competing, err := s.CreateMarketOffer(ctx, otherBidder.ID, listing.ID, "offer_refund_"+suffix, []string{"stash_other_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListMarketOffersForSeller(ctx, bidder.ID, listing.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign list offers err = %v, want ErrNotFound", err)
	}
	offers, err := s.ListMarketOffersForSeller(ctx, seller.ID, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 3 {
		t.Fatalf("seller offers = %+v, want 3", offers)
	}
	summary, err := s.GetMarketSummary(ctx, seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.PublishedListings != 1 || summary.IncomingBids != 3 {
		t.Fatalf("seller market summary = %+v, want 1 listing and 3 bids", summary)
	}
	accepted, err := s.AcceptMarketOffer(ctx, seller.ID, listing.ID, offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if accepted.Status != store.MarketOfferAccepted || accepted.AcceptedAt == nil || len(accepted.Items) != 2 {
		t.Fatalf("accepted offer = %+v", accepted)
	}
	sellerStash, err := s.ListAccountStashItems(ctx, seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !accountStashItemsContainStore(sellerStash, "stash_bidder_item_a_"+suffix) || !accountStashItemsContainStore(sellerStash, "stash_bidder_item_b_"+suffix) {
		t.Fatalf("seller stash after accept = %+v, want accepted offer items", sellerStash)
	}
	sellerItems, err := s.ListCharacterItems(ctx, seller.ID, sellerChar.ID)
	if err != nil {
		t.Fatal(err)
	}
	if characterItemsContainStore(sellerItems, "stash_bidder_item_a_"+suffix) || characterItemsContainStore(sellerItems, "stash_bidder_item_b_"+suffix) {
		t.Fatalf("seller character after accept = %+v, want accepted offer items in account stash", sellerItems)
	}
	bidderStash, err := s.ListAccountStashItems(ctx, bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	bidderItems, err := s.ListCharacterItems(ctx, bidder.ID, bidderChar.ID)
	if err != nil {
		t.Fatal(err)
	}
	if characterItemsContainStore(bidderItems, listing.StashItemID) {
		t.Fatalf("bidder character after accept = %+v, want listed item in account stash", bidderItems)
	}
	if characterItemsContainStore(bidderItems, "stash_bidder_item_a_"+suffix) || characterItemsContainStore(bidderItems, "stash_bidder_item_b_"+suffix) {
		t.Fatalf("bidder kept accepted offered item after accept: %+v", bidderItems)
	}
	if !accountStashItemsContainStore(bidderStash, listing.StashItemID) || !accountStashItemsContainStore(bidderStash, secondBidderOffer.Items[0].StashItemID) {
		t.Fatalf("bidder stash after accept = %+v, want listed item and refunded competing bidder offer", bidderStash)
	}
	otherStash, err := s.ListAccountStashItems(ctx, otherBidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(otherStash) != 1 || otherStash[0].StashItemID != competing.Items[0].StashItemID {
		t.Fatalf("other bidder stash after competing refund = %+v", otherStash)
	}
	audit, err := s.ListMarketAuditRecords(ctx, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit) < 4 || audit[len(audit)-1].Action != "offer_accepted" {
		t.Fatalf("market audit after accept = %+v", audit)
	}
}

func TestMarketListingCancelRefundsActiveOffers(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_cancel_seller_"+suffix, "market-cancel-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_cancel_seller_"+suffix, seller.ID, "Market Cancel Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_cancel_bidder_"+suffix, "market-cancel-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_cancel_bidder_"+suffix, bidder.ID, "Market Cancel Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "cancel_seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "cancel_seller_item_"+suffix, "cancel_seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "cancel_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "cancel_bidder_item_"+suffix, "cancel_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "cancel_seller_stash_"+suffix, "cancel_listing_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "cancel_offer_"+suffix, []string{"cancel_bidder_stash_" + suffix}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CancelMarketListing(ctx, seller.ID, listing.ID); err != nil {
		t.Fatal(err)
	}
	sellerStash, err := s.ListAccountStashItems(ctx, seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellerStash) != 1 || sellerStash[0].StashItemID != listing.StashItemID {
		t.Fatalf("seller stash after cancel = %+v", sellerStash)
	}
	bidderStash, err := s.ListAccountStashItems(ctx, bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "cancel_bidder_stash_"+suffix {
		t.Fatalf("bidder stash after cancel refund = %+v", bidderStash)
	}
}

func TestMarketOfferCancelRefundsBidderItems(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_offer_cancel_seller_"+suffix, "market-offer-cancel-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_offer_cancel_seller_"+suffix, seller.ID, "Market Offer Cancel Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_offer_cancel_bidder_"+suffix, "market-offer-cancel-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_offer_cancel_bidder_"+suffix, bidder.ID, "Market Offer Cancel Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "offer_cancel_seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "offer_cancel_seller_item_"+suffix, "offer_cancel_seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "offer_cancel_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "offer_cancel_bidder_item_"+suffix, "offer_cancel_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "offer_cancel_seller_stash_"+suffix, "offer_cancel_listing_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	offer, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "offer_cancel_offer_"+suffix, []string{"offer_cancel_bidder_stash_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CancelMarketOffer(ctx, seller.ID, listing.ID, offer.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign offer cancel err = %v, want ErrNotFound", err)
	}
	canceled, err := s.CancelMarketOffer(ctx, bidder.ID, listing.ID, offer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if canceled.Status != store.MarketOfferCanceled || canceled.CanceledAt == nil {
		t.Fatalf("canceled offer = %+v", canceled)
	}
	bidderStash, err := s.ListAccountStashItems(ctx, bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "offer_cancel_bidder_stash_"+suffix {
		t.Fatalf("bidder stash after offer cancel = %+v", bidderStash)
	}
}

func characterItemsContainStore(items []store.CharacterItemInstance, itemID string) bool {
	for _, item := range items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func accountStashItemsContainStore(items []store.AccountStashItem, stashItemID string) bool {
	for _, item := range items {
		if item.StashItemID == stashItemID {
			return true
		}
	}
	return false
}

func TestMarketListingExpirationRefundsListingAndOffers(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_expire_seller_"+suffix, "market-expire-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_expire_seller_"+suffix, seller.ID, "Market Expire Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_expire_bidder_"+suffix, "market-expire-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_expire_bidder_"+suffix, bidder.ID, "Market Expire Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "expire_seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "expire_seller_item_"+suffix, "expire_seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "expire_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "expire_bidder_item_"+suffix, "expire_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "expire_seller_stash_"+suffix, "expire_listing_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "expire_offer_"+suffix, []string{"expire_bidder_stash_" + suffix}); err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(ctx, testDatabaseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, `UPDATE market_listings SET expires_at = now() - INTERVAL '1 second' WHERE id = $1`, listing.ID); err != nil {
		t.Fatal(err)
	}
	expired, err := s.ExpireMarketListings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if expired != 1 {
		t.Fatalf("expired count = %d, want 1", expired)
	}
	sellerStash, err := s.ListAccountStashItems(ctx, seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellerStash) != 1 || sellerStash[0].StashItemID != listing.StashItemID {
		t.Fatalf("seller stash after expiration = %+v", sellerStash)
	}
	bidderStash, err := s.ListAccountStashItems(ctx, bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "expire_bidder_stash_"+suffix {
		t.Fatalf("bidder stash after expiration refund = %+v", bidderStash)
	}
	active, err := s.ListActiveMarketListings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range active {
		if rec.ID == listing.ID {
			t.Fatalf("expired listing still active: %+v", active)
		}
	}
	audit, err := s.ListMarketAuditRecords(ctx, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit) < 3 || audit[len(audit)-1].Action != "listing_expired" {
		t.Fatalf("market audit after expiration = %+v", audit)
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
