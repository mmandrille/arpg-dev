# v478 Spec: Shared Member Setup (live ↔ replay parity)

Status: Complete
Date: 2026-09-29
Codename: `shared-member-setup`
Baseline: v477 `clip-owned-death-pose` (commit `a4490493`); server work built on v476 `recorded-load-shed`

## Problem

Live and replay put session members into the sim through two hand-maintained copies of the
same setup code, and the copies had drifted:

| Setup step | Live host (`buildSessionSim`) | Live late join (`playerIDForMember`) | Replay (`sessionStartSim` / `addMemberToSim`) |
|---|---|---|---|
| Resource wallet | loaded | loaded | **missing** |
| Resource bag | loaded | loaded | **missing** |
| Recoverable corpses | loaded (live table) | loaded (live table) | **missing** |
| Class override from `characters` | applied | **missing** | **missing** |
| Item `WeaponSet` | kept | kept | **dropped** (replay's own `persistedItems`) |
| Guest display name | `"Guest"` | `"Guest <suffix>"` | `"Guest <suffix>"` |
| Member-less legacy host metadata | set | n/a | **not set** |

The missing loads change the entity ID allocator. `LoadAccountResourceBag` and
`LoadCharacterCorpses` raise `nextID` past the persisted IDs. `spawnCorpseOnLevel` then calls
`alloc()` when the corpse's death level is generated, or at once if the level already exists
(late join). Live allocates IDs that replay never does, so every later entity ID and event
diverges. Session resume goes through `replay.Reconstruct`, so a server restart silently dropped
the wallet, bag, and corpses from live play too.

Corpses had a second, independent problem. Live read them from
`ListRecoverableCharacterCorpses` whenever the sim was built. That table is mutable: bodies get
looted and other characters die. So even a replay that loaded corpses the same way would rebuild
them from the table's state at replay time, not session start.

## Decision

1. **One setup path.** New package `server/internal/sessionsetup` owns member setup:
   - `Members`: list members, synthesize the legacy host, sort host-first.
   - `Resolve`: all I/O for a member (start snapshot, class override).
   - `NewHostSim`: the host setup sequence.
   - `AddGuest`: pure sim mutation, so it is safe under the live loop's sim lock.

   Live build, live late join, and replay all call it. The store-row → sim converters move there,
   and the duplicate copies in `realtime/hub.go` and `replay/replay.go` are deleted.
2. **Setup order is part of the replay contract.** Host: start state (inventory, hotbar, skill
   bindings, waypoints, shop, stash, wallet, bag), then mercenary roster, restore hired companion,
   then corpses. Guest: start state, then corpses. This matches the pre-v478 live order, so live
   entity IDs are unchanged.
3. **Corpses get a session-start snapshot. Yes, it is required.** New table
   `session_start_character_corpses` (migration `0031`). `createSessionStartSnapshot` reads
   `ListRecoverableCharacterCorpses` once, when the member creates or joins the session, and
   inserts the result in the same transaction as the rest of the member's snapshot. `ordinal` keeps list order, because
   corpses on an already-generated level spawn in load order. Live and replay both read
   `SessionStartSnapshot.Corpses`. Neither reads the live corpse table during a session.
4. **Gameplay debug is applied before member loads** through a `NewHostSim` `configure` hook,
   matching the pre-v478 live order. Enabling debug seeds unique test chests, and a late call
   re-seeds them after `usePlayer` swaps the chest state out.
5. **`CreateSessionStartSnapshot` takes a `SessionStartSnapshot` value** instead of 14
   positional arguments. The positional list was how new snapshot fields ended up half-wired.
6. **Class override stays, and is now applied everywhere.** The snapshot already records the
   character's class (since v70). The override only fixes pre-v70 snapshots that stored the default.
   A character's class never changes and deleting a character deletes its sessions, so reading
   `characters.character_class` at replay time is reproducible.

## Non-goals / known follow-ups

- **Legacy sessions** created before migration 0031 have no corpse snapshot rows. They now
  reconstruct with no corpses in both live and replay. Before, live loaded them and replay did
  not, which was already irreproducible.
- **Mercenary roster has the same flaw as corpses did.** `mercenaryroster.LoadIntoSim` reads live
  alt-character progression and items at build and replay time. It is not snapshotted, and was
  left out of scope here.
- **Never-connected guests.** A guest that HTTP-joined while the loop was running and never
  attached keeps `joined_tick = -1`. Replay adds such a guest at tick 0, but live never added it.
  Separate slice.

## Acceptance

- A live `sessionLoop` where the host has a corpse on level -1 plus a resource bag item, and walks
  down the stairs through recorded intents, reproduces under `replay.Verify`. Live and replay
  entity IDs and resource bags also match.
- The same holds when a late-joining guest brings the corpse and bag, and the corpse spawns on an
  already-generated level at admission.
- Neither live setup nor replay calls `ListRecoverableCharacterCorpses`. The test repo panics if
  it is called.
- Both tests fail on the v476 baseline. The host test fails `replay.Verify` with `entity_id`
  `1023` against `880011`.
- `replay.go` drops below 600 lines (off the ratchet). `session_loop.go` and `repos.go` baselines
  go down.
