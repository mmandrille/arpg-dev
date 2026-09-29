package realtime

import (
	"context"
	"encoding/json"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// Co-op membership changes mutate the sim between ticks, so each one is
// recorded as a server-authored stored input (v479) stamped with the last tick
// that finished, CurrentTick()-1 (-1 before the first tick). Replay applies it
// after that tick, in sequence order. These helpers run with l.mu held and
// persist immediately. Their callers already write the session_members row
// under the same lock, and these paths are rare, unlike the per-tick load-shed
// row.

// recordMemberJoinLocked marks the point where playerIDForMember added a late
// guest, so replay adds it in the same order relative to other membership
// changes in this inter-tick gap.
func (l *sessionLoop) recordMemberJoinLocked(member store.SessionMember, playerID uint64) {
	l.recordMemberLifecycleLocked(game.SystemMemberJoinInputType, game.MemberLifecycle{
		PlayerID:    playerID,
		AccountID:   member.AccountID,
		CharacterID: member.CharacterID,
	})
}

// applyMemberLeaveLocked removes a departed co-op player and records it.
func (l *sessionLoop) applyMemberLeaveLocked(playerID uint64) {
	if !l.recordMemberLifecycleLocked(game.SystemMemberLeaveInputType, game.MemberLifecycle{PlayerID: playerID}) {
		return
	}
	l.sim.ApplyMemberLeave(playerID)
}

// applyMemberRejoinLocked reconnects a member and records the admission
// decision. It records nothing when admission would not change the sim.
func (l *sessionLoop) applyMemberRejoinLocked(playerID uint64, respawn bool) {
	if !respawn && l.sim.PlayerConnected(playerID) {
		return
	}
	if !l.recordMemberLifecycleLocked(game.SystemMemberRejoinInputType, game.MemberLifecycle{PlayerID: playerID, Respawn: respawn}) {
		return
	}
	if err := l.sim.ApplyMemberRejoin(playerID, respawn); err != nil {
		l.log.Error("respawn reconnecting player", "player_id", playerID, "error", err)
	}
}

// recordMemberLifecycleLocked persists the membership row. It is stamped with
// the finished tick, like system_load_shed, so the final leave before the loop
// stops never names a tick live did not run. It reports false only when the row
// cannot be encoded. The caller must then skip the mutation, because an
// unrecordable change would break replay.
func (l *sessionLoop) recordMemberLifecycleLocked(typ string, m game.MemberLifecycle) bool {
	messageID := ids.New("sys")
	payload, err := inputdecode.EncodeStoredMemberLifecycle(typ, messageID, m)
	if err != nil {
		l.log.Error("encode member lifecycle input", "type", typ, "player_id", m.PlayerID, "error", err)
		return false
	}
	l.persistSystemInput(l.systemInputRowLocked(int64(l.sim.CurrentTick())-1, messageID, payload))
	return true
}

// systemInputRowLocked allocates the sequence for an actor-less, server-authored
// input row. Caller holds l.mu.
func (l *sessionLoop) systemInputRowLocked(tick int64, messageID string, payload json.RawMessage) *store.SessionInput {
	sequence := l.seq
	l.seq++
	l.seen[messageID] = true
	l.noteDurableLocked(tick)
	return &store.SessionInput{
		ID:        ids.New("inp"),
		SessionID: l.sess.ID,
		Tick:      tick,
		Sequence:  sequence,
		MessageID: messageID,
		Payload:   payload,
	}
}

func (l *sessionLoop) persistSystemInput(rec *store.SessionInput) {
	if rec == nil {
		return
	}
	if err := l.hub.store.AppendInput(context.Background(), *rec); err != nil {
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist system input", "type", systemInputType(rec.Payload), "tick", rec.Tick, "error", err)
	}
}

func systemInputType(payload json.RawMessage) string {
	var env struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(payload, &env)
	return env.Type
}
