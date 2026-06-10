# v40 Plan — Reachable Dungeon Obstacles

Status: Implemented
Goal: Add deterministic generated dungeon interior obstacles while proving all generated targets remain reachable.
Architecture: The Go sim remains the sole owner of dungeon PCG, collision, navigation, and replay. Normal generated dungeon floors gain data-driven AABB obstacle groups; boss floors keep the v35 compact arena path. Protocol v3 exposes authoritative current-level static walls through snapshots and a complete `wall_layout_update` delta on level transitions, so Godot renders server-provided layout instead of duplicating generation.
Tech stack: shared JSON rules/schemas/goldens, Go `server/internal/game`, JSON WebSocket protocol, Godot GDScript client, Python protocol and client bots.

## Baseline and shortcut decision

Baseline is v39 `ui-currency-and-mana-polish`, with v9 wall collision, v11 server-owned A* navigation, v18 generated levels/stairs, v21 dungeon monsters, v25 guarded chests, v35 boss floors, and v38+ co-op/session replay already in place.

Godot shortcut decision: reject plugin adoption. v40 only needs simple AABB wall-box rendering from authoritative server layout; no UI framework, camera plugin, asset pack, imported dungeon art, or external collision tool is needed for this proof.

Implementation notes:

- Work stays on the current branch; do not create branches.
- Keep boss floors out of obstacle generation by default.
- Prefer semantic/range tests for density and exact fixtures only for named wall-layout golden contracts.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add obstacle-generation tuning. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate obstacle rules. |
| Add | `shared/protocol/session_snapshot.v3.schema.json` | Add snapshot `walls` layout contract. |
| Add | `shared/protocol/state_delta.v3.schema.json` | Add `wall_layout_update` delta contract. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Include representative wall layout. |
| Modify | `shared/protocol/examples/state_delta.json` | Include representative wall-layout update. |
| Add | `shared/golden/dungeon_obstacles.json` | Pin one deterministic obstacle/reachability fixture. |
| Add | `shared/golden/dungeon_obstacles.v0.schema.json` | Validate the obstacle golden. |
| Modify | `tools/validate_shared.py` | Validate obstacle rules, protocol examples, and golden drift. |
| Modify | `server/internal/game/rules.go` | Parse and validate obstacle-generation rules. |
| Modify | `server/internal/game/dungeon_gen.go` | Generate obstacle groups and validate reachability. |
| Modify | `server/internal/game/pathfind.go` | Reuse or expose path reachability helpers. |
| Modify | `server/internal/game/types.go` | Add wall-layout view and change fields. |
| Modify | `server/internal/game/sim.go` | Populate levels with walls and expose layout in snapshots/deltas. |
| Modify | `server/internal/game/*_test.go` | Cover generation, reachability, wall solidity, and boss exclusion. |
| Modify | `server/internal/replay/*` | Keep replay/timeline wall layout parity if needed. |
| Modify | `server/internal/http/*_test.go` | Keep `/state` wall layout parity if needed. |
| Modify | `client/scripts/main.gd` | Render authoritative wall layout from snapshots and deltas. |
| Modify/Add | `client/tests/*` | Cover wall-layout application/debug state if helper extraction is useful. |
| Modify | `tools/bot/run.py` | Add wall-layout assertions/helpers if needed. |
| Add | `tools/bot/scenarios/28_reachable_dungeon_obstacles.json` | Protocol bot proof. |
| Add | `tools/bot/scenarios/client/14_dungeon_wall_rendering.json` | Client bot proof if reliable. |
| Modify | `PROGRESS.md` | Lifecycle update when v40 ships. |

## Task 1 — Shared Contracts

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Add: `shared/protocol/session_snapshot.v3.schema.json`
- Add: `shared/protocol/state_delta.v3.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Add: `shared/golden/dungeon_obstacles.json`
- Add: `shared/golden/dungeon_obstacles.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add an `obstacle_generation` block with `enabled`, `max_attempts`, group-count range, wall segment range, solid block range, shape weights, and clearance rules.

```bash
make validate-shared
```

- [x] Step 1.2: Extend the dungeon-generation schema to reject impossible dimensions, negative clearances, empty shape weights, and invalid min/max ranges.

```bash
make validate-shared
```

- [x] Step 1.3: Create protocol v3 schemas from v2 and add top-level snapshot `walls[]`, plus a `wall_layout_update` state-delta op carrying a complete `walls[]` replacement for the destination level.

```bash
make validate-shared
```

- [x] Step 1.4: Update protocol examples to include a generated wall and a level-change wall-layout update.

```bash
make validate-shared
```

- [x] Step 1.5: Add `shared/golden/dungeon_obstacles.json` with one pinned seed/level that owns wall count floor, shape-family coverage, stable wall ordering, and reachable generated target positions.

```bash
make validate-shared
```

- [x] Step 1.6: Update `tools/validate_shared.py` to validate obstacle rule semantics and compare the golden against rule-derived expectations without over-locking density tuning.

```bash
make validate-shared
```

## Task 2 — Go Rules and Generator

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/pathfind.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add Go rule structs for obstacle generation and parse them into `Rules.DungeonGeneration`.

```bash
cd server && go test ./internal/game/... -run TestLoadRules
```

- [x] Step 2.2: Add strict rule validation for min/max ranges, positive sizes, nonzero weight totals, and clearances compatible with the configured floor size.

```bash
cd server && go test ./internal/game/... -run 'TestLoadRules|TestInvalid'
```

- [x] Step 2.3: Introduce a stable obstacle RNG stream such as `seed|obstacles|abs(level)` and generate `line`, `l`, `t`, and `block` AABB groups for non-boss levels only.

```bash
cd server && go test ./internal/game/... -run TestDungeonObstacleGeneration
```

- [x] Step 2.4: Keep generated AABBs in deterministic order with stable debug ids/sources and preserve v35 boss-floor wall output.

```bash
cd server && go test ./internal/game/... -run 'TestDungeonObstacleGeneration|TestBossFloorGenerationGolden'
```

- [x] Step 2.5: Add placement clearances so obstacles do not overlap spawn, stairs, teleporter, chest, guaranteed loot, or generated monster spawn points.

```bash
cd server && go test ./internal/game/... -run TestDungeonObstacleGeneration
```

## Task 3 — Reachability and Collision Proofs

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/pathfind.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/game/pathfind_test.go`

- [x] Step 3.1: Add a reachability helper that uses the same grid/corner-cutting assumptions as authoritative auto-pathing.

```bash
cd server && go test ./internal/game/... -run TestPlanPath
```

- [x] Step 3.2: Validate generated floors from the player spawn or arrival marker to stairs up/down, teleporter, chest, guaranteed loot, and every generated monster spawn.

```bash
cd server && go test ./internal/game/... -run TestGeneratedDungeonTargetsReachable
```

- [x] Step 3.3: Retry obstacle generation deterministically on failed connectivity and return a clear error after exhausting attempts.

```bash
cd server && go test ./internal/game/... -run TestGeneratedDungeonUnreachableTuningFailsClearly
```

- [x] Step 3.4: Add tests proving player movement, auto-pathing, monster chase, projectile sweep, loot/drop placement, and travel arrival all treat generated walls as solid.

```bash
cd server && go test ./internal/game/... -run 'TestGeneratedObstacle.*|TestProjectile|TestTravelArrival|TestDrop'
```

- [x] Step 3.5: Update existing dungeon/boss tests only where they intentionally touch generated layout contracts; keep tuning assertions semantic.

```bash
make test-go
```

## Task 4 — Protocol and Server Exposure

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/replay/*`
- Modify: `server/internal/http/*_test.go`
- Modify: `server/internal/game/*_test.go`

- [x] Step 4.1: Add a `WallView` or equivalent type with stable `id`, `position`, `size`, and optional `source`.

```bash
cd server && go test ./internal/game/... -run TestSnapshot
```

- [x] Step 4.2: Include complete current-level wall layout in `Snapshot` for fresh attach, reconnect, `/state`, and replay timeline.

```bash
cd server && go test ./internal/game/... -run 'TestSnapshot|TestReplay|TestState'
```

- [x] Step 4.3: Emit `wall_layout_update` as the first destination-level change for level transitions before entity spawn/update changes are applied client-side.

```bash
cd server && go test ./internal/game/... -run 'TestDungeonLevel|TestTeleporter|TestBossFloor'
```

- [x] Step 4.4: Preserve co-op recipient scoping: each player receives wall layout only for their current level.

```bash
cd server && go test ./internal/game/... -run 'TestCoop|TestSessionBrowser|TestSnapshotForPlayer'
```

- [x] Step 4.5: Verify JSON output validates against the new v3 schemas and replay reconstruction returns identical wall layouts.

```bash
make validate-shared
make test-go
```

## Task 5 — Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Add: `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 5.1: Add bot assertions for wall layout count, generated wall presence, optional source filtering, and "non-perimeter wall exists".

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.2: Add traversal helpers only if existing `use_stair`, `move_until_player_position`, `action_entity`, and `pick_up_loot` cannot prove the scenario cleanly.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.3: Create `28_reachable_dungeon_obstacles.json` with pinned `world_id: dungeon_levels` and a seed that descends to a normal generated floor with interior obstacles.

```bash
make bot scenario=reachable_dungeon_obstacles
```

- [x] Step 5.4: Scenario steps must prove interior walls exist, route around at least one wall to a generated target, interact with or pick up something beyond the obstacle structure, then verify `/state`, reconnect, and replay.

```bash
make bot scenario=reachable_dungeon_obstacles
```

- [x] Step 5.5: Keep existing boss, dungeon, teleporter, projectile, and co-op scenarios green under the new layout.

```bash
make bot
```

## Task 6 — Godot Client Rendering

Files:
- Modify: `client/scripts/main.gd`
- Modify/Add: `client/tests/*`

- [x] Step 6.1: Replace generated-level local perimeter-only rendering with `_render_wall_layout(walls)` driven by snapshot/delta payloads.

```bash
make client-unit
```

- [x] Step 6.2: Keep preset/town wall rendering intact for `worlds.v0.json` worlds and support generated dungeon perimeter plus interior walls from the server.

```bash
make client-unit
```

- [x] Step 6.3: Apply `wall_layout_update` before entity spawn/update changes on level transition and on visual replay playback.

```bash
make client-smoke
```

- [x] Step 6.4: Expose wall layout debug state, such as total wall count and generated wall count, for client bot assertions.

```bash
make client-unit
```

- [x] Step 6.5: Add or update focused client tests for snapshot wall rendering, delta wall replacement, teardown cleanup, and preset world fallback.

```bash
make client-unit
make client-smoke
```

## Task 7 — Client Bot Scenario

Files:
- Add: `tools/bot/scenarios/client/14_dungeon_wall_rendering.json`
- Modify: `client/scripts/bot_controller.gd` or related bot state helpers if needed
- Modify: `client/tests/test_client_bot.gd` if scenario validation changes

- [x] Step 7.1: Add client bot state fields for wall count, generated wall count, and current level if they are not already exposed.

```bash
make client-unit
```

- [x] Step 7.2: Add `14_dungeon_wall_rendering.json` to start `dungeon_levels`, descend to a pinned obstacle floor, and assert non-perimeter generated walls render.

```bash
HEADLESS=1 make bot-client scenario=14_dungeon_wall_rendering
```

- [x] Step 7.3: If headless client bot cannot reliably assert rendered nodes, document the limitation in this plan during execution and replace the scenario with client unit coverage plus `make client-smoke`.

```bash
make client-unit
make client-smoke
```

Execution note: the headless client bot reliably asserted rendered wall-layout state; no fallback replacement was needed.

- [x] Step 7.4: Run the complete client bot catalog.

```bash
make bot-client
```

## Task 8 — Golden, Replay, and Regression Sweep

Files:
- Modify: `shared/golden/dungeon_obstacles.json`
- Modify: `client/tests/test_golden.gd` if the golden is consumed by GDScript
- Modify: `server/internal/game/game_replay_test.go`
- Modify: `tools/replay/*` if replay wrapper assumptions change

- [x] Step 8.1: Add Go golden tests for `dungeon_obstacles.json` covering stable wall order, shape family coverage, and reachability.

```bash
cd server && go test ./internal/game/... -run TestDungeonObstaclesGolden
```

- [x] Step 8.2: Add GDScript golden checks only if the client consumes the obstacle golden directly; otherwise keep client tests focused on protocol wall payload application.

```bash
make client-unit
```

- [x] Step 8.3: Verify replay reconstruction and timeline output include matching wall layout for generated levels.

```bash
make bot scenario=reachable_dungeon_obstacles
make replay SESSION_ID=<session_id_from_bot_output>
```

- [x] Step 8.4: Run targeted existing regression scenarios: dungeon levels, teleporter lab, ranged/projectile lab, boss floor gate, combat control, and session browser.

```bash
make bot scenario=dungeon_levels
make bot scenario=teleporter_lab
make bot scenario=ranged_lab
make bot scenario=boss_floor_gate
make bot scenario=combat_control_and_boss_ai_fixes
make bot scenario=session_browser_uncapped_coop
```

## Task 9 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v40_2026-06-09-reachable-dungeon-obstacles.md` if execution discovers an approved deviation

- [x] Step 9.1: Update `PROGRESS.md` lifecycle table with v40 after implementation is complete.

```bash
rg -n "v40|reachable-dungeon-obstacles|Latest completed slice|Open gaps" PROGRESS.md
```

- [x] Step 9.2: Add the v40 "What each slice proved" summary and any deferred gaps, preserving the explicit non-goals from the spec.

```bash
rg -n "reachable dungeon|obstacle|wall layout|boss floor" PROGRESS.md
```

- [x] Step 9.3: Run full CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot scenario=reachable_dungeon_obstacles`
- [x] `make bot`
- [x] `make bot-client`
- [x] `make ci`

## Deferred scope

- Generated doors in obstacle walls.
- Full room/corridor PCG.
- Rotated/polygon/destructible/secret obstacles.
- Production dungeon art, imported asset packs, lighting, texture, or sound work.
- Boss-floor obstacle generation.
- Durable dungeon map snapshots across fresh sessions.
- Final density, biome, or difficulty balance.
