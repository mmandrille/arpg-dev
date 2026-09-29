package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sort"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v477 regression: live member setup loaded the resource wallet, resource bag,
// and recoverable corpses, but replay did not. Both bump or allocate entity
// IDs (the corpse spawns when its death level is generated), so every later
// entity ID and event diverged.
const (
	memberSetupWorld      = "dungeon_levels"
	memberSetupSeed       = "v477_member_setup_seed"
	memberSetupCorpseDef  = "hero_corpse"
	memberSetupStairsDef  = "stairs_down"
	memberSetupCorpseLvl  = -1
	memberSetupMaxTravel  = 200
	memberSetupSettleTick = 80
	memberSetupReach      = 1.5
)

func TestHostCorpseAndResourceBagReplayMatchesLive(t *testing.T) {
	ctx := context.Background()
	repo := newMemberSetupRepo(t)
	withMemberCorpseAndBag(repo, repo.session.AccountID, repo.session.CharacterID, "host")
	loop := newMemberSetupLoop(t, repo)
	host := repo.memberByRole(store.SessionMemberHost)
	client := attachMemberSetupClient(ctx, loop, host)

	descendToCorpseLevel(t, loop, client)
	settleMemberSetupLoop(t, loop, client)

	if !liveLevelHasEntity(loop, memberSetupCorpseLvl, memberSetupCorpseDef) {
		t.Fatal("host corpse did not spawn on its death level; fixture no longer exercises corpse allocation")
	}
	assertReplayMatchesLive(t, repo, loop)
}

func TestLateJoinGuestCorpseAndResourceBagReplayMatchesLive(t *testing.T) {
	ctx := context.Background()
	repo := newMemberSetupRepo(t)
	guestAcct, guestChar := "acct_guest", "char_guest"
	withMemberCorpseAndBag(repo, guestAcct, guestChar, "guest")
	loop := newMemberSetupLoop(t, repo)
	host := repo.memberByRole(store.SessionMemberHost)
	client := attachMemberSetupClient(ctx, loop, host)

	// The host generates the corpse level first, so the guest's corpse spawns
	// (allocates) at the moment the guest is admitted.
	descendToCorpseLevel(t, loop, client)
	admitLoadShedTestMember(ctx, loop, repo.joinGuest())
	settleMemberSetupLoop(t, loop, client)

	if !liveLevelHasEntity(loop, memberSetupCorpseLvl, memberSetupCorpseDef) {
		t.Fatal("guest corpse did not spawn on the generated level; fixture no longer exercises join-time allocation")
	}
	assertReplayMatchesLive(t, repo, loop)
}

// Gameplay-debug servers seed unique test chests when the sim is created (the
// env flag) and again if SetGameplayDebug runs after the per-player loads.
// Live must apply the flag before member state loads, like replay effectively
// does, or its allocator runs ahead and the late guest's entity ID diverges.
func TestGameplayDebugLiveBuildReplayMatches(t *testing.T) {
	t.Setenv("ARPG_GAMEPLAY_DEBUG", "true")
	ctx := context.Background()
	repo := newMemberSetupRepo(t)
	repo.session.WorldID = "vendor_lab"
	repo.session.Seed = "town_vendor_gold_sink_1075"
	loop := newMemberSetupLoopWithDebug(t, repo, true)
	host := repo.memberByRole(store.SessionMemberHost)
	client := attachMemberSetupClient(ctx, loop, host)
	sendMemberSetupInput(t, loop, client, "move_intent", map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 1})
	loop.doTick()
	admitLoadShedTestMember(ctx, loop, repo.joinGuest())
	loop.doTick()
	drainLoadShedTestClient(client)

	assertReplayMatchesLive(t, repo, loop)
}

// memberSetupRepo extends the v476 in-memory repo. Corpses live only in the
// frozen session-start snapshot; the mutable live corpse table panics if read.
type memberSetupRepo struct {
	*loadShedMemRepo
}

func newMemberSetupRepo(t *testing.T) *memberSetupRepo {
	t.Helper()
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	base := newLoadShedMemRepo(rules)
	base.session.ID = "sess_v477_member_setup"
	base.session.Seed = memberSetupSeed
	base.session.WorldID = memberSetupWorld
	for key, start := range base.starts {
		start.SessionID = base.session.ID
		start.Progression = nil
		base.starts[key] = start
	}
	for i := range base.members {
		base.members[i].SessionID = base.session.ID
	}
	return &memberSetupRepo{loadShedMemRepo: base}
}

// withMemberCorpseAndBag gives a member a same-account corpse on the first
// dungeon level plus a resource bag item, both with high persisted IDs so a
// missing load shifts the sim's entity ID allocator.
func withMemberCorpseAndBag(repo *memberSetupRepo, accountID, characterID, tag string) {
	start := repo.starts[characterID]
	start.AccountID = accountID
	start.CharacterID = characterID
	start.ResourceBagItems = []store.AccountResourceBagItem{{
		AccountID:   accountID,
		BagItemID:   "880001",
		ItemDefID:   "renew_stone",
		RolledStats: json.RawMessage(`{}`),
	}}
	start.Corpses = []store.CharacterCorpse{{
		CharacterID: "char_dead_" + tag,
		Name:        "Fallen " + tag,
		Level:       3,
		DeathLevel:  memberSetupCorpseLvl,
		Items: []store.CharacterItemInstance{{
			ID:          "870001",
			AccountID:   accountID,
			CharacterID: "char_dead_" + tag,
			ItemDefID:   "rusty_sword",
			Location:    store.ItemLocationInventory,
			RolledStats: json.RawMessage(`{}`),
		}},
	}}
	repo.starts[characterID] = start
}

// ListRecoverableCharacterCorpses is the mutable live table: corpses get looted
// or added after the session starts, so neither live setup nor replay may read
// it. Both must use the session-start snapshot.
func (r *memberSetupRepo) ListRecoverableCharacterCorpses(context.Context, string, string) ([]store.CharacterCorpse, error) {
	panic("session setup must load corpses from the session start snapshot, not the live corpse table")
}

func newMemberSetupLoop(t *testing.T, repo *memberSetupRepo) *sessionLoop {
	t.Helper()
	return newMemberSetupLoopWithDebug(t, repo, false)
}

func newMemberSetupLoopWithDebug(t *testing.T, repo *memberSetupRepo, gameplayDebug bool) *sessionLoop {
	t.Helper()
	hub := &Hub{store: repo, rules: repo.rules, log: slog.New(slog.NewTextHandler(io.Discard, nil)), metrics: metrics.New(), gameplayDebug: gameplayDebug}
	loop, err := newSessionLoop(context.Background(), hub, repo.session)
	if err != nil {
		t.Fatalf("new session loop: %v", err)
	}
	return loop
}

func attachMemberSetupClient(ctx context.Context, loop *sessionLoop, member store.SessionMember) *loopClient {
	client := &loopClient{
		loop:   loop,
		key:    memberKey(member),
		member: member,
		sendCh: make(chan outEnvelope, 4096),
		done:   make(chan struct{}),
	}
	client.playerID = admitLoadShedTestMember(ctx, loop, member)
	loop.clients[client.key] = client
	return client
}

// descendToCorpseLevel walks the host to the town stairs and descends, all
// through recorded client intents (move_to_intent, then descend_intent).
func descendToCorpseLevel(t *testing.T, loop *sessionLoop, client *loopClient) {
	t.Helper()
	stairs, ok := liveEntity(loop, 0, memberSetupStairsDef)
	if !ok {
		t.Fatal("town has no down stairs")
	}
	sendMemberSetupInput(t, loop, client, "move_to_intent", map[string]any{"position": stairs.Position})
	descendSent := false
	for i := 0; i < memberSetupMaxTravel; i++ {
		if !descendSent && livePlayerNear(loop, client.playerID, stairs.Position) {
			sendMemberSetupInput(t, loop, client, "descend_intent", map[string]any{})
			descendSent = true
		}
		loop.doTick()
		drainLoadShedTestClient(client)
		if level, _ := loop.sim.PlayerCurrentLevel(client.playerID); level == memberSetupCorpseLvl {
			return
		}
	}
	t.Fatalf("host did not reach level %d within %d ticks (descend sent=%v)", memberSetupCorpseLvl, memberSetupMaxTravel, descendSent)
}

// settleMemberSetupLoop walks the host at the nearest monster so combat events
// carrying entity IDs allocated after the corpse reach the recorded stream.
func settleMemberSetupLoop(t *testing.T, loop *sessionLoop, client *loopClient) {
	t.Helper()
	monster, ok := nearestLiveMonster(loop, client.playerID)
	if !ok {
		t.Fatal("corpse level has no monsters to engage")
	}
	sendMemberSetupInput(t, loop, client, "move_to_intent", map[string]any{"position": monster.Position})
	for i := 0; i < memberSetupSettleTick; i++ {
		loop.doTick()
		drainLoadShedTestClient(client)
	}
}

func nearestLiveMonster(loop *sessionLoop, playerID uint64) (game.EntityView, bool) {
	loop.mu.Lock()
	defer loop.mu.Unlock()
	entities := loop.sim.SnapshotForPlayer(playerID).Entities
	var player game.EntityView
	for _, e := range entities {
		if e.ID == fmt.Sprint(playerID) {
			player = e
		}
	}
	best, found := game.EntityView{}, false
	bestDist := math.MaxFloat64
	for _, e := range entities {
		if e.MonsterDefID == "" {
			continue
		}
		if d := math.Hypot(e.Position.X-player.Position.X, e.Position.Y-player.Position.Y); d < bestDist {
			best, bestDist, found = e, d, true
		}
	}
	return best, found
}

func sendMemberSetupInput(t *testing.T, loop *sessionLoop, client *loopClient, typ string, payload map[string]any) {
	t.Helper()
	tick := loop.sim.CurrentTick()
	env, err := json.Marshal(map[string]any{
		"type":       typ,
		"message_id": fmt.Sprintf("msg-%s-%d", typ, tick),
		"session_id": loop.sess.ID,
		"tick":       tick,
		"payload":    payload,
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	loop.handleClientMessage(client, env)
}

func liveEntity(loop *sessionLoop, level int, interactableDefID string) (game.EntityView, bool) {
	for _, e := range liveLevelEntities(loop, level) {
		if e.InteractableDefID == interactableDefID {
			return e, true
		}
	}
	return game.EntityView{}, false
}

func liveLevelHasEntity(loop *sessionLoop, level int, interactableDefID string) bool {
	_, ok := liveEntity(loop, level, interactableDefID)
	return ok
}

// livePlayerNear reports whether the player is within interaction reach of pos.
func livePlayerNear(loop *sessionLoop, playerID uint64, pos game.Vec2) bool {
	loop.mu.Lock()
	defer loop.mu.Unlock()
	for _, e := range loop.sim.SnapshotForPlayer(playerID).Entities {
		if e.ID == fmt.Sprint(playerID) {
			return math.Hypot(e.Position.X-pos.X, e.Position.Y-pos.Y) <= memberSetupReach
		}
	}
	return false
}

func liveLevelEntities(loop *sessionLoop, level int) []game.EntityView {
	loop.mu.Lock()
	defer loop.mu.Unlock()
	for _, playerID := range loop.sim.PlayerIDs() {
		if current, _ := loop.sim.PlayerCurrentLevel(playerID); current == level {
			return loop.sim.SnapshotForPlayer(playerID).Entities
		}
	}
	return nil
}

func assertReplayMatchesLive(t *testing.T, repo *memberSetupRepo, loop *sessionLoop) {
	t.Helper()
	report, err := replay.Verify(context.Background(), repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !report.Match {
		t.Fatalf("replay mismatch: %s", report.Mismatch)
	}
	recon, err := replay.Reconstruct(context.Background(), repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("reconstruct: %v", err)
	}
	loop.mu.Lock()
	defer loop.mu.Unlock()
	for _, playerID := range loop.sim.PlayerIDs() {
		live := loop.sim.SnapshotForPlayer(playerID)
		replayed := recon.Sim.SnapshotForPlayer(playerID)
		if got, want := sortedEntityIDs(replayed.Entities), sortedEntityIDs(live.Entities); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("player %d entity IDs: replay %v != live %v", playerID, got, want)
		}
		if got, want := fmt.Sprint(replayed.ResourceBagItems), fmt.Sprint(live.ResourceBagItems); got != want {
			t.Fatalf("player %d resource bag: replay %s != live %s", playerID, got, want)
		}
	}
}

func sortedEntityIDs(entities []game.EntityView) []string {
	ids := make([]string, 0, len(entities))
	for _, e := range entities {
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	return ids
}
