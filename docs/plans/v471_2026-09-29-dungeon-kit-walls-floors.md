# v471 Plan — Dungeon kit walls and floors (ADR-0018 P2)

Status: Implemented 2026-09-29. See the [as-built](../as-built/v471_dungeon-kit-walls-floors.md).

Goal: dungeon walls, columns and floors render from vendored KayKit Dungeon Remastered pieces
without touching server, protocol, collision or occlusion contracts.

Architecture: kit logic lives in four new client files: a loader, a piece library, a wall builder
and a floor builder. `WallRenderer.make_wall_node` keeps building the `StaticBody3D` and collision,
and swaps only the visual child when the kit is active. Occlusion registration moves from one
`MeshInstance3D` per wall to a list. Asset budgets become a manifest-side data file enforced by
`validate_assets.py`.

Tech stack: Godot 4 client (GDScript), shared presentation JSON, Python asset validator.

Spec: [`v471_spec-dungeon-kit-walls-floors.md`](../specs/v471_spec-dungeon-kit-walls-floors.md)

## Baseline and shortcut decision

- **Reuses:** `WallRenderer` body, collision and occlusion structure; `equipment_visuals.gd`'s
  `runtime_path` → `res://` rule; the v469 `glb_reader`; the v470 `scenes` capture suite.
- **Adopt:** KayKit Dungeon Remastered 1.0.
- **Reject:** `GridMap` and auto-tile addons.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/assets/environment/kaykit_dungeon/*.glb`, `LICENSE.txt` | Vendored kit pieces |
| Modify | `assets/manifests/assets.v0.json`, `assets.v0.schema.json` | `environment` type and entries |
| Create | `assets/manifests/asset_budgets.v0.json` (+ `.schema.json`) | D8 budgets and exemptions |
| Modify | `tools/assets/validate_assets.py`, `tools/assets/test_validate_assets.py` | Budget enforcement |
| Create | `shared/assets/dungeon_kit_presentation.v0.json` (+ schema) | Kit catalog |
| Create | `client/scripts/dungeon_kit_presentation_loader.gd` | Static catalog loader |
| Create | `client/scripts/kit_piece_library.gd` | asset_id → PackedScene / Mesh / AABB (cached) |
| Create | `client/scripts/dungeon_kit_wall_builder.gd` | Wall rectangle → piece runs; column → pillar |
| Create | `client/scripts/dungeon_kit_floor.gd` | Floor tile MultiMeshes with skip rectangles |
| Modify | `client/scripts/wall_renderer.gd` | Kit dispatch, multi-mesh occlusion, legacy skips |
| Modify | `client/scripts/dungeon_surface_detail_presentation.gd` | Skip under kit |
| Create | `client/tests/test_dungeon_kit.gd` | Kit geometry, floor, occlusion, fallback |
| Modify | `client/tests/test_factories.gd`, `scripts/client_smoke.sh` | Kit-semantic wall checks; register test |

## Maintenance ratchet

- [x] No grandfathered file touched (`wall_renderer.gd` is 508 lines, not grandfathered, and must
  stay ≤ 600).
- [x] New files ≤ 600 lines.

```bash
make maintainability
```

## Task 1 — Vendor pieces, manifest, budgets
- [x] Copy `wall`, `wall_half`, `pillar` and `floor_tile_small` (+ `_broken_A`, `_broken_B`,
  `_weeds_A`, `_decorated`) from `.artifacts/kaykit/extracted/.../gltf/*.gltf.glb` to
  `client/assets/environment/kaykit_dungeon/<name>.glb`, together with `LICENSE.txt`.
- [x] Add the `environment` type to the manifest schema. Add entries with provenance and sha256.
- [x] Add the budgets data file and schema. Enforce budgets in `validate_assets.py`, and add tests.
- [x] Run `godot --headless --path client --import` so the import sidecars exist.
```bash
.venv/bin/python tools/assets/validate_assets.py && .venv/bin/pytest tools/assets/test_validate_assets.py -q
```

## Task 2 — Catalog and loader
- [x] `dungeon_kit_presentation.v0.json` plus schema; the loader follows the `ensure_loaded` pattern.
```bash
.venv/bin/python tools/validate_shared.py
```

## Task 3 — Piece library, wall builder, floor
- [x] Add the library, wall builder and floor builder; write `test_dungeon_kit.gd` first.
```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_dungeon_kit.gd
```

## Task 4 — WallRenderer integration and legacy skips
- [x] Kit dispatch for `wall` and `column` in the dungeon.
- [x] Occlusion registration becomes a list.
- [x] Add the floor node to `render_wall_layout`.
- [x] Skip corners and surface details under the kit.
- [x] Update `test_factories.gd`.
```bash
GODOT=godot CLIENT_UNIT_ONLY=1 ./scripts/client_smoke.sh
```

## Task 5 — Captures, docs, CI
- [x] Run the `scenes` suite; curate `docs/as-built/assets/v471/`.
- [x] Update the as-built, the ADR D6 note, PROGRESS, lifecycle and CODEMAP.
```bash
make ci
```

## Deferred
Props and torches (P2b); junction and corner pieces; water and hole art; the town; the v347 perf
floor re-check.
