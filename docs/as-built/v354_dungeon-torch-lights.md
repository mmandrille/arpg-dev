# v354 As-Built — Dungeon Torch Lights

Date: 2026-06-26  
Spec: [`docs/specs/v354_spec-dungeon-torch-lights.md`](../specs/v354_spec-dungeon-torch-lights.md)  
Plan: [`docs/plans/v354_2026-06-26-dungeon-torch-lights.md`](../plans/v354_2026-06-26-dungeon-torch-lights.md)

## Shipped behavior

- **`dungeon_torch_presentation.v0.json`**: spacing, max count, colors, omni/fog radii.
- **`DungeonTorchPlacement`**: derives perimeter torch mounts from wall layout (max 8).
- **`DungeonTorchLights`**: emissive flame meshes + OmniLight3D per torch; syncs fog punch-through.
- **Fog compositor**: combines hero and torch visibility in isometric mode; exposes `torch_count`.
- **Extended bot proof**: `87_dungeon_torch_lights` (descend + assert torches/fog).

## Boundaries

- Client-only; gameplay visibility formulas unchanged.
- Interior obstacle torches deferred.

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_torch_placement.gd
HEADLESS=1 make bot-client SCENARIO=87_dungeon_torch_lights
```
