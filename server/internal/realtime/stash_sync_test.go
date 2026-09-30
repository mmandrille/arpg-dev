package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sort"
	"sync"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// stashMemRepo adds account stash reads to the in-memory session repo, standing in for the rows an
// HTTP market/stash route commits while the session is live.
type stashMemRepo struct {
	*loadShedMemRepo
	stashMu  sync.Mutex
	stash    map[string][]store.AccountStashItem
	gold     map[string]int
	readErr  error
	readHits int
}

func (r *stashMemRepo) ListAccountStashItems(_ context.Context, accountID string) ([]store.AccountStashItem, error) {
	r.stashMu.Lock()
	defer r.stashMu.Unlock()
	r.readHits++
	if r.readErr != nil {
		return nil, r.readErr
	}
	return append([]store.AccountStashItem(nil), r.stash[accountID]...), nil
}

func (r *stashMemRepo) GetOrCreateAccountStashGold(_ context.Context, accountID string) (store.AccountStashGold, error) {
	r.stashMu.Lock()
	defer r.stashMu.Unlock()
	return store.AccountStashGold{AccountID: accountID, Gold: r.gold[accountID]}, nil
}

// commitHTTPStash mimics an HTTP route writing the account stash in the store.
func (r *stashMemRepo) commitHTTPStash(accountID string, gold int, items ...store.AccountStashItem) {
	r.stashMu.Lock()
	defer r.stashMu.Unlock()
	r.stash[accountID] = items
	r.gold[accountID] = gold
}

func newStashSyncTestLoop(t *testing.T) (*stashMemRepo, *Hub, *sessionLoop) {
	t.Helper()
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	repo := &stashMemRepo{loadShedMemRepo: newLoadShedMemRepo(rules), stash: map[string][]store.AccountStashItem{}, gold: map[string]int{}}
	hub := &Hub{store: repo, rules: rules, log: slog.New(slog.NewTextHandler(io.Discard, nil)), metrics: metrics.New(), loops: map[string]*sessionLoop{}}
	loop, err := newSessionLoop(context.Background(), hub, repo.session)
	if err != nil {
		t.Fatalf("new session loop: %v", err)
	}
	hub.loops[repo.session.ID] = loop
	return repo, hub, loop
}

func attachStashSyncClient(ctx context.Context, loop *sessionLoop, member store.SessionMember) *loopClient {
	client := &loopClient{loop: loop, key: memberKey(member), member: member, sendCh: make(chan outEnvelope, 4096), done: make(chan struct{})}
	client.playerID = admitLoadShedTestMember(ctx, loop, member)
	loop.clients[client.key] = client
	return client
}

func stashRowFor(accountID, id, itemDefID string) store.AccountStashItem {
	return store.AccountStashItem{AccountID: accountID, StashItemID: id, ItemDefID: itemDefID, RolledStats: json.RawMessage(`{}`)}
}

func stashIDs(views []game.StashItemView) []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		out = append(out, v.StashItemID)
	}
	sort.Strings(out)
	return out
}

func recordedStashSyncRows(repo *stashMemRepo) int {
	n := 0
	for _, in := range repo.inputs {
		if systemInputType(in.Payload) == game.SystemAccountStashSyncInputType {
			n++
		}
	}
	return n
}

// An HTTP stash write reaches the owning player's live stash on the next tick, is recorded, and
// replay rebuilds the same stash; other players' stashes are untouched.
func TestHTTPStashChangeSyncsLiveSessionAndReplays(t *testing.T) {
	ctx := context.Background()
	repo, hub, loop := newStashSyncTestLoop(t)
	host := attachStashSyncClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	guest := attachStashSyncClient(ctx, loop, repo.joinGuest())
	loop.doTick()

	repo.commitHTTPStash(host.member.AccountID, 7, stashRowFor(host.member.AccountID, "7010", "long_sword"))
	hub.NotifyAccountStashChanged(host.member.AccountID)
	loop.doTick()
	loop.doTick()
	drainLoadShedTestClient(host)
	drainLoadShedTestClient(guest)

	hostSnap := loop.sim.SnapshotForPlayer(host.playerID)
	if got := stashIDs(hostSnap.StashItems); len(got) != 1 || got[0] != "7010" {
		t.Fatalf("host live stash = %v, want [7010]", got)
	}
	if hostSnap.StashGold != 7 {
		t.Fatalf("host live stash gold = %v, want 7", hostSnap.StashGold)
	}
	if got := stashIDs(loop.sim.SnapshotForPlayer(guest.playerID).StashItems); len(got) != 0 {
		t.Fatalf("guest stash changed by the host's sync: %v", got)
	}
	if n := recordedStashSyncRows(repo); n != 1 {
		t.Fatalf("recorded %d stash sync rows, want 1", n)
	}

	recon, err := replay.Reconstruct(ctx, repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	if got := stashIDs(recon.Sim.SnapshotForPlayer(host.playerID).StashItems); len(got) != 1 || got[0] != "7010" {
		t.Fatalf("replayed host stash = %v, want [7010]", got)
	}
	report, err := replay.Verify(ctx, repo, repo.rules, repo.session.ID)
	if err != nil || !report.Match {
		t.Fatalf("verify err=%v match=%v mismatch=%s", err, report.Match, report.Mismatch)
	}
}

// The fixture must detect an unrecorded sync; otherwise the parity check above proves nothing.
func TestUnrecordedStashSyncDivergesInReplay(t *testing.T) {
	ctx := context.Background()
	repo, hub, loop := newStashSyncTestLoop(t)
	host := attachStashSyncClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	loop.doTick()
	repo.commitHTTPStash(host.member.AccountID, 0, stashRowFor(host.member.AccountID, "7010", "long_sword"))
	hub.NotifyAccountStashChanged(host.member.AccountID)
	loop.doTick()
	drainLoadShedTestClient(host)

	kept := repo.inputs[:0]
	for _, in := range repo.inputs {
		if systemInputType(in.Payload) != game.SystemAccountStashSyncInputType {
			kept = append(kept, in)
		}
	}
	repo.inputs = kept

	recon, err := replay.Reconstruct(ctx, repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	live := stashIDs(loop.sim.SnapshotForPlayer(host.playerID).StashItems)
	replayed := stashIDs(recon.Sim.SnapshotForPlayer(host.playerID).StashItems)
	if len(live) == len(replayed) {
		t.Fatalf("unrecorded sync replayed the same stash (live %v, replay %v); fixture is not sensitive", live, replayed)
	}
}

func TestStashSyncIgnoresAccountsWithoutAPlayer(t *testing.T) {
	ctx := context.Background()
	repo, hub, loop := newStashSyncTestLoop(t)
	host := attachStashSyncClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	hub.NotifyAccountStashChanged("acct_not_in_session")
	loop.doTick()
	drainLoadShedTestClient(host)
	if repo.readHits != 0 || recordedStashSyncRows(repo) != 0 {
		t.Fatalf("absent account read the store %d times and recorded %d rows", repo.readHits, recordedStashSyncRows(repo))
	}
}

// A failed store read is metered, records nothing, and drops the flag instead of retrying every tick.
func TestStashSyncReadFailureIsDroppedNotRetried(t *testing.T) {
	ctx := context.Background()
	repo, hub, loop := newStashSyncTestLoop(t)
	host := attachStashSyncClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	repo.readErr = errors.New("db down")
	before := counterValue(t, hub.metrics.PersistenceErrors)
	hub.NotifyAccountStashChanged(host.member.AccountID)
	loop.doTick()
	loop.doTick()
	drainLoadShedTestClient(host)
	if repo.readHits != 1 {
		t.Fatalf("store read %d times, want exactly 1 (no per-tick retry)", repo.readHits)
	}
	if got := counterValue(t, hub.metrics.PersistenceErrors) - before; got != 1 {
		t.Fatalf("persistence errors delta = %v, want 1", got)
	}
	if recordedStashSyncRows(repo) != 0 {
		t.Fatal("a failed read recorded a stash sync row")
	}
}
