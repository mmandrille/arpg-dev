# v354 Spec: Dungeon Torch Lights

Status: Complete  
Date: 2026-06-26  
Codename: `dungeon-torch-lights`  
Baseline: v353 `mobility-skill-smoothing`

## Purpose

Add client-only perimeter wall torches in dungeon levels: emissive flame meshes plus local
OmniLight3D circles, with fog compositor punch-through so torches remain visible outside the
hero light radius.

## Non-goals

- No server, protocol, gameplay `light_radius`, or golden changes.
- No torch placement on generated interior obstacles (perimeter only in v354).

## Acceptance criteria

- `dungeon_torch_presentation.v0.json` owns spacing, counts, colors, and light radii.
- Perimeter wall layout drives deterministic torch mount points (capped at 8 for fog shader).
- Fog compositor combines hero visibility with torch visibility in isometric mode.
- Bot debug exposes `dungeon_torch_lights` and fog `torch_count`.
- Extended client bot scenario descends one floor and asserts torches are active.

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_torch_placement.gd
HEADLESS=1 make bot-client SCENARIO=87_dungeon_torch_lights
```
