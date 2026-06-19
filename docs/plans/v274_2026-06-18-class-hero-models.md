# v274 Plan: Class Hero Models

## Context

The supplied barbarian, paladin, rogue, and sorcerer GLBs are static mesh models: no skins and no embedded animations. The current class presentation contract assumes generated skinned humanoids with hand bones. This plan keeps the class asset IDs stable, makes static character support explicit, and preserves equipment sockets through fallback `Node3D` sockets.

## Tasks

- [x] Replace the four runtime class GLBs and update manifest provenance.
- [x] Extend class presentation metadata with optional `scale` and `height_offset`; set a paladin scale correction so the small authored model renders at hero size.
- [x] Apply class presentation transform metadata when swapping `ModelRoot`.
- [x] Add static hand socket fallback support to `character_visual.gd`.
- [x] Relax asset validation only for character entries that explicitly declare `required_nodes: []`, and add validator tests proving rigged entries still require hand bones.
- [x] Update the Godot animation/class model smoke to accept static class meshes and assert sockets after class replacement.
- [x] Verify with focused asset/client tests and record the as-built notes.

## Verification

- `.venv/bin/python -m pytest tools/assets/test_validate_assets.py`
- `make validate-assets`
- `godot --headless --path client --script res://tests/test_animation.gd`
- `make client-unit`
- Manual visual check available: `make bot-visual scenario=20_menu_create_join_flow`
