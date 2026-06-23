# v318 As Built - Dungeon Depth Lighting

Date: 2026-06-23
Spec: [`docs/specs/v318_spec-dungeon-depth-lighting.md`](../specs/v318_spec-dungeon-depth-lighting.md)
Plan: [`docs/plans/v318_2026-06-23-dungeon-depth-lighting.md`](../plans/v318_2026-06-23-dungeon-depth-lighting.md)

## What Shipped

- Extended `biome_palettes` in `dungeon_generation.v0.json` with directional/ambient lighting fields per depth band.
- Added `DungeonDepthLighting` to resolve town vs dungeon profiles and apply them to scene lights.
- `main.gd` now owns `DirectionalLight3D` + `WorldEnvironment` and refreshes lighting on level changes.
- Headless unit test proves shallow vs deep profiles differ and scene nodes receive the profile.

## Proof

```bash
make validate-shared
godot --headless --path client --script res://tests/test_dungeon_depth_lighting.gd
make client-unit
make maintainability
```

Result: green on 2026-06-23. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=78_wall_floor_shader_polish
```

## Deferred

- Fog overlay tuning, light-radius gameplay, imported lighting assets, shader packages, and camera rebalance
  remain deferred.
