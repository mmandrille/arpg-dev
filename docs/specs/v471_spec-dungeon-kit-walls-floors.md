# v471 — Dungeon kit walls and floors (ADR-0018 P2)

- **Status:** Implemented (v471)
- **Date:** 2026-09-29
- **Codename:** `dungeon-kit-walls-floors`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D2, D6, D8
- **Baseline:** v470 render baseline (`4d3d4e28`); `make ci` green per the owner

## Purpose

Dungeon floors (level < 0) stop rendering as `BoxMesh` walls and pixel-painted planes. They render
from vendored **KayKit Dungeon Remastered 1.0** pieces instead:

- **Walls:** each server wall rectangle (`kind` absent / `wall`) becomes a run of kit `wall` (4 m)
  and `wall_half` (2 m) pieces along its long axis.
  - The last partial segment is a half piece stretched along its length.
  - Walls thicker than the kit piece depth are built from parallel rows.
  - Kit pieces are natively 4 tall × 1 deep, which matches `ceiling_height` 4.0 and
    `wall_thickness` 1.0. That means **scale 1.0**, not the 0.5 from the v469 findings, so the
    bricks keep their authored proportions.
- **Columns** (`kind: column`) render as the kit `pillar`, scaled to the rectangle and wall height.
- **Floors** become a deterministic 2 × 2 kit floor-tile grid.
  - It is drawn with one `MultiMeshInstance3D` per tile variant, with weighted variants chosen by
    cell hash.
  - Cells under walls, holes and water are skipped, so pits and water keep their existing
    presentation.

The **server, protocol, collision, reachability and wall-occlusion contracts do not change.**
- Every wall keeps its `StaticBody3D` + `BoxShape3D` and its `wall_id` / `source` / `kind` metadata.
- Every wall keeps its per-wall occlusion registration. Occlusion now fades all kit meshes of the
  wall.
- The town (level ≥ 0), holes, water, rocks, wood palisades and the ceiling are unchanged.

## Why per-wall tiling instead of the ADR's global occupancy auto-tiler

ADR-0018 D6 sketched rasterizing into a grid and auto-tiling. Tiling each wall rectangle instead
keeps the **per-wall identity** that occlusion fade, picking blocks and debug state are keyed on,
and it needs no junction solver. v469's audit showed that walls already meet on a 0.5/1.0 grid.
Corner and junction pieces stay deferred until captures show seams. The ADR is updated to record
this.

## Non-goals

- Torches, doors, chests, stairs, barrels and other dungeon props (P2b).
- Hero, monster and weapon kit models (P3, P4).
- Wall junction or corner kit pieces, and end caps.
- Water and hole art.
- Town walls and props.
- Any server snap rule.

## Acceptance criteria

1. **Vendored assets.** The selected kit GLBs live under `client/assets/environment/kaykit_dungeon/`
   with the pack's `LICENSE.txt`. Each has a manifest entry with the new `type: environment` and
   provenance (origin, pack version, commit, source URL, `CC0-1.0`, sha256). `make validate-assets`
   passes.
2. **Budgets enforced (ADR-0018 D8).**
   - `assets/manifests/asset_budgets.v0.json` plus its schema define triangle and texture budgets
     per asset type.
   - `validate_assets.py` fails any asset over budget unless it is listed in `exemptions` with a
     reason.
   - Current over-budget assets are exempt with named reasons: the two AI monsters (P4) and the two
     ring meshes (D5 removes ring world visuals).
3. **Catalog.** `shared/assets/dungeon_kit_presentation.v0.json` plus schema holds:
   - the enable flag and the minimum dungeon depth where the kit applies
   - wall full/half piece asset ids
   - the pillar asset id
   - weighted floor variants
   - which legacy v462/v463 presentations to disable under the kit

   Piece dimensions are **measured from the imported mesh AABB at runtime**, not duplicated in data.
4. **Walls.** For a dungeon wall rectangle of length L:
   - The kit pieces cover [−L/2, L/2] along the long axis, with no gap larger than epsilon.
   - Pieces stay inside the rectangle's footprint.
   - The collision body is unchanged: same `BoxShape3D` size as before.
5. **Occlusion.** `apply_occlusion_fades` fades every kit mesh of a faded wall and restores it when
   the wall is no longer faded. The existing `test_wall_occlusion_fade.gd` still passes.
6. **Floors.**
   - Floor tiles cover the perimeter-bounded floor area.
   - No tile center falls inside a wall, hole or water rectangle.
   - The variant choice is deterministic for the same layout and level.
7. **Legacy presentations.** Under the kit, the v462 rounded-corner cylinders and the v463
   surface-detail overlays are not built for dungeon levels, as each catalog flag says.
8. **Capture.** `make regen-screenshots SUITE=scenes` passes 9/9. The `dungeon-room-*` captures
   show kit walls, pillars and tiles under the v470 lighting, and the as-built records them against
   the v470 baseline.
9. **Existing gates.**
   - The client unit suite passes; `test_factories.gd` wall expectations move to kit-semantic
     checks.
   - Client bot wall scenarios (14, 78, 79, 101, 102) are unaffected, because they assert layout
     data.
   - `make ci` passes.

## Scope and files

| Area | Files |
|------|-------|
| Assets | `client/assets/environment/kaykit_dungeon/*.glb` (+ `.import`, texture sidecars), `LICENSE.txt`; `assets/manifests/assets.v0.json`, `assets.v0.schema.json` (`environment` type); new `assets/manifests/asset_budgets.v0.json` (+ schema) |
| Tools | `tools/assets/validate_assets.py` (budgets, environment type), `tools/assets/test_validate_assets.py` |
| Shared | new `shared/assets/dungeon_kit_presentation.v0.json` (+ schema) |
| Client | new `client/scripts/dungeon_kit_presentation_loader.gd`, `client/scripts/kit_piece_library.gd`, `client/scripts/dungeon_kit_wall_builder.gd`, `client/scripts/dungeon_kit_floor.gd`; `client/scripts/wall_renderer.gd` (kit dispatch, multi-mesh occlusion), `client/scripts/dungeon_surface_detail_presentation.gd` (kit skip) |
| Tests | new `client/tests/test_dungeon_kit.gd` (registered in `scripts/client_smoke.sh`); `client/tests/test_factories.gd` |
| Docs | ADR-0018 D6 note, as-built, PROGRESS, lifecycle, CODEMAP |

`wall_renderer.gd` (508 lines, not grandfathered) must stay ≤ 600; the kit logic lives in the new
files.

### Asset/plugin decision

- **Adopt** KayKit Dungeon Remastered 1.0 (CC0; ADR-0018 D1/D2), staged and verified in v469.
- **Reuse** the manifest `runtime_path` → `res://` rule from `equipment_visuals.gd` and the existing
  `WallRenderer` body/occlusion structure.
- **Reject** Godot `GridMap` (per-wall occlusion identity) and third-party auto-tile addons (D3).

## Test proof

- `test_dungeon_kit.gd` (headless, node-render component test):
  - pieces cover the wall length
  - thick walls produce multiple rows
  - pillar footprint matches the rectangle
  - floor cells skip wall, hole and water rectangles
  - variant selection is deterministic
  - occlusion fade reaches every kit mesh
  - catalog-disabled mode falls back to legacy `BoxMesh` walls
- `test_validate_assets.py`: a budget violation fails, exemptions pass, and an `environment` entry
  validates.
- Visual: the `scenes` suite; captures go in `docs/as-built/assets/v471/`.
- No bot scenario changes: no gameplay, protocol or movement changes.

## Risks

- **Draw cost.** Perimeter walls of 100–120 m become roughly 30 pieces each. The expected total is
  a few hundred mesh instances per level, plus floor MultiMeshes with about 1–2k instances. That's
  fine for Forward+, but the v347 performance floor should be re-checked.
- **Headless unit tests** run `gl_compatibility`. GLB import sidecars must exist
  (`godot --headless --import`) before the tests load kit scenes.
