# Spec: `gear-before-combat-scenario`

Status: Complete
Branch: `feature/gear-before-combat-scenario`
Related:

- [`spec-visual-bot-scenario-runner.md`](spec-visual-bot-scenario-runner.md)
- [`spec-resume-authoritative-state.md`](spec-resume-authoritative-state.md)
- [`../PROGRESS.md`](../PROGRESS.md)

## 1. Purpose

Add a second visual bot scenario that proves a more ARPG-shaped progression loop:

```text
start far from monster
walk to sword
pick up sword
equip sword
kill monster while sword is equipped
monster drops another item
pick up second item
assert inventory has 2 items
```

The scenario must be visually available through `make bot-visual` automatically, and it must use the same authoritative protocol and replay path as every other bot scenario.

## 2. Current Problems

This cannot be implemented correctly as only a bot JSON file today.

### 2.1 The initial world is hardcoded in Go

`server/internal/game/sim.go` currently starts every session with:

```text
player at (10, 5)
training_dummy at (12, 5)
no initial loot entity
```

The requested scenario needs:

```text
player farther away
rusty_sword loot in the middle
monster farther beyond the sword
```

If the bot or client fakes that sword locally, it violates the authoritative boundary. The server must spawn the initial sword as real loot in the authoritative snapshot.

### 2.2 The monster currently drops the same sword

`shared/rules/loot_tables.v0.json` has one loot table, `basic_drop`, and it always drops `rusty_sword`.

The requested scenario needs the monster to drop a second, different item after the player already owns the sword.

### 2.3 Item rules only support weapon slot items

`shared/rules/items.v0.schema.json` currently requires every item to have:

```json
{
  "name": "...",
  "slot": "weapon",
  "equippable": true
}
```

Actually, the schema allows `equippable: false`, but still requires `slot`, and that slot can only be `"weapon"`. A non-equippable quest/material/drop item should not need a weapon slot.

### 2.4 Sessions do not persist world/scenario identity

Replay currently reconstructs a session from:

```text
session seed + recorded inputs + current hardcoded initial world
```

If we introduce multiple world setups, the chosen setup must be persisted with the session. Otherwise replay, resume, `/state`, and visual timeline reconstruction can drift from the original run.

### 2.5 Bot scenario steps cannot target world positions yet

The current scenario runner supports simple scripted actions such as:

```text
move
attack_until_event
pick_up_first_loot
equip_first_inventory_item
```

The new scenario needs higher-level steps:

```text
walk to specific loot/item
pick up loot with item_def_id == rusty_sword
equip item_def_id == rusty_sword
walk to monster
assert inventory_count == 2
```

The bot should still drive only normal protocol messages, not mutate state directly.

## 3. Non-goals

- No client-side fake sword spawn.
- No special Godot-only scenario scripting.
- No combat stat modifiers from equipped weapons yet; the acceptance only requires that the sword is equipped before combat.
- No generalized quest system.
- No character-scoped inventory.
- No new item visual requirement for the second item unless the item is equippable.

## 4. Required Design

### 4.1 World presets as data

Create a shared data file for deterministic initial world setup.

Proposed file:

```text
shared/rules/worlds.v0.json
shared/rules/worlds.v0.schema.json
```

Shape:

```json
{
  "version": 0,
  "worlds": {
    "vertical_slice": {
      "player": {"position": {"x": 10, "y": 5}},
      "entities": [
        {
          "type": "monster",
          "monster_def_id": "training_dummy",
          "position": {"x": 12, "y": 5}
        }
      ]
    },
    "gear_before_combat": {
      "player": {"position": {"x": 0, "y": 5}},
      "entities": [
        {
          "type": "loot",
          "item_def_id": "rusty_sword",
          "position": {"x": 6, "y": 5}
        },
        {
          "type": "monster",
          "monster_def_id": "training_dummy_reward",
          "position": {"x": 12, "y": 5}
        }
      ]
    }
  }
}
```

The existing behavior becomes the `vertical_slice` preset.

### 4.2 Session creation chooses a world preset

Extend `POST /v0/sessions` with optional `world_id`:

```json
{
  "mode": "solo",
  "world_id": "gear_before_combat"
}
```

Default:

```text
world_id = vertical_slice
```

Resume behavior:

- `resume_session_id` ignores incoming `world_id`.
- The resumed session uses the world ID persisted when it was first created.

### 4.3 Persist world ID with the session

Add a nullable/defaulted session column:

```sql
ALTER TABLE sessions ADD COLUMN world_id TEXT NOT NULL DEFAULT 'vertical_slice';
```

Update:

```text
store.Session
CreateSession
GetSession
sessionResponse
replay.Reconstruct
replay.BuildTimeline
realtime.Hub resume path
```

Replay must instantiate the same initial world as the original session.

### 4.4 Sim construction becomes world-aware

Preserve current compatibility with a default constructor, but introduce explicit world construction:

```go
game.NewSim(sessionID, seed, rules) // default vertical_slice
game.NewSimWithWorld(sessionID, seed, rules, worldID)
```

`NewSimWithWorld` should:

1. Look up `worldID` in `rules.Worlds`.
2. Spawn the player from the preset.
3. Spawn preset entities in deterministic listed order.
4. Allocate IDs using the existing deterministic counter.

For the proposed `gear_before_combat` world:

```text
1001 player
1002 initial rusty_sword loot
1003 monster
1004 inventory item created when sword is picked up
1005 monster drop loot
1006 inventory item created when monster drop is picked up
```

Bot scenarios must not assume these exact IDs unless documented by a golden. Prefer item/monster selectors.

### 4.5 Add a second monster definition or configurable loot table

Current `training_dummy` should keep dropping `rusty_sword` so the existing scenario remains stable.

Add a second monster:

```json
{
  "training_dummy_reward": {
    "name": "Training Dummy",
    "max_hp": 3,
    "loot_table": "reward_drop",
    "retaliation_damage": {"min": 1, "max": 1}
  }
}
```

Add a second loot table:

```json
{
  "reward_drop": {
    "entries": [
      {"item_def_id": "training_badge", "weight": 1}
    ]
  }
}
```

### 4.6 Add non-equippable item support

Revise `items.v0.schema.json` so non-equippable items do not need a weapon slot.

Proposed item:

```json
{
  "training_badge": {
    "name": "Training Badge",
    "equippable": false
  }
}
```

Validation rules:

- If `equippable == true`, `slot` is required and must be one of known equip slots.
- If `equippable == false`, `slot` may be omitted.

Server behavior:

- Pickup works for all items.
- Equip rejects non-equippable items with existing `not_equippable`.
- Inventory views may expose empty slot for non-equippable items.

### 4.7 Bot scenario catalog grows world/selectors/assertions

Add a new scenario:

```text
tools/bot/scenarios/gear_before_combat.json
```

Proposed shape:

```json
{
  "id": "gear_before_combat",
  "title": "Gear before combat",
  "world_id": "gear_before_combat",
  "description": "Pick up and equip the sword before fighting, then loot a second item.",
  "steps": [
    {"action": "walk_to_loot", "item_def_id": "rusty_sword"},
    {"action": "pick_up_loot", "item_def_id": "rusty_sword"},
    {"action": "equip_inventory_item", "item_def_id": "rusty_sword", "slot": "weapon"},
    {"action": "walk_to_monster", "monster_def_id": "training_dummy_reward"},
    {"action": "attack_until_event", "monster_def_id": "training_dummy_reward", "event_type": "monster_killed"},
    {"action": "pick_up_loot", "item_def_id": "training_badge"}
  ],
  "assertions": [
    {"type": "inventory_count", "equals": 2},
    {"type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": true},
    {"type": "inventory_contains", "item_def_id": "training_badge", "equipped": false},
    {"type": "monster_dead", "monster_def_id": "training_dummy_reward"}
  ]
}
```

The existing string assertions can remain supported for `vertical_slice`, but new structured assertions are easier to extend.

### 4.8 Bot movement should be selector-driven

Add bot helpers that inspect snapshots/deltas and target authoritative entities by attributes:

```text
loot where item_def_id == rusty_sword
monster where monster_def_id == training_dummy_reward
```

Movement should still send normal `move_intent`.

For v0/v7 simplicity, movement can be axis-aligned and deterministic:

1. Compute entity position from latest authoritative state.
2. Send `move_intent` toward the target for a bounded number of ticks.
3. Stop when the player is close enough or after a timeout.

Important: pickup currently does not enforce distance. This scenario may visually walk to the sword, but pickup is still accepted anywhere. Adding pickup range is a future combat/world interaction slice, not required here.

## 5. Implementation Order

### Step 1: Shared data and validation

Files:

```text
shared/rules/worlds.v0.schema.json
shared/rules/worlds.v0.json
shared/rules/items.v0.schema.json
shared/rules/items.v0.json
shared/rules/loot_tables.v0.json
shared/rules/monsters.v0.json
tools/validate_shared.py
```

Work:

1. Add world preset schema and data.
2. Add `training_badge`.
3. Add `reward_drop`.
4. Add `training_dummy_reward`.
5. Validate cross-references:
   - world monster refs exist in monsters.
   - world loot item refs exist in items.
   - monster loot tables exist.
   - loot table item refs exist.

Why first:

The server, bot, and tests should consume data, not invent scenario-specific state.

### Step 2: Persist and expose session world ID

Files:

```text
server/migrations/0002_session_world_id.sql
server/internal/store/models.go
server/internal/store/repos.go
server/internal/http/session.go
server/internal/http/auth_session_test.go
```

Work:

1. Add `world_id` to sessions.
2. Accept optional `world_id` on session create.
3. Default to `vertical_slice`.
4. Reject unknown world IDs with a client error.
5. Include `world_id` in session response for tooling/debug visibility.
6. Ensure resume ignores requested world and returns persisted world.

Why second:

Replay and resume need a durable world identity before the sim becomes world-aware.

### Step 3: Make sim/replay world-aware

Files:

```text
server/internal/game/rules.go
server/internal/game/sim.go
server/internal/game/game_test.go
server/internal/replay/replay.go
server/internal/realtime/hub.go
server/internal/http/inspect.go
server/internal/http/replay_test.go
```

Work:

1. Load `WorldDef` into `game.Rules`.
2. Add `NewSimWithWorld`.
3. Spawn initial world entities from data in deterministic order.
4. Update all session-runner and replay construction paths to use `sess.WorldID`.
5. Keep `NewSim` as a default `vertical_slice` wrapper for existing tests.
6. Add tests proving replay reconstructs `gear_before_combat` initial sword and monster.

Why third:

At this point sessions can persist world identity, so reconstruction remains deterministic.

### Step 4: Bot scenario runner support

Files:

```text
tools/bot/scenarios/gear_before_combat.json
tools/bot/run.py
tools/bot/test_protocol.py
```

Work:

1. Add `world_id` support to scenario files.
2. Pass `world_id` when creating sessions.
3. Track latest authoritative entity positions from snapshot/deltas.
4. Add selector helpers for loot/item/monster.
5. Add actions:
   - `walk_to_loot`
   - `pick_up_loot`
   - `equip_inventory_item`
   - `walk_to_monster`
6. Add structured assertions:
   - inventory count
   - inventory contains item def
   - equipped state
   - monster dead by monster def

Why fourth:

The bot can only be honest once the server owns the initial sword and world layout.

### Step 5: Visual replay verification

Files:

```text
client/scripts/main.gd
scripts/bot_visual.sh
docs/PROGRESS.md
```

Expected work:

No scenario-specific Godot code should be needed.

Only update docs if `make bot-visual` automatically plays:

```text
vertical_slice
gear_before_combat
```

Why last:

Visual replay already consumes protocol-shaped envelopes. If the server and bot are correct, the new scenario becomes visually accessible by adding the scenario JSON.

## 6. Acceptance Criteria

1. `make bot` runs both `vertical_slice` and `gear_before_combat`.
2. `make bot-visual` records both scenarios and plays both in order.
3. In `gear_before_combat`, the first snapshot includes:
   - player far from monster
   - a `rusty_sword` loot entity between player and monster
   - a monster using `training_dummy_reward`
4. Bot picks up and equips `rusty_sword` before the first attack.
5. Monster death drops `training_badge`.
6. Bot picks up `training_badge`.
7. Final authoritative state has exactly two inventory items:
   - `rusty_sword`, equipped
   - `training_badge`, not equipped
8. Replay verification succeeds for both scenarios.
9. `/state`, WebSocket resume, and replay timeline agree for `gear_before_combat`.
10. `make ci` remains green.

## 7. Tests

### Shared validation

```bash
make validate-shared
```

Must validate:

- world schema
- world-to-monster references
- world-to-item references
- monster-to-loot-table references
- loot-table-to-item references

### Go tests

```bash
cd server && go test ./internal/game ./internal/http ./internal/replay
```

Required coverage:

- default world remains compatible with existing vertical slice.
- gear-before-combat world spawns player, sword loot, and monster in deterministic order.
- session create persists selected world.
- resume ignores conflicting requested world.
- replay timeline reconstructs selected world.

### Python tests

```bash
make tools
.venv/bin/python -m pytest -q tools
```

Required coverage:

- scenario loader accepts `world_id`.
- structured assertions work.
- unknown scenario/world/action/assertion fails clearly.

### End-to-end

```bash
make ci
GODOT_FLAGS="--headless" make bot-visual
```

`make bot-visual` should close normally after replaying both scenarios.

## 8. Open Questions

| # | Question | Proposed answer |
|---|----------|-----------------|
| 1 | Should the second item have visuals? | No for this slice; it is a non-equippable inventory proof. |
| 2 | Should pickup require distance? | No for this slice; visual walking proves flow, server range checks are a future interaction slice. |
| 3 | Should equipped sword affect damage? | No for this slice; acceptance is ordering and state, not combat math. |
| 4 | Should `world_id` be exposed in protocol snapshots? | Not required; session response and replay/debug state are enough for tooling. |
| 5 | Should scenario order be configurable? | Later; file-name/catalog order is enough for v7. |

## 9. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Replay drift from new world presets | Persist `world_id` and use it in every sim construction path. |
| Existing vertical slice changes accidentally | Keep `vertical_slice` as the default world and preserve current monster/drop data. |
| Bot starts depending on hardcoded entity IDs | Add selectors by `item_def_id` and `monster_def_id`. |
| Non-equippable item breaks inventory/equip assumptions | Add schema and server tests for `equippable: false`. |
| Visual replay needs custom scenario code | Do not add Godot scenario branches; rely on replay timeline envelopes. |
