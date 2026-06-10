# v17 Plan — Monster chase movement

Status: Ready for implementation (2026-06-06) — **shipped 2026-06-06**

## 1. Goal

Add deterministic, server-authoritative monster chase movement for opt-in monster definitions while
keeping all existing monsters static and preserving replay/resume determinism.

## 2. Baseline and Clarifications Needed

Resolved baseline facts:

- `docs/specs/v17_spec-monster-chase-movement.md` exists and is draft.
- This plan file was created during planning review.
- v16 `use-consumable` is accepted as complete.
- `PROGRESS.md` has been updated to list v16 as the latest completed slice and v17 as planned.
- Active branch is `feature/monster-chase-movement`.

Remaining implementation note:

- Update `PROGRESS.md` again only when v17 ships, replacing planned status with complete
  lifecycle notes.

## 3. Design Decisions Closed During Review

- **Server authority:** monster AI and movement live only in `server/internal/game`.
- **Opt-in behavior:** existing monster defs default to static. Only `training_dummy_chase` moves.
- **Tick order:** inputs, player movement, monster movement, projectiles, tick increment.
- **Pathing:** replan every tick and consume one step.
- **Blocked player:** player is solid for monsters, so chase must select a reachable adjacent goal
  around the player rather than pathing directly to `player.pos`.
- **Speed:** v17 requires chase `move_speed == navigation.cell_size`; fractional speeds are a
  follow-up.
- **Events:** emit `monster_aggro` and `monster_leashed` only on mode transitions.
- **Client:** Godot only renders authoritative position deltas; no LimboAI, NavMesh, or client AI.

## 4. File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.schema.json` | Add behavior fields and schema constraints |
| Modify | `shared/rules/monsters.v0.json` | Add `training_dummy_chase` |
| Modify | `shared/rules/worlds.v0.json` | Add `chase_lab`, `chase_maze`, `leash_lab` |
| Modify | `shared/protocol/state_delta.v0.schema.json` | Require `entity_id` for chase transition events |
| Add | `shared/golden/monster_chase.v0.schema.json` | Golden schema |
| Add | `shared/golden/monster_chase.json` | Pinned chase/leash fixture |
| Modify | `tools/validate_shared.py` | Behavior/golden/world drift guards |
| Modify | `server/internal/game/rules.go` | Parse and validate monster behavior fields |
| Modify | `server/internal/game/sim.go` | Monster AI state, chase movement, collision, events |
| Modify | `server/internal/game/game_test.go` | Golden and behavioral tests |
| Modify | `tools/bot/run.py` | `wait_ticks`, monster movement assertions |
| Add | `tools/bot/scenarios/09_chase_lab.json` | Open-field chase proof |
| Add | `tools/bot/scenarios/10_chase_maze.json` | Maze pathing proof |
| Add | `tools/bot/scenarios/11_leash_lab.json` | Leash/return proof |
| Modify | `client/tools/build_animations.gd` | Add monster `walk` clip generation |
| Modify | `client/animations/monster_anims.tres` | Regenerated animation library |
| Modify | `client/scripts/main.gd` | Monster walk/idle from authoritative position deltas |
| Modify | `client/tests/test_golden.gd` | Monster chase fixture drift check |
| Optional | `tools/bot/scenarios/client/07_monster_chase_idle.json` | Client-bot presentation smoke |
| Modify | `PROGRESS.md` | v17 lifecycle after implementation ships |

## 5. Implementation Tasks

### Task 1 — Shared Rules and Validation

- [ ] Extend `MonsterDef` schema with `behavior`, `aggro_radius`, `leash_radius`, `move_speed`.
- [ ] Add `training_dummy_chase` with `behavior: "chase"`, `aggro_radius`, `leash_radius`, and
  `move_speed: 1.0`.
- [ ] Add `chase_lab`, `chase_maze`, and `leash_lab` worlds using only the chase monster.
- [ ] Add `monster_chase` golden schema and initial fixture.
- [ ] Update `tools/validate_shared.py`:
  - behavior enum validation
  - chase requires positive `aggro_radius`
  - `leash_radius >= aggro_radius`
  - `move_speed == navigation.cell_size`
  - chase worlds reference valid chase monster defs
  - golden references valid worlds/monsters/navigation

Focused check:

```bash
make validate-shared
```

### Task 2 — Server Rule Loading

- [ ] Add typed behavior fields to `server/internal/game/rules.go`.
- [ ] Default missing behavior to static in Go, not by mutating shared JSON.
- [ ] Mirror shared validation in `LoadRules`.
- [ ] Add unit coverage for invalid chase definitions if local test helpers support rule fixtures.

Focused check:

```bash
cd server && go test ./internal/game/... -run Rules
```

### Task 3 — Server Monster AI Movement

- [ ] Extend internal `entity` with `spawnPos Vec2` and `aiMode string`.
- [ ] Set `spawnPos` for monsters when constructing worlds.
- [ ] Add `advanceMonsterMovement(res)` after `applyMovement` and before `advanceProjectiles`.
- [ ] Iterate live monsters by `sortedEntityIDs`.
- [ ] Implement transition logic:
  - idle/return to chase when alive player is within `aggro_radius`
  - chase to return when player is beyond `leash_radius` from spawn
  - return to idle when near spawn
- [ ] Emit `monster_aggro`/`monster_leashed` only on transition edges.
- [ ] Emit `entity_update` when a monster position changes.

Focused check:

```bash
cd server && go test ./internal/game/... -run MonsterChase
```

### Task 4 — Monster Pathfinding and Collision

- [ ] Add `buildMonsterBlockedFn(excludeMonsterID uint64)`.
- [ ] Add `monsterPositionBlocked(pos Vec2, excludeMonsterID uint64)`.
- [ ] Add `resolveMonsterMovement(monster *entity, delta Vec2)`.
- [ ] Reuse wall, closed-door, live-monster, and player collision with stable ordering.
- [ ] Add `findMonsterChaseGoal(monster, player)`:
  - search cells around player with `ringCells`
  - skip blocked cells
  - prefer adjacent/reachable cells
  - call `PlanPath` and use the first reachable candidate
- [ ] Use spawn position directly for return path, with no movement if blocked/unreachable.

Focused check:

```bash
cd server && go test ./internal/game/... -run 'MonsterChase|Collision|Path'
```

### Task 5 — Protocol Bot and Scenarios

- [ ] Add `wait_ticks` action to `tools/bot/run.py`.
- [ ] Track initial monster positions by entity id and monster def id.
- [ ] Track seen events already present in `RuntimeState`.
- [ ] Add assertions:
  - `monster_moved`
  - `monster_within_player_distance`
  - `monster_near_spawn`
  - `event_seen`
- [ ] Add scenarios `09_chase_lab.json`, `10_chase_maze.json`, `11_leash_lab.json`.
- [ ] Verify existing scenarios remain unchanged and green.

Focused check:

```bash
make db-up
make bot
```

### Task 6 — Godot Presentation

- [ ] Add a monster `walk` clip in `client/tools/build_animations.gd`.
- [ ] Regenerate committed animation libraries with `make gen-anims`.
- [ ] In `client/scripts/main.gd`, store previous monster position before each `entity_update`.
- [ ] Call `controller.set_locomotion(position_changed && hp > 0)` for monsters.
- [ ] Preserve priority: death terminal > hit one-shot > walk/idle.
- [ ] Keep server events as the source of hit/death reactions.

Focused check:

```bash
make client-unit
```

### Task 7 — Optional Client Bot Smoke

- [ ] If low-cost after protocol bot is green, add `tools/bot/scenarios/client/07_monster_chase_idle.json`.
- [ ] Use existing client-bot world selection through scenario `world_id`.
- [ ] Add/extend bot state to compare initial and current monster position if needed.

Focused check:

```bash
HEADLESS=1 make bot-client
```

### Task 8 — Docs and Final Verification

- [ ] Update `PROGRESS.md` after v17 implementation ships:
  - current status
  - lifecycle row
  - v17 summary
  - closed deferred monster AI/pathfinding gap
  - v17 non-goals/follow-ups
- [ ] Run full CI once, instead of asking users to run every focused gate separately.

Final gate:

```bash
make ci
```

Optional visual inspection:

```bash
make bot-visual scenario=10_chase_maze.json
```

## 6. Known Risks

- **Goal-cell blocking:** direct pathing to `player.pos` will fail because player is solid. The
  adjacent chase-goal helper is required.
- **Scenario timing:** chase movement may make older bot helpers that assume stationary monsters
  flaky if any existing monster accidentally opts into chase. Keep all old monsters static.
- **Replay drift:** monster movement must not use map iteration, wall-clock time, or unseeded RNG.
- **Projectile order:** projectiles intentionally see post-chase monster positions. Tests should pin
  this order if ranged/chase worlds overlap later.
- **Leash tuning:** if leash worlds use too small a map or blocked spawn cells, the monster can fail
  to return. Tune worlds before pinning golden coordinates.
