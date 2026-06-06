# Spec: `teleporters-and-waypoint-ui`

Status: Complete; `make ci` green on 2026-06-06
Branch: `feature/teleporters-and-waypoint-ui`
Slice: v19 — generated dungeon teleporters, session discovery, waypoint panel, level teleport intent
Baseline: v18 `dungeon-levels-and-stairs`
Related:

- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) — multi-level Sim, generated stairs, scoped transition deltas
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — D4 waypoints deferred from v18
- [`v10_spec-click-action-and-melee-range.md`](v10_spec-click-action-and-melee-range.md) — interactable activation pattern
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) — client UI shortcut checklist
- [`../PROGRESS.md`](../PROGRESS.md)

## 1. Purpose

v18 proved multiple dungeon levels and stair transitions. v19 adds the first fast-travel loop:
each generated dungeon floor gets one teleporter/waypoint. The player discovers a level's
teleporter by clicking it, then can use any discovered teleporter to travel to another discovered
level's teleporter.

After this slice:

- Dungeon generation places one `teleporter` interactable per generated dungeon level.
- Discovery is session-scoped and replayed from authoritative inputs. It is not persisted across
  new sessions or characters yet.
- A v1 `session_snapshot` includes `discovered_teleporters`, keyed by level.
- A v1 `state_delta` can emit `teleporter_discovered` events and `teleporter_discovery_update`
  changes.
- Clicking a teleporter first discovers it. Clicking an already discovered teleporter opens a
  left-side client panel.
- The panel lists all generated/visited levels known to the session. Undiscovered levels are shown
  disabled; discovered levels are enabled.
- If the list exceeds nine rows, the panel scrolls vertically.
- Choosing an enabled level sends `teleport_intent { target_level }`.
- The server validates discovery and moves the player to the destination level's teleporter,
  emitting the same old-level/remove and new-level/full-spawn transition shape as stairs.

## 2. Non-goals

- No character-scoped waypoint persistence.
- No town waypoint, safe-zone rules, vendors, or NPCs.
- No production teleporter art, VFX, sound, or transition animation.
- No arbitrary target IDs in `teleport_intent`; the target is a level number.
- No hidden infinite level catalog. The UI lists only generated/visited levels known to the
  current session.
- No plugin adoption for this UI slice.

## 3. Files to create or modify

```text
shared/rules/dungeon_generation.v0.json        - teleporter placement rules
shared/rules/dungeon_generation.v0.schema.json - schema for teleporter placement rules
shared/rules/interactables.v0.json             - teleporter interactable definition
shared/rules/interactables.v0.schema.json      - ready transition enum allows waypoint
shared/protocol/messages.v1.schema.json        - teleport_intent payload
shared/protocol/session_snapshot.v1.schema.json - discovered_teleporters
shared/protocol/state_delta.v1.schema.json     - teleporter events/changes
shared/golden/dungeon_stairs.json              - extend pinned generated level fixture
tools/validate_shared.py                       - validation for teleporter rules/protocol
server/internal/game/dungeon_gen.go            - deterministic teleporter placement
server/internal/game/sim.go                    - discovery state and teleport transition
server/internal/game/types.go                  - protocol views for discovery
server/internal/inputdecode/inputdecode.go     - teleport_intent decode
server/internal/realtime/runner.go             - unchanged shape, validates new state through schemas
server/internal/replay/replay.go               - input decode/replay through shared decoder
tools/bot/run.py                               - teleporter action and assertions
tools/bot/scenarios/13_teleporter_lab.json     - discover/teleport round trip
client/scripts/main.gd                         - teleporter rendering, click behavior, left panel
client/tests/test_golden.gd                    - teleporter golden fixture checks
docs/PROGRESS.md                               - lifecycle update when v19 ships
```

## 4. Data Shapes

### Rules

`dungeon_generation.v0.json` adds:

```json
{
  "teleporter_placement": {
    "margin_from_wall": 2.0,
    "min_stair_distance": 4.0,
    "max_attempts": 64
  }
}
```

### Interactable

```json
{
  "teleporter": {
    "name": "Teleporter",
    "initial_state": "ready",
    "transition": "waypoint"
  }
}
```

### Snapshot

```json
{
  "discovered_teleporters": [
    { "level": -1, "discovered": true },
    { "level": -2, "discovered": false }
  ]
}
```

The list includes generated/visited dungeon levels only. It is empty for single-level worlds.

### Delta Change

```json
{
  "op": "teleporter_discovery_update",
  "level": -1,
  "discovered": true
}
```

### Event

```json
{
  "event_type": "teleporter_discovered",
  "level": -1,
  "entity_id": "1003"
}
```

`level_changed` remains the transition event for actual travel.

### Intent

```json
{
  "type": "teleport_intent",
  "payload": { "target_level": -2 }
}
```

Reject reasons: `invalid_payload`, `not_dungeon_world`, `player_dead`, `no_teleporter_in_range`,
`teleporter_not_discovered`, `target_level_not_discovered`, `invalid_level`.

## 5. Architecture and Flow

Discovery:

```text
player clicks teleporter entity
  -> client sends action_intent target_id
  -> server validates player is near that teleporter
  -> if level undiscovered: mark discovered, emit discovery update + teleporter_discovered
  -> if already discovered: ack only; client opens panel locally from current snapshot/delta state
```

Travel:

```text
player selects enabled level in waypoint panel
  -> client sends teleport_intent target_level
  -> server validates current-level teleporter in range and target discovered
  -> ensure destination level exists
  -> move player to destination teleporter
  -> emit from-level level_changed + player remove
  -> emit to-level complete spawn set
```

The same deterministic transition delta shape from v18 is reused so the client does not need a
second level-loading model.

## 6. Acceptance Criteria

1. `dungeon_levels` level -1 and -2 generation includes stairs plus one teleporter each, pinned by
   `shared/golden/dungeon_stairs.json`.
2. First click on level -1 teleporter discovers level -1 and updates snapshot/delta discovery state.
3. After descending to -2, the panel can show level -1 enabled and level -2 disabled until the
   level -2 teleporter is clicked.
4. Clicking the level -2 teleporter enables it.
5. Selecting level -1 from level -2 teleports the player to the level -1 teleporter.
6. Selecting an undiscovered or invalid level is rejected by the server.
7. Reconnect resume and replay reconstruct current level and discovered teleporters.
8. The Godot panel appears on the left side, shows disabled undiscovered levels, and uses a scroll
   container when more than nine known levels are present.

## 7. Testing Plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Dungeon|Teleport|Rules'`
3. `make client-unit`
4. `make bot` after local DB is available
5. `make ci` final gate when DB-dependent infrastructure is available
