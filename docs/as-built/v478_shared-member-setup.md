# v478 As-Built: Shared Member Setup

Date: 2026-09-29
Status: Complete
Commit: pending

## What shipped

- **`internal/sessionsetup`** is now the only code that puts a session member into a sim:
  - `Members`: list members, synthesize the legacy host, sort host-first.
  - `Resolve`: all I/O, meaning the start snapshot plus the `characters.character_class` override.
  - `NewHostSim`: start state → mercenary roster → restore hired companion → corpses.
  - `AddGuest`: start state → corpses. It does no I/O, so it runs under the live sim lock.

  `realtime.buildSessionSim`, `realtime.playerIDForMember`, and `replay.sessionStartSim` /
  `addMemberToSim` all call it. The store → sim converters live in `sessionsetup/convert.go`.
  Both duplicate copies are gone: replay's had silently dropped `WeaponSet`.
- **Corpse session-start snapshot.** Migration `0031_session_start_character_corpses.sql` adds
  `SessionStartSnapshot.Corpses`. HTTP session create/join captures
  `ListRecoverableCharacterCorpses` and inserts it in the snapshot transaction, with `ordinal`
  preserving load order. `LoadSessionStartSnapshotForMember` returns it. Neither live nor replay
  reads the live corpse table during a session. The table is included in delete-character and
  stale-session cleanup.
- **`CreateSessionStartSnapshot(ctx, SessionStartSnapshot)`** replaces the 14-positional-argument
  signature. It moved from `repos.go` to `store/session_start_snapshot.go`, together with the corpse
  insert/load helpers.
- **Behavior aligned to one definition:**
  - Guests added at session build are named `Guest <suffix>`, as late-join and replay already did.
  - Late-join guests get the class override.
  - Legacy member-less replay sets host metadata like live does.

- **Gameplay-debug ordering is part of setup.** `NewHostSim` takes a `configure` hook that runs
  right after construction and before any member state loads. Live applies `SetGameplayDebug`
  there, at the same point it always did.
  - Found by `make ci`: the first v478 cut called `SetGameplayDebug` *after* setup.
  - On a debug server, `NewSim` has already seeded the unique test chests from the env flag. The
    per-player loads (`usePlayer`) then swap that chest state out, so a late `SetGameplayDebug`
    re-seeds and allocates about 35 IDs that replay never does.
  - Result: `account_stash_storage`, `true_coop_session`, `town_vendor_gold_sink`, and
    `mystery_seller_core` failed their `/state` replay checks.
  - `TestGameplayDebugLiveBuildReplayMatches` pins this. With the call moved back after setup it
    fails with `player_entity_id=1091 reconstructed=1056`.

## Proof

- `realtime/member_setup_replay_test.go` drives a live `sessionLoop` on `dungeon_levels`. The
  member has a level -1 corpse and a resource bag item (`renew_stone`, ID `880001`). The host
  walks to the town stairs with a recorded `move_to_intent`, then sends `descend_intent`, then
  walks at the nearest monster. The test then requires `replay.Verify` to match and requires
  identical live/replay entity IDs and resource bags.
  - Host variant: the corpse spawns when level -1 is generated.
  - Late-join guest variant: the host generates level -1 first, so the corpse spawns at admission.
  - The test repo's `ListRecoverableCharacterCorpses` panics, proving neither path reads the live
    table.
- **Red on v476** (corpse fed through the live table, as before the fix):
  - host: `replay mismatch: event 2 payload differs: derived {"event_type":"monster_aggro","entity_id":"1023"} != recorded {... "880011"}`
  - guest: `Verify` passed but the entity IDs diverged (live `… 880002`). The corpse allocated
    after every monster on the level, so no recorded event carried a shifted ID within 80 ticks.
    That is why the test also compares snapshots.
- `sessionsetup/sessionsetup_test.go` imports the package directly. It covers member ordering and
  the legacy host, the class override, and `AddGuest` loading the bag and corpses: the allocator
  advances past the persisted IDs, and `WeaponSet` is kept.
- `store/session_start_snapshot_test.go` (Postgres) freezes a corpse, loots it through
  `TransferCorpseItemToCharacter`, and checks that the snapshot still returns the original body
  with slot, weapon set, and rolled stats intact.

## Ratchet

- `replay/replay.go`: 765 → 572. It is off the grandfather list.
- `realtime/session_loop.go`: baseline 887 → 795.
- `store/repos.go`: baseline 2369 → 2193.
- `replay/replay_test.go` stays at its 1622 baseline. Its fake repo changed in place: the
  signature, plus the member-less `LoadSessionStartSnapshotForMember` fallback now matching
  Postgres, which returns an empty snapshot rather than `ErrNotFound`.

## Known gaps (not fixed here)

- **Mercenary roster** (`mercenaryroster.LoadIntoSim`) still reads live alt-character progression
  and items at build and replay time. This is the same flaw corpses had.
- **Never-attached guest.** A guest that HTTP-joins a running loop and never attaches keeps
  `joined_tick = -1`. Replay adds it at tick 0; live never does. Reproduced with a scratch
  live-loop test: `event count: derived 43, recorded 38`. To fix it, have the live build persist
  `joined_tick = 0` for the members it adds, and have replay skip `joined_tick < 0`.
- **Legacy sessions** from before migration 0031 have no corpse rows. They now reconstruct with no
  corpses on both sides.
