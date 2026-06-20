# v307 Plan - Line-of-Sight Blockers

Status: Complete
Goal: Add tall obstacle LOS metadata so rock/column blockers affect fog visibility and client shadow masks.
Architecture: Extend existing wall layout views with optional `blocks_line_of_sight` metadata instead of adding a new obstacle protocol shape. The server keeps authoritative fog filtering and treats explicit tall blockers plus legacy walls as LOS occluders. The Godot fog overlay consumes the same normalized layout metadata for presentation-only shadow masks.
Tech stack: shared JSON schemas/rules, Go sim fog filtering and deterministic dungeon layout, Godot wall/fog rendering tests, Python protocol bot, Godot client bot, SDD docs.

## Baseline and Shortcut Decision

Baseline: v306 `obstacle-variety-pack` is committed on `codex/world-detail-navigation` as `6c81a4c9`.

Autoloop note: final batch `make ci` and the due review/refactor handoff remain deferred until the selected World Detail/Navigation queue completes.

Asset/plugin decision: reject imported assets, shader packages, and Godot addons. Borrow the existing wall renderer, fog overlay shadow-mask path, closed-door occluder sync, and protocol/client bot fog assertions.

Protocol decision: extend only the latest v8 wall view schemas with optional `blocks_line_of_sight`. No new message op, payload version, or replay input shape.

Rules decision: do not add a new `dungeon_generation` tuning field in this slice because `rules.go` is already at its maintainability allowance. Touch `rules.go` only for the preset-world parser field. Treat rock and column as tall LOS-capable obstacle kinds in focused helpers; rubble remains low.

## Spec Review

Spec OK. It builds on v306, keeps server authority for visibility, lists the shared schema change, includes concrete preset and bot proof, and keeps movement/projectile/flying/Leap behavior out of scope. No blocking questions.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/worlds.v0.schema.json` | Allow optional preset wall LOS metadata. |
| Modify | `shared/rules/worlds.v0.json` | Add `line_of_sight_blocker_lab` with tall column/rock blockers and hidden/reveal monster setup. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow optional wall-view LOS metadata in snapshots. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow optional wall-view LOS metadata in wall-layout deltas. |
| Modify | `server/internal/game/obstacle_blocking.go` | Route LOS blocking through explicit metadata plus legacy wall defaults. |
| Modify | `server/internal/game/dungeon_obstacle_variety.go` | Add focused solid-kind LOS helper for generated rock/column. |
| Modify | `server/internal/game/dungeon_gen.go` | Mark generated rock/column obstacles as LOS blockers. |
| Modify | `server/internal/game/rules.go` | Parse optional preset wall LOS metadata. |
| Modify | `server/internal/game/sim.go` | Store preset/generated LOS metadata and emit it in wall views. |
| Modify | `server/internal/game/types.go` | Add optional `blocks_line_of_sight` to `WallView` while offsetting line growth. |
| Modify | `server/internal/game/fog_of_war_test.go` | Prove tall rock/column occlusion, rubble/water/hole transparency, and reveal transitions. |
| Modify | `server/internal/game/dungeon_obstacle_variety_test.go` | Update hard-blocker semantics now that rock/column are LOS blockers. |
| Modify | `client/scripts/wall_renderer.gd` | Preserve `blocks_line_of_sight` in normalized layout rows. |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Include explicit LOS blockers and legacy walls in shadow-mask occluders. |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Prove rock/column metadata shadows and rubble/water/hole transparency. |
| Modify | `client/tests/test_factories.gd` | Prove layout normalization preserves LOS metadata. |
| Add | `tools/bot/scenarios/102_line_of_sight_blockers.json` | Protocol bot proof for hidden-then-revealed monster. |
| Add | `tools/bot/scenarios/client/77_line_of_sight_blocker_shadow.json` | Visual client bot proof for fog shadow mask from tall blocker layout. |
| Modify | `docs/specs/v307_spec-line-of-sight-blockers.md` | Mark complete after proof. |
| Modify | `docs/plans/v307_2026-06-19-line-of-sight-blockers.md` | Track execution checkboxes. |
| Add | `docs/as-built/v307_line-of-sight-blockers.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v307 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep autoloop handoff due. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go` is grandfathered; add only LOS field threading and wall-view emission.
- [x] `server/internal/game/types.go` is at its +25 allowance; offset the `WallView` field addition by removing one existing blank/comment line.
- [x] `server/internal/game/rules.go` is at its allowance; add only the preset-world field and keep the file within allowance.
- [x] `server/internal/game/dungeon_gen.go` is grandfathered; keep edits to generated obstacle metadata.
- [x] `tools/bot/run.py` is over-limit but should not be touched; use existing entity-count assertions.
- [x] Client files touched by this slice must stay under 600 lines.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Use existing focused helper/test files; do not grow large rules or bot coordinators.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Schemas and Preset Lab

Files:
- Modify: `shared/rules/worlds.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`

- [x] Step 1.1: Add optional `blocks_line_of_sight` boolean to preset and latest protocol wall schemas.
- [x] Step 1.2: Add `line_of_sight_blocker_lab` with player, tall column/rock occluders, low rubble/water/hole controls, and an occluded monster that can be revealed from a clear angle.

```bash
make validate-shared
```

## Task 2 - Server LOS Semantics

Files:
- Modify: `server/internal/game/obstacle_blocking.go`
- Modify: `server/internal/game/dungeon_obstacle_variety.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`

- [x] Step 2.1: Add `blocksLineOfSight` to internal wall obstacles and optional `blocks_line_of_sight` to wall views.
- [x] Step 2.2: Preserve legacy behavior: omitted metadata on kind `wall` blocks LOS; water/hole/rubble do not.
- [x] Step 2.3: Mark generated rock and column obstacles as LOS blockers when emitted by dungeon generation.
- [x] Step 2.4: Thread preset `blocks_line_of_sight` into authoritative walls and wall views.

```bash
cd server && go test ./internal/game -run 'FogOfWar|ObstacleVariety'
```

## Task 3 - Server Fog Tests

Files:
- Modify: `server/internal/game/fog_of_war_test.go`
- Modify: `server/internal/game/dungeon_obstacle_variety_test.go`

- [x] Step 3.1: Prove rock and column blockers hide monsters inside light radius when marked as LOS blockers.
- [x] Step 3.2: Prove rubble, water, and holes remain visible-through.
- [x] Step 3.3: Prove moving to a clear angle emits a spawn transition for a previously occluded monster.
- [x] Step 3.4: Update v306 obstacle-variety assertions to reflect rock/column now block LOS while rubble does not.

```bash
cd server && go test ./internal/game -run 'FogOfWar|ObstacleVariety|GeneratedObstacleCollisionPaths'
```

## Task 4 - Client Fog and Layout Proof

Files:
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/scripts/fog_of_war_overlay.gd`
- Modify: `client/tests/test_fog_of_war_overlay.gd`
- Modify: `client/tests/test_factories.gd`

- [x] Step 4.1: Preserve `blocks_line_of_sight` in normalized wall layout rows.
- [x] Step 4.2: Update the fog overlay to include explicit LOS blockers plus legacy walls and skip explicit false/non-tall floor features.
- [x] Step 4.3: Extend tests for rock/column LOS shadows, rubble transparency, and normalized metadata.

```bash
godot --headless --path client --script res://tests/test_fog_of_war_overlay.gd
make client-unit
```

## Task 5 - Bot and Visual Proof

Files:
- Add: `tools/bot/scenarios/102_line_of_sight_blockers.json`
- Add: `tools/bot/scenarios/client/77_line_of_sight_blocker_shadow.json`

- [x] Step 5.1: Add a protocol bot scenario with a `fog_of_war` seed that first asserts the occluded monster is absent.
- [x] Step 5.2: Move the player to a clear angle and assert the monster becomes visible.
- [x] Step 5.3: Add a headless client visual scenario that asserts wall layout and fog shadow counts for tall blockers.

```bash
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blockers ./scripts/bot_local.sh
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blocker_shadow ./scripts/bot_visual.sh
```

## Task 6 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v307_spec-line-of-sight-blockers.md`
- Modify: `docs/plans/v307_2026-06-19-line-of-sight-blockers.md`
- Add: `docs/as-built/v307_line-of-sight-blockers.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 6.1: Mark spec and plan complete after proof.
- [x] Step 6.2: Write as-built summary with shipped LOS metadata, proof commands, and deferred scope.
- [x] Step 6.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'FogOfWar|ObstacleVariety|GeneratedObstacleCollisionPaths'`
- [x] `godot --headless --path client --script res://tests/test_fog_of_war_overlay.gd`
- [x] `make client-unit`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blockers ./scripts/bot_local.sh`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blocker_shadow ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [ ] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- Minimap occlusion, durable map memory, monster AI fog awareness, stealth, lighting equipment, and line-of-sight combat targeting.
- True polygon, rotated, destructible, or per-piece visibility geometry.
- Production obstacle assets, imported models, wall/floor shader polish, and material art passes.
