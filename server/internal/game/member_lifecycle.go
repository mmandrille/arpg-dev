package game

// Co-op membership changes are server-authored stored inputs (v479). The live
// runner applies them between ticks and records them with no actor, stamped
// with the tick that just finished (CurrentTick()-1, or -1 before the first
// tick), like SystemLoadShedInputType. Replay applies a tick-T row after tick T
// in sequence order, so both paths mutate the sim at the same point. Clients can
// never send them: inputdecode only accepts them from the persisted stream.
const (
	// SystemMemberJoinInputType marks where a late guest was added. Replay owns
	// the add itself, because it needs the member's start snapshot from the store.
	SystemMemberJoinInputType = "system_member_join"
	// SystemMemberLeaveInputType removes a departed co-op player from play.
	SystemMemberLeaveInputType = "system_member_leave"
	// SystemMemberRejoinInputType reconnects a member, optionally respawning it
	// in town.
	SystemMemberRejoinInputType = "system_member_rejoin"
)

// MemberLifecycle is the payload of a membership stored input. AccountID and
// CharacterID identify the member for joins. Respawn is the live admission
// decision for rejoins, which replay applies verbatim instead of re-deriving it.
type MemberLifecycle struct {
	PlayerID    uint64
	AccountID   string
	CharacterID string
	Respawn     bool
}

// IsMemberLifecycleInput reports whether in is a membership stored input.
func IsMemberLifecycleInput(in Input) bool {
	switch in.Type {
	case SystemMemberJoinInputType, SystemMemberLeaveInputType, SystemMemberRejoinInputType:
		return true
	}
	return false
}

// ApplyMemberLeave takes a departed co-op player out of play. Callers must
// record a SystemMemberLeaveInputType input at the current tick.
func (s *Sim) ApplyMemberLeave(playerID uint64) {
	s.RemovePlayerEntity(playerID)
}

// ApplyMemberRejoin reconnects a member, respawning it in town first when
// respawn is set. Callers must record a SystemMemberRejoinInputType input at the
// current tick. A respawn error still leaves the player connected, as admission
// always did. Replay reproduces the same error deterministically.
func (s *Sim) ApplyMemberRejoin(playerID uint64, respawn bool) error {
	var err error
	if respawn {
		err = s.RespawnPlayerInTown(playerID)
	}
	s.SetPlayerConnected(playerID, true)
	return err
}
