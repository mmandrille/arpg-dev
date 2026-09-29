# v476 Spec: Recorded Load Shedding (replay determinism)

Status: Complete
Date: 2026-09-29
Codename: `recorded-load-shed`
Baseline: v475 `kit-heroes` (commit `a76953f7`)

## Problem

`make benchmark` (scenario `sorcerer_multigroup_perf_probe`) leaves listed co-op sessions that
cannot be reconstructed: `GET /v0/sessions/{id}/state` returns 500 and `make replay` reports

```
member acct_…/char_… player_entity_id=1737871872194362512 reconstructed=1737871872194362509
```

The failure reproduces at v469 and at HEAD. The brief assumed the co-op member join was applied at
a tick replay does not reproduce.

## Root cause

The join is where the bug surfaces, not what causes it. `sessionLoop.doTick` fed **wall-clock**
timings back into the deterministic sim after every tick, and never recorded them:

- `sim.ApplyOverloadDegradation()` when the tick overran its 100 ms budget, and
- `sim.SetCombatMovementThrottle(combatPhaseOverBudget(profiler, …))`.

Both switch monster movement to LOD skipping. Replay has no clock, so it never degraded. Monster
positions drifted, a different attack landed, loot drops differed, and entity IDs shifted. A late
co-op join is simply the first place anything asserts on an entity ID. Single-player sessions that
hit one slow tick diverged too, but silently.

Evidence (session `sess_01M3NQ9T18WH57E8AB1EQ4SMBG`, same failure):

- Server log: ticks 3 and 47 logged `session_tick_budget_overrun` with `degradation_applied:true`.
  Tick 3 had `sim_ms` 5 ms of a 304 ms total, so the trigger was a host hiccup, not sim load.
- The first replay divergence is at tick 26 (a missing `monster_attack_windup`), long before the
  tick-49 join.
- A scratch harness that applied degradation after ticks 3 and 47 reproduced all 173 recorded
  events, and the guest entity ID matched.

Secondary bug (latent, not the cause here): `attach` persisted `joined_tick` as `CurrentTick()`
*after* releasing the lock and writing to the DB, not at the moment `AddGuestPlayer` ran. Any tick
that runs in that window records a later join than actually happened.

## Decision

Record load shedding as a **server-authored stored input**, `system_load_shed`.

- After tick T, the runner asks a `loadShedPolicy` (production: the existing wall-clock rules) for a
  `LoadShedDirective{OverloadDegrade, CombatMovementThrottle}`. It applies the directive with
  `Sim.ApplyLoadShed` and persists a `session_inputs` row at **tick T** with no actor, but only when
  `Sim.LoadShedChanges` reports that it would mutate state.
- Replay passes the row in with tick T's inputs. `TickResultsProfiled` strips load-shed inputs
  before player processing and applies them **after** `finalizeResults`, the same point where the
  live runner applies them. Semantics are unchanged: the window starts at T+1, and throttle is
  `degradation applied || combat phase over budget`.
- `ApplyOverloadDegradation` / `SetCombatMovementThrottle` become unexported. `ApplyLoadShed` is the
  only runtime entry point, so the compiler blocks any new unrecorded mutation.
- `system_load_shed` is stored-only. It is not in `IsClientIntent` and client `Decode` rejects it.
  Only `DecodeStored` accepts it, and the sim ignores any directive that carries an actor.
- `joined_tick` comes from the sim: `playerState.JoinedTick` is set in `AddGuestPlayer`, and
  `admitMemberLocked` persists `Sim.PlayerJoinedTick`.
- `http/inspect.go`: the state, replay, and timeline endpoints now log the reconstruct error
  (`session_replay_failed`, with correlation ID and session ID). Response bodies stay generic.

### Why recording instead of the alternatives (owner decision, 2026-09-29)

- **Deterministic sim signals only** (entity or path-request counters, no clock) would remove the
  input but stop reacting to real host load. That defeats the v271/v384 overload guardrail.
- **Removing adaptive throttling** is the simplest option, but it drops the safety net.
- Recording is the standard lockstep rule: every nondeterministic external signal that affects the
  sim becomes an input. ADR-0016 already requires budget deferral to preserve replay determinism.

### Why a tick-T row applied post-tick instead of a tick-T+1 row

A T+1 row that is recorded just before the session stops would make replay simulate a tick the live
server never ran, producing a false event-count mismatch. A tick-T row never dangles.

## Non-goals

- Old sessions recorded before v476 still fail replay. Their degradation was never recorded and
  cannot be recovered.
- Replaying co-op leave/reconnect: `RemovePlayerEntity` on detach and `RespawnPlayerInTown` on
  reconnect are still not recorded. See the as-built gaps.
- Live-vs-replay guest setup parity: the live path loads the resource wallet, resource bag, and
  corpses, and replay's `addMemberToSim` does not. See the as-built gaps.
- No protocol schema change. The wire protocol is untouched; the new type exists only in persisted
  `session_inputs.payload`.

## Acceptance

- A live `sessionLoop` driven through `handleClientMessage` → `doTick`, with degradation forced at
  tick 3, loot dropped before a guest joins at tick 49, and the guest admitted through
  `playerIDForMember` + `admitMemberLocked`, passes `replay.Verify`.
- The same run with the degradation applied but **not** recorded fails `replay.Verify`, which
  proves the fixture is sensitive.
- `joined_tick` equals the tick the guest entity was added even when ticks run before admission.
- `cd server && go test ./...`, `make lint-determinism`, `make ci` green.
