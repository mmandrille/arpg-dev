package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
)

func staleSessionTestDatabaseURL() string {
	if v := os.Getenv("ARPG_TEST_DATABASE_URL"); v != "" {
		return v
	}
	if v := os.Getenv("ARPG_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"
}

func newStaleSessionStore(t *testing.T) *Store {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := Connect(ctx, staleSessionTestDatabaseURL())
	if err != nil {
		t.Skipf("skipping store integration test: cannot connect to Postgres: %v", err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func createCleanupSession(t *testing.T, ctx context.Context, s *Store, label string) Session {
	t.Helper()
	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "stale-"+label+"+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	char, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Cleanup "+label, "barbarian")
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	sess := Session{
		ID:          ids.New("sess"),
		AccountID:   acct.ID,
		CharacterID: char.ID,
		Seed:        "cleanup-seed-" + label,
		WorldID:     "dungeon_levels",
		Mode:        SessionModeCoop,
		Listed:      true,
		Status:      SessionActive,
	}
	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.CreateSessionHostMember(ctx, SessionMember{
		SessionID:   sess.ID,
		AccountID:   acct.ID,
		CharacterID: char.ID,
	}); err != nil {
		t.Fatalf("create host member: %v", err)
	}
	return sess
}

func TestDeleteStaleEmptySessions(t *testing.T) {
	s := newStaleSessionStore(t)
	ctx := context.Background()

	oldEmpty := createCleanupSession(t, ctx, s, "old-empty")
	recentEmpty := createCleanupSession(t, ctx, s, "recent-empty")
	oldConnected := createCleanupSession(t, ctx, s, "old-connected")
	if err := s.SetSessionMemberConnected(ctx, oldConnected.ID, oldConnected.AccountID, oldConnected.CharacterID, "1001", 0, 0); err != nil {
		t.Fatalf("connect old session: %v", err)
	}

	if err := s.AppendInput(ctx, SessionInput{
		ID:               ids.New("inp"),
		SessionID:        oldEmpty.ID,
		Tick:             1,
		Sequence:         1,
		MessageID:        ids.New("msg"),
		ActorAccountID:   oldEmpty.AccountID,
		ActorCharacterID: oldEmpty.CharacterID,
		Payload:          json.RawMessage(`{"type":"move_intent"}`),
	}); err != nil {
		t.Fatalf("append input: %v", err)
	}
	if err := s.AppendEvent(ctx, SessionEvent{
		ID:        ids.New("ev"),
		SessionID: oldEmpty.ID,
		Tick:      1,
		Sequence:  1,
		EventType: "test_event",
		Payload:   json.RawMessage(`{"ok":true}`),
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	if err := s.CreateSessionStartSnapshot(ctx, oldEmpty.ID, oldEmpty.AccountID, oldEmpty.CharacterID, nil, nil, nil, CharacterSkillBindings{}, nil, nil, AccountStashGold{AccountID: oldEmpty.AccountID}, CharacterProgression{
		AccountID:         oldEmpty.AccountID,
		CharacterID:       oldEmpty.CharacterID,
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		Stats:             CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5},
		SkillRanks:        map[string]int{"magic_bolt": 1},
	}); err != nil {
		t.Fatalf("create start snapshot: %v", err)
	}

	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = now() - interval '13 hours' WHERE id = ANY($1)`,
		[]string{oldEmpty.ID, oldConnected.ID},
	); err != nil {
		t.Fatalf("backdate old sessions: %v", err)
	}
	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = now() - interval '11 hours' WHERE id = $1`,
		recentEmpty.ID,
	); err != nil {
		t.Fatalf("backdate recent session: %v", err)
	}

	deleted, err := s.DeleteStaleEmptySessions(ctx, time.Now().Add(-12*time.Hour))
	if err != nil {
		t.Fatalf("delete stale empty sessions: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("deleted = %d, want at least 1", deleted)
	}
	if _, err := s.GetSession(ctx, oldEmpty.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("old empty session err = %v, want ErrNotFound", err)
	}
	if _, err := s.GetSession(ctx, recentEmpty.ID); err != nil {
		t.Fatalf("recent empty session was deleted: %v", err)
	}
	if _, err := s.GetSession(ctx, oldConnected.ID); err != nil {
		t.Fatalf("old connected session was deleted: %v", err)
	}

	deleted, err = s.DeleteStaleEmptySessions(ctx, time.Now().Add(-12*time.Hour))
	if err != nil {
		t.Fatalf("delete stale empty sessions again: %v", err)
	}
	if deleted != 0 {
		t.Fatalf("second deleted = %d, want 0", deleted)
	}
}

func TestResetConnectedSessionMembersBeforeStartupCleanup(t *testing.T) {
	s := newStaleSessionStore(t)
	ctx := context.Background()

	oldConnected := createCleanupSession(t, ctx, s, "startup-old-connected")
	if err := s.SetSessionMemberConnected(ctx, oldConnected.ID, oldConnected.AccountID, oldConnected.CharacterID, "1001", 0, 0); err != nil {
		t.Fatalf("connect old session: %v", err)
	}
	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = now() - interval '13 hours' WHERE id = $1`,
		oldConnected.ID,
	); err != nil {
		t.Fatalf("backdate old connected session: %v", err)
	}

	reset, err := s.ResetConnectedSessionMembers(ctx)
	if err != nil {
		t.Fatalf("reset connected members: %v", err)
	}
	if reset < 1 {
		t.Fatalf("reset = %d, want at least 1", reset)
	}
	deleted, err := s.DeleteStaleEmptySessions(ctx, time.Now().Add(-12*time.Hour))
	if err != nil {
		t.Fatalf("delete stale empty sessions: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("deleted = %d, want at least 1", deleted)
	}
	if _, err := s.GetSession(ctx, oldConnected.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("old connected session err = %v, want ErrNotFound", err)
	}
}
