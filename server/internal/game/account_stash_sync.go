package game

import (
	"encoding/json"
	"sort"
)

// SystemAccountStashSyncInputType is a stored-only, server-authored input (v488). It carries the
// account stash rows a live runner read from the store after an HTTP route (market, account-stash
// upgrade/merge/renew) changed them, so the open session sees the change and replay applies the
// same payload at the same tick. Clients can never send it: inputdecode accepts it only from the
// persisted input stream.
//
// A row recorded at tick T is applied at the start of tick T, before player inputs, and applies
// even when the player is dead (the stash is account state). It never produces acks or rejects.
const SystemAccountStashSyncInputType = "system_account_stash_sync"

// AccountStashSync replaces one player's account stash with the authoritative rows.
type AccountStashSync struct {
	PlayerID uint64
	Items    []PersistedStashItem
	Gold     int
}

// PlayerIDsForAccount returns, sorted, the players in this sim that belong to accountID.
func (s *Sim) PlayerIDsForAccount(accountID string) []uint64 {
	var out []uint64
	if accountID == "" {
		return out
	}
	for _, id := range sortedPlayerIDs(s.players) {
		if ps := s.players[id]; ps != nil && ps.AccountID == accountID {
			out = append(out, id)
		}
	}
	return out
}

// splitAccountStashSyncInputs removes stash-sync rows from a tick's inputs. Only actor-less rows
// with a payload count; a row carrying an actor is dropped rather than honored.
func splitAccountStashSyncInputs(inputs []Input) ([]Input, []AccountStashSync) {
	found := false
	for _, in := range inputs {
		if in.Type == SystemAccountStashSyncInputType {
			found = true
			break
		}
	}
	if !found {
		return inputs, nil
	}
	players := make([]Input, 0, len(inputs))
	var syncs []AccountStashSync
	for _, in := range inputs {
		if in.Type != SystemAccountStashSyncInputType {
			players = append(players, in)
			continue
		}
		if in.ActorPlayerID == 0 && in.StashSync != nil {
			syncs = append(syncs, *in.StashSync)
		}
	}
	return players, syncs
}

// applyAccountStashSyncs applies each sync in input order and restores the active player.
func (s *Sim) applyAccountStashSyncs(ctx *simTickCtx, syncs []AccountStashSync) {
	if len(syncs) == 0 {
		return
	}
	previous := s.players[s.playerID]
	for _, sync := range syncs {
		ps := s.players[sync.PlayerID]
		if ps == nil {
			continue
		}
		s.usePlayer(ps)
		s.replaceAccountStash(sync, ctx.resultFor(ps.CurrentLevel, ps.PlayerID))
		s.savePlayer(ps)
	}
	if previous != nil {
		s.usePlayer(previous)
	}
}

// replaceAccountStash swaps the active player's stash for the synced rows and emits the difference
// as stash change ops: removals, then additions or in-place updates (both by ascending stash item
// ID), then the gold update.
func (s *Sim) replaceAccountStash(sync AccountStashSync, res *TickResult) {
	before := make(map[uint64]string, len(s.stashItems))
	for _, item := range s.stashItems {
		if item != nil {
			before[item.stashItemID] = stashItemFingerprint(item)
		}
	}
	oldGold := s.stashGold
	s.LoadAccountStash(sync.Items, sync.Gold, s.stashCapacity)

	after := make(map[uint64]*stashItem, len(s.stashItems))
	for _, item := range s.stashItems {
		if item != nil {
			after[item.stashItemID] = item
		}
	}
	removed := make([]uint64, 0)
	for id := range before { //nolint:determinism collected then sorted below
		if _, ok := after[id]; !ok {
			removed = append(removed, id)
		}
	}
	sort.Slice(removed, func(i, j int) bool { return removed[i] < removed[j] })
	for _, id := range removed {
		res.Changes = append(res.Changes, Change{Op: OpStashItemRemove, StashItemID: idStr(id)})
	}
	for _, item := range s.stashItems { // sorted by stash item ID in LoadAccountStash
		if item == nil {
			continue
		}
		if fingerprint, ok := before[item.stashItemID]; ok && fingerprint == stashItemFingerprint(item) {
			continue
		}
		res.Changes = append(res.Changes, Change{Op: OpStashItemAdd, StashItem: ptrStashItemView(s.stashItemView(item))})
	}
	if s.stashGold != oldGold {
		res.Changes = append(res.Changes, Change{Op: OpStashGoldUpdate, StashGold: intPtr(s.stashGold)})
	}
}

// stashItemFingerprint identifies an item's own data (definition and rolls), not the
// equipment-dependent annotations stashItemView adds.
func stashItemFingerprint(item *stashItem) string {
	raw, err := json.Marshal(item.view())
	if err != nil {
		return item.itemDefID
	}
	return string(raw)
}
