package realtime

import (
	"context"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/sessionsetup"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// admitMemberLocked puts a resolved member player into play: reconnecting
// co-op members respawn in town, and the member row records the connection and
// the exact tick the player entity joined. Caller holds l.mu.
func (l *sessionLoop) admitMemberLocked(member store.SessionMember, playerID uint64) {
	isCoopMember := isCoopSession(l.sess) ||
		member.AccountID != l.sess.AccountID ||
		member.CharacterID != l.sess.CharacterID
	currentLevel, ok := l.sim.PlayerCurrentLevel(playerID)
	if !ok {
		return
	}
	respawn := (isCoopMember || playerID != l.sim.DefaultPlayerID()) && (!member.Connected || member.CurrentLevel != 0 || currentLevel != 0 || !l.sim.PlayerConnected(playerID))
	// Recorded as system_member_rejoin (v479) so replay respawns at this point.
	l.applyMemberRejoinLocked(playerID, respawn)
	level, _ := l.sim.PlayerCurrentLevel(playerID)
	// joined_tick must be the tick the entity was added, not "now": the lock
	// is released between AddGuestPlayer and here, so ticks may have run.
	joinedTick, _ := l.sim.PlayerJoinedTick(playerID)
	if err := l.hub.store.SetSessionMemberConnected(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), level, int64(joinedTick)); err != nil {
		// joined_tick is a replay input (v480): a lost write must be visible.
		l.hub.metrics.PersistenceErrors.Inc()
		l.log.Error("persist member connected", "account_id", member.AccountID, "character_id", member.CharacterID, "joined_tick", joinedTick, "error", err)
	}
}

// persistMemberDisconnected clears the member's connected flag. A dropped write
// leaves connected=true, which rejects the next attach as member_already_connected.
func (h *Hub) persistMemberDisconnected(sessionID, accountID, characterID string, level int, tick int64) {
	if err := h.store.SetSessionMemberDisconnected(context.Background(), sessionID, accountID, characterID, level, tick); err != nil {
		h.metrics.PersistenceErrors.Inc()
		h.log.Error("persist member disconnected", "session_id", sessionID, "account_id", accountID, "character_id", characterID, "error", err)
	}
}

func (l *sessionLoop) playerIDForMember(ctx context.Context, member store.SessionMember) uint64 {
	if id, ok := game.ParseEntityID(member.PlayerEntityID); ok && id != 0 {
		l.mu.Lock()
		_, exists := l.sim.PlayerCurrentLevel(id)
		l.mu.Unlock()
		if exists {
			return id
		}
	}
	l.mu.Lock()
	if playerID, ok := l.sim.PlayerIDForCharacter(member.CharacterID); ok {
		l.mu.Unlock()
		return playerID
	}
	l.mu.Unlock()

	guest, err := sessionsetup.Resolve(ctx, l.hub.store, l.hub.rules, member.SessionID, member)
	if err != nil {
		l.log.Error("resolve late-joined member", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}

	l.mu.Lock()
	if playerID, ok := l.sim.PlayerIDForCharacter(member.CharacterID); ok {
		l.mu.Unlock()
		return playerID
	}
	playerID, err := sessionsetup.AddGuest(l.sim, guest)
	var joinedTick uint64
	if err == nil {
		joinedTick, _ = l.sim.PlayerJoinedTick(playerID)
		l.recordMemberJoinLocked(member, playerID)
	}
	l.mu.Unlock()
	if err != nil {
		l.log.Error("add late-joined guest player", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}
	if err := l.hub.store.SetSessionMemberPlayer(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), 0, int64(joinedTick)); err != nil && err != store.ErrNotFound {
		l.log.Error("set late-joined member player", "account_id", member.AccountID, "character_id", member.CharacterID, "player_id", playerID, "error", err)
	}
	return playerID
}
