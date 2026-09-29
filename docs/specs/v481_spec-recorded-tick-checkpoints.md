# v481 Spec: Recorded Tick Checkpoints (quiet trailing ticks)

Status: Complete
Date: 2026-09-29
Codename: `recorded-tick-checkpoints`
Baseline: v479 `recorded-member-lifecycle` (commit `7c020abd`)

## Problem

Replay and resume simulate only through the highest tick that has a durable row:
`maxTick = max(session_inputs.tick, session_events.tick)`. Live ticks after the last row are
**quiet**: they emit no events, but they still change state (monster movement, regen, projectiles
in flight, a long `move_intent`, cooldowns). They are never replayed.

Consequences:

- **Resume rewinds.** `newSessionLoop` rebuilds through `replay.Reconstruct`, so a resumed session
  restarts at `maxTick+1` instead of the tick live reached. Players see monsters and themselves
  snap back. Client intents that carry the live tick land in the loop buffer as far-future ticks
  and stall for the missing stretch (`handleClientMessage` only clamps ticks that are in the past).
- **Inspect lies.** `/state` for a stopped session shows state from an earlier tick.
- **When it happens:**
  - Solo stop. A solo `detach` records nothing. v479 only covered the co-op "everyone left" stop.
  - Server shutdown. `main.go` calls only `httpServer.Shutdown`. That does not touch hijacked
    websockets or session loops, so a graceful restart or deploy behaves like a crash for every
    live session.
  - Crash. Everything after the last row is lost.
- **A v479 race.** `detach` unlocks before `stop()`. A tick can therefore run after the final leave
  row, and replay will never reproduce it.

`replay.Verify` stays self-consistent, because a resumed loop is itself a `Reconstruct`. That is
why no test caught this. The divergence is between what players saw live and what the session
resumes into.

## Decision

Record a server-authored, stored-only `system_tick_checkpoint` input. A row at tick T means "live
finished tick T". It has no actor and no payload. Replay strips it before `TickResults`. Its only
effect is that `StoredInputs` counts it in `maxTick`, so replay and resume run through T.

Checkpoints are written in three places:

1. **Loop stop (exact).** `stop()` closes `done` and then waits for the tick goroutine to exit, so
   no tick can run afterwards. It then records a checkpoint at `CurrentTick()-1`. This also closes
   the v479 race.
2. **Quiet stretches (bounded crash loss).** After tick T, if nothing durable covers the last
   `quietCheckpointTicks` ticks, record a checkpoint at T. "Durable" means any input or system row,
   or any tick that persisted events. A crash then loses at most `quietCheckpointTicks` ticks. The
   interval is 50 ticks (5 s).
3. **Graceful shutdown (exact).** `Hub.Shutdown` stops every loop, and each one records its stop
   checkpoint. `main.go` calls it after `httpServer.Shutdown` returns.

Two lifecycle rules make the stop checkpoint trustworthy:

- **A stopped loop takes no new work.** Intents that arrive after stop are rejected with
  `session_stopped`. Before this, they were buffered and persisted at a tick that never runs, a
  phantom tick for replay. `attach` refuses a stopping loop.
- **Handoff waits for the final tick.** `detach` marks the loop stopping as soon as its last client
  leaves. `loopForSession` waits for a stopping loop's final checkpoint (`stopDone`) before
  building its successor from storage. `Hub.Run` retries `attach` on the successor. Before this,
  a reconnect in that window attached to the dying loop. The window was tiny on v479; waiting for
  the tick goroutine widened it, and `TestResumeSnapshotMatchesStateEndpoint` caught it.

A checkpoint is written only when it extends coverage: its tick must be greater than the loop's
`durableThrough`. A loop rebuilt by `Reconstruct` starts with `durableThrough = CurrentTick()-1`,
so stopping a loop that never ticked writes nothing.

### Why a stored input instead of a `sessions.last_tick` column

- **No migration and no new write path.** Replay already derives `maxTick` from inputs.
- **Crash safety needs periodic writes either way.** An `UPDATE` per tick costs more than an
  append every 5 s during quiet stretches only.
- **One story.** Every fact about what live simulated lives in the ordered input stream, next to
  load shed and membership rows.

### Why the interval is a code constant

This is a persistence cadence, not gameplay tuning. It does not change outcomes; it only bounds
how much a crash loses. Keeping it next to the runner is the right ownership boundary, per the
Data-Driven Configuration Policy's documented exception.

## Non-goals

- Sessions recorded before v481 keep the old truncation.
- A crash still loses up to `quietCheckpointTicks` ticks.
- Client-side tick re-sync after resume: once the resume tick is exact, client ticks line up again.

## Acceptance

- Solo live loop: a long `move_intent` keeps running quiet ticks after its last row. When the loop
  stops, the rebuilt loop resumes at exactly the stop tick, the player position matches live, and
  `replay.Verify` matches.
- A quiet stretch longer than `quietCheckpointTicks` records a checkpoint, so a simulated crash
  (loop dropped without stop) rebuilds no more than `quietCheckpointTicks` ticks behind.
- Negative control: the same solo stop without checkpoints resumes behind live.
- `Hub.Shutdown` stops loops and records their checkpoints.
- Client `Decode` rejects the type, `DecodeStored` accepts it, and replay drops a checkpoint that
  carries an actor.
- `cd server && go test ./...`, `make lint-determinism`, `make ci` green, and the ratchet holds.
