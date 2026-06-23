# v318 Spec - Dungeon Depth Lighting

Status: Draft
Date: 2026-06-23
Codename: dungeon-depth-lighting

## Purpose

Give dungeon floors more mood by shifting client-only directional and ambient lighting with depth/biome
bands, without changing fog, line-of-sight, or visibility rules.

## Non-goals

- No fog overlay parameter changes, light-radius gameplay changes, protocol, or server changes.
- No imported lighting assets, shader packages, or external plugins.
- No camera rebalance.

## Acceptance Criteria

- Town (`level >= 0`) uses a warm hub lighting profile.
- Shallow and deep dungeon depth bands use distinct data-backed lighting profiles from `biome_palettes`.
- `main.gd` updates lighting when the current level changes via snapshot or `level_changed` events.
- A focused Godot unit test asserts shallow vs deep lighting profiles differ and application updates
  the scene light nodes.

## Scope and Files

- Modify `shared/rules/dungeon_generation.v0.json` and schema with optional palette lighting fields.
- Create `client/scripts/dungeon_depth_lighting.gd`.
- Modify `client/scripts/main.gd`.
- Create `client/tests/test_dungeon_depth_lighting.gd` and register in `scripts/client_smoke.sh`.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_dungeon_depth_lighting.gd
make validate-shared
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=78_wall_floor_shader_polish
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt code-native `DirectionalLight3D` + `WorldEnvironment` tuning;
  reject external lighting assets/plugins.
