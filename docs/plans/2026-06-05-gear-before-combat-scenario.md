# Gear Before Combat Scenario (Slice v7) — Implementation Plan

Status: Complete

Goal: Add a second bot scenario that proves an ARPG-shaped loop — walk to sword,
pick up, equip, kill monster, loot a second item — with the server owning the
initial world layout and replay/resume reconstructing the same preset.

Architecture: World presets live in `shared/rules/worlds.v0.json`; sessions persist
`world_id`; sim construction becomes world-aware; bot scenarios drive selector-based
movement and structured assertions over the existing WebSocket protocol.

Tech stack: shared JSON rules + validator, Go sim/store/replay/realtime, Python bot,
Godot visual replay playlist (no scenario-specific client code).

Spec: [`docs/specs/spec-gear-before-combat-scenario.md`](../specs/spec-gear-before-combat-scenario.md)
Baseline: slice v6 `visual-bot-scenario-runner` (complete)
Branch: `feature/gear-before-combat-scenario`

## Review findings incorporated

- **Replay drift is the primary risk.** Every sim construction path must use
  `sess.WorldID`, not the hardcoded `NewSim` default. Grep targets:
  `hub.go`, `replay.ReconstructFromInputs`, `replay.BuildTimeline`.
- **v5 resume stays replay-only.** World ID affects only the initial spawn;
  mid-session resume still replays inputs on top of the correct preset.
- **Entity IDs must not leak into scenario JSON.** Migrate `vertical_slice` to
  selector-based steps alongside the new scenario.
- **Scenario order:** prefix filenames so regression runs first:
  `01_vertical_slice.json`, then `02_gear_before_combat.json`.
- **No golden fixture changes** unless a pinned-seed regression is added later;
  structured bot assertions suffice for v7.
- **`make client-smoke` unchanged** — smoke creates sessions without `world_id`
  and continues to use the default `vertical_slice` preset.

## Constants (pin in plan, reuse in code/tests)

| Constant | Value | Used by |
|----------|-------|---------|
| Default world | `vertical_slice` | session create, `NewSim` wrapper, migration default |
| Walk stop distance | Chebyshev ≤ 1 tile | bot `walk_to_*` steps |
| Walk max ticks | 40 per step | bot timeout before failure |
| Walk tick direction | axis-aligned, one axis at a time (|dx| before |dy|) | deterministic bot movement |

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/worlds.v0.schema.json` | World preset schema |
| Create | `shared/rules/worlds.v0.json` | `vertical_slice` + `gear_before_combat` presets |
| Modify | `shared/rules/items.v0.schema.json` | Conditional `slot` when `equippable: true` |
| Modify | `shared/rules/items.v0.json` | Add `training_badge` |
| Modify | `shared/rules/loot_tables.v0.json` | Add `reward_drop` |
| Modify | `shared/rules/monsters.v0.json` | Add `training_dummy_reward` |
| Modify | `tools/validate_shared.py` | World cross-ref + entity-type guards |
| Create | `server/migrations/0002_session_world_id.sql` | Persist `world_id` on sessions |
| Modify | `server/internal/store/models.go` | `Session.WorldID` field |
| Modify | `server/internal/store/repos.go` | Read/write `world_id` |
| Modify | `server/internal/store/store_test.go` | Round-trip `world_id` |
| Modify | `server/internal/http/session.go` | Accept/return `world_id`; validate preset |
| Modify | `server/internal/http/auth_session_test.go` | Create/resume world coverage |
| Modify | `server/internal/game/rules.go` | Load worlds; validate item equippable rules |
| Modify | `server/internal/game/sim.go` | `NewSimWithWorld`; spawn from preset |
| Modify | `server/internal/game/game_test.go` | World spawn + non-equippable equip reject |
| Modify | `server/internal/replay/replay.go` | Pass `worldID` into reconstruction |
| Modify | `server/internal/realtime/hub.go` | Fresh session uses `sess.WorldID` |
| Modify | `server/internal/http/replay_test.go` | Timeline for `gear_before_combat` world |
| Rename | `tools/bot/scenarios/vertical_slice.json` → `01_vertical_slice.json` | Regression-first order |
| Create | `tools/bot/scenarios/02_gear_before_combat.json` | New scenario |
| Modify | `tools/bot/run.py` | World ID, selectors, walk actions, structured assertions |
| Modify | `tools/bot/test_protocol.py` | Loader, assertions, unknown action/world tests |
| Modify | `docs/PROGRESS.md` | Record v7 when complete |

No changes expected: `client/scripts/main.gd`, `scripts/bot_visual.sh`, `client/scripts/smoke.gd`.

---

## Task 1: Shared Data and Validation

Files:
- Create: `shared/rules/worlds.v0.schema.json`, `shared/rules/worlds.v0.json`
- Modify: `shared/rules/items.v0.schema.json`, `shared/rules/items.v0.json`
- Modify: `shared/rules/loot_tables.v0.json`, `shared/rules/monsters.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `worlds.v0.schema.json` with:
  - `version: 0`
  - `worlds` map keyed by world id
  - each world: `player.position`, `entities[]`
  - entity `type` enum: `monster` | `loot`
  - `monster` requires `monster_def_id`; `loot` requires `item_def_id`
- [x] Step 1.2: Add `worlds.v0.json` with presets from spec §4.1:
  - `vertical_slice`: player `(10, 5)`, one `training_dummy` at `(12, 5)`
  - `gear_before_combat`: player `(0, 5)`, `rusty_sword` loot at `(6, 5)`,
    `training_dummy_reward` at `(12, 5)`
- [x] Step 1.3: Relax `items.v0.schema.json` using `if/then`:
  - when `equippable: true` → require `slot` with enum `["weapon"]`
  - when `equippable: false` → forbid `slot` (or omit; no slot required)
- [x] Step 1.4: Add `training_badge` (`equippable: false`, no slot).
- [x] Step 1.5: Add `reward_drop` → `training_badge`; add `training_dummy_reward`
  monster (same stats as `training_dummy`, different loot table).
- [x] Step 1.6: Extend `validate_shared.py` cross-checks:
  - load and schema-validate `worlds.v0.json`
  - each world entity ref resolves (monster → monsters, loot item → items)
  - each world's `player.position` has numeric `x`/`y`
  - reject unknown entity types or missing required fields per type

Verification:

```bash
make validate-shared
```

---

## Task 2: Persist and Expose Session World ID

Files:
- Create: `server/migrations/0002_session_world_id.sql`
- Modify: `server/internal/store/models.go`, `repos.go`, `store_test.go`
- Modify: `server/internal/http/session.go`, `auth_session_test.go`

- [x] Step 2.1: Migration adds `world_id TEXT NOT NULL DEFAULT 'vertical_slice'`
  to `sessions`.
- [x] Step 2.2: Add `WorldID string` to `store.Session`; wire `CreateSession`,
  `GetSession`, and any list/touch paths that read sessions.
- [x] Step 2.3: Extend `createSessionRequest` with optional `world_id`.
- [x] Step 2.4: On create:
  - default `world_id` to `vertical_slice` when omitted
  - reject unknown world ids with `400 invalid_world_id`
  - persist selected world on the session row
- [x] Step 2.5: Extend create/resume JSON response with `world_id`.
- [x] Step 2.6: On resume (`resume_session_id` set): ignore incoming `world_id`;
  return persisted value.
- [x] Step 2.7: Tests:
  - create without `world_id` → `vertical_slice`
  - create with `gear_before_combat` → persisted and returned
  - create with unknown world → 400
  - resume returns original world even if caller sends a different one

Verification:

```bash
cd server && go test ./internal/store ./internal/http -run 'Session|World'
```

---

## Task 3: World-Aware Sim and Replay

Files:
- Modify: `server/internal/game/rules.go`, `sim.go`, `game_test.go`
- Modify: `server/internal/replay/replay.go`, `replay_test.go` (create if missing)
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/http/replay_test.go`

- [x] Step 3.1: Add Go types and loader for worlds in `rules.go`:
  ```go
  type WorldDef struct { Player WorldPlayer; Entities []WorldEntity }
  type WorldEntity struct { Type string; MonsterDefID string; ItemDefID string; Position Vec2 }
  ```
  Load `worlds.v0.json` into `Rules.Worlds map[string]WorldDef`.
  Validate at load: unknown world refs, entity type fields.
- [x] Step 3.2: Add `DefaultWorldID = "vertical_slice"` constant.
- [x] Step 3.3: Implement `NewSimWithWorld(sessionID, seed, rules, worldID)`:
  1. Resolve preset (error if missing)
  2. Spawn player at preset position (hp unchanged: 10)
  3. Iterate `entities` in listed order; allocate IDs via existing `alloc()`
  4. `monster`: load def from rules, set hp/maxHP/lootTable
  5. `loot`: spawn loot entity with `item_def_id`
- [x] Step 3.4: Change `NewSim(...)` to delegate to `NewSimWithWorld(..., DefaultWorldID)`.
- [x] Step 3.5: Remove hardcoded `playerStartPos` / `monsterStartPos` / `monsterDefID`
  spawn logic from `NewSim` body (constants may remain for tests or be deleted).
- [x] Step 3.6: Extend `ReconstructFromInputs` signature with `worldID string`;
  call `NewSimWithWorld` instead of `NewSim`.
- [x] Step 3.7: Update `Reconstruct` to pass `sess.WorldID`.
- [x] Step 3.8: Update `BuildTimeline` to pass `sess.WorldID`.
- [x] Step 3.9: Update `hub.Run` fresh-session path:
  `game.NewSimWithWorld(sess.ID, sess.Seed, h.rules, sess.WorldID)`.
  Resume path already uses `Reconstruct` — no separate change beyond Step 3.7.
- [x] Step 3.10: Add `game_test.go` coverage:
  - `vertical_slice` snapshot matches current layout (player 1001, monster 1002, positions)
  - `gear_before_combat` snapshot has player 1001, loot 1002 (`rusty_sword`), monster 1003
  - equip `training_badge` rejects with `not_equippable`
- [x] Step 3.11: Add replay test: empty-input reconstruction of `gear_before_combat`
  session includes initial sword loot entity in snapshot.
- [x] Step 3.12: Confirm `/state` and timeline endpoint inherit world via `Reconstruct` /
  `BuildTimeline` (no direct `NewSim` left outside tests).

Verification:

```bash
cd server && go test ./internal/game ./internal/replay ./internal/http -run 'World|Replay|Timeline|Gear'
```

---

## Task 4: Bot Scenario Runner — Selectors, Walk, Structured Assertions

Files:
- Rename: `tools/bot/scenarios/vertical_slice.json` → `01_vertical_slice.json`
- Create: `tools/bot/scenarios/02_gear_before_combat.json`
- Modify: `tools/bot/run.py`, `test_protocol.py`

### 4A: Scenario model and session create

- [x] Step 4.1: Extend `Scenario` dataclass with optional `world_id: str | None`.
- [x] Step 4.2: `load_scenarios` reads `world_id` from JSON (default `vertical_slice`).
- [x] Step 4.3: `create_session(client, token, world_id)` sends `world_id` in POST body.
- [x] Step 4.4: Rename scenario file; update `id` field stays `vertical_slice`.

### 4B: Authoritative entity tracking

- [x] Step 4.5: Extend `RuntimeState` with `entities: dict[str, dict]` keyed by entity id.
- [x] Step 4.6: On initial `session_snapshot` and each `state_delta`, merge entity
  create/update/remove into `entities` (position, type, `item_def_id`, `monster_def_id`, hp).
- [x] Step 4.7: Add selector helpers:
  - `find_loot(state, item_def_id) -> entity | None`
  - `find_monster(state, monster_def_id) -> entity | None`
  - `find_inventory_item(inv, item_def_id) -> item | None`
  - `find_player(state) -> entity`

### 4C: Movement actions

- [x] Step 4.8: Implement `walk_toward(ws, state, target_pos, max_ticks=40)`:
  - read player position from `find_player`
  - each tick: choose axis-aligned direction reducing Chebyshev distance
  - send `move_intent` with `duration_ticks: 1`
  - stop when Chebyshev distance ≤ 1 or ticks exhausted (raise on timeout)
- [x] Step 4.9: Add actions:
  - `walk_to_loot` — `{item_def_id}`
  - `walk_to_monster` — `{monster_def_id}`
- [x] Step 4.10: Add `pick_up_loot` — resolve loot entity by `item_def_id`, send
  `pick_up_intent`, wait for matching `inventory_add`.
- [x] Step 4.11: Add `equip_inventory_item` — resolve inventory row by `item_def_id`,
  send `equip_intent`, wait for `equipped_update`.
- [x] Step 4.12: Extend `attack_until_event` to accept optional `monster_def_id`
  (resolve target id from state) while keeping `target_id` for backward compat.

### 4D: Migrate vertical_slice scenario

- [x] Step 4.13: Update `01_vertical_slice.json`:
  - replace fixed `move` step with optional keep-as-is (simple move still valid)
  - replace `"target_id": "1002"` with `"monster_def_id": "training_dummy"`
  - keep string assertions unchanged

### 4E: New gear_before_combat scenario

- [x] Step 4.14: Add `02_gear_before_combat.json` per spec §4.7 with structured assertions:
  ```json
  {"type": "inventory_count", "equals": 2}
  {"type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": true}
  {"type": "inventory_contains", "item_def_id": "training_badge", "equipped": false}
  {"type": "monster_dead", "monster_def_id": "training_dummy_reward"}
  ```

### 4F: Assertion engine

- [x] Step 4.15: Refactor `run_assertions` to accept `list[str | dict]`.
- [x] Step 4.16: Implement structured assertion handlers (used for `/state`,
  reconnect, and end-of-scenario checks).
- [x] Step 4.17: Update `check_persistence` to pass structured assertions through
  the same engine (gear scenario: inventory count 2, sword equipped, badge present).
- [x] Step 4.18: Replace hardcoded `MONSTER_ID = "1002"` in monster-dead string
  assertion with monster lookup by `training_dummy` def id.
- [x] Step 4.19: Manifest output includes `world_id` per scenario entry.

### 4G: Python tests

- [x] Step 4.20: Tests for:
  - loader reads `world_id`
  - structured assertion parsing (valid + unknown type)
  - selector helpers with fixture snapshot/delta payloads
  - unknown action/world/scenario fails clearly
  - catalog order: `01_vertical_slice` before `02_gear_before_combat`

Verification:

```bash
make tools
.venv/bin/python -m pytest -q tools/bot/test_protocol.py -v
make bot   # server + db required
```

---

## Task 5: End-to-End and Visual Replay

Files:
- Modify: `docs/PROGRESS.md`

- [x] Step 5.1: Run full bot suite — both scenarios pass including reconnect/resume.
- [x] Step 5.2: Run replay verification for both recorded sessions.
- [x] Step 5.3: Confirm `make bot-visual` records and plays scenarios in order
  (`01_vertical_slice`, then `02_gear_before_combat`) without Godot changes.
- [x] Step 5.4: Confirm `make client-smoke` still green (default world unchanged).
- [x] Step 5.5: Update `docs/PROGRESS.md`:
  - lifecycle row for v7
  - summary of world presets + gear-before-combat flow
  - note scenario catalog now has two entries

Verification:

```bash
make ci
GODOT_FLAGS="--headless" make bot-visual
```

---

## Acceptance Checklist (from spec §6)

- [x] `make bot` runs `vertical_slice` and `gear_before_combat`
- [x] `make bot-visual` records and replays both in filename order
- [x] Initial `gear_before_combat` snapshot: player at `(0,5)`, sword loot at `(6,5)`,
  monster `training_dummy_reward` at `(12,5)`
- [x] Bot equips `rusty_sword` before first attack on reward dummy
- [x] Monster death drops `training_badge`; bot picks it up
- [x] Final state: two inventory items — sword equipped, badge not equipped
- [x] Replay verification passes for both scenarios
- [x] `/state`, WebSocket resume, and replay timeline agree for `gear_before_combat`
- [x] `make ci` green

---

## Rollback / regression guards

| If this breaks… | Check… |
|-----------------|--------|
| Existing vertical slice bot | `vertical_slice` preset positions match old hardcoded spawn; default `world_id` |
| Golden Go tests | `NewSim` still wraps `vertical_slice`; no combat formula changes |
| Resume (v5) | `ReconstructFromInputs` uses persisted `world_id`, not default |
| Visual replay | Timeline built with same world as live session |
| Item visuals | `training_badge` has no `item_visuals` entry (intentional) |

---

## Out of scope (defer)

- Pickup range / distance gate
- Equipped sword changing damage
- `world_id` on WebSocket snapshots
- Configurable scenario run order beyond filename sort
- New golden fixtures for gear-before-combat pinned seed
- Godot inventory UI for non-visual items (inventory exists in protocol only)
