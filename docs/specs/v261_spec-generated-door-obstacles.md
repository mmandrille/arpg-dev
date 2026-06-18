# v261 Spec - Generated Door Obstacles

Status: Complete
Date: 2026-06-18
Codename: generated-door-obstacles

## Purpose

Add deterministic closed wooden doors to generated dungeon obstacle layouts so some interior wall
segments become player-openable doorway blockers. Doors should be authoritative server entities,
use existing interactable behavior, and be generated from shared dungeon-generation rules.

## Non-goals

- No full room/corridor PCG, secret doors, locked/keyed doors, destructible doors, rotated door
  visuals, door-specific minimap icons, fog/LOS doorway behavior, or boss-floor door generation.
- No client-only decorative doors and no protocol shape change beyond existing entity snapshots and
  deltas carrying generated `wooden_door` interactables.
- No reconnect/resume, database, replay-format, loot, monster, or combat tuning change.

## Acceptance Criteria

- `shared/rules/dungeon_generation.v0.json` owns schema-backed generated-door tuning under
  obstacle generation.
- Non-boss generated dungeon floors can include at least one deterministic closed `wooden_door`
  placed in a generated wall opening when eligible wall segments exist.
- Door generation splits only supported horizontal generated wall segments, leaving two wall pieces
  and a gap occupied by the closed door barrier.
- Generated doors are populated as normal `interactable` entities with `interactable_def_id:
  "wooden_door"` and initial state `closed`.
- Dungeon reachability validation treats generated doors as passable planning points, while live
  simulation still uses the existing closed-door barrier until the player opens the door.
- Existing generated-wall rendering, pathing, projectile, and door-opening behavior remain green.

## Scope and Likely Files

- Shared rules/schema:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
- Server generation/rules:
  - `server/internal/game/dungeon_gen.go`
  - `server/internal/game/dungeon_population.go`
  - `server/internal/game/rules.go`
  - New focused helper files if needed for ratchets.
- Tests/goldens:
  - `server/internal/game/game_test.go`
  - `server/internal/game/dungeon_obstacles_golden_test.go`
  - `shared/golden/dungeon_obstacles.json`
  - `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`
- Docs:
  - `docs/plans/v261_2026-06-18-generated-door-obstacles.md`
  - `docs/as-built/v261_generated-door-obstacles.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external art, plugins, and imported door assets. Borrow the existing
`wooden_door` interactable, barrier, server action, and client presentation.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game`
- `make bot scenario=reachable_dungeon_obstacles`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```

## Open Questions and Risks

- No required questions.
- Risk: door placement can accidentally make generated targets practically unreachable until a door
  is manually opened. This slice should keep validation semantics explicit and add a bot proof that
  generated doors can be opened on the pinned obstacle seed.
