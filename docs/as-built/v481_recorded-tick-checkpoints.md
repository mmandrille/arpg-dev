# v481 As-Built: Recorded Tick Checkpoints

Date: 2026-09-29
Status: Complete
Commit: pending
Baseline: v479 `recorded-member-lifecycle` (`7c020abd`)

## What shipped

- **`system_tick_checkpoint` stored input.** Stored-only, with no actor or payload. A row at tick T
  means "live finished tick T". Replay drops it before `TickResults`, even when it carries an
  actor. It only extends the replayed range, because `StoredInputs` counts it in `maxTick`.
- **Checkpoints are written in three places** (`realtime/loop_lifecycle.go`):
  - **Stop.** `stop()` waits for the tick goroutine to exit, then checkpoints at `CurrentTick()-1`.
  - **Quiet stretch.** After tick T, the loop checkpoints once nothing durable (an input, a system
    row, or persisted events) has covered 50 ticks. `durableThrough` tracks this.
  - **Graceful shutdown.** `Hub.Shutdown` stops every loop. `main.go` calls it after
    `httpServer.Shutdown`, because hijacked websockets outlive that call.
  A checkpoint is only written when it extends coverage.
- **Stopped loops refuse work.** Intents are rejected with `session_stopped` instead of being
  persisted at a tick that never runs. `attach` refuses a stopping loop, and `Hub.Run` retries on
  the successor.
- **Exact handoff.** `detach` marks the loop stopping when the last client leaves.
  `loopForSession` waits for `stopDone`, meaning the final checkpoint is persisted, before
  rebuilding from storage.
- **Extraction.** Loop start/stop/tick and its state moved to `loop_lifecycle.go` as an embedded
  `loopLifecycle` struct. `session_loop.go` baseline went from 795 to 788.

## Proof

- `realtime/tick_checkpoint_replay_test.go` (in `collision_lab`, which emits no events, so the
  single long `move_intent` is followed only by quiet ticks):
  - **Solo stop:** the rebuilt loop resumes on the exact stop tick with the host at the live
    position, and `replay.Verify` matches. Before the fix it failed with `resumed at tick 1, live
    stopped at 40`.
  - **Crash:** the loop is dropped without stop. After a long quiet stretch it rebuilds within 50
    ticks of live. A crash inside the first window still loses ticks, which is the control.
  - **`Hub.Shutdown`:** it stops a running loop, no tick runs after it returns, and the final
    checkpoint names the last tick that ran.
- `TestResumeSnapshotMatchesStateEndpoint` (http suite) failed on the first cut: a quick reconnect
  attached to the dying loop. That regression drove the handoff fix above.
- Mutation checks caught by a failing test:
  - dropping the stop checkpoint,
  - dropping the quiet checkpoint,
  - replay keeping checkpoint rows (`replay/members_test.go`).
- **Not caught:** skipping the wait for the tick goroutine in `stop()`. The race it prevents is a
  tick already picked by `select` but not yet holding `l.mu`, a window of well under a millisecond.
  Any hook inside `doTick` holds `l.mu` and serializes with `stop()` anyway. The wait stays because
  it is the correct fix, but no test proves it.
- Unit tests: decode round-trip and client rejection (`inputdecode/system_test.go`).

## Gaps / follow-ups

- A crash still loses up to 50 quiet ticks. Sessions recorded before v481 keep the truncation.
- Checkpoint rows cost one append per 5 s per session, and only during quiet stretches.
- The dying-loop handoff now waits up to one tick plus one DB append. `attach` retries at most
  three times, then closes the socket.
