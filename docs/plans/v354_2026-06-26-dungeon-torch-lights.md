# v354 Plan: Dungeon Torch Lights

Date: 2026-06-26  
Spec: [`docs/specs/v354_spec-dungeon-torch-lights.md`](../specs/v354_spec-dungeon-torch-lights.md)

## Tasks

- [x] 1.1 Shared dungeon torch presentation JSON + schema
- [x] 1.2 Placement helper + torch light presenter
- [x] 1.3 Fog compositor torch visibility + main.gd sync
- [x] 1.4 Bot scenario 87 + wait/assert handlers + unit tests
- [x] 1.5 Docs: as-built, lifecycle, PROGRESS

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_torch_placement.gd
HEADLESS=1 make bot-client SCENARIO=87_dungeon_torch_lights
```
