package realtime

import (
	"context"
	"sort"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
	"github.com/mmandrille_meli/arpg-dev/server/internal/sessionsetup"
)

// Live stash sync (v488). HTTP routes (market, account-stash upgrade/merge/renew) write account
// stash rows directly in the store, while each live session holds its own copy of the stash in the
// sim, loaded at session start. After a successful stash write, HTTP calls
// NotifyAccountStashChanged. Each live loop then reads the account's stash at the start of its next
// tick and records a system_account_stash_sync row for that tick, which the sim applies before
// player inputs. Reading inside the loop (not in HTTP), after flushDeferredPersist and after the
// previous tick's persistTick, means every sim-originated stash write is already in the store, so the
// sync can never drop a deposit the sim applied but had not yet persisted.

// stashSyncReadTimeout bounds the store read on the tick path. On failure the sync is skipped and
// the flag dropped: retrying every tick would stall ticks for the whole outage, and the next HTTP
// stash change or session build resyncs.
const stashSyncReadTimeout = 250 * time.Millisecond

// NotifyAccountStashChanged marks the accounts' stash dirty on every live loop. Loops without a
// player of that account ignore the flag when they take it.
func (h *Hub) NotifyAccountStashChanged(accountIDs ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, loop := range h.loops {
		for _, accountID := range accountIDs {
			if accountID == "" {
				continue
			}
			if h.stashDirty == nil {
				h.stashDirty = make(map[*sessionLoop]map[string]bool)
			}
			if h.stashDirty[loop] == nil {
				h.stashDirty[loop] = make(map[string]bool)
			}
			h.stashDirty[loop][accountID] = true
		}
	}
}

// takeDirtyStashAccounts returns, sorted, and clears the accounts marked for loop.
func (h *Hub) takeDirtyStashAccounts(loop *sessionLoop) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.stashDirty[loop]
	if len(set) == 0 {
		return nil
	}
	delete(h.stashDirty, loop)
	out := make([]string, 0, len(set))
	for accountID := range set {
		out = append(out, accountID)
	}
	sort.Strings(out)
	return out
}

// stashSyncInputsLocked reads the stash of every dirty account that has a player in this sim,
// records one sync row per player at tick, and returns the inputs to run with that tick. Caller
// holds l.mu and has already flushed deferred persistence.
func (l *sessionLoop) stashSyncInputsLocked(tick uint64) []game.Input {
	if l.hub == nil || l.sim == nil {
		return nil
	}
	var out []game.Input
	for _, accountID := range l.hub.takeDirtyStashAccounts(l) {
		players := l.sim.PlayerIDsForAccount(accountID)
		if len(players) == 0 {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), stashSyncReadTimeout)
		items, err := l.hub.store.ListAccountStashItems(ctx, accountID)
		gold := 0
		if err == nil {
			stashGold, goldErr := l.hub.store.GetOrCreateAccountStashGold(ctx, accountID)
			gold, err = stashGold.Gold, goldErr
		}
		cancel()
		if err != nil {
			l.hub.metrics.PersistenceErrors.Inc()
			l.log.Error("read account stash for live sync", "account_id", accountID, "error", err)
			continue
		}
		persisted := sessionsetup.PersistedStashItems(items)
		for _, playerID := range players {
			sync := game.AccountStashSync{PlayerID: playerID, Items: persisted, Gold: gold}
			messageID := ids.New("sys")
			payload, err := inputdecode.EncodeStoredAccountStashSync(messageID, sync)
			if err != nil {
				l.log.Error("encode account stash sync input", "account_id", accountID, "player_id", playerID, "error", err)
				continue
			}
			rec := l.systemInputRowLocked(int64(tick), messageID, payload)
			l.persistSystemInput(rec)
			out = append(out, game.Input{
				MessageID: messageID,
				Sequence:  rec.Sequence,
				Type:      game.SystemAccountStashSyncInputType,
				StashSync: &sync,
			})
		}
	}
	return out
}
