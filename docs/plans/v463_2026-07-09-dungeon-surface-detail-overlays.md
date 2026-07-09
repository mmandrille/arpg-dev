# v463 Plan - Dungeon Surface Detail Overlays

Date: 2026-07-09
Spec: `docs/specs/v463_spec-dungeon-surface-detail-overlays.md`
Baseline: v462 `rounded-dungeon-corners`

## Goal

Add a client-only, data-driven detail overlay layer for dungeon floors and walls so generated
dungeon surfaces read with more texture and authored detail without changing any authoritative
geometry or behavior.

## Architecture

Introduce a focused presentation helper that attaches two overlay roots:

- a floor-detail root under `ground_node`
- a wall-detail root under `walls_root`

The helper consumes existing wall layout + entity state plus a small shared presentation config.
`main.gd` remains a thin orchestrator and only calls the helper alongside the existing room-tint
sync.

## Adopt / Borrow / Reject

- Adopt: current procedural wall/floor materials, room-divider segmentation, dungeon rollout
  scenario pattern, and the presentation-helper extraction style used by v462.
- Borrow: none for this slice; external CC0 assets are explicitly deferred.
- Reject: plugins, external kits, runtime downloads, and server-authored placement data.

## Tasks

1. Add shared client-presentation config for floor motifs, wall motifs, and density knobs.
2. Add a loader for the config with safe defaults.
3. Add `dungeon_surface_detail_presentation.gd` to:
   - split dungeon floor space into room-like cells
   - place mixed-scale floor marks per cell
   - place modest wall artifacts on eligible wall segments
   - clear all roots outside dungeon levels
4. Wire the helper into `main.gd` beside `DungeonRoomFloorTint.sync(...)`.
5. Extend `client/tests/test_factories.gd` with focused assertions for:
   - no roots in town
   - floor detail root present on dungeon floors
   - mixed-size floor sign generation
   - wall artifact root present on eligible walls
6. Add `tools/bot/scenarios/client/101_dungeon_surface_detail_overlays.json` using the generated
   wall lab and a stairs-down transition.
7. Run targeted verification and write as-built/progress updates.

## Verification

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=dungeon_surface_detail_overlays
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
make maintainability
```

Visual verification:

```bash
make bot-visual scenario=dungeon_surface_detail_overlays
```

## Notes

- Keep overlay materials alpha-blended and shallow so they read as surface detail, not pickups or
  interactables.
- Reuse existing room-divider segmentation logic rather than introducing authoritative room IDs.
- If file-size pressure appears, prefer extracting a second helper over growing `main.gd`.
