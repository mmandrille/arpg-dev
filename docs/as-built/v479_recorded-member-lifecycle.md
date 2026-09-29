# v479 As-Built: Recorded Co-op Member Lifecycle

Date: 2026-09-29
Status: Complete
Commit: pending
Baseline: v478 `shared-member-setup` (`abf3fa2a`). Rebased onto it: the replay roster builds members
through `internal/sessionsetup`

## What shipped

- **Three stored-only system inputs:**
  - `system_member_join`: `playerIDForMember` added a late guest.
  - `system_member_leave`: co-op `detach`.
  - `system_member_rejoin`: `admitMemberLocked`, with the live `respawn` decision recorded verbatim.
  Rows have no actor and carry `player_entity_id` (plus the account and character on joins).
  `IsClientIntent` and client `Decode` reject them, and only `DecodeStored` accepts them.
- **One mutation path.** Live and replay both call `Sim.ApplyMemberLeave` /
  `Sim.ApplyMemberRejoin` (`game/member_lifecycle.go`). A rejoin row is written only when admission
  would change the sim (a respawn, or a disconnected player). An ordinary first join records just
  the join row.
- **Timing matches `system_load_shed`.** A row is stamped with the tick that just finished,
  `CurrentTick()-1`, or `-1` before the first tick. Replay applies it after that tick, in
  `sequence` order.
- **Replay stepper.** `replay.replayTicks` replaces three copies of the tick loop: plain
  reconstruct, reconstruct with pending members, and `BuildTimeline`. It strips membership rows
  before `TickResults` and applies them after it.
  `replay/members.go` owns the roster: start members, legacy `joined_tick` joins, recorded joins,
  and leave/rejoin. A member with a join row joins only at that row.
  `applyCurrentMemberConnectivity` is now a legacy fallback, and it skips players that recorded
  rows govern.
- **Shared runner helpers.** `systemInputRowLocked` / `persistSystemInput`
  (`realtime/member_lifecycle.go`) replace the load-shed-specific row and persist code.

## Why join is recorded too

Player entities block town spawn cells (`spawnPositionBlocked`). A late join and a reconnect
respawn in the same inter-tick gap therefore depend on their order. `joined_tick` records the tick,
but not the order within the gap.

## Why post-tick stamping

The first cut stamped rows at `CurrentTick()` and applied them before that tick. When the last
member leaves, the loop stops with tick T unrun. A tick-T leave row made replay, and every resume
through `Reconstruct`, simulate tick T anyway, so the session resumed one tick ahead of live.
`TestMemberLifecycleSurvivesSessionResume` pins this: with the pre-tick stamp it fails with
`resumed at tick 41, live stopped at 40`.

## Proof

- `realtime/member_lifecycle_replay_test.go` drives a live `sessionLoop` on the v476 in-memory repo
  harness in `combat_control_lab`. The lab has one out-of-range `dungeon_mob`.
  - **Leave mid-fight:** the guest walks into the mob and is attacked, then leaves at tick 40. The
    host walks in, and live retargets the host. `replay.Verify` matches.
  - **Rejoin:** same run, then a rejoin at tick 80 respawns the guest, which walks back.
    `replay.Verify` matches.
  - **Resume:** both players leave and the loop stops. `newSessionLoop` rebuilds through
    `Reconstruct` onto the exact stop tick, both players rejoin, and `replay.Verify` matches the
    whole history.
  - **Negative controls:** applying the leave or the rejoin without a row fails `Verify`. The
    unrecorded leave reproduces the reported bug: the ghost guest keeps the mob's sticky target, so
    replay derives 13 events against 11 recorded.
- Mutation checks, each caught by a failing test:
  - dropping the replay leave,
  - replaying a rejoin without its respawn,
  - stamping rows at `CurrentTick()`.
- Unit tests: `inputdecode/system_test.go` (round-trip, client rejection, and a missing player
  entity) and `game/member_lifecycle_test.go` (apply semantics).
- Ratchet: `replay.go` 765 → 525 lines, baseline entry dropped. `session_loop.go` stays at its 887
  baseline.

## Gaps / follow-ups

- **Quiet trailing ticks.** Replay and resume simulate only through the highest recorded
  input/event tick. A live session that ran quiet ticks after its last row, then stopped, resumes
  at a lower tick than live reached. v479 does not fix this: it only makes the final leave rows name
  the exact stop tick, which covers the common all-players-left stop. A general fix needs a
  persisted "last simulated tick".
- Guest setup parity (wallet, bag, corpses) was closed by v478 `sessionsetup`. v479 reuses it for
  recorded joins.
- Sessions recorded before v479 keep the final-connectivity approximation.
- `combat_control_lab`'s town respawn point is walled off from the arena, so the resume test proves
  continuity but not post-resume combat. Post-rejoin combat is covered without a resume.
- Crash recovery (no `detach`, so the row stays `connected=true`) is unchanged.
