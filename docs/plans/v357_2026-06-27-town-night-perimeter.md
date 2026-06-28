# v357 Plan — Town Night Perimeter

Status: Complete  
Goal: Enclose level-0 town in a night-time wooden palisade with a locked south gate, dungeon-mood fog/lighting, and a respaced hub layout.  
Architecture: Perimeter geometry is committed deterministically in shared data (generator script → `worlds.v0.json` wall segments with `kind: "wood"`). A new `town_exit_gate` interactable carries `locked_exit_reason` in rules; server rejects `action_intent` before the generic door-open path. Client already enables fog at town when walls exist (`_lab_world_fog_at_town_level`); this slice extends **lighting suppression** to match dungeon fog and adds wood fence rendering plus reject feedback. Town services keep existing interactable IDs on a wider ring around center `(11, 12)`.  
Tech stack: shared JSON rules/assets, Go sim, Godot client, Python validate + optional generator script, client bot + Go unit tests.

Spec: [`docs/specs/v357_spec-town-night-perimeter.md`](../specs/v357_spec-town-night-perimeter.md)  
Baseline: v356 `player-navigation-guardrails`

## Spec review (gate)

| Area | Result |
|------|--------|
| Baseline v357 / builds on v356 | OK |
| Scope / non-goals | OK — hub enclosure + presentation; no overworld |
| Contracts | `interactables.v0.json` + schema (`locked_exit_reason`, optional `wood` wall kind in `worlds`); **no protocol bump** |
| Determinism | Fixed perimeter recipe / committed segments; no RNG in `game/` |
| Shared rules | Perimeter tuning in `town_presentation.v0.json`; gate reason data-driven |
| Server authority | Walls + gate barrier + reject reason owned by sim |
| Animation (ADR-0007) | N/A — floating text on existing `intent_rejected` |
| World presets | `dungeon_levels` entity list + wall segments + player spawn |
| Bot proof | Extended client `90_town_night_perimeter`; Go tests for gate/walls |
| Replay | No input-shape change |
| Client assets | Borrow door mesh; adopt procedural wood; reject external assets |
| Maintainability | Thin `main.gd` edits; extract `town_perimeter.go` if needed |
| As-built drift | `_lab_world_fog_at_town_level()` already true when town has walls; lighting path still uses bright `TOWN_PROFILE` at `level >= 0` — confirmed gap |

**Schema note:** `interactables.v0.schema.json` does not yet allow `locked_exit_reason`; plan adds it as an optional field on `initial_state: "closed"` defs with `barrier_when_closed` (still forbids `service`/`transition` on closed defs).

## Baseline and shortcut decision

Reuse patterns from:

- `server/internal/game/interactables.go` — closed-door barrier + toggle (gate intercepts **before** open)
- `server/internal/game/dungeon_generation.v0.json` `boss_floor.locked_exit_reason` — reject-reason precedent (town uses interactable-level field instead)
- `client/scripts/dungeon_depth_lighting.gd` — fog suppression already applied for `level < 0`
- `client/scripts/main.gd` — `_sync_fog_and_dungeon_lighting`, `_lab_world_fog_at_town_level`
- `client/scripts/town_node_factory.gd` — `make_door_node`, `make_town_preview_scene`, `make_interactable_node` fallback to door
- `tools/bot/scenarios/client/87_dungeon_torch_lights.json` — `assert_fog_of_war` at town/dungeon
- `docs/as-built/v96_town-presentation-polish.md` — prior hub ring layout (this slice supersedes positions)

| Option | Decision |
|--------|----------|
| `TownNodeFactory.make_door_node()` for south gate | **Borrow** |
| Procedural wood via `GroundWallFactory` / `WallRenderer` | **Adopt** |
| `town_presentation.v0.json` for center/radius/segment tuning | **Adopt** |
| Runtime circular PCG in sim tick | **Reject** — commit segments or expand once at populate |
| External fence GLBs / night plugins | **Reject** |
| Town fence torches | **Reject** (deferred) |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/town_presentation.v0.json` | `center`, `radius_m`, `segment_count`, `gate_gap_segments`, wood palette hints |
| Create | `shared/assets/town_presentation.v0.schema.json` | Schema for presentation tuning |
| Create | `tools/assets/gen_town_perimeter.py` | Deterministic wall-segment + gate position generator (stdout or patch helper) |
| Modify | `shared/rules/worlds.v0.json` | `dungeon_levels` walls, gate, respaced interactables, player spawn |
| Modify | `shared/rules/worlds.v0.schema.json` | Add `wood` to wall `kind` enum if not present |
| Modify | `shared/rules/interactables.v0.json` | `town_exit_gate` def |
| Modify | `shared/rules/interactables.v0.schema.json` | `locked_exit_reason` optional on closed barrier defs |
| Modify | `server/internal/game/rules.go` | `InteractableDef.LockedExitReason`, validation, `wood` obstacle kind |
| Modify | `server/internal/game/obstacle_blocking.go` | Treat `wood` like `wall` for movement/projectiles |
| Modify | `server/internal/game/interactables.go` | Reject when `LockedExitReason` set and closed |
| Create | `server/internal/game/town_perimeter_test.go` | Gate reject, barrier block, service reachability |
| Modify | `server/internal/game/sim.go` | Optional: `source: "town_perimeter"` on preset walls if snapshot tagging needed |
| Modify | `client/scripts/dungeon_depth_lighting.gd` | Fog suppression when `suppress_for_fog` at town (`level >= 0`) |
| Modify | `client/scripts/main.gd` | Pass town-fog flag to lighting; gate reject feedback; optional rename helper |
| Modify | `client/scripts/client_constants.gd` | `TOWN_EXIT_LOCKED_TEXT` |
| Modify | `client/scripts/wall_renderer.gd` | `wood` kind → palisade material |
| Modify | `client/scripts/ground_wall_factory.gd` | `wall_material_for_level` wood/town perimeter branch |
| Modify | `client/scripts/town_node_factory.gd` | `town_exit_gate` node; refresh preview layout + prop positions |
| Create | `client/tests/test_town_night_lighting.gd` | Fog-suppressed profile at level 0 |
| Modify | `client/tests/test_factories.gd` | Wood wall material smoke |
| Modify | `client/tests/test_coop_client.gd` | `town_exit_locked` feedback unit test |
| Modify | `client/scripts/main.gd` `get_bot_state()` | Expose `last_intent_reject_reason` for bot assertions |
| Modify | `client/scripts/bot_assertion_handlers.gd` | `assert_intent_rejected` step |
| Modify | `client/scripts/bot_step_catalog.gd` | Register new step |
| Create | `tools/bot/scenarios/client/90_town_night_perimeter.json` | Extended client proof |
| Modify | `tools/bot/scenarios/client/18_town_floor_click_to_move.json` | Click stays **inside** fence; add note if coords change |
| Audit | `tools/bot/scenarios/**/*.json` | Fix hardcoded `(16,10)` / vendor-adjacent coords that break after layout |
| Modify | `tools/validate_shared.py` | Cross-check `town_exit_gate` + presentation asset if pattern exists |
| Modify | `scripts/client_smoke.sh` | Register new GDScript test |
| Modify | `docs/CODEMAP.md` | Town perimeter / presentation entries |
| Create | `docs/as-built/v357_town-night-perimeter.md` | On `/finish` |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | On `/finish` |

### Target layout (defaults — adjust only if playtest blocks pathing)

**Center anchor:** `(11, 12)`

| Entity | Position `(x, y)` | Notes |
|--------|-------------------|-------|
| `stairs_down` | `(11, 12)` | Center |
| `teleporter` | `(12, 12)` | Adjacent east of stairs |
| `town_blacksmith` | `(5, 12)` | West ring |
| `town_stash` | `(6, 8)` | NW |
| `town_bishop` | `(16, 8)` | NE |
| `town_quest_giver` | `(11, 5)` | North |
| `town_vendor` | `(20, 12)` | East |
| `town_mystery_seller` | `(18, 17)` | SE |
| `town_market_board` | `(7, 18)` | SW |
| `town_mercenary_board` | `(11, 20)` | South inner |
| `town_unique_chest` | `(8, 10)` | Debug only |
| `town_exit_gate` | `(11, 27)` | South perimeter gap |
| Player spawn | `(11, 18)` | Inside, south of center |

**Perimeter:** radius `15.0`, ~32 tangential `wood` wall segments, 2-segment gap at south (`gate_gap_segments: 2`).

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [ ] `client/scripts/main.gd` (baseline 6232) — feedback + bot state + thin lighting wire only
- [ ] `client/scripts/ground_wall_factory.gd` — wood material branch only
- [ ] `server/internal/game/sim.go` — no new domains inline
- [ ] Other over-limit file: none expected beyond thin edits

Decision:

- [x] Extract perimeter math to `tools/assets/gen_town_perimeter.py` + optional small `town_perimeter_test.go`
- [x] Defer `main.gd` extraction — gate message is ~10 lines in `_handle_intent_rejected`
- [ ] Touched grandfathered files stay at or below baseline

Verification:

```bash
make maintainability
```

## Task 1 — Shared contracts (perimeter, gate, layout)

Files:

- Create: `shared/assets/town_presentation.v0.json`, `town_presentation.v0.schema.json`
- Create: `tools/assets/gen_town_perimeter.py`
- Modify: `shared/rules/worlds.v0.json`, `worlds.v0.schema.json`
- Modify: `shared/rules/interactables.v0.json`, `interactables.v0.schema.json`

- [ ] Step 1.1: Add `town_presentation.v0.json` with `center: {x:11,y:12}`, `radius_m: 15`, `segment_count: 32`, `gate_heading: "south"`, `gate_gap_segments: 2`, `wall_kind: "wood"`.
- [ ] Step 1.2: Implement `gen_town_perimeter.py` — emit `wall` entities (kind `wood`) as tangential segments + `town_exit_gate` at south gap; print JSON fragment for review.
- [ ] Step 1.3: Run generator; merge wall segments + gate into `dungeon_levels.entities`; apply layout table above; set `player.position` to `(11, 18)`.
- [ ] Step 1.4: Add `wood` to `worlds.v0.schema.json` wall `kind` enum; validate segment sizes are positive.
- [ ] Step 1.5: Add `town_exit_gate` to interactables:
  ```json
  "town_exit_gate": {
    "name": "Town Gate",
    "initial_state": "closed",
    "barrier_when_closed": { "size": { "x": 2.0, "y": 0.25 } },
    "locked_exit_reason": "town_exit_locked"
  }
  ```
- [ ] Step 1.6: Extend `interactables.v0.schema.json` — optional `locked_exit_reason` (pattern `^[a-z0-9_]+$`) allowed only when `initial_state` is `closed` and `barrier_when_closed` is present.

```bash
python3 tools/assets/gen_town_perimeter.py
make validate-shared
```

## Task 2 — Server authority (wood walls, locked gate)

Files:

- Modify: `server/internal/game/rules.go`, `obstacle_blocking.go`, `interactables.go`
- Create: `server/internal/game/town_perimeter_test.go`

- [ ] Step 2.1: Add `LockedExitReason string \`json:"locked_exit_reason,omitempty"\`` to `InteractableDef`; validate non-empty reason only on closed defs with barriers.
- [ ] Step 2.2: Map `wood` obstacle kind → same blocking/LOS behavior as `wall` in `obstacle_blocking.go` / navigation helpers.
- [ ] Step 2.3: In `activateInteractable`, before closed→open toggle:
  ```go
  if def.LockedExitReason != "" && e.state == interactableClosed {
      res.reject(in.MessageID, def.LockedExitReason)
      return
  }
  ```
- [ ] Step 2.4: Tests in `town_perimeter_test.go`:
  - `TestTownExitGateRejectsWhileClosed` — `action_intent` → `town_exit_locked`, state stays `closed`
  - `TestTownPerimeterBlocksMovement` — move intent beyond fence → `no_path` or no position change
  - `TestTownServicesReachableFromSpawn` — path exists to `town_vendor` and `stairs_down` from spawn
- [ ] Step 2.5: Optional — tag preset perimeter walls `source: "town_perimeter"` in `populatePresetLevel` when `kind == wood` for client/bot diagnostics.

```bash
cd server && go test ./internal/game/... -run 'TownExit|TownPerimeter|TownService' -count=1
```

## Task 3 — Client night mood (fog lighting + wood fence)

Files:

- Modify: `client/scripts/dungeon_depth_lighting.gd`, `main.gd`
- Modify: `client/scripts/wall_renderer.gd`, `ground_wall_factory.gd`, `town_node_factory.gd`
- Create: `client/tests/test_town_night_lighting.gd`

- [ ] Step 3.1: Change `DungeonDepthLighting.apply_for_level` — apply `apply_fog_suppression` when `suppress_for_fog` is true **regardless of level sign** (town included).
- [ ] Step 3.2: Change `profile_for_level` — when `level >= 0` **and** caller indicates town-fog mode, return dungeon-like baseline palette (depth-1 fallback or dedicated night-town constants in presentation JSON) instead of `TOWN_PROFILE`.
- [ ] Step 3.3: In `main._sync_fog_and_dungeon_lighting`, pass `town_fog_active := current_level == 0 and _lab_world_fog_at_town_level()` into lighting helper (optional rename to `_town_fog_active()`).
- [ ] Step 3.4: `WallRenderer.make_wall_node` — `wood` kind uses warm brown procedural texture via `GroundWallFactory` (palisade height `TOWN_WALL_HEIGHT`).
- [ ] Step 3.5: `TownNodeFactory` — map `town_exit_gate` → `make_door_node()`; update `make_town_preview_scene` positions to match layout table; move campfire to `(11, 13)` or remove from exact center if it occludes stairs/teleporter.
- [ ] Step 3.6: Unit test — when `suppress_for_fog` + level `0`, ambient energy ≤ fog-suppressed dungeon scale.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_town_night_lighting.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_factories.gd
```

## Task 4 — Client gate feedback

Files:

- Modify: `client/scripts/client_constants.gd`, `main.gd`
- Modify: `client/tests/test_coop_client.gd`

- [ ] Step 4.1: Add `TOWN_EXIT_LOCKED_TEXT := "You can't leave for now."` to `client_constants.gd`.
- [ ] Step 4.2: Track `_last_intent_reject_reason` in `main.gd` inside `_handle_intent_rejected`.
- [ ] Step 4.3: On `reason == "town_exit_locked"`, call `_show_damage_number(player_id, …, text_override=TOWN_EXIT_LOCKED_TEXT)` (or inventory-style floating text).
- [ ] Step 4.4: Expose `last_intent_reject_reason` in `get_bot_state()`.
- [ ] Step 4.5: Unit test in `test_coop_client.gd` for message + reason tracking.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_coop_client.gd
```

## Task 5 — Bot scenarios and scenario audit

Files:

- Create: `tools/bot/scenarios/client/90_town_night_perimeter.json`
- Modify: `client/scripts/bot_assertion_handlers.gd`, `bot_step_catalog.gd`
- Audit: town-coordinate scenarios (see list below)

- [ ] Step 5.1: Add `assert_intent_rejected` bot step — checks `get_bot_state()["last_intent_reject_reason"]` equals expected reason (with optional timeout wait wrapper `wait_intent_rejected`).
- [ ] Step 5.2: Create `90_town_night_perimeter.json` (`ci_tier: extended`):
  1. `wait_ws_open`
  2. `wait_wall_layout` `current_level: 0`, `at_least: 24`
  3. `assert_fog_of_war` `enabled: true`, `active: true`, `wall_count_min: 24`
  4. `click_entity` `interactable_def_id: town_exit_gate`
  5. `wait_intent_rejected` / `assert_intent_rejected` `reason: town_exit_locked`
  6. `click_entity` `interactable_def_id: town_vendor` → shop opens or `wait_event` shop-related ack
  7. `click_floor` outside fence (e.g. `x: 11, z: 29`) → `wait_player_near` **fails** or player stays inside (`assert_player_position_unchanged` within timeout) — pick one stable assertion
- [ ] Step 5.3: Audit and fix scenarios that hardcode old vendor coords `(16, 10)` or bishop `(15, 6)` when not using `interactable_def_id` approach:
  - `tools/bot/scenarios/13_teleporter_lab.json`
  - `tools/bot/scenarios/15_character_persistence.json`
  - `tools/bot/scenarios/17_treasure_classes_and_guarded_chests.json`
  - `tools/bot/scenarios/18_character_stats_and_leveling.json`
  - `tools/bot/scenarios/20_dungeon_equipment_drops.json`
  - `tools/bot/scenarios/21_monster_rarity_loot_scaling.json`
  - `tools/bot/scenarios/40_paladin_heal_skill.json`
  - `tools/bot/scenarios/45_town_bishop_respec.json`
  - `tools/bot/scenarios/62_barbarian_earthbreaker.json`
  - Prefer replacing coord moves with `action_entity` / `interactable_def_id` where possible.
- [ ] Step 5.4: Update `18_town_floor_click_to_move.json` — floor click must remain inside fence (e.g. `(11, 16)`).

```bash
HEADLESS=1 make bot-client SCENARIO=90_town_night_perimeter
make bot scenario=town_vendor_gold_sink
make bot scenario=45_town_bishop_respec
```

**CI pack:** keep `90_town_night_perimeter` as **extended**; do not add to `ci_pack.json` unless merge gate lacks town layout coverage.

## Task 6 — Showme and docs (optional polish)

Files:

- Modify: `skills/showme/scripts/visual_capture.gd` (if town framing is off after enclosure)

- [ ] Step 6.1: Refresh `$showme --focus town` camera/framing for perimeter visibility.
- [ ] Step 6.2: Update `docs/CODEMAP.md` with `town_presentation.v0.json`, generator, and gate def.

```bash
python3 skills/showme/scripts/render_focus.py --focus town
```

## Task 7 — Lifecycle docs and CI

- [x] Update `PROGRESS.md` and `docs/progress/slice-lifecycle.md` on `/finish`
- [x] Write `docs/as-built/v357_town-night-perimeter.md`

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -count=1`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-client SCENARIO=90_town_night_perimeter`
- [x] `make maintainability`
- [x] `make ci`

Manual:

```bash
make play
# Town is dark; fence visible; south gate shows lock message; stairs/teleporter centered; services reachable.
```

## Deferred (explicit)

- Opening gate / overworld travel
- Town fence torches and lantern audio
- Renaming `_lab_world_fog_at_town_level` (behavior correct; rename only if clarity wins without churn)
- Promoting `90_town_night_perimeter` into merge CI pack
