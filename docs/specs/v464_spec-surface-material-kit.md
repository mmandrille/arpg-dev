# v464 Spec - Surface Material Kit

Status: Draft
Date: 2026-07-09
Codename: surface-material-kit
Baseline: v463 `dungeon-surface-detail-overlays`

## Purpose

Improve the readability and richness of the game's major environmental surfaces by turning the
current procedural floor, wall, obstacle, and water treatment into a small data-driven material kit.
Town ground, dungeon floors, dungeon walls, columns, and water should each gain clearer texture
identity while remaining client-only presentation.

This slice also adds a focused "little room" visual capture that renders a compact review scene with
the improved materials, water, columns, walls, and the existing dungeon detail overlays. The capture
is the implementation-time feedback checkpoint before final finish.

## Non-goals

- No server, protocol, collision, pathfinding, or authoritative layout changes.
- No imported texture pack, marketplace plugin, shader addon, or runtime-downloaded art.
- No final production art pipeline, Blender workflow, or remote patcher.
- No biome/depth variant expansion beyond preserving the existing dungeon palette hook.
- No lighting, fog, camera, or full town decoration overhaul.
- No gameplay-visible water or obstacle behavior changes.

## Acceptance criteria

- Town ground renders with a data-configured grass/soil material style instead of one hardcoded green
  procedural recipe.
- Dungeon floors render with a data-configured stone/slab material style that remains compatible
  with v463 floor detail overlays.
- Dungeon walls and columns use related but distinguishable data-configured stone material styles.
- Water surfaces render with a data-configured surface style and visible foam/ripple bands without
  changing blocking or traversal semantics.
- The material parameters live in a shared client-presentation asset file with schema validation.
- Existing procedural generation remains deterministic and local to the client; no network or server
  state is added.
- Focused client tests prove material config loading and material assignment for town ground, dungeon
  floor, wall/column, and water paths.
- A focused little-room screenshot script produces a PNG that includes floor, walls, columns, water,
  and dungeon surface details for implementation-time feedback.
- A client bot scenario remains available for visual verification with `make bot-visual`.

## Scope and likely files

- Shared client-presentation config:
  - `shared/assets/surface_material_presentation.v0.json`
  - `shared/assets/surface_material_presentation.v0.schema.json`
- Client presentation:
  - `client/scripts/surface_material_loader.gd`
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/wall_renderer.gd`
  - `client/scripts/rounded_dungeon_corner_capture.gd` or a new focused capture script
  - `client/tests/test_factories.gd`
- Bot / visual proof:
  - `tools/bot/scenarios/client/102_surface_material_kit.json`
- Docs:
  - `docs/plans/v464_2026-07-09-surface-material-kit.md`
  - `docs/as-built/v464_surface-material-kit.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Asset / plugin decision

- Adopt: existing in-repo procedural material generation in `ground_wall_factory.gd`, water/hole
  surface overlays in `wall_renderer.gd`, v463 dungeon surface detail overlays, and the focused
  rounded-room capture pattern.
- Borrow: no external assets for v464; use shared JSON to parameterize the existing procedural
  texture algorithms.
- Reject: imported texture packs, Godot shader addons, marketplace dungeon kits, and runtime asset
  downloads for this slice.

## Test and bot proof

```bash
make validate-shared
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=surface_material_kit
```

Implementation-time visual feedback checkpoint:

```bash
godot --headless --path client --script res://scripts/surface_material_room_capture.gd -- --output .artifacts/showme/surface-material-kit-room.png
```

Expected manual visual verification command:

```bash
make bot-visual scenario=surface_material_kit
```

## Open questions and risks

- Screenshot checkpoint: the implementation should pause after the focused little-room PNG is
  generated so the user can review the screenshot before final finish.
- Maintainability risk: `ground_wall_factory.gd` and `wall_renderer.gd` already own the relevant
  material creation paths, so keep loader/config code separate and avoid growing coordinators more
  than necessary.
- Visual risk: stronger textures can become noisy under gameplay entities. Default to moderate
  contrast and prove v463 overlays still read on dungeon floors.
