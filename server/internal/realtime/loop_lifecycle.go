package realtime

import (
	"sync"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// quietCheckpointTicks bounds how many quiet ticks a crash can lose (v481).
// Replay only runs through the last durable row, so after this many ticks with
// no input, system row, or event, the loop records a tick checkpoint. This is
// a persistence cadence, not gameplay tuning, so it stays next to the runner.
const quietCheckpointTicks = 50

// loopLifecycle is the start/stop and durable-coverage state of a session
// loop. The bool and int fields are guarded by sessionLoop.mu.
type loopLifecycle struct {
	done       chan struct{} // closed by stop: the tick goroutine exits
	tickExited chan struct{} // closed when the tick goroutine has exited
	stopDone   chan struct{} // closed once the final checkpoint is persisted
	closeOnce  sync.Once
	started    bool
	stopped    bool // refuses intents and attaches
	// durableThrough is the last tick covered by a durable row or events.
	durableThrough int64
}

func newLoopLifecycle(sim *game.Sim) loopLifecycle {
	return loopLifecycle{
		done:       make(chan struct{}),
		tickExited: make(chan struct{}),
		stopDone:   make(chan struct{}),
		// A rebuilt loop's Reconstruct already replayed through here.
		durableThrough: int64(sim.CurrentTick()) - 1,
	}
}

func (l *sessionLoop) start() {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return
	}
	l.started = true
	l.mu.Unlock()
	go l.tickLoop()
}

// stop ends the loop and records the last tick it ran. It waits for the tick
// goroutine to exit first, so no tick can run after the checkpoint. Intents
// that arrive afterwards are rejected instead of being buffered for a tick
// that will never run. Must not be called from the tick goroutine.
func (l *sessionLoop) stop() {
	l.closeOnce.Do(func() {
		close(l.done)
		l.mu.Lock()
		l.stopped = true
		started := l.started
		l.mu.Unlock()
		if started {
			<-l.tickExited
		}
		l.mu.Lock()
		rec := l.checkpointLocked(int64(l.sim.CurrentTick()) - 1)
		l.mu.Unlock()
		l.persistSystemInput(rec)
		close(l.stopDone)
	})
}

// stopping reports whether the loop refuses new work: its last client left or
// stop was called.
func (l *sessionLoop) stopping() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stopped
}

func (l *sessionLoop) tickLoop() {
	defer close(l.tickExited)
	ticker := time.NewTicker(time.Second / tickHz)
	defer ticker.Stop()
	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			l.doTick()
		}
	}
}

// Shutdown stops every session loop, recording each one's final tick, so a
// graceful restart resumes every session on the tick it stopped at.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	loops := make([]*sessionLoop, 0, len(h.loops))
	for _, loop := range h.loops {
		loops = append(loops, loop)
	}
	h.loops = map[string]*sessionLoop{}
	h.mu.Unlock()
	for _, loop := range loops {
		loop.stop()
	}
}

// noteDurableLocked records that tick is covered by a durable input row,
// system row, or persisted events. Caller holds l.mu.
func (l *sessionLoop) noteDurableLocked(tick int64) {
	if tick > l.durableThrough {
		l.durableThrough = tick
	}
}

// quietCheckpointLocked records a checkpoint after tick when nothing durable
// covered the last quietCheckpointTicks ticks. Caller holds l.mu.
func (l *sessionLoop) quietCheckpointLocked(tick uint64) *store.SessionInput {
	if int64(tick)-l.durableThrough < quietCheckpointTicks {
		return nil
	}
	return l.checkpointLocked(int64(tick))
}

// checkpointLocked returns a system_tick_checkpoint row for tick when it
// extends durable coverage, and nil otherwise. Caller holds l.mu and persists
// the row after unlocking.
func (l *sessionLoop) checkpointLocked(tick int64) *store.SessionInput {
	if tick < 0 || tick <= l.durableThrough {
		return nil
	}
	messageID := ids.New("sys")
	payload, err := inputdecode.EncodeStoredTickCheckpoint(messageID)
	if err != nil {
		l.log.Error("encode tick checkpoint", "tick", tick, "error", err)
		return nil
	}
	return l.systemInputRowLocked(tick, messageID, payload)
}
