# v476 As-Built: Recorded Load Shedding

Date: 2026-09-29
Status: Complete
Commit: pending

## What shipped

- **`system_load_shed` stored input.** After each tick the live runner turns its wall-clock
  guardrail decision into a `game.LoadShedDirective`. It applies the directive with
  `Sim.ApplyLoadShed` and persists a `session_inputs` row at that tick with no actor. The row is
  written only when `Sim.LoadShedChanges` says state would change: typically once when shedding
  starts and once when the throttle clears.
- **Replay parity.** `TickResultsProfiled` strips load-shed inputs before player processing and
  applies them after `finalizeResults`, the same point where the live runner applies them.
  `ApplyOverloadDegradation`/`SetCombatMovementThrottle` are now unexported, so an unrecorded
  runtime mutation no longer compiles.
- **Client-proof.** `inputdecode.DecodeStored` accepts the type through `inputdecode/system.go`,
  and `IsClientIntent`/`Decode` reject it. The sim drops any load-shed input that has an actor.
- **Exact `joined_tick`.** `AddGuestPlayer` stamps `playerState.JoinedTick`, and
  `admitMemberLocked` persists `Sim.PlayerJoinedTick` instead of sampling `CurrentTick()` after
  the lock and a DB write.
- **Diagnosable 500s.** The state, replay, and timeline inspect endpoints log
  `session_replay_failed` with the correlation ID, op, session ID, and error.
- **Extractions.** The `Input` struct moved to `game/input.go` (`sim.go` baseline 6782 → 6731).
  `attach` was split into connection wiring plus `admitMemberLocked`, and admission moved to
  `realtime/session_admission.go` (`session_loop.go` baseline 951 → 887).

## Root cause (proved against the benchmark sessions left in local Postgres)

The co-op join was the symptom. `doTick` called `ApplyOverloadDegradation` and
`SetCombatMovementThrottle` based on wall-clock tick and combat-phase durations, and recorded
nothing. In `sess_01M3NQ9T18WH57E8AB1EQ4SMBG` the server log shows over-budget ticks 3 and 47 with
`degradation_applied:true` (tick 3: `sim_ms` 5 of `total_ms` 304, a host hiccup from Godot running
alongside). Replay first diverged at tick 26. Injecting degradation after ticks 3 and 47 in a
scratch harness reproduced all 173 recorded events, including the tick-49 guest entity ID. Both
failing benchmark sessions were off by exactly 3 IDs, which matches the loot drops that went
missing on replay.

## Tests

- `realtime/load_shed_replay_test.go`
  - `TestRecordedLoadShedKeepsLateJoinReplayDeterministic` drives a real `sessionLoop`
    (benchmark world and seed, level-12 sorcerer host) through `handleClientMessage` → `doTick`.
    The policy stub forces degradation at tick 3, lightning kills drop loot at tick 2, and a guest
    joins at tick 49 via `playerIDForMember` + `admitMemberLocked`. `replay.Verify` matches.
  - `TestUnrecordedLoadShedBreaksReplay` is the negative control: the same run with degradation
    applied but not recorded fails `Verify`, so the fixture really exercises load shedding.
  - `TestLateJoinRecordsTickGuestEntityWasAdded` runs two ticks between `AddGuestPlayer` and
    admission and checks that `joined_tick` is still the add tick and `Verify` matches.
- `game/load_shed_test.go` covers stripping of actor-bearing directives and post-tick application
  timing. `inputdecode/system_test.go` covers the round trip and client rejection.
- Mutation checks: disabling recording fails the positive test, and reverting `joined_tick` to
  `CurrentTick()` fails the join test (12 vs 10).

## Gaps found (not fixed here)

- **Live/replay member setup parity.** Live (`buildSessionSim`, `playerIDForMember`) loads the
  resource wallet, resource bag, and recoverable corpses for the host and guests. Replay's
  `sessionStartSim`/`addMemberToSim` loads none of them. Corpse entities `alloc()` IDs when their
  death level is generated, so a character with a recoverable corpse diverges replay.
- **Co-op leave/reconnect is not replayed.** `detach` calls `RemovePlayerEntity` and reconnects
  can `RespawnPlayerInTown`. Neither is recorded, so replay keeps departed guests in the world.
- **`DefaultPlayerID()` drifts to the last-added guest** when `buildSessionSim` adds pre-existing
  guest rows (no-input rebuild path). It is harmless for actor-tagged inputs but fragile.
- Sessions recorded before v476 still fail replay if they ever shed load; that data is
  unrecoverable.
