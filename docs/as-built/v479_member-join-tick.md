# v479 As-Built: Member Join Tick

Date: 2026-09-29
Status: Complete
Spec: [`v479_spec-member-join-tick.md`](../specs/v479_spec-member-join-tick.md) ·
Plan: [`v479_2026-09-29-member-join-tick.md`](../plans/v479_2026-09-29-member-join-tick.md)

## What shipped

- **`joined_tick < 0` now only means "never in the sim".** The constant is
  `store.SessionMemberNotJoinedTick` (-1), and the HTTP join handler inserts guests with it.
- **`SetSessionMemberPlayer(..., joinedTick)`** writes `joined_tick` together with
  `player_entity_id`. Every live path that adds a member entity goes through it:
  - `buildSessionSim` passes `0` for the host and for every guest row it finds, including HTTP
    guests that have not attached yet.
  - `playerIDForMember` (late join) passes `Sim.PlayerJoinedTick`, read under the same lock as
    `AddGuest`. This also closes the crash window where the entity existed but `joined_tick` was
    written only later by `admitMemberLocked`.
  - `SetSessionMemberConnected` keeps its "only while negative" rule, so a reconnect never moves
    the join tick.
- **Replay skips never-joined members.** `sessionsetup.JoinedSim(member)` is false for a negative
  `joined_tick`. `replay.sessionStartSim` skips those members before resolving their snapshot, so
  replay no longer allocates a guest entity that live never did.

## Proof

- `realtime/member_join_tick_replay_test.go`, a live `sessionLoop` on the v478 harness:
  - `TestHTTPJoinedGuestNeverAttachedReplayMatchesLive`: the host descends, then
    `repo.joinGuest()` runs with no attach. The guest is not in the live sim, its `joined_tick`
    stays negative, and `replay.Verify` matches. On v478 this failed with
    `event count: derived 16, recorded 17`.
  - `TestGuestJoinedBeforeBuildNeverAttachedReplayMatchesLive`: the guest row exists before the
    build. The guest is in the live sim with `joined_tick = 0`, and replay matches. On v478 this
    failed on `joined_tick = -1`.
- `store/session_member_join_tick_test.go` (Postgres): `SetSessionMemberPlayer` writes the add
  tick, and a later first `SetSessionMemberConnected` does not move it.
- `assertReplayMatchesLive` compares end-state entity IDs only for members whose row is connected,
  and fails if it compared none. `Reconstruct` is also the resume path and removes disconnected
  co-op members (`applyCurrentMemberConnectivity`), so a never-attached build-time guest has no
  replayed counterpart by design. `replay.Verify` still covers that guest's effect on the event
  stream.
- `go test ./...`, `make lint-determinism`, and `make ci` (isolated port `:18232`) are green.

## Known gaps (not fixed here)

- **Legacy sessions.** Sessions from before v479 where a guest was built in at tick 0 but never
  attached still store `joined_tick = -1`. They now replay without that guest, although live had
  it. The data cannot tell them apart from the running-loop case, so there is no migration.
- **A live guest that never attached keeps its entity.** A build-time guest that never attaches
  stays in the live sim (`PlayerConnected` true) until the session ends. This predates v479 and
  is left unchanged.
- Leave/rejoin within one sim (`RemovePlayerEntity`) is still not recorded (v476 gap).
