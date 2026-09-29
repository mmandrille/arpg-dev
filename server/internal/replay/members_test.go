package replay

import (
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

// Checkpoints only extend the replayed tick range. They must never reach
// TickResults, not even with an actor, and must not count as membership rows.
func TestSplitMemberInputsDropsTickCheckpoints(t *testing.T) {
	leave := &game.MemberLifecycle{PlayerID: 7}
	players, members := splitMemberInputs([]game.Input{
		{Type: "move_intent", ActorPlayerID: 1001, Sequence: 0},
		{Type: game.SystemTickCheckpointInputType, Sequence: 1},
		{Type: game.SystemTickCheckpointInputType, ActorPlayerID: 1001, Sequence: 2},
		{Type: game.SystemMemberLeaveInputType, Member: leave, Sequence: 3},
		{Type: game.SystemMemberLeaveInputType, ActorPlayerID: 1001, Member: leave, Sequence: 4},
	})
	if len(players) != 1 || players[0].Type != "move_intent" {
		t.Fatalf("player inputs = %+v, want only move_intent", players)
	}
	if len(members) != 1 || members[0].Sequence != 3 {
		t.Fatalf("member rows = %+v, want only the actor-less leave", members)
	}
}
