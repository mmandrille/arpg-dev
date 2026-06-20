# v308 As-Built - Wall/Floor Shader Polish

Date: 2026-06-19
Spec: [`docs/specs/v308_spec-wall-floor-shader-polish.md`](../specs/v308_spec-wall-floor-shader-polish.md)
Plan: [`docs/plans/v308_2026-06-19-wall-floor-shader-polish.md`](../plans/v308_2026-06-19-wall-floor-shader-polish.md)

## Shipped

- Added deterministic, palette-aware generated normal maps for dungeon floor materials.
- Kept town ground on the previous simple material path, so the polish is scoped to dungeon floors.
- Added deterministic, palette-aware generated normal maps for cave wall materials.
- Centralized standard cave wall material creation in `GroundWallFactory` while preserving generated,
  perimeter, and preset wall tinting.
- Kept wall, water, hole, obstacle, fog, movement, navigation, protocol, and server behavior
  unchanged.
- Added a headless client visual scenario for generated dungeon floor/wall rendering across a
  stairs-down transition.
- Stabilized final full-CI residuals found by the selected-batch gate: generated obstacle LOS
  metadata pointer identity, generated-door test seed coverage, minimap float-bound tolerance, and
  long-running bot scenario budgets.

## Proof

- `godot --headless --path client --script res://tests/test_factories.gd`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=wall_floor_shader_polish ./scripts/bot_visual.sh`
- `make maintainability`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=dungeon_elite_side_objective ./scripts/bot_local.sh`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=elite_objective_minimap_pin ./scripts/bot_client_local.sh`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=mercenary_recovery_ui ./scripts/bot_client_local.sh`
- `COMPOSE_PROJECT_NAME=arpg-dev make ci`

## Deferred

- Production dungeon art, imported textures/models, Godot shader packages/plugins, and a real asset
  pipeline.
- Lighting/camera rebalance, water/hole/obstacle material expansion, and gameplay visibility or
  collision changes.
- Due review/refactor handoff now that the World Detail/Navigation queue is complete and final
  selected-batch `make ci` is green.
