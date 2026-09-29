package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v479 regression: co-op leave and reconnect mutated the live sim between
// ticks without being recorded, so replay kept a departed guest in play (and
// monsters kept targeting it) and never re-applied the reconnect respawn.
//
// combat_control_lab has one dungeon_mob out of aggro range of the spawn. The
// guest walks into it and is mid-fight when it leaves. The host then walks in,
// so live retargets the host while an unreplayed ghost guest would keep the
// mob's sticky target. The rejoined guest walks back into the fight.
const (
	memberLifecycleWorld      = "combat_control_lab"
	memberLifecycleJoinTick   = 2
	memberLifecycleLeaveTick  = 40
	memberLifecycleRejoinTick = 80
	memberLifecycleLastTick   = 140
)

type memberLifecycleRun struct {
	rejoin bool
	// unrecordedLeave and unrecordedRejoin apply the change straight to the
	// sim, without a row, to prove the fixture is sensitive to it.
	unrecordedLeave  bool
	unrecordedRejoin bool
}

func TestRecordedMemberLeaveReplaysMidFight(t *testing.T) {
	repo := runMemberLifecycleSession(t, memberLifecycleRun{})
	assertMemberLifecycleRows(t, repo, game.SystemMemberJoinInputType, game.SystemMemberLeaveInputType)
	guestID := repo.memberByRole(store.SessionMemberGuest).PlayerEntityID
	hostID := repo.memberByRole(store.SessionMemberHost).PlayerEntityID
	if !repoHasTargetedEvent(repo, guestID, 0, memberLifecycleLeaveTick) {
		t.Fatal("scenario must have the monster fighting the guest before it leaves")
	}
	if repoHasTargetedEvent(repo, guestID, memberLifecycleLeaveTick, memberLifecycleLastTick+1) {
		t.Fatal("live must not target the guest after it left")
	}
	if !repoHasTargetedEvent(repo, hostID, memberLifecycleLeaveTick, memberLifecycleLastTick+1) {
		t.Fatal("scenario must have the monster turn on the host after the guest leaves")
	}
	assertReplayMatches(t, repo)
}

func TestRecordedMemberRejoinReplays(t *testing.T) {
	repo := runMemberLifecycleSession(t, memberLifecycleRun{rejoin: true})
	assertMemberLifecycleRows(t, repo, game.SystemMemberJoinInputType, game.SystemMemberLeaveInputType, game.SystemMemberRejoinInputType)
	assertReplayMatches(t, repo)
}

// When everyone leaves the loop stops, and the next attach rebuilds it through
// replay.Reconstruct. The rebuilt sim must resume at the exact tick live
// stopped at: the final leave rows must not make replay run a tick live never
// ran. The whole history, including the rejoins after the resume, must still
// verify.
func TestMemberLifecycleSurvivesSessionResume(t *testing.T) {
	ctx := context.Background()
	repo, loop := newMemberLifecycleTestLoop(t)
	host := attachMemberLifecycleClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	var guest *loopClient
	for tick := uint64(0); tick < memberLifecycleLeaveTick; tick++ {
		if tick == memberLifecycleJoinTick {
			guest = attachMemberLifecycleClient(ctx, loop, repo.joinGuest())
		}
		if tick%5 == 0 && guest != nil && tick > memberLifecycleJoinTick {
			sendMemberLifecycleInput(t, loop, guest, tick, "move_intent", walkIntoMob)
		}
		loop.doTick()
	}
	for _, client := range []*loopClient{guest, host} {
		client.once.Do(func() {})
		loop.detach(client)
	}
	stoppedAt := loop.sim.CurrentTick()

	resumed, err := newSessionLoop(ctx, loop.hub, repo.session)
	if err != nil {
		t.Fatalf("resume session loop: %v", err)
	}
	if got := resumed.sim.CurrentTick(); got != stoppedAt {
		t.Fatalf("resumed at tick %d, live stopped at %d", got, stoppedAt)
	}
	resumed.loadShed = noLoadShed
	host = attachMemberLifecycleClient(ctx, resumed, repo.memberByRole(store.SessionMemberHost))
	guest = attachMemberLifecycleClient(ctx, resumed, repo.memberByRole(store.SessionMemberGuest))
	// The lab's town respawn point is walled off from the arena, so this part
	// proves resume continuity. The rejoin-in-combat contract is
	// TestRecordedMemberRejoinReplays.
	for i := 0; i < 20; i++ {
		tick := resumed.sim.CurrentTick()
		if tick%5 == 0 {
			sendMemberLifecycleInput(t, resumed, guest, tick, "move_intent", walkIntoMob)
		}
		resumed.doTick()
	}
	assertMemberLifecycleRows(t, repo,
		game.SystemMemberJoinInputType, game.SystemMemberLeaveInputType, game.SystemMemberLeaveInputType,
		game.SystemMemberRejoinInputType, game.SystemMemberRejoinInputType)
	assertReplayMatches(t, repo)
}

// Without the recorded rows replay keeps a ghost guest, or never respawns the
// rejoined one. The fixture must notice, otherwise the tests above prove
// nothing.
func TestUnrecordedMemberLifecycleBreaksReplay(t *testing.T) {
	for name, run := range map[string]memberLifecycleRun{
		"leave":  {unrecordedLeave: true},
		"rejoin": {rejoin: true, unrecordedRejoin: true},
	} {
		t.Run(name, func(t *testing.T) {
			repo := runMemberLifecycleSession(t, run)
			report, err := replay.Verify(context.Background(), repo, repo.rules, repo.session.ID)
			if err == nil && report.Match {
				t.Fatalf("unrecorded %s still replayed cleanly; fixture no longer exercises it", name)
			}
		})
	}
}

func newMemberLifecycleTestLoop(t *testing.T) (*loadShedMemRepo, *sessionLoop) {
	t.Helper()
	repo, loop := newLoadShedTestLoopFor(t, func(r *loadShedMemRepo) { r.session.WorldID = memberLifecycleWorld })
	loop.loadShed = noLoadShed
	return repo, loop
}

// noLoadShed keeps these fixtures about membership; load shedding is v476's.
func noLoadShed(loadShedSample) game.LoadShedDirective { return game.LoadShedDirective{} }

// walkIntoMob heads east from the spawn toward the lab's only monster.
var walkIntoMob = map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 5}

func runMemberLifecycleSession(t *testing.T, run memberLifecycleRun) *loadShedMemRepo {
	t.Helper()
	ctx := context.Background()
	repo, loop := newMemberLifecycleTestLoop(t)
	host := attachMemberLifecycleClient(ctx, loop, repo.memberByRole(store.SessionMemberHost))
	var guest *loopClient
	for tick := uint64(0); tick <= memberLifecycleLastTick; tick++ {
		switch {
		case tick == memberLifecycleJoinTick:
			guest = attachMemberLifecycleClient(ctx, loop, repo.joinGuest())
		case tick == memberLifecycleLeaveTick && run.unrecordedLeave:
			loop.mu.Lock()
			delete(loop.clients, guest.key)
			loop.sim.RemovePlayerEntity(guest.playerID)
			loop.mu.Unlock()
		case tick == memberLifecycleLeaveTick:
			// The test client has no websocket; mark it closed so detach
			// skips conn.Close.
			guest.once.Do(func() {})
			loop.detach(guest)
		case tick == memberLifecycleRejoinTick && run.unrecordedRejoin:
			guest = &loopClient{loop: loop, key: guest.key, member: guest.member, playerID: guest.playerID, sendCh: make(chan outEnvelope, 4096), done: make(chan struct{})}
			loop.mu.Lock()
			loop.clients[guest.key] = guest
			_ = loop.sim.ApplyMemberRejoin(guest.playerID, true)
			loop.mu.Unlock()
		case tick == memberLifecycleRejoinTick && run.rejoin:
			guest = attachMemberLifecycleClient(ctx, loop, repo.memberByRole(store.SessionMemberGuest))
		}
		guestActive := tick < memberLifecycleLeaveTick || (run.rejoin && tick > memberLifecycleRejoinTick)
		if tick%5 == 0 && guest != nil && tick > memberLifecycleJoinTick && guestActive {
			sendMemberLifecycleInput(t, loop, guest, tick, "move_intent", walkIntoMob)
		}
		if tick%5 == 0 && tick > memberLifecycleLeaveTick {
			sendMemberLifecycleInput(t, loop, host, tick, "move_intent", walkIntoMob)
		}
		loop.doTick()
		drainLoadShedTestClient(host)
		if guest != nil {
			drainLoadShedTestClient(guest)
		}
	}
	return repo
}

// attachMemberLifecycleClient mirrors attach without a websocket.
func attachMemberLifecycleClient(ctx context.Context, loop *sessionLoop, member store.SessionMember) *loopClient {
	client := &loopClient{
		loop:   loop,
		key:    memberKey(member),
		member: member,
		sendCh: make(chan outEnvelope, 4096),
		done:   make(chan struct{}),
	}
	client.playerID = admitLoadShedTestMember(ctx, loop, member)
	loop.mu.Lock()
	loop.clients[client.key] = client
	loop.mu.Unlock()
	return client
}

func sendMemberLifecycleInput(t *testing.T, loop *sessionLoop, client *loopClient, tick uint64, typ string, payload map[string]any) {
	t.Helper()
	env, err := json.Marshal(map[string]any{
		"type":       typ,
		"message_id": fmt.Sprintf("msg-%s-%d", client.member.CharacterID, tick),
		"session_id": loop.sess.ID,
		"tick":       tick,
		"payload":    payload,
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	loop.handleClientMessage(client, env)
}

func assertReplayMatches(t *testing.T, repo *loadShedMemRepo) {
	t.Helper()
	report, err := replay.Verify(context.Background(), repo, repo.rules, repo.session.ID)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !report.Match {
		t.Fatalf("replay mismatch: %s", report.Mismatch)
	}
}

func assertMemberLifecycleRows(t *testing.T, repo *loadShedMemRepo, want ...string) {
	t.Helper()
	var got []string
	for _, in := range repo.inputs {
		var env struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(in.Payload, &env) != nil || in.ActorPlayerEntityID != "" {
			continue
		}
		switch env.Type {
		case game.SystemMemberJoinInputType, game.SystemMemberLeaveInputType, game.SystemMemberRejoinInputType:
			got = append(got, env.Type)
		}
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("recorded member lifecycle rows = %v, want %v", got, want)
	}
}

// repoHasTargetedEvent reports a recorded event aimed at playerID in [from, to).
func repoHasTargetedEvent(repo *loadShedMemRepo, playerID string, from, to int64) bool {
	for _, ev := range repo.events {
		if ev.Tick < from || ev.Tick >= to {
			continue
		}
		var payload struct {
			Target string `json:"target_entity_id"`
		}
		if json.Unmarshal(ev.Payload, &payload) == nil && payload.Target == playerID {
			return true
		}
	}
	return false
}

// Disconnect and listing side effects of detach. SetSessionMemberDisconnected
// mirrors the Postgres UPDATE (connected=false, left_tick=tick).
func (r *loadShedMemRepo) SetSessionMemberDisconnected(_ context.Context, _, accountID, characterID string, currentLevel int, tick int64) error {
	return r.updateMember(accountID, characterID, func(m *store.SessionMember) {
		m.Connected = false
		m.CurrentLevel = currentLevel
		left := tick
		m.LeftTick = &left
	})
}

func (r *loadShedMemRepo) EndListedSessionIfNoConnected(context.Context, string) (bool, error) {
	return false, nil
}
