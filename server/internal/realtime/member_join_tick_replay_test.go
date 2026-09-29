package realtime

import (
	"context"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v480 regression: joined_tick = -1 must mean "never in the sim". A guest that
// joins over HTTP while the loop is running, and never attaches over the
// websocket, is never added live. Replay used to treat JoinedTick <= 0 as
// "present at tick 0" and allocated a guest entity, shifting every later
// entity ID and event.
func TestHTTPJoinedGuestNeverAttachedReplayMatchesLive(t *testing.T) {
	ctx := context.Background()
	repo := newMemberSetupRepo(t)
	loop := newMemberSetupLoop(t, repo)
	client := attachMemberSetupClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))

	descendToCorpseLevel(t, loop, client)
	guest := repo.joinGuest()
	settleMemberSetupLoop(t, loop, client)

	if _, inSim := loop.sim.PlayerIDForCharacter(guest.CharacterID); inSim {
		t.Fatal("HTTP-only guest was added to the live sim; fixture no longer exercises a never-attached member")
	}
	if got := repo.memberByRole(store.SessionMemberGuest).JoinedTick; got >= 0 {
		t.Fatalf("never-attached guest joined_tick = %d, want negative", got)
	}
	assertReplayMatchesLive(t, repo, loop)
}

// The live build adds every member row it finds at tick 0, including guests
// that joined over HTTP before the loop started and have not attached yet. It
// must persist joined_tick = 0 for them, or the replay rule above would drop a
// guest the live sim really has.
func TestGuestJoinedBeforeBuildNeverAttachedReplayMatchesLive(t *testing.T) {
	ctx := context.Background()
	repo := newMemberSetupRepo(t)
	guest := repo.joinGuest()
	loop := newMemberSetupLoop(t, repo)
	client := attachMemberSetupClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))

	descendToCorpseLevel(t, loop, client)
	settleMemberSetupLoop(t, loop, client)

	if _, inSim := loop.sim.PlayerIDForCharacter(guest.CharacterID); !inSim {
		t.Fatal("guest present at build was not added to the live sim")
	}
	if got := repo.memberByRole(store.SessionMemberGuest).JoinedTick; got != 0 {
		t.Fatalf("build-time guest joined_tick = %d, want 0", got)
	}
	assertReplayMatchesLive(t, repo, loop)
}
