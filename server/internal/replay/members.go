package replay

import (
	"context"
	"fmt"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/sessionsetup"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type memberPlayer struct {
	member   store.SessionMember
	playerID uint64
}

type pendingMember struct {
	sessionsetup.Member
	// recordedJoin members join at their system_member_join row, never
	// through joined_tick.
	recordedJoin bool
}

// memberRoster replays co-op membership over the recorded ticks: members that
// were present when the session was built, late joins, and the recorded
// leave/rejoin rows (v479).
type memberRoster struct {
	sim     *game.Sim
	pending []pendingMember
	joined  []memberPlayer
	// governed holds players with recorded lifecycle rows. Those rows, not the
	// final session_members connectivity, decide whether the player is in play.
	governed map[uint64]bool
}

func newMemberRoster(sim *game.Sim, pending []pendingMember) *memberRoster {
	return &memberRoster{sim: sim, pending: pending, governed: map[uint64]bool{}}
}

// memberKey identifies a member across session_members rows and join inputs.
func memberKey(accountID, characterID string) string {
	return accountID + "\x00" + characterID
}

// recordedJoins returns the members that have a system_member_join row. They
// join at that row, not at joined_tick.
func recordedJoins(inputs []RecordedInput) map[string]bool {
	joins := map[string]bool{}
	for _, rec := range inputs {
		if rec.Input.Type == game.SystemMemberJoinInputType && rec.Input.Member != nil {
			joins[memberKey(rec.Input.Member.AccountID, rec.Input.Member.CharacterID)] = true
		}
	}
	return joins
}

// splitMemberInputs sorts a tick's inputs and separates the membership rows.
// Only actor-less rows count, as for load shed: a row with an actor is dropped.
// Tick checkpoints (v481) are dropped too: they only extend the replayed range.
func splitMemberInputs(inputs []game.Input) ([]game.Input, []game.Input) {
	sortInputs(inputs)
	var players, members []game.Input
	for _, in := range inputs {
		switch {
		case in.Type == game.SystemTickCheckpointInputType: // range marker only
		case !game.IsMemberLifecycleInput(in):
			players = append(players, in)
		case in.ActorPlayerID == 0 && in.Member != nil:
			members = append(members, in)
		}
	}
	return players, members
}

// joinLegacyThrough adds members that have no join row and whose joined_tick
// is at or before tick. They join right before that tick runs.
func (r *memberRoster) joinLegacyThrough(tick int64) error {
	remaining := r.pending[:0]
	for _, item := range r.pending {
		if item.recordedJoin || item.Member.Member.JoinedTick > tick {
			remaining = append(remaining, item)
			continue
		}
		player, err := addMemberToSim(r.sim, item.Member)
		if err != nil {
			return err
		}
		r.joined = append(r.joined, player)
	}
	r.pending = remaining
	return nil
}

// apply replays recorded membership rows in order.
func (r *memberRoster) apply(members []game.Input) error {
	for _, in := range members {
		m := *in.Member
		switch in.Type {
		case game.SystemMemberJoinInputType:
			if err := r.joinRecorded(m); err != nil {
				return err
			}
		case game.SystemMemberLeaveInputType:
			r.sim.ApplyMemberLeave(m.PlayerID)
		case game.SystemMemberRejoinInputType:
			// Live logged and tolerated a respawn error, and replay hits the same one.
			_ = r.sim.ApplyMemberRejoin(m.PlayerID, m.Respawn)
		}
		r.governed[m.PlayerID] = true
	}
	return nil
}

func (r *memberRoster) joinRecorded(m game.MemberLifecycle) error {
	for i, item := range r.pending {
		if item.Member.Member.AccountID != m.AccountID || item.Member.Member.CharacterID != m.CharacterID {
			continue
		}
		player, err := addMemberToSim(r.sim, item.Member)
		if err != nil {
			return err
		}
		if player.playerID != m.PlayerID {
			return fmt.Errorf("replay: join row for %s/%s player_entity_id=%d reconstructed=%d", m.AccountID, m.CharacterID, m.PlayerID, player.playerID)
		}
		r.pending = append(r.pending[:i], r.pending[i+1:]...)
		r.joined = append(r.joined, player)
		return nil
	}
	return fmt.Errorf("replay: join row for unknown member %s/%s", m.AccountID, m.CharacterID)
}

// applyCurrentMemberConnectivity is the legacy fallback for sessions without
// lifecycle rows. It applies the final session_members connectivity once, at
// the end. Players governed by recorded rows are left as the rows put them.
func (r *memberRoster) applyCurrentMemberConnectivity(sess store.Session, players []memberPlayer) {
	coop := sess.Mode == store.SessionModeCoop || sess.JoinCodeHash != ""
	for _, player := range players {
		if r.governed[player.playerID] {
			continue
		}
		if coop && (player.member.Status != store.SessionMemberActive || !player.member.Connected) {
			r.sim.RemovePlayerEntity(player.playerID)
			continue
		}
		r.sim.SetPlayerConnected(player.playerID, true)
	}
}

// sessionStartSim builds the tick-0 sim exactly like the live session build:
// host first, then guests already in the session at tick 0. Guests that joined
// later are returned as pending. joins names the members with a recorded join
// row: they are always pending and join at that row. Members with neither a
// join row nor a non-negative joined_tick never entered the live sim and are
// skipped.
func sessionStartSim(ctx context.Context, repo store.Repository, rules *game.Rules, sess store.Session, joins map[string]bool) (*game.Sim, []memberPlayer, []pendingMember, error) {
	members, err := sessionsetup.Members(ctx, repo, sess)
	if err != nil {
		return nil, nil, nil, err
	}
	hostMember := sessionsetup.Host(members)
	host, err := sessionsetup.Resolve(ctx, repo, rules, sess.ID, hostMember)
	if err != nil {
		return nil, nil, nil, err
	}
	sim, err := sessionsetup.NewHostSim(rules, sess, host, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	hostID := sim.DefaultPlayerID()
	if err := assertStoredPlayerID(hostMember, hostID); err != nil {
		return nil, nil, nil, err
	}
	players := []memberPlayer{{member: hostMember, playerID: hostID}}

	var pending []pendingMember
	for _, member := range members {
		if sessionsetup.IsHost(member, hostMember) {
			continue
		}
		recordedJoin := joins[memberKey(member.AccountID, member.CharacterID)]
		if !recordedJoin && !sessionsetup.JoinedSim(member) {
			continue
		}
		guest, err := sessionsetup.Resolve(ctx, repo, rules, sess.ID, member)
		if err != nil {
			return nil, nil, nil, err
		}
		if recordedJoin || member.JoinedTick > 0 {
			pending = append(pending, pendingMember{Member: guest, recordedJoin: recordedJoin})
			continue
		}
		player, err := addMemberToSim(sim, guest)
		if err != nil {
			return nil, nil, nil, err
		}
		players = append(players, player)
	}
	return sim, players, pending, nil
}

func addMemberToSim(sim *game.Sim, guest sessionsetup.Member) (memberPlayer, error) {
	playerID, err := sessionsetup.AddGuest(sim, guest)
	if err != nil {
		return memberPlayer{}, err
	}
	if err := assertStoredPlayerID(guest.Member, playerID); err != nil {
		return memberPlayer{}, err
	}
	return memberPlayer{member: guest.Member, playerID: playerID}, nil
}

func assertStoredPlayerID(member store.SessionMember, actual uint64) error {
	if member.PlayerEntityID == "" {
		return nil
	}
	want, ok := game.ParseEntityID(member.PlayerEntityID)
	if !ok {
		return fmt.Errorf("replay: invalid member player_entity_id %q for %s/%s", member.PlayerEntityID, member.AccountID, member.CharacterID)
	}
	if want != actual {
		return fmt.Errorf("replay: member %s/%s player_entity_id=%d reconstructed=%d", member.AccountID, member.CharacterID, want, actual)
	}
	return nil
}
