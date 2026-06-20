# v299 Plan - Obstacle Variety Pack

Status: Complete
Goal: Add rock, column, and rubble obstacle kinds that keep existing solid collision while rendering as distinct dungeon blockers.
Architecture: Extend the existing wall layout `kind` enum and generated solid-obstacle rules instead of adding a new obstacle protocol shape. Server generation chooses one solid visual kind per generated obstacle group, while movement/pathing keeps every solid variant blocked and fog/LOS stays unchanged. The client maps the new kinds to code-native non-rectangular visuals using the existing wall renderer.
Tech stack: shared JSON schemas/rules/goldens, Go deterministic dungeon generation and collision helpers, Godot wall rendering/unit tests, Python protocol bot scenario, SDD docs.

## Baseline and Decisions

Baseline: v298 `barbarian-leap-obstacle-crossing` is committed on `codex/world-detail-navigation` as `d123dd05`.

Autoloop note: final batch `make ci` and the due review/refactor handoff remain deferred until the selected World Detail/Navigation queue completes.

Asset/plugin decision: reject external assets, imported obstacle models, shaders, and Godot addons. Borrow the existing wall renderer, generated wall layout path, water/hole kind metadata pattern, and code-native mesh/material approach.

Protocol decision: extend only the latest v8 wall-kind enum for the existing optional `kind` field. Do not add a new message shape or wall-layout payload version.

LOS decision: rock/column/rubble remain out of fog occluder layout in this slice; the later line-of-sight blocker slice owns tall obstacle perception rules.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add solid obstacle kind weights for generated blockers. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate `solid_kind_weights`. |
| Modify | `shared/rules/worlds.v0.json` | Add `obstacle_variety_lab` preset with rock/column/rubble blockers and reachable loot. |
| Modify | `shared/rules/worlds.v0.schema.json` | Allow preset wall kinds `rock`, `column`, and `rubble`. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow rock/column/rubble in authoritative snapshot walls. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow rock/column/rubble in wall layout updates. |
| Modify | `shared/golden/dungeon_obstacles.v0.schema.json` | Allow solid variety kinds in generated obstacle goldens. |
| Modify | `shared/golden/dungeon_obstacles.json` | Update the focused generated obstacle golden after deterministic kind selection. |
| Modify | `server/internal/game/rules.go` | Load and validate solid kind weights with minimal line growth. |
| Modify | `server/internal/game/obstacle_blocking.go` | Add solid kind constants and projectile blocking semantics. |
| Add | `server/internal/game/dungeon_obstacle_variety.go` | Keep solid-kind selection and validation helpers out of large coordinators. |
| Modify | `server/internal/game/dungeon_doors.go` | Allow generated doors to split solid line variants so existing door proofs remain viable. |
| Modify | `server/internal/game/dungeon_gen.go` | Thread chosen solid kind into existing generated obstacle groups. |
| Modify | `server/internal/game/dungeon_obstacles_golden_test.go` | Count and write solid variety kind proof. |
| Add | `server/internal/game/dungeon_obstacle_variety_test.go` | Focused server tests for generated variety and hard-block semantics. |
| Modify | `client/scripts/wall_renderer.gd` | Render rock/column/rubble with distinct non-rectangular code-native visuals. |
| Modify | `client/tests/test_factories.gd` | Unit proof for new wall renderer kinds and preset lab rendering. |
| Add | `tools/bot/scenarios/101_obstacle_variety_pack.json` | Protocol and visual proof for all three kinds and pathing around them. |
| Modify | `docs/specs/v299_spec-obstacle-variety-pack.md` | Mark complete after proof. |
| Modify | `docs/plans/v299_2026-06-19-obstacle-variety-pack.md` | Track execution checkboxes. |
| Add | `docs/as-built/v299_obstacle-variety-pack.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v299 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep autoloop review/refactor handoff due. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go` is grandfathered; add only the field/call-site needed for validation.
- [x] `server/internal/game/dungeon_gen.go` is grandfathered and close to its allowance; keep edits to signature/field threading only.
- [x] Do not grow `server/internal/game/game_test.go`; add a focused new test file instead.
- [x] `client/scripts/wall_renderer.gd`, `client/tests/test_factories.gd`, and the new Go test/helper files must stay under 600 lines.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Add `dungeon_obstacle_variety.go` and `dungeon_obstacle_variety_test.go` so new logic/proof does not expand large coordinators.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Rules, Schemas, and Lab

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/worlds.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/golden/dungeon_obstacles.v0.schema.json`

- [x] Step 1.1: Add `obstacle_generation.solid_kind_weights` with `wall`, `rock`, `column`, and `rubble`.
- [x] Step 1.2: Extend wall-kind enums in latest protocol schemas, world rules, and dungeon obstacle golden schema.
- [x] Step 1.3: Add `obstacle_variety_lab` with rock/column/rubble blockers and loot reachable only by routing around the blocker strip.

```bash
make validate-shared
```

## Task 2 - Server Generation and Collision Semantics

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/obstacle_blocking.go`
- Add: `server/internal/game/dungeon_obstacle_variety.go`
- Modify: `server/internal/game/dungeon_doors.go`
- Modify: `server/internal/game/dungeon_gen.go`

- [x] Step 2.1: Add solid obstacle kind constants and a `SolidObstacleKindWeights` data type.
- [x] Step 2.2: Validate non-negative solid kind weights and reject all-zero solid kind weights.
- [x] Step 2.3: Choose one solid kind per generated obstacle group using the deterministic dungeon RNG.
- [x] Step 2.4: Preserve existing movement/pathing blocking for rock/column/rubble.
- [x] Step 2.5: Block solid projectiles on rock/column/rubble while leaving fog/LOS helpers unchanged for this slice.
- [x] Step 2.6: Let generated doors split solid line variants so door placement does not depend on rolling only the default wall kind.
- [x] Step 2.7: Ensure flying monsters and Barbarian Leap ignore only water/hole, not the new solid variants.

```bash
cd server && go test ./internal/game -run 'ObstacleVariety|FlyingNavigationTrait|LeapObstacle|MobilityObstacle'
```

## Task 3 - Focused Go and Golden Proof

Files:
- Add: `server/internal/game/dungeon_obstacle_variety_test.go`
- Modify: `server/internal/game/dungeon_obstacles_golden_test.go`
- Modify: `shared/golden/dungeon_obstacles.json`

- [x] Step 3.1: Prove generated production rules emit at least one solid variety kind for a pinned seed.
- [x] Step 3.2: Prove rock/column/rubble are hard blockers for pathing/movement and projectile collision.
- [x] Step 3.3: Prove Leap/flying exceptions do not treat rock/column/rubble as ignorable floor features.
- [x] Step 3.4: Add solid variety counts/kinds to the generated dungeon obstacle golden.

```bash
cd server && go test ./internal/game -run 'ObstacleVariety|DungeonObstacle|GeneratedObstacleCollisionPaths|FlyingNavigationTrait|LeapObstacle|MobilityObstacle'
cd server && go test ./internal/game -run TestDungeonObstaclesGolden
```

## Task 4 - Client Rendering Proof

Files:
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/tests/test_factories.gd`

- [x] Step 4.1: Add rock, column, and rubble render paths with stable node names and `kind` metadata.
- [x] Step 4.2: Use clustered/split code-native meshes so each variant is visibly non-rectangular while staying inside its layout rectangle.
- [x] Step 4.3: Extend factory tests for direct layout rendering and `obstacle_variety_lab`.

```bash
make client-unit
```

## Task 5 - Bot and Visual Proof

Files:
- Add: `tools/bot/scenarios/101_obstacle_variety_pack.json`

- [x] Step 5.1: Add a protocol bot scenario that observes rock, column, and rubble wall kinds.
- [x] Step 5.2: Prove the player can path around the blocker strip to pick up loot.
- [x] Step 5.3: Add visual replay framing for the obstacle variety lab.

```bash
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_local.sh
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_visual.sh
```

## Task 6 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v299_spec-obstacle-variety-pack.md`
- Modify: `docs/plans/v299_2026-06-19-obstacle-variety-pack.md`
- Add: `docs/as-built/v299_obstacle-variety-pack.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 6.1: Mark spec and plan complete after proof.
- [x] Step 6.2: Write as-built summary with shipped obstacle kinds, proof commands, and deferred LOS scope.
- [x] Step 6.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'ObstacleVariety|DungeonObstacle|GeneratedObstacleCollisionPaths|FlyingNavigationTrait|LeapObstacle|MobilityObstacle'`
- [x] `cd server && go test ./internal/game -run TestDungeonObstaclesGolden`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_local.sh`
- [x] `make client-unit`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=obstacle_variety_pack ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [ ] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- True non-rectangular server collision, rotated/polygon/destructible blockers, terrain costs, and
  boss-floor obstacle generation remain out of scope.
- Fog/visibility/minimap occlusion for tall obstacles remains the later line-of-sight blocker slice.
- Production obstacle assets, shaders, imported models, and material polish remain out of scope.
