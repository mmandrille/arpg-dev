# v300 As-Built - Line-of-Sight Blockers

Date: 2026-06-19
Spec: [`docs/specs/v300_spec-line-of-sight-blockers.md`](../specs/v300_spec-line-of-sight-blockers.md)
Plan: [`docs/plans/v300_2026-06-19-line-of-sight-blockers.md`](../plans/v300_2026-06-19-line-of-sight-blockers.md)

## Shipped

- Added optional `blocks_line_of_sight` metadata to preset world walls and latest v8 wall layout schemas.
- Threaded optional LOS metadata through server preset walls, generated wall obstacles, and authoritative `WallView` snapshots/deltas.
- Preserved legacy behavior: omitted metadata on normal walls still blocks LOS; water, holes, and rubble remain visible-through terrain.
- Marked generated rock and column obstacle variants as tall LOS blockers while keeping rubble as a low solid blocker.
- Extended fog-of-war filtering so tall rock/column blockers hide living monsters inside light radius and emit spawn transitions when the player reaches a clear angle.
- Preserved closed-door occlusion/reveal behavior through the existing interactable barrier path.
- Updated Godot wall normalization and fog overlay filtering so explicit tall blockers feed the shadow-mask occluder layout and low/floor features do not.
- Added `line_of_sight_blocker_lab`, a protocol bot scenario, and a headless client visual scenario for hidden-then-revealed monster and fog-shadow proof.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'FogOfWar|ObstacleVariety|GeneratedObstacleCollisionPaths'`
- `godot --headless --path client --script res://tests/test_fog_of_war_overlay.gd`
- `make client-unit`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blockers ./scripts/bot_local.sh`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=line_of_sight_blocker_shadow ./scripts/bot_visual.sh`
- `make maintainability`

## Deferred

- Minimap occlusion, durable map memory, monster AI fog awareness, stealth, lighting equipment, and LOS combat targeting.
- True polygon, rotated, destructible, or per-piece visibility geometry.
- Production obstacle assets, imported models, wall/floor shader polish, and material art passes.
- Final selected-batch `make ci` and the due review/refactor handoff after the World Detail/Navigation queue completes.
