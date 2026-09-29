# v479 Spec: Member Join Tick (never-attached guests)

Status: Complete
Date: 2026-09-29
Codename: `member-join-tick`
Baseline: v478 `shared-member-setup` (merge `11831b3f`)

## Problem

`session_members.joined_tick` is the replay contract for when a member's player entity entered
the sim. The HTTP join handler inserts guest rows with `joined_tick = -1`, and
`SetSessionMemberConnected` sets it only while it is still negative, on the first websocket
attach. That leaves `-1` ambiguous:

| Case | Live sim | Stored `joined_tick` | Replay (v478) |
|---|---|---|---|
| Guest HTTP-joins before the loop builds, never attaches | added at tick 0 by `buildSessionSim` | `-1` (`SetSessionMemberPlayer` does not write it) | added at tick 0 (correct by accident) |
| Guest HTTP-joins a running loop, never attaches | **never added** (`playerIDForMember` runs only on attach) | `-1` | **added at tick 0** |

In the second case replay allocates a guest entity that live never did, so every later entity ID
and event diverges (`event count: derived 16, recorded 17` in the regression test). The
`applyCurrentMemberConnectivity` removal afterwards cannot undo the allocation.

## Contract

1. `joined_tick >= 0` means the member has a player entity in the sim, and the value is the tick
   the entity was added. `joined_tick < 0` (`store.SessionMemberNotJoinedTick`) means the member
   was never put into the sim.
2. Every live path that adds a member entity records the tick together with the entity ID:
   `SetSessionMemberPlayer` gains a `joinedTick` argument and writes `joined_tick`. The build
   passes `0`. A late join passes `Sim.PlayerJoinedTick`. `SetSessionMemberConnected` keeps its
   "only while negative" rule, so a reconnect never moves the original join tick.
3. Replay skips members with `joined_tick < 0`, through `sessionsetup.JoinedSim`, the shared
   setup package. It does not resolve their start snapshot either.

This also closes a crash window in the late join: before, `joined_tick` was written only by
`admitMemberLocked`, after the lock was released following `AddGuest`.

## Non-goals

- Old sessions where a guest was built in at tick 0 but never attached still store `-1`, and now
  replay without that guest. A migration cannot tell this case apart from the running-loop case.
  This is recorded as a known gap.
- Leave and rejoin within one sim (`RemovePlayerEntity`) is still not recorded (v476 gap).

## Acceptance

- `realtime/member_join_tick_replay_test.go`: a running-loop HTTP guest that never attaches is
  not in the live sim, keeps `joined_tick < 0`, and `replay.Verify` matches (fails on v478). A
  guest present at build that never attaches is in the live sim with `joined_tick = 0`, and
  replay matches.
- `go test ./...`, `make lint-determinism`, `make ci` green.
