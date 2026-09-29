# v479 Spec: Recorded Co-op Member Lifecycle (replay determinism)

Status: Complete
Date: 2026-09-29
Codename: `recorded-member-lifecycle`
Baseline: v476 `recorded-load-shed` (commit `b164365c`); landed on top of v478 `shared-member-setup`

## Problem

The live runner changes co-op membership between ticks, but replay only reproduces the first join:

- `sessionLoop.detach` calls `Sim.RemovePlayerEntity` for co-op sessions, and persists only
  `session_members.left_tick`. That column holds just the *latest* leave.
- `admitMemberLocked` may call `Sim.RespawnPlayerInTown` + `SetPlayerConnected` on reconnect.
  Nothing records it.
- `replay.Reconstruct` adds late members before `joined_tick`, then applies the *final*
  connectivity once at the end (`applyCurrentMemberConnectivity`). It never removes a departed
  guest mid-session, and never replays a reconnect respawn.

Any session where a guest leaves while monsters are active diverges: in replay the monsters keep
targeting a ghost player that live had already removed. Resumed sessions compound it, because the
resume-time `Reconstruct` removes the departed guest at the resume tick, and nothing records that
either.

A second ordering hazard: player entities block town spawn cells (`spawnPositionBlocked`). So a join
(`AddGuestPlayer`) and a reconnect respawn that land in the same inter-tick gap depend on their
order. `joined_tick` records the tick of a join, but not its order relative to other membership
changes in that gap.

## Decision

Record every membership change as a **server-authored stored input** (the v476 pattern). Each row
is stamped at the exact tick and sequence the live runner applied it, and has no actor.

| Type | Live trigger | Mutation |
|------|--------------|----------|
| `system_member_join` | `playerIDForMember` adds a late guest (`AddGuestPlayer`) | replay adds the pending member from its start snapshot |
| `system_member_leave` | co-op `detach` | `Sim.ApplyMemberLeave` (`RemovePlayerEntity`) |
| `system_member_rejoin` | `admitMemberLocked` when admission would change sim state | `Sim.ApplyMemberRejoin(playerID, respawn)` |

- **Timing: post-tick, like `system_load_shed`.** Changes happen between ticks, so a row is stamped
  with the tick that just finished, `CurrentTick()-1`. A change made before the first tick is
  stamped `-1`. Replay applies a tick-T row **after** tick T. (The first draft stamped rows at
  `CurrentTick()` and applied them before that tick. That was rejected: the final leave before the
  loop stops would name a tick live never ran. Replay, and every resume through `Reconstruct`,
  would then simulate a phantom tick and resume one tick ahead of live.) Load-shed rows for the same
  tick apply inside `TickResults`, before membership rows. Load shed only mutates the overload
  window and throttle flags, so it commutes with membership changes.
- **Ordering.** Within one gap, replay applies membership rows in `sequence` order, the order in
  which the runner allocated them under `l.mu`.
- **One mutation path.** Live and replay call the same `Sim.ApplyMemberLeave` /
  `Sim.ApplyMemberRejoin`. The respawn decision (which reads DB member state) is made only in live
  and recorded as `respawn: true|false`, so replay never re-derives it.
- **Rejoin is recorded only when it changes state:** when a respawn is needed or the player is
  disconnected. The v476 first-join case (member already connected, in town) records nothing.
- **Stored-only.** The new types are not in `IsClientIntent`, client `Decode` rejects them, and
  only `DecodeStored` accepts them. Replay drops any membership row that carries an actor, and it
  strips membership rows before `TickResults`, so they never reach player-input processing.
- **Join rows own the join point.** A member that has a `system_member_join` row is added at that
  row, whatever its `joined_tick`. Members with no join row (pre-v479 sessions, and members present
  when the session was built) keep the `joined_tick` path.
- **Final connectivity becomes a legacy fallback.** `applyCurrentMemberConnectivity` still runs,
  but only for members that have no recorded lifecycle rows. Recorded rows are authoritative.
- `left_tick` keeps being written (it feeds session listings), but replay no longer depends on it.

## Non-goals

- Sessions recorded before v479 keep their old replay behavior. Their leaves were never recorded
  and cannot be recovered.
- Live-vs-replay guest setup parity (resource wallet, resource bag, corpses) is v478's
  `sessionsetup`. Recorded joins reuse it.
- Crash recovery (no `detach` ran, so the member row stays `connected=true`) is unchanged.
- No protocol schema change. The types exist only in persisted `session_inputs.payload`.

## Acceptance

- A live `sessionLoop` (the v476 in-memory repo harness) runs: a guest joins, the fight starts,
  the guest leaves mid-fight while monsters are active, and then (in a second case) rejoins. Both
  cases pass `replay.Verify`.
- When everyone leaves, the loop stops. The rebuilt loop resumes at exactly the tick live stopped
  at, and the full history, including the post-resume rejoins, passes `replay.Verify`.
- Negative controls: the same leave or rejoin applied without a recorded row fails
  `replay.Verify`, which proves the fixture is sensitive.
- Client `Decode` rejects the new types, and `DecodeStored` round-trips them.
- `cd server && go test ./...`, `make lint-determinism`, `make ci` green. `session_loop.go` and
  `replay.go` end at or below their ratchet baselines.
