# v493 As-Built — Town floor detail

- **Date:** 2026-09-30
- **Spec:** [`v493_spec-town-floor-detail.md`](../specs/v493_spec-town-floor-detail.md) · **Plan:** [`v493_2026-09-30-town-floor-detail.md`](../plans/v493_2026-09-30-town-floor-detail.md)
- **Scope:** client presentation + presentation data + nine vendored CC0 KayKit assets. No server,
  protocol or gameplay change.

## What shipped

- **Planner + builders in one module.** `client/scripts/town_ground_detail.gd` (364 lines) turns
  `town_presentation.v0.json` → `dressing` into tile layers (core / rim / edge), service path
  capsules and a deterministic scatter, then builds one `MultiMeshInstance3D` per distinct asset per
  layer. The planner takes explicit `frame` data, so it is unit-tested without a scene.
  `TownDressing` (`town_dressing.gd`, 140 → 93 lines) now only syncs the root and places props.
- **Stone rim** around the plaza with weedy wear tiles (weights 5 plain : 2 weeds).
- **Dirt edge band**, one tile wide, outside the rim (`dressing.edge.enabled`).
- **Stone service path to the vendor** (capsule from the plaza to the service anchor, 3 m wide so
  diagonal paths stay connected). The mystery-seller path is still configured but redundant: plaza
  radius 7.5 m + path half-width 1.5 m already reaches 8.6 m, so only the vendor path is visible.
- **Grass scatter** of rocky dirt patches, weed patches and stone chunks on a jittered 3 m grid,
  excluded around gameplay anchors, props, the plaza and the fence ring. `scatter.radius_m` = 28
  (see tuning below).
- **Nine assets vendored and registered** (seven dirt/weed tile variants incl. rocky, plus small and
  large stone chunks), CC0 KayKit, within the ADR-0018 D8 budgets.
- **Data-driven throughout:** every distance, weight, jitter and scale is in the `dressing` block;
  the new `anchors` key (14 world-preset gameplay points) lets the client avoid gameplay positions
  without reading the server's world preset.
- **Deviations from the spec text, on purpose:** `width_m` distances are Chebyshev between cell
  centres; `path_width_m` is 3.0 (connected diagonals); `anchors` is a new data key.

## What it proved

| Check | Result |
|-------|--------|
| `test_town_ground_detail.gd` (new) | PASS: frame/cell maths, layer partition, path connectivity to the vendor, scatter exclusions (anchors, props, plaza, fence ring), determinism, rock/patch seating, loader pin |
| `test_town_dressing.gd`, `test_item_visuals.gd` | PASS unchanged behaviour: root attaches only in town, removed on dungeon levels, `TownPlaza` name kept |
| `.venv/bin/pytest tools/test_town_dressing.py tools/assets` | 56 passed (anchors == world preset, path targets, manifest ids, fence clearance) |
| `make validate-shared` / `make validate-assets` | OK (2204 checks / 435 checks) |
| `make client-unit` | PASS |
| `make bot-client SCENARIO=15_town_vendor_shop_panel / 23_account_stash_panel / 07_town_teleporter_auto_approach` (HEADLESS=1) | PASS ×3 (4.6 s / 4.7 s / 5.7 s): picking and approach in town are unaffected by the extra ground nodes |
| `make maintainability` | PASS; grandfathered 35 files / 66 115 lines; `town_dressing.gd` 93, `town_ground_detail.gd` 364 |
| `make ci` | PASS first run, 7m09s, no flake (`paladin_class_foundation` included) |

## Outcome vs spec (honest)

- **The soft-edge goal is NOT met.** The KayKit dirt tiles are the same taupe as the stone tiles, so
  the edge band reads as more pavement with a hard, stair-stepped outline (bigger than v491's
  octagon), not as a transition into grass. Net effect on the plaza silhouette: roughly neutral. A
  real blend needs a terrain shader or a nature pack (spec non-goals). The band is kept enabled as
  a neutral-to-marginal choice; `dressing.edge.enabled = false` removes it.
- **Service path to the vendor is the clear win**: the plaza is no longer an isolated disc.
- **Rim wear and weed/pebble detail add life** close up.
- **Grass variation is only partly delivered.** The scatter is sparse, and the grass ring *inside*
  the palisade is still almost empty (the fence clearance and the plaza leave a thin ring): spec
  problem 3 is not fully solved.
- **Scatter radius tuning:** `scatter.radius_m` went 13 → 22 → 28. At 13 the inside-fence ring held a
  single placement; 22 suffices at the default zoom; at max zoom (about 32 m visible from the fence)
  22 visibly thinned out; 28 fills the view and stays on the 140×90 ground plane. The scatter
  deliberately extends past the palisade.

Play-camera evidence (same camera before/after; the palisade is a box stand-in in the scratch
capture):

| Before | After |
|--------|-------|
| ![before](assets/v493/town-before.png) | ![after](assets/v493/town-after.png) |

| Overview | Max zoom |
|----------|----------|
| ![overview](assets/v493/town-overview-after.png) | ![max zoom](assets/v493/town-maxzoom-after.png) |

| Edge close-up | Vendor path close-up |
|---------------|----------------------|
| ![edge](assets/v493/town-edge-closeup.png) | ![path](assets/v493/town-path-closeup.png) |

## Measurements

One-shot headless script (not committed) calling `TownGroundDetail.build(TownPresentationLoader.dressing())`:

| Layer | `MultiMeshInstance3D` | Instances |
|-------|-----------------------|-----------|
| `TownPlaza` (core) | 1 | 26 |
| `TownPlazaRim` | 2 | 37 |
| `TownPlazaEdge` | 5 | 47 |
| `TownScatter` | 4 | 86 |
| **Total** | **12** | **196** |

Build time: 50.9 ms cold (first call, includes loading the glb meshes), 3.4 ms on a second call,
on a developer laptop in headless (dummy renderer) mode. The 12 multimesh nodes are the draw-call
floor for the ground detail (one per distinct asset per layer); the GPU cost was not profiled
(headless).

## Files

- New: `client/scripts/town_ground_detail.gd`, `client/tests/test_town_ground_detail.gd`,
  `shared/assets/town_presentation.v0.schema.json`, nine `.glb` under
  `client/assets/environment/kaykit_dungeon/` and `kaykit_resource_bits/` (with `LICENSE.txt`) and
  their entries in `assets/manifests/assets.v0.json`, `docs/as-built/assets/v493/*.png`.
- Changed: `client/scripts/town_dressing.gd` (delegates; 140 → 93 lines),
  `shared/assets/town_presentation.v0.json` (`anchors`, `plaza.rim`, `edge`, `service_paths`,
  `scatter`), `tools/test_town_dressing.py`, `tools/assets/validate_assets.py`,
  `client/tests/test_town_dressing.gd`, `scripts/client_smoke.sh`, `docs/CODEMAP.md`.

## Known limits and follow-ups

- **Dirt tiles are not a terrain blend** (see outcome). Needs a terrain shader or the KayKit Forest
  Nature pack.
- `kaykit_dungeon_floor_dirt_large_v0` is vendored and registered but unused by data. Kept on purpose
  (the spec acceptance criterion lists nine entries); removal is a separate cleanup.
- Rocks cannot be removed by data alone: `test_town_ground_detail.gd` requires both patch and rock
  kinds.
- The "buried slab" look of the dirt patches depends on `scale_min >= 0.8` at `surface_y` 0.0 and is
  not guarded by a test.
- Weed leaves from rim/edge tiles poke through the vendor / quest-giver / stash bases (cosmetic).
  The stone chunks' grey clashes mildly with the beige kit stone.
- `TownDressing.plaza_cells` / `in_plaza` are dead in production now (kept only for
  `test_town_dressing.gd`).
- **showme's `town` focus** still uses an overview camera (28×22 m ground, no palisade), so it does
  not show what the player sees. The v493 images came from an uncommitted scratch capture with the
  real play camera. Fix in a follow-up.
- **Ground plane:** void past the 140×90 plane is visible from the west/north fence at max zoom
  (v491 issue, not this slice).
- **Next: v494**, random kit props on every dungeon floor (separate spec).
