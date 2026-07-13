# v464 Plan - Surface Material Kit

Status: Ready for implementation
Goal: Add a data-driven client material kit for town ground, dungeon floors, walls, columns, and water.
Architecture: Keep this slice entirely in the client presentation boundary. Shared asset JSON owns
material colors, contrast, scale, and overlay tuning; Godot loaders expose safe defaults; existing
procedural texture generation remains the runtime renderer. A focused little-room capture provides
the requested feedback checkpoint before final finish.
Tech stack: shared JSON/schema, Godot GDScript client, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v463 `dungeon-surface-detail-overlays`, v338 wall/floor rollout, v317 water/hole material
motion, and the existing rounded-room capture script.

Asset/plugin decision:
- Adopt: existing procedural texture generation in `ground_wall_factory.gd`, existing water/hole
  overlay rendering in `wall_renderer.gd`, v463 detail overlays, and focused Godot capture scripts.
- Borrow: no external assets; use shared JSON to parameterize the in-repo algorithms.
- Reject: imported texture packs, shader addons, marketplace kits, and runtime downloads.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/surface_material_presentation.v0.json` | Material style tuning for grass, dungeon stone, columns, and water |
| Create | `shared/assets/surface_material_presentation.v0.schema.json` | Validate the client-only material config |
| Create | `client/scripts/surface_material_loader.gd` | Load config with safe defaults |
| Modify | `client/scripts/ground_wall_factory.gd` | Apply config to ground, wall, column, and procedural texel generation |
| Modify | `client/scripts/wall_renderer.gd` | Apply water overlay tuning and column/water metadata if needed |
| Create | `client/scripts/surface_material_room_capture.gd` | Focused little-room screenshot with floor, walls, columns, water, overlays |
| Modify | `client/tests/test_factories.gd` | Loader/material/capture-facing unit coverage |
| Create | `tools/bot/scenarios/client/102_surface_material_kit.json` | Client visual proof scenario |
| Modify | `docs/CODEMAP.md` | Discoverability for the new material kit |
| Modify | `docs/specs/v464_spec-surface-material-kit.md` | Mark implemented at close-out |
| Create | `docs/as-built/v464_surface-material-kit.md` | As-built proof |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Modify | `PROGRESS.md` | Current status and deferred notes |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd` - not expected to be touched.
- [ ] `server/internal/game/game_test.go` - not touched.
- [ ] `tools/bot/run.py` - not touched.
- [ ] `tools/validate_shared.py` - not touched.
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [ ] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice: `surface_material_loader.gd`
  and `surface_material_room_capture.gd`.
- [ ] Defer extraction with rationale: not expected.

Verification:

```bash
make maintainability
```

## Task 1 - Shared material config

Files:
- Create: `shared/assets/surface_material_presentation.v0.json`
- Create: `shared/assets/surface_material_presentation.v0.schema.json`

- [ ] Step 1.1: Define versioned material styles for `town_ground`, `dungeon_floor`,
  `dungeon_wall`, `dungeon_column`, and `water`, including colors, contrast, uv scale, normal scale,
  and overlay tuning.
- [ ] Step 1.2: Add a schema that constrains colors and numeric ranges.

```bash
make validate-shared
```

## Task 2 - Loader and material wiring

Files:
- Create: `client/scripts/surface_material_loader.gd`
- Modify: `client/scripts/ground_wall_factory.gd`
- Modify: `client/scripts/wall_renderer.gd`

- [ ] Step 2.1: Implement `SurfaceMaterialLoader` with safe defaults and accessors by style id.
- [ ] Step 2.2: Route ground, wall, obstacle/column, and water materials through the loader while
  preserving existing procedural texture generation and biome palette hooks.
- [ ] Step 2.3: Keep v463 floor/wall details rendering above the updated base materials.

```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 3 - Focused tests

Files:
- Modify: `client/tests/test_factories.gd`

- [ ] Step 3.1: Add tests for loader defaults/config values.
- [ ] Step 3.2: Add tests that prove material assignment for town ground, dungeon floor, dungeon
  wall, dungeon column, and water.

```bash
make client-unit
```

## Task 4 - Little-room screenshot checkpoint

Files:
- Create: `client/scripts/surface_material_room_capture.gd`

- [ ] Step 4.1: Build a compact capture scene with dungeon floor, room walls, at least one column,
  water, and v463 surface details.
- [ ] Step 4.2: Generate the requested screenshot and pause for user feedback before final finish.

```bash
godot --headless --path client --script res://scripts/surface_material_room_capture.gd -- --output .artifacts/showme/surface-material-kit-room.png
```

## Task 5 - Client bot visual proof

Files:
- Create: `tools/bot/scenarios/client/102_surface_material_kit.json`

- [ ] Step 5.1: Add a client scenario that reaches a generated dungeon view suitable for visual
  proof of the material kit.
- [ ] Step 5.2: Keep it extended visual proof rather than promoting it into the CI pack.

```bash
HEADLESS=1 make bot-client scenario=surface_material_kit
```

Manual visual check:

```bash
make bot-visual scenario=surface_material_kit
```

## Task 6 - Lifecycle docs and final verification

Files:
- Modify: `docs/CODEMAP.md`
- Modify: `docs/specs/v464_spec-surface-material-kit.md`
- Create: `docs/as-built/v464_surface-material-kit.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [ ] Step 6.1: Update CODEMAP, spec status, as-built notes, lifecycle row, and PROGRESS current
  status/deferred scope.
- [ ] Step 6.2: Run final focused verification, then `make ci` during finish before commit.

```bash
make maintainability
make validate-shared
make client-unit
HEADLESS=1 make bot-client scenario=surface_material_kit
make ci
```

## Final verification

- [ ] `make maintainability`
- [ ] `make validate-shared`
- [ ] `godot --headless --path client --script res://tests/test_factories.gd`
- [ ] `make client-unit`
- [ ] `HEADLESS=1 make bot-client scenario=surface_material_kit`
- [ ] `make ci`

## Deferred scope

- External/CC0 texture assets.
- Depth-specific biome material expansion beyond the existing palette hook.
- Production water shaders and lighting/fog integration.
- Gameplay-visible water/obstacle behavior.
