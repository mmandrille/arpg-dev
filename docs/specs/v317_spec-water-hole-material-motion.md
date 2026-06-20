# v317 Spec - Water Hole Material Motion

Status: Draft
Date: 2026-06-20
Codename: water-hole-material-motion

## Purpose

Make existing dungeon water and chasm obstacles feel less static by adding lightweight code-native
overlay bands to their client presentation.

## Non-goals

- No obstacle generation, collision, pathing, line-of-sight, fog, protocol, or server changes.
- No shader packages, imported textures, particle plugins, or external assets.
- No lighting/camera rebalance.

## Acceptance Criteria

- Rendered water obstacle nodes include a `WaterMotionBands` overlay child.
- Rendered hole/chasm nodes include a `HoleParallaxBands` overlay child.
- Existing wall layout normalization, metadata, and mesh kinds remain unchanged.
- Existing factory test covers the new overlay nodes.

## Scope and Files

- Modify `client/scripts/wall_renderer.gd`.
- Modify `client/tests/test_factories.gd`.
- Add lifecycle/as-built docs when the slice ships.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
make maintainability
```

Manual visual command:

```bash
make bot-visual scenario=78_wall_floor_shader_polish
```

## Open Questions and Risks

- None. Asset/plugin decision: adopt existing code-native obstacle material rendering; reject
  external textures/shaders/plugins.
