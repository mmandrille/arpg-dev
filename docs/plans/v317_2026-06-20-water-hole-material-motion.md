# v317 Plan - Water Hole Material Motion

Status: Ready for implementation
Goal: Add lightweight motion/parallax overlay affordances to water and hole obstacle presentation.
Architecture: Keep the overlays inside `WallRenderer` node construction. Server-owned layout,
pathing, collision, fog, and LoS semantics remain unchanged.
Tech stack: Godot 4 GDScript obstacle renderer and factory unit test.

## Baseline and Shortcut Decision

Builds on v316. Asset/plugin decision: adopt code-native mesh/material overlays; reject external
texture, shader, and particle packages.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/wall_renderer.gd` | Add water shimmer and hole parallax overlay children |
| Modify | `client/tests/test_factories.gd` | Assert overlay nodes on rendered water/hole obstacles |
| Create | `docs/as-built/v317_water-hole-material-motion.md` | Completion proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] None

Decision:
- [x] Keep changes in the existing focused renderer; no large coordinator touched.

Verification:
```bash
make maintainability
```

## Tasks

- [ ] Add shimmer/parallax overlay mesh helpers to `WallRenderer`.
- [ ] Assert overlays in `test_factories.gd`.
- [ ] Update lifecycle docs and as-built proof.

## Verification

- [ ] `godot --headless --path client --script res://tests/test_factories.gd`
- [ ] `make client-unit`
- [ ] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
