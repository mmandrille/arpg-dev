package realtime

import (
	"context"
	"encoding/json"
	"fmt"
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

// v476 regression: the live runner used to apply wall-clock load shedding
// directly to the sim without recording it, so replay diverged and a co-op
// member joining later reconstructed with a different entity ID
// (make benchmark, sorcerer_multigroup_perf_probe).
const (
	loadShedTestWorld = "benchmark_mixed_arena"
	loadShedTestSeed  = "sorcerer_multigroup_perf_probe_seed"
	loadShedDegradeAt = 3
	loadShedJoinTick  = 49
	loadShedLastTick  = 90
)

// loadShedHostScript mirrors the opening of the benchmark bot's sorcerer run:
// casts that kill monsters and drop loot before the guest joins.
var loadShedHostScript = map[uint64][]struct {
	typ     string
	payload map[string]any
}{
	0:  {{"cast_skill_intent", map[string]any{"skill_id": "lightning", "direction": map[string]any{"x": 1, "y": 0}}}},
	5:  {{"cast_skill_intent", map[string]any{"skill_id": "ice_shard", "direction": map[string]any{"x": 1, "y": 0}}}},
	34: {{"cast_skill_intent", map[string]any{"skill_id": "magic_bolt", "direction": map[string]any{"x": 1, "y": 0}}}},
	44: {{"move_intent", map[string]any{"direction": map[string]any{"x": -1, "y": 0}, "duration_ticks": 25}}},
	45: {{"cast_skill_intent", map[string]any{"skill_id": "lightning", "direction": map[string]any{"x": 1, "y": 0}}}},
	59: {{"cast_skill_intent", map[string]any{"skill_id": "ice_shard", "direction": map[string]any{"x": 1, "y": 0}}}},
	75: {{"cast_skill_intent", map[string]any{"skill_id": "lightning", "direction": map[string]any{"x": 1, "y": 0}}}},
}

func TestRecordedLoadShedKeepsLateJoinReplayDeterministic(t *testing.T) {
	repo := runLoadShedLiveSession(t, false)

	if !repoHasSystemLoadShedInput(repo) {
		t.Fatal("live run did not record a system_load_shed input")
	}
	if !repoHasEventBefore(repo, "loot_dropped", loadShedJoinTick) {
		t.Fatal("scenario must allocate loot entity IDs before the guest joins")
	}
	guest := repo.memberByRole(store.SessionMemberGuest)
	if guest.JoinedTick != loadShedJoinTick || guest.PlayerEntityID == "" {
		t.Fatalf("guest member = %+v, want joined_tick=%d with a player entity", guest, loadShedJoinTick)
	}

	report, err := replay.Verify(context.Background(), repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !report.Match {
		t.Fatalf("replay mismatch: %s", report.Mismatch)
	}
}

// The fixture must be sensitive to load shedding; otherwise the test above
// would pass even if recording were removed.
func TestUnrecordedLoadShedBreaksReplay(t *testing.T) {
	repo := runLoadShedLiveSession(t, true)

	report, err := replay.Verify(context.Background(), repo, repo.rules, repo.session.ID)
	if err == nil && report.Match {
		t.Fatal("unrecorded degradation still replayed cleanly; fixture no longer exercises load shedding")
	}
}

func newLoadShedTestLoop(t *testing.T) (*loadShedMemRepo, *sessionLoop) {
	t.Helper()
	return newLoadShedTestLoopFor(t, nil)
}

// newLoadShedTestLoopFor lets a test adjust the session (world, seed) before
// the loop builds its sim.
func newLoadShedTestLoopFor(t *testing.T, configure func(*loadShedMemRepo)) (*loadShedMemRepo, *sessionLoop) {
	t.Helper()
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	repo := newLoadShedMemRepo(rules)
	if configure != nil {
		configure(repo)
	}
	hub := &Hub{store: repo, rules: rules, log: slog.New(slog.NewTextHandler(io.Discard, nil)), metrics: metrics.New()}
	loop, err := newSessionLoop(context.Background(), hub, repo.session)
	if err != nil {
		t.Fatalf("new session loop: %v", err)
	}
	return repo, loop
}

// joined_tick must be the tick the guest entity was added, even when ticks run
// between AddGuestPlayer and the attach that persists joined_tick.
func TestLateJoinRecordsTickGuestEntityWasAdded(t *testing.T) {
	ctx := context.Background()
	repo, loop := newLoadShedTestLoop(t)
	const addedAt = 10
	for loop.sim.CurrentTick() < addedAt {
		loop.doTick()
	}
	guest := repo.joinGuest()
	guest.Connected = true
	playerID := loop.playerIDForMember(ctx, guest)
	loop.doTick()
	loop.doTick()
	loop.mu.Lock()
	loop.admitMemberLocked(guest, playerID)
	loop.mu.Unlock()

	if got := repo.memberByRole(store.SessionMemberGuest).JoinedTick; got != addedAt {
		t.Fatalf("joined_tick = %d, want %d (tick the entity was added)", got, addedAt)
	}
	report, err := replay.Verify(ctx, repo, repo.rules, repo.session.ID)
	if err != nil || !report.Match {
		t.Fatalf("verify err=%v match=%v mismatch=%s", err, report.Match, report.Mismatch)
	}
}

func runLoadShedLiveSession(t *testing.T, bypassRecording bool) *loadShedMemRepo {
	t.Helper()
	ctx := context.Background()
	repo, loop := newLoadShedTestLoop(t)
	policyCalls := 0
	loop.loadShed = func(loadShedSample) game.LoadShedDirective {
		defer func() { policyCalls++ }()
		if policyCalls != loadShedDegradeAt {
			return game.LoadShedDirective{}
		}
		if bypassRecording {
			loop.sim.ApplyLoadShed(game.LoadShedDirective{OverloadDegrade: true})
			return game.LoadShedDirective{CombatMovementThrottle: true}
		}
		return game.LoadShedDirective{OverloadDegrade: true}
	}
	host := repo.memberByRole(store.SessionMemberHost)
	client := &loopClient{
		loop:   loop,
		key:    memberKey(host),
		member: host,
		sendCh: make(chan outEnvelope, 4096),
		done:   make(chan struct{}),
	}
	client.playerID = admitLoadShedTestMember(ctx, loop, host)
	loop.clients[client.key] = client
	for tick := uint64(0); tick <= loadShedLastTick; tick++ {
		if tick == loadShedJoinTick {
			// Mirrors POST /v0/sessions/{id}/join followed by the WS attach.
			admitLoadShedTestMember(ctx, loop, repo.joinGuest())
		}
		for i, step := range loadShedHostScript[tick] {
			env, err := json.Marshal(map[string]any{
				"type":       step.typ,
				"message_id": fmt.Sprintf("msg-%d-%d", tick, i),
				"session_id": repo.session.ID,
				"tick":       tick,
				"payload":    step.payload,
			})
			if err != nil {
				t.Fatalf("marshal input: %v", err)
			}
			loop.handleClientMessage(client, env)
		}
		loop.doTick()
		drainLoadShedTestClient(client)
	}
	return repo
}

func repoHasSystemLoadShedInput(repo *loadShedMemRepo) bool {
	for _, in := range repo.inputs {
		var env struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(in.Payload, &env) == nil && env.Type == game.SystemLoadShedInputType && in.ActorPlayerEntityID == "" {
			return true
		}
	}
	return false
}

func repoHasEventBefore(repo *loadShedMemRepo, eventType string, tick int64) bool {
	for _, ev := range repo.events {
		if ev.EventType == eventType && ev.Tick < tick {
			return true
		}
	}
	return false
}

// loadShedMemRepo is the minimal in-memory store for one live co-op session
// plus its replay. Unused Repository methods panic via the nil embed.
type loadShedMemRepo struct {
	store.Repository
	mu      sync.Mutex
	rules   *game.Rules
	session store.Session
	members []store.SessionMember
	starts  map[string]store.SessionStartSnapshot
	inputs  []store.SessionInput
	events  []store.SessionEvent
}

func newLoadShedMemRepo(rules *game.Rules) *loadShedMemRepo {
	sess := store.Session{
		ID:          "sess_v476_load_shed",
		AccountID:   "acct_host",
		CharacterID: "char_host",
		Seed:        loadShedTestSeed,
		WorldID:     loadShedTestWorld,
		Mode:        store.SessionModeCoop,
		Status:      store.SessionActive,
		Listed:      true,
	}
	hostProgression := &store.CharacterProgression{
		AccountID:      sess.AccountID,
		CharacterID:    sess.CharacterID,
		CharacterClass: "sorcerer",
		Level:          12,
		Stats:          store.CharacterBaseStats{Str: 3, Dex: 5, Vit: 40, Magic: 20},
		SkillRanks:     map[string]int{"magic_bolt": 1, "ice_shard": 1, "lightning": 1},
	}
	return &loadShedMemRepo{
		rules:   rules,
		session: sess,
		members: []store.SessionMember{
			{SessionID: sess.ID, AccountID: sess.AccountID, CharacterID: sess.CharacterID, Role: store.SessionMemberHost, Status: store.SessionMemberActive},
		},
		starts: map[string]store.SessionStartSnapshot{
			sess.CharacterID: {SessionID: sess.ID, AccountID: sess.AccountID, CharacterID: sess.CharacterID, Progression: hostProgression},
			"char_guest":     {SessionID: sess.ID, AccountID: "acct_guest", CharacterID: "char_guest"},
		},
	}
}

// joinGuest inserts the guest row the way the HTTP join handler does
// (joined_tick = -1 until the first WS attach).
func (r *loadShedMemRepo) joinGuest() store.SessionMember {
	guest := store.SessionMember{
		SessionID:   r.session.ID,
		AccountID:   "acct_guest",
		CharacterID: "char_guest",
		Role:        store.SessionMemberGuest,
		Status:      store.SessionMemberActive,
		JoinedTick:  store.SessionMemberNotJoinedTick,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members = append(r.members, guest)
	return guest
}

func (r *loadShedMemRepo) memberByRole(role string) store.SessionMember {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.members {
		if m.Role == role {
			return m
		}
	}
	return store.SessionMember{}
}

func (r *loadShedMemRepo) updateMember(accountID, characterID string, fn func(*store.SessionMember)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.members {
		if r.members[i].AccountID == accountID && r.members[i].CharacterID == characterID {
			fn(&r.members[i])
			return nil
		}
	}
	return store.ErrNotFound
}

func (r *loadShedMemRepo) GetSession(context.Context, string) (store.Session, error) {
	return r.session, nil
}

func (r *loadShedMemRepo) GetCharacter(_ context.Context, characterID string) (store.Character, error) {
	return store.Character{ID: characterID}, nil
}

func (r *loadShedMemRepo) ListCharacters(context.Context, string) ([]store.CharacterSummary, error) {
	return nil, nil
}

func (r *loadShedMemRepo) ListSessionMembers(context.Context, string) ([]store.SessionMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]store.SessionMember(nil), r.members...), nil
}

func (r *loadShedMemRepo) LoadSessionStartSnapshotForMember(_ context.Context, _, _, characterID string) (store.SessionStartSnapshot, error) {
	return r.starts[characterID], nil
}

func (r *loadShedMemRepo) SetSessionMemberPlayer(_ context.Context, _, accountID, characterID, playerEntityID string, currentLevel int, joinedTick int64) error {
	return r.updateMember(accountID, characterID, func(m *store.SessionMember) {
		m.PlayerEntityID = playerEntityID
		m.CurrentLevel = currentLevel
		m.JoinedTick = joinedTick
	})
}

// SetSessionMemberConnected mirrors the Postgres rule: joined_tick is only set
// the first time (while it is still negative).
func (r *loadShedMemRepo) SetSessionMemberConnected(_ context.Context, _, accountID, characterID, playerEntityID string, currentLevel int, tick int64) error {
	return r.updateMember(accountID, characterID, func(m *store.SessionMember) {
		m.Connected = true
		m.PlayerEntityID = playerEntityID
		m.CurrentLevel = currentLevel
		if m.JoinedTick < 0 {
			m.JoinedTick = tick
		}
	})
}

func (r *loadShedMemRepo) ListRecoverableCharacterCorpses(context.Context, string, string) ([]store.CharacterCorpse, error) {
	return nil, nil
}

func (r *loadShedMemRepo) AppendInput(_ context.Context, in store.SessionInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inputs = append(r.inputs, in)
	return nil
}

func (r *loadShedMemRepo) ListInputs(context.Context, string) ([]store.SessionInput, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := append([]store.SessionInput(nil), r.inputs...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Tick != out[j].Tick {
			return out[i].Tick < out[j].Tick
		}
		return out[i].Sequence < out[j].Sequence
	})
	return out, nil
}

func (r *loadShedMemRepo) AppendEvent(_ context.Context, ev store.SessionEvent) error {
	return r.AppendEvents(context.Background(), []store.SessionEvent{ev})
}

func (r *loadShedMemRepo) AppendEvents(_ context.Context, events []store.SessionEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, events...)
	return nil
}

func (r *loadShedMemRepo) ListEvents(context.Context, string) ([]store.SessionEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := append([]store.SessionEvent(nil), r.events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Tick != out[j].Tick {
			return out[i].Tick < out[j].Tick
		}
		return out[i].Sequence < out[j].Sequence
	})
	return out, nil
}

// Persistence side effects the live loop emits during combat; the replay
// contract under test does not read them back.
func (r *loadShedMemRepo) UpsertCharacterProgression(context.Context, string, store.CharacterProgression) error {
	return nil
}

func (r *loadShedMemRepo) SetCharacterGold(context.Context, string, string, int) error { return nil }

func (r *loadShedMemRepo) TouchSession(context.Context, string) error { return nil }

// admitLoadShedTestMember runs the sim side of attach (no websocket). Like
// Hub.Run, the co-op connection is claimed first, so the member is connected.
func admitLoadShedTestMember(ctx context.Context, loop *sessionLoop, member store.SessionMember) uint64 {
	member.Connected = true
	playerID := loop.playerIDForMember(ctx, member)
	loop.mu.Lock()
	loop.admitMemberLocked(member, playerID)
	loop.mu.Unlock()
	return playerID
}

func drainLoadShedTestClient(client *loopClient) {
	for {
		select {
		case <-client.sendCh:
		default:
			return
		}
	}
}
