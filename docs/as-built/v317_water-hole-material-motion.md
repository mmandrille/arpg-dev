# v317 As Built - Water Hole Material Motion

Date: 2026-06-20
Spec: [`docs/specs/v317_spec-water-hole-material-motion.md`](../specs/v317_spec-water-hole-material-motion.md)
Plan: [`docs/plans/v317_2026-06-20-water-hole-material-motion.md`](../plans/v317_2026-06-20-water-hole-material-motion.md)

## What Shipped

- `WallRenderer` now adds `WaterMotionBands` overlay children to rendered water obstacles.
- `WallRenderer` now adds `HoleParallaxBands` overlay children to rendered hole/chasm obstacles.
- Existing wall layout normalization, metadata, mesh kinds, collision, fog, and LoS semantics are unchanged.
- `test_factories.gd` asserts both overlay node types on rendered water/hole obstacles.

## Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
make maintainability
```

Result: green on 2026-06-23. Full `make ci` is deferred to the enclosing `$autoloop` batch gate.

## Manual Visual Command

```bash
make bot-visual scenario=78_wall_floor_shader_polish
```

## Deferred

- Imported water textures, shader packages, particle plugins, obstacle generation changes, and
  lighting/camera rebalance remain deferred.
