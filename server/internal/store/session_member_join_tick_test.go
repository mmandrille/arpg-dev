package store_test

import (
	"context"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v479: SetSessionMemberPlayer records the tick the member's entity entered the
// sim, so joined_tick < 0 only ever means "never in the sim". A later first
// websocket connect must not move it.
func TestSetSessionMemberPlayerRecordsJoinedTick(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	hostAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "host-join-tick+"+ids.Token()[:12]+"@example.test")
	hostChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), hostAcct.ID, "Host")
	guestAcct, _ := s.UpsertAccountByEmail(ctx, ids.New("acct"), "guest-join-tick+"+ids.Token()[:12]+"@example.test")
	guestChar, _ := s.GetOrCreateDefaultCharacter(ctx, ids.New("char"), guestAcct.ID, "Guest")
	sess := store.Session{
		ID:          ids.New("sess"),
		AccountID:   hostAcct.ID,
		CharacterID: hostChar.ID,
		Seed:        "join-tick",
		WorldID:     "dungeon_levels",
		Mode:        store.SessionModeCoop,
		Status:      store.SessionActive,
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
		SessionID: sess.ID, AccountID: guestAcct.ID, CharacterID: guestChar.ID, Role: store.SessionMemberGuest,
		JoinedTick: store.SessionMemberNotJoinedTick,
	}); err != nil {
		t.Fatalf("create guest member: %v", err)
	}

	const addedAt = 7
	if err := s.SetSessionMemberPlayer(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1009", 0, addedAt); err != nil {
		t.Fatalf("set member player: %v", err)
	}
	member, err := s.GetSessionMember(ctx, sess.ID, guestAcct.ID, guestChar.ID)
	if err != nil {
		t.Fatalf("get member: %v", err)
	}
	if member.JoinedTick != addedAt || member.PlayerEntityID != "1009" {
		t.Fatalf("after SetSessionMemberPlayer: joined_tick=%d entity=%q, want %d and 1009", member.JoinedTick, member.PlayerEntityID, addedAt)
	}

	if err := s.SetSessionMemberConnected(ctx, sess.ID, guestAcct.ID, guestChar.ID, "1009", 0, addedAt+5); err != nil {
		t.Fatalf("connect: %v", err)
	}
	member, err = s.GetSessionMember(ctx, sess.ID, guestAcct.ID, guestChar.ID)
	if err != nil {
		t.Fatalf("get connected member: %v", err)
	}
	if member.JoinedTick != addedAt {
		t.Fatalf("joined_tick after first connect = %d, want the add tick %d", member.JoinedTick, addedAt)
	}
}
