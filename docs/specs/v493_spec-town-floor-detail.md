# v493 — Town floor detail: soft plaza edge, paths to services, grass variation (ADR-0018 follow-up)

- **Status:** Draft
- **Date:** 2026-09-30
- **Codename:** `town-floor-detail`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D6/D10
- **Baseline:** v492 (`e8973608`)
- **Spec gate:** client presentation + presentation data only; exempt, written anyway (new manifest
  entries and a data schema change).
- **Slice order:** first of two. v494 (random kit props on every dungeon floor) is a separate spec.

## Purpose

v491 gave the town a paved plaza and a road, but the capture (`docs/as-built/assets/v491/town-after.png`)
still shows three problems:

1. **Hard edge.** The plaza is a stone octagon cut straight into flat green ground. There is no
   transition, so it reads as a texture pasted on a plane.
2. **Stranded services.** The vendor (20, 12) and mystery seller (18, 17) sit on bare grass outside
   the 7.5 m plaza, and nothing connects them to it. Only the gate road exists.
3. **Empty grass.** Everything outside the plaza is one uniform noise texture.

After this slice the live town (and the town capture, which shares the code path) gets:

1. **A soft plaza edge.** A band of KayKit dirt tiles (`floor_dirt_small_A/B/C/D`, `_weeds`) one
   tile wide around the plaza ∪ road ∪ service paths, and a stone rim inside the boundary where a
   share of the tiles swap to the intact decorated/weedy stone variants.
2. **Paths to services.** A stone path of `service_path_width_m` from the plaza to each service
   listed in data (the vendor and mystery seller today), using the same tile grid as the plaza.
3. **Grass variation.** Deterministic scatter of large dirt patches (`floor_dirt_large`,
   `floor_dirt_large_rocky`, `floor_dirt_small_weeds`) and stone chunks (Resource Bits
   `Stone_Chunks_Small/Large`) over the grass out to `scatter_radius_m`, kept clear of gameplay
   positions, the plaza, paths, edge band and the v491 props.

All of it hangs under the existing `TownDressing` root, so it inherits the ground-offset
compensation and the removal on every non-town level.

## Non-goals

- Trees, bushes, buildings, terrain shaders, grass blades: they need the Forest Nature / Medieval
  Hexagon packs, which are not downloaded (see the v491 follow-up packs).
- Wood-plank zones under stalls (deferred; not picked for this slice).
- Any dungeon change (v494), collision for town dressing, lighting change, world-preset change,
  server or protocol change.
- Moving or re-laying the v491 props. They are inputs to the scatter exclusion only.

## Design

Data-driven per the Data-Driven Configuration Policy: every count, width and weight lives in
`shared/assets/town_presentation.v0.json` under `dressing`; no tuning value is hardcoded in GDScript.

```
dressing.plaza            (existing) + rim: { width_m, tile_variants[] }
dressing.edge             { enabled, width_m, surface_y, tile_variants[{asset_id, weight}] }
dressing.service_paths    { path_width_m, targets[{ interactable_def_id, position{x,y} }] }
dressing.scatter          { radius_m, cell_m, min_clearance_m, patches{...}, rocks{...} }
```

- **Region model.** Plaza disc ∪ gate road ∪ service-path capsules is one "paved region"
  (`in_plaza` generalised to a capsule list; the existing function stays the disc/road case).
  Edge band = grid cells not paved but within `edge.width_m` of the region. Rim = paved cells
  within `rim.width_m` of the boundary. Tile grid, `KitFloorScript.pick` and
  `KitFloorScript.tile_transform` are reused, so dirt tiles seat on the plain tile's slab (the v491
  seating fix) and align with the stone.
- **Scatter.** A jittered grid of `cell_m` cells inside `scatter.radius_m` of the town centre. Each
  cell runs a `hash`-based roll (same scheme as the dungeon floor, no `randf`) for occupancy, kind
  and yaw/scale jitter. A cell is dropped when its point is within `min_clearance_m` of a gameplay
  position or a v491 prop, or inside paved ∪ edge, or in the gate gap.
- **Rendering.** MultiMesh per asset for tiles and patches. Rocks are few and use MultiMesh too, so
  the whole feature adds one draw call per distinct asset (about 12).
- **Structure.** New `client/scripts/town_ground_detail.gd` owns edge, rim, paths and scatter
  planning and build. `TownDressing` (140 lines) only calls it and adds the result under its root.
  The planning functions are pure (data in, cell lists out) and unit-testable without a scene.
- **Loader.** `TownPresentationLoader` gets accessors for the new keys, following the existing
  `dressing()` pattern. Its v491 bug (dropping every key but `night_lighting`) must not recur, so a
  test pins that the new keys survive the merge.

### Asset decision (adopt / borrow / reject)

- **Adopt** seven CC0 KayKit Dungeon Remastered 1.0 pieces already staged in
  `.artifacts/kaykit/extracted/` (pre-converted `.gltf.glb`, same provenance commit as the vendored
  set): `floor_dirt_small_A/B/C/D`, `floor_dirt_small_weeds`, `floor_dirt_large`,
  `floor_dirt_large_rocky`.
- **Adopt** two CC0 KayKit Resource Bits 1.0 pieces (`Stone_Chunks_Small` 554 tris,
  `Stone_Chunks_Large` 1578 tris, both under the 4k environment budget). They ship as external
  `.gltf` + `resource_bits_texture.png`, so they go through `tools/assets/gltf_to_glb.py` before
  vendoring. Provenance and sha256 are recorded like the existing entries.
- **Reuse** `floor_tile_small_decorated` and `floor_tile_small_weeds_A` (already vendored) for the
  rim. The broken variants stay excluded (holes are meant for a dark dungeon base).
- **Reject** new plugins and shaders: no dependency needed.

## Acceptance criteria

- [ ] In town, `TownDressing` contains the edge band, rim, service paths and scatter nodes; on a
  dungeon level none exist (existing removal test extended).
- [ ] Planner output is deterministic: two calls with the same data return identical cell lists and
  placements; changing one catalog weight changes the result (proves it is data-driven).
- [ ] Every edge cell centre is outside the paved region and within `edge.width_m` of it; every rim
  cell is paved; no cell appears in two layers (derived from catalog data, no pinned coordinates).
- [ ] Each configured service target is reachable by a path: some path tile centre lies within
  `path_width_m` of the target, and the path connects to the plaza.
- [ ] Scatter placements are ≥ `min_clearance_m` from every gameplay position and v491 prop, outside
  paved ∪ edge, inside `scatter.radius_m`, and their count lies within the range implied by the
  catalog (cell count × occupancy bounds), not an exact number.
- [ ] `pytest tools/test_town_dressing.py`: service-path targets equal the matching interactable
  positions in the world preset; scatter/edge config keeps clearance; every `asset_id` used exists in
  the manifest.
- [ ] `TownPresentationLoader` returns the new keys (regression pin for the v491 merge bug).
- [ ] `make validate-shared` / `make validate-assets` accept the schema change and the nine new CC0
  manifest entries within D8 budgets.
- [ ] Visual gate: `scenes/town` before/after captures in the as-built, covering the plaza edge, the
  vendor and mystery-seller paths, and the grass.
- [ ] `make client-unit` green; town and stash client scenarios green; `make ci` green (pre-PR gate
  because the asset manifest and a shared data file change).

## Scope and files likely touched

- **Assets:** `assets/manifests/assets.v0.json`; nine `client/assets/environment/kaykit_dungeon/*.glb`;
  `tools/assets/validate_assets.py` only if the environment list is enumerated there.
- **Data:** `shared/assets/town_presentation.v0.json` and its schema.
- **Client:** new `client/scripts/town_ground_detail.gd`; `client/scripts/town_dressing.gd`;
  `client/scripts/town_presentation_loader.gd`.
- **Tools:** `tools/test_town_dressing.py` (extend); `tools/validate_shared.py` cross-check that
  service-path targets match the world preset, if the schema alone cannot express it.
- **Tests:** new `client/tests/test_town_ground_detail.gd`, extend `client/tests/test_town_dressing.gd`,
  register the new test with the client unit runner.
- **Docs:** `docs/as-built/v493_town-floor-detail.md`, `docs/as-built/assets/v493/`, `PROGRESS.md`,
  `docs/progress/slice-lifecycle.md`, `docs/CODEMAP.md` (Town services row).
- **Not touched:** `server/`, `shared/protocol/`, `shared/golden/`, `shared/rules/`, bot scenarios.

## Test and bot proof

- **Unit (GDScript):** planner determinism, layer disjointness, region membership, path
  connectivity, scatter clearance and bounds, loader key survival, dressing add/remove per level.
  All expectations are derived from the loaded catalog, per the Test Locking Policy. No test pins a
  generated coordinate or an exact tile count.
- **Python:** the extended `tools/test_town_dressing.py` gates world-preset alignment and manifest
  ids.
- **Visual:** showme `scenes/town` captures, before/after. There is no bot proof for pure
  presentation; the existing town, vendor and stash client scenarios must stay green as regression.
- **Maintainability:** new files stay under 600 lines; `make maintainability` must not raise any
  baseline. `town_dressing.gd` must stay under its current size, since the new logic lives in
  `town_ground_detail.gd`.
- **Perf:** confirm the build cost at town entry on the capture path (one-shot, no per-frame work) and
  record the draw-call delta in the as-built.
- **CI scope:** targeted tests during development; `make ci` once before the PR. Not `ci-full`.

## Open questions and risks

- **Resource Bits texture.** The stone chunks use a separate `resource_bits_texture.png` atlas, unlike
  the Dungeon pieces. After `gltf_to_glb` the texture is embedded in each glb, adding two textures to
  the town. If the palette clashes with the dungeon atlas in the capture, fall back to
  `floor_tile_large_rocks`-style pieces from Dungeon Pack 1.1 or drop the rocks. This is decided at
  the visual gate, not now.
- **Dirt-versus-grass contrast.** The kit dirt is a flat brown atlas colour against a noisy grass
  plane; it may read as pasted tiles rather than a blend. Mitigation is the tile mix and the rim
  swap; the honest limit is that a real blend needs a terrain shader or a nature pack (non-goal).
  The visual gate decides whether the edge is good enough to ship.
- **Fence overlap.** The fence circle (radius 15 at the centre) is inside `scatter.radius_m` only if
  the radius stays under 15 + margin. The plan must confirm scatter never lands under the palisade
  or its outer approach. The test asserts distance to the fence line.
- **Camera coverage.** The grass plane is 140 × 90 m, but the camera only ever shows a few metres
  past the fence. `scatter.radius_m` should be sized from the observed play view, not the plane.
  The plan must measure it.
