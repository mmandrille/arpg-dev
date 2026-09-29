package realtime

import (
	"context"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// admitMemberLocked puts a resolved member player into play: reconnecting
// co-op members respawn in town, and the member row records the connection and
// the exact tick the player entity joined. Caller holds l.mu.
func (l *sessionLoop) admitMemberLocked(member store.SessionMember, playerID uint64) {
	isCoopMember := isCoopSession(l.sess) ||
		member.AccountID != l.sess.AccountID ||
		member.CharacterID != l.sess.CharacterID
	currentLevel, _ := l.sim.PlayerCurrentLevel(playerID)
	if (isCoopMember || playerID != l.sim.DefaultPlayerID()) && (!member.Connected || member.CurrentLevel != 0 || currentLevel != 0 || !l.sim.PlayerConnected(playerID)) {
		if err := l.sim.RespawnPlayerInTown(playerID); err != nil {
			l.log.Error("respawn reconnecting player", "player_id", playerID, "error", err)
		}
	}
	if level, ok := l.sim.PlayerCurrentLevel(playerID); ok {
		// joined_tick must be the tick the entity was added, not "now": the lock
		// is released between AddGuestPlayer and here, so ticks may have run.
		joinedTick, _ := l.sim.PlayerJoinedTick(playerID)
		_ = l.hub.store.SetSessionMemberConnected(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), level, int64(joinedTick))
		l.sim.SetPlayerConnected(playerID, true)
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

	start, err := l.hub.store.LoadSessionStartSnapshotForMember(ctx, member.SessionID, member.AccountID, member.CharacterID)
	if err != nil {
		l.log.Error("load late-joined member start snapshot", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}

	l.mu.Lock()
	if playerID, ok := l.sim.PlayerIDForCharacter(member.CharacterID); ok {
		l.mu.Unlock()
		return playerID
	}
	playerID, err := l.sim.AddGuestPlayer(member.AccountID, member.CharacterID, displayNameForMember(member), progressionStateFromStore(l.hub.rules, start.Progression))
	if err != nil {
		l.mu.Unlock()
		l.log.Error("add late-joined guest player", "account_id", member.AccountID, "character_id", member.CharacterID, "error", err)
		return l.sim.DefaultPlayerID()
	}
	l.sim.LoadInventoryForPlayer(playerID, persistedItems(start.Items))
	l.sim.LoadHotbarForPlayer(playerID, persistedHotbar(start.Hotbar))
	l.sim.LoadSkillBindingsForPlayer(playerID, persistedSkillBindings(start.SkillBinds))
	l.sim.LoadDiscoveredTeleportersForPlayer(playerID, waypointLevels(start.Waypoints))
	l.sim.LoadShopStockForPlayer(playerID, persistedShopStock(start.ShopStock))
	l.sim.LoadAccountStashForPlayer(playerID, persistedStashItems(start.StashItems), start.StashGold.Gold, 0)
	l.sim.LoadResourceWalletForPlayer(playerID, persistedResources(start.Resources))
	l.sim.LoadAccountResourceBagForPlayer(playerID, persistedResourceBagItems(start.ResourceBagItems))
	l.hub.loadCharacterCorpses(context.Background(), l.log, l.sim, member)
	l.mu.Unlock()
	if err := l.hub.store.SetSessionMemberPlayer(context.Background(), member.SessionID, member.AccountID, member.CharacterID, idStr(playerID), 0); err != nil && err != store.ErrNotFound {
		l.log.Error("set late-joined member player", "account_id", member.AccountID, "character_id", member.CharacterID, "player_id", playerID, "error", err)
	}
	return playerID
}
