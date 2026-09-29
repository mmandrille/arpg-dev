package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// v481 regression: replay and resume only ran through the last durable row,
// so quiet ticks after it (a long move, monster movement, regen) were lost and
// a resumed session restarted behind live. collision_lab is event-free, so a
// single long move_intent is followed only by quiet ticks.
const tickCheckpointWorld = "collision_lab"

func newTickCheckpointTestLoop(t *testing.T) (*loadShedMemRepo, *sessionLoop, *loopClient) {
	t.Helper()
	repo, loop := newLoadShedTestLoopFor(t, func(r *loadShedMemRepo) {
		r.session.WorldID = tickCheckpointWorld
		r.session.Mode = store.SessionModeSolo
		r.session.Listed = false
	})
	loop.loadShed = noLoadShed
	host := attachMemberLifecycleClient(context.Background(), loop, repo.memberByRole(store.SessionMemberHost))
	sendMemberLifecycleInput(t, loop, host, 0, "move_intent", map[string]any{
		"direction": map[string]any{"x": 0, "y": 1}, "duration_ticks": 60,
	})
	return repo, loop, host
}

func runQuietTicks(loop *sessionLoop, host *loopClient, n int) {
	for i := 0; i < n; i++ {
		loop.doTick()
		drainLoadShedTestClient(host)
	}
}

// A solo stop records nothing else, so without a checkpoint the resume would
// land on tick 1 and undo the whole walk.
func TestSoloStopResumesOnExactTick(t *testing.T) {
	ctx := context.Background()
	repo, loop, host := newTickCheckpointTestLoop(t)
	runQuietTicks(loop, host, 40)
	stoppedAt, livePos := loop.sim.CurrentTick(), hostPosition(t, loop)

	host.once.Do(func() {})
	loop.detach(host)

	resumed, err := newSessionLoop(ctx, loop.hub, repo.session)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if got := resumed.sim.CurrentTick(); got != stoppedAt {
		t.Fatalf("resumed at tick %d, live stopped at %d", got, stoppedAt)
	}
	if got := hostPosition(t, resumed); got != livePos {
		t.Fatalf("resumed host at %+v, live left it at %+v", got, livePos)
	}
	assertReplayMatches(t, repo)
}

// A crash skips the stop checkpoint. Quiet checkpoints bound the loss to
// quietCheckpointTicks; a crash inside the first window still loses it, which
// is the control that shows the fixture is sensitive.
func TestCrashLosesAtMostOneQuietCheckpointInterval(t *testing.T) {
	for name, quiet := range map[string]int{
		"within first interval": quietCheckpointTicks - 10,
		"long quiet stretch":    3*quietCheckpointTicks + 7,
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo, loop, host := newTickCheckpointTestLoop(t)
			runQuietTicks(loop, host, quiet)
			crashedAt := loop.sim.CurrentTick() // the loop is dropped: no detach, no stop

			rebuilt, err := newSessionLoop(ctx, loop.hub, repo.session)
			if err != nil {
				t.Fatalf("rebuild: %v", err)
			}
			got := rebuilt.sim.CurrentTick()
			if got > crashedAt || crashedAt-got > quietCheckpointTicks {
				t.Fatalf("rebuilt at tick %d after crash at %d, want within %d ticks", got, crashedAt, quietCheckpointTicks)
			}
			if quiet < quietCheckpointTicks && got == crashedAt {
				t.Fatal("control: a crash inside the first interval must still lose ticks")
			}
			if quiet > quietCheckpointTicks && countTickCheckpoints(repo) == 0 {
				t.Fatal("a long quiet stretch must record checkpoints")
			}
			assertReplayMatches(t, repo)
		})
	}
}

// Graceful shutdown must stop the tick goroutine first, so the final
// checkpoint names the last tick that really ran.
func TestHubShutdownCheckpointsRunningLoops(t *testing.T) {
	repo, loop, _ := newTickCheckpointTestLoop(t)
	loop.hub.loops = map[string]*sessionLoop{repo.session.ID: loop}
	loop.start()
	deadline := time.Now().Add(5 * time.Second)
	for loop.currentTick() < 5 {
		if time.Now().After(deadline) {
			t.Fatal("tick loop did not advance")
		}
		time.Sleep(20 * time.Millisecond)
	}

	loop.hub.Shutdown()
	stoppedAt := loop.currentTick()
	time.Sleep(3 * time.Second / tickHz)
	if loop.currentTick() != stoppedAt {
		t.Fatal("a tick ran after Shutdown returned")
	}
	if len(loop.hub.loops) != 0 {
		t.Fatalf("hub still tracks %d loops after Shutdown", len(loop.hub.loops))
	}
	if got := lastTickCheckpoint(repo); got != int64(stoppedAt)-1 {
		t.Fatalf("final checkpoint at tick %d, want %d (last tick that ran)", got, int64(stoppedAt)-1)
	}
	assertReplayMatches(t, repo)
}

func hostPosition(t *testing.T, loop *sessionLoop) game.Vec2 {
	t.Helper()
	loop.mu.Lock()
	defer loop.mu.Unlock()
	for _, e := range loop.sim.Snapshot().Entities {
		if e.Type == "player" && e.CharacterID == loop.sess.CharacterID {
			return e.Position
		}
	}
	t.Fatal("host entity not found")
	return game.Vec2{}
}

func tickCheckpointTicks(repo *loadShedMemRepo) []int64 {
	var ticks []int64
	for _, in := range repo.inputs {
		var env struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(in.Payload, &env) == nil && env.Type == game.SystemTickCheckpointInputType && in.ActorPlayerEntityID == "" {
			ticks = append(ticks, in.Tick)
		}
	}
	return ticks
}

func countTickCheckpoints(repo *loadShedMemRepo) int { return len(tickCheckpointTicks(repo)) }

func lastTickCheckpoint(repo *loadShedMemRepo) int64 {
	ticks := tickCheckpointTicks(repo)
	if len(ticks) == 0 {
		return -1
	}
	return ticks[len(ticks)-1]
}
