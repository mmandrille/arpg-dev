# Spec: `character-scoped-persistence`

Status: Draft — pending review
Branch: `feature/character-scoped-persistence`
Slice: v22 — durable character item instances, equipment, and waypoints across fresh play sessions
Baseline: v21 `dungeon-monster-combat`
Related:

- [`v13_spec-inventory-ui.md`](v13_spec-inventory-ui.md) — inventory mutation and presentation
- [`v16_spec-use-consumable.md`](v16_spec-use-consumable.md) — consumable removal from inventory
- [`v19_spec-teleporters-and-waypoint-ui.md`](v19_spec-teleporters-and-waypoint-ui.md) — waypoint discovery and travel
- [`v20_spec-play-session-loop.md`](v20_spec-play-session-loop.md) — fresh town-to-dungeon run loop
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) — current dungeon threat baseline
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — D1 character persistence, D4 persistent waypoints
- [`../PROGRESS.md`](../PROGRESS.md)

## 1. Purpose

The current game has real accounts and a default character row, but inventory and waypoint
progression are still effectively session-scoped. Starting a fresh `make play` session gives the
same account a new dungeon run, but durable player progress does not come along.

This slice makes the default character the durable owner of item instances, equipped weapon state,
and unlocked waypoints. A fresh session for the same authenticated account loads the character's
saved items and discovered waypoints into a new authoritative run. Dungeon maps, monsters, floor
loot, player HP, and live session replay remain session-scoped.

The result is the first cross-session progression loop: pick up and equip gear, discover a dungeon
teleporter, close the session, start a new session, and keep the character-owned gear and waypoint
access while receiving a fresh generated dungeon layout. The persistence model must anticipate
future per-item randomly rolled stats: every item the character owns is a durable instance, whether
it is equipped, in inventory, or later moved to stash.

## 2. Non-goals

- No character select UI or multiple character management. Use the existing default character.
- No persistent dungeon map layout, monster state, dropped floor loot, corpses, or opened doors
  across sessions.
- No player-facing resume of an old live session; fresh play creates a new session from character
  state.
- No stash UI or stash interactions. The durable item-instance model may include a `location`
  value that can later represent stash, but v22 only exercises inventory and equipped items.
- No item stat rolling implementation. v22 stores a future-ready `rolled_stats` payload but does
  not generate randomized affixes or stat values yet.
- No vendors, gold economy, crafting, quest state, character stats, skills, level/XP, or respawn.
- No migration compatibility for old local development rows beyond what the implementation needs
  to keep tests and local startup healthy.
- No protocol version bump unless implementation discovers an unavoidable schema change; snapshots
  already carry inventory and discovered teleporters.

## 3. Files to create or modify

```text
docs/specs/v22_spec-character-scoped-persistence.md        - this slice contract
docs/plans/v22_2026-06-07-character-scoped-persistence.md  - implementation plan
server/migrations/0003_character_progression.sql           - character item-instance/waypoint ownership migration
server/internal/store/models.go                            - character item-instance and waypoint persistence models
server/internal/store/interfaces.go                        - character-scoped item/waypoint repo methods
server/internal/store/repos.go                             - Postgres implementation
server/internal/store/store_test.go                        - persistence unit coverage
server/internal/http/session.go                            - fresh sessions load default character progression
server/internal/realtime/hub.go                            - initial Sim receives character progression on fresh attach
server/internal/realtime/runner.go                         - persist inventory and waypoint mutations by character
server/internal/replay/replay.go                           - keep replay reconstruction session-input-owned
server/internal/http/ws_test.go                            - cross-session inventory/equipment/waypoint tests
server/internal/http/auth_session_test.go                  - create/resume behavior around default character
tools/bot/run.py                                           - multi-session same-account scenario helper
tools/bot/scenarios/15_character_persistence.json          - end-to-end persistence proof
docs/PROGRESS.md                                           - lifecycle update when v22 ships
```

No new Godot UI is required. Existing inventory and waypoint panels should render the snapshot
state they already receive.

## 4. Data shapes

### Durable item instances

Current `inventory_items` rows include `session_id`, `account_id`, and `character_id`, but repo
methods query and mutate by `session_id`. v22 changes the durable ownership key to `character_id`
and treats every owned item as a durable instance, not just an inventory row.

Target logical shape:

```sql
character_item_instances (
  id TEXT PRIMARY KEY,
  character_id TEXT NOT NULL REFERENCES characters(id),
  account_id TEXT NOT NULL REFERENCES accounts(id),
  item_def_id TEXT NOT NULL,
  location TEXT NOT NULL, -- inventory | equipped | stash (stash reserved for future use)
  slot TEXT NULL,
  equipped BOOLEAN NOT NULL DEFAULT FALSE,
  rolled_stats JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)
```

Implementation may either migrate the existing `inventory_items` table in place or introduce a new
table. The required contract is that item instance IDs are unique per character across sessions,
item mutations are persisted by character, and the schema can retain per-instance rolled stat data
when loot generation grows beyond fixed `item_def_id` rules.

For v22, loaded protocol inventory can continue to expose the existing fields. `rolled_stats` is
durable server data reserved for future stat/affix slices unless implementation chooses to include
it in debug-only inspection output.

### Durable waypoints

Target logical shape:

```sql
character_waypoints (
  character_id TEXT NOT NULL REFERENCES characters(id),
  level INTEGER NOT NULL,
  discovered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (character_id, level)
)
```

Level `0` town is always available. The implementation may persist it explicitly or synthesize it
when loading character waypoints, but snapshots for fresh `dungeon_levels` sessions must include
town as discovered.

### Fresh session bootstrap

`POST /v0/sessions` keeps its current request/response shape. On fresh create:

```text
authenticate account
get or create default character
create new session row with account_id + character_id + fresh seed
load character item instances + equipped state
load character waypoint unlocks
start Sim from fresh world/session seed
inject durable character item instances and discovered waypoints into initial snapshot
```

Same-session WebSocket reconnect and replay reconstruction continue to use recorded
`session_inputs` first. They must not load a newer character item snapshot that would rewrite the
historical session being replayed.

## 5. Architecture and flow

### Item mutation flow

```text
pickup/equip/unequip/drop/use intent
  -> Sim validates and emits inventory_add/update/remove changes
  -> realtime runner persists each item-instance change against session.character_id
  -> current session snapshot/deltas continue to reflect Sim state
  -> next fresh session loads durable character item instances into a new Sim
```

Dropping or using an item removes it from the durable character item table for v22. Dropped floor
loot itself remains session-scoped; if the session ends before re-pickup, that dropped item is
gone. Future stash behavior should move the same item instance to `location = 'stash'` rather than
creating a separate representation.

### Waypoint mutation flow

```text
action_intent on teleporter
  -> Sim discovers current level
  -> state_delta emits teleporter_discovery_update / teleporter_discovered
  -> realtime runner persists (character_id, level)
  -> next fresh session lists persisted levels as discovered destinations
```

Using a persisted waypoint in a fresh session generates that level from the new session seed if the
level has not been visited in the current session. This follows ADR-0008 D4: waypoint access
persists; map layout does not.

### Replay boundary

Replay is a historical session verifier, not a "current character state" loader. Given a recorded
session seed and ordered inputs, replay must reconstruct the session from its own inputs and
persisted session metadata. Character item instances and waypoints are used only for the initial
fresh session snapshot that was current when the session began; later out-of-session character
changes must not affect replay output.

If the implementation needs a session-start character snapshot for replay stability, persist that
snapshot on session create rather than reading live character state during replay.

## 6. Acceptance criteria

1. A fresh session for an authenticated account loads the account's default character item
   instances, inventory view, and equipped weapon state.
2. Picking up `rusty_sword`, equipping it, ending the session process, and creating a new session
   for the same account shows `rusty_sword` in the initial snapshot inventory and equipped weapon.
3. Using a `red_potion` removes it from durable character item instances; a later fresh session
   does not restore the consumed item.
4. Dropping an item removes it from durable character item instances; a later fresh session does not
   restore the dropped item unless it was picked up again before session end.
5. Discovering a dungeon teleporter persists that level for the character; a later fresh session
   includes the discovered level in `discovered_teleporters`.
6. Teleporting to a persisted negative dungeon level in a fresh session generates that level from
   the new session seed, not from prior-session map state.
7. Same-session reconnect still reconstructs current live session state from recorded inputs and
   does not duplicate inventory rows.
8. Replay verification for old and new sessions remains deterministic and does not read mutable
   live character state in a way that changes historical output.
9. Another account cannot load, inspect, or mutate the first account's character item instances,
   sessions, or waypoints.
10. Character-owned item persistence includes a future-ready `rolled_stats` payload and a location
    model that can represent inventory/equipped now and stash later, without implementing stat
    rolling or stash UI in v22.
11. Bot scenario `15_character_persistence.json` proves same-account cross-session inventory,
    equipped weapon, waypoint persistence, `/state`, reconnect, and replay.
12. `make ci` green.

## 7. Testing plan

1. `cd server && go test ./internal/store/...` or `go test ./internal/store` — migration/repo tests
   for character item instances and waypoints.
2. `cd server && go test ./internal/http/... -run 'Character|Persistence|Waypoint|Inventory'` —
   HTTP/WebSocket cross-session behavior.
3. `make bot` — scenario `15_character_persistence.json` plus regression scenarios.
4. `make client-unit` — unchanged client data handling still passes.
5. `make ci` — final gate.
6. Manual: `make play`, pick up/equip an item and discover a teleporter, restart play with the same
   dev account, confirm the inventory panel and waypoint panel retain character-owned progress.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Use the existing default character; no character picker. | The repo already creates one character per account and this slice is about persistence semantics, not UI. |
| 2 | Item instances and equipped weapon become character-owned. | ADR-0008 D1 requires cross-session character state; future rolled gear needs durable per-instance rows. |
| 3 | Waypoint unlocks become character-owned; maps stay session-owned. | Matches ADR-0008 D4 and keeps PCG replay/fresh-run behavior clean. |
| 4 | Same-session replay remains session-input-owned. | Deterministic debugging must not drift when live character state changes later. |
| 5 | Dropped and consumed items are removed durably. | Prevents duplication and keeps the server authoritative for inventory loss. |
| 6 | Town level `0` remains always unlocked. | ADR-0008 D4 says town is the default fast-travel hub. |

## 9. Open questions

| # | Question | Default if unanswered |
|---|----------|----------------------|
| Q-1 | Should v22 migrate existing local `inventory_items` rows into character ownership, or is a dev DB reset acceptable? | Prefer a migration that preserves rows where possible, but do not overfit old local data. |
| Q-2 | Should session create persist an immutable starting character snapshot for replay, or can replay rely only on input reconstruction? | Persist a minimal session-start snapshot if tests show live character state can affect replay. |
| Q-3 | Should the first persisted waypoint beyond town be any discovered floor, or only activated teleporters? | Only activated/discovered teleporters persist, matching v19 behavior. |
| Q-4 | Should player HP persist across sessions? | No. HP remains session-scoped until a future death/respawn/checkpoint slice. |
| Q-5 | Should v22 generate random item stats? | No. Store `rolled_stats` for future item-instance data, but keep current fixed item behavior. |
