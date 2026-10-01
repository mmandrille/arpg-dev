# v500 Plan — Town terrain and nature landmarks

Status: Complete; combined `make ci` passed in 10m33s.

Goal: Make the town's stone paths meet surrounding ground naturally and add readable KayKit nature landmarks without changing gameplay.

Architecture: Keep all terrain and placement data in the town presentation catalog and schema. Blend the existing town ground material by distance from the plaza and configured path capsules; retain the existing stone meshes and use a dedicated town-only shader only if catalog/material tuning cannot remove the hard edge. Plan nature placement deterministically from data, build repeated small meshes with the existing MultiMesh path, and keep trees outside the palisade and clear of camera sightlines. Server collision, world presets, protocol, and dungeon rendering remain untouched.

Tech stack: Godot 4/GDScript and shader, shared JSON/schema, existing Python asset validation and screenshot tooling.

## Baseline and shortcut decision

- `PROGRESS.md` reports v493 as completed and v494–v500 as **To do**. Execute v500 only after its preceding slice gates; compare performance against the actual v497/v498 result, not the draft specs.
- v493 supplies `TownGroundDetail`'s paved region, service-path capsules, clearance rules, and `MultiMesh` builders. The current town ground is a 140 × 90 m plane; v493 records a visible void at maximum zoom from west/north fence positions.
- The Forest Nature 1.0 CC0 source is present at `/Users/mmandrille/git/arpg-dev/.artifacts/kaykit/itch/KayKit_Forest_Nature_Pack_1.0_FREE`, but absent from this worktree. Read source there and put only selected converted runtime GLBs in this worktree. Do not make CI depend on `.artifacts/`. `SOURCE.txt` records archive SHA-256 `2ee83e63bb7695f2d884ec27ddf6fce020789a452e7d5c5b0bbdfc4f6ea1fc8c`.
- **Adopt:** a small, visually checked set of Forest Nature tree, bush, and grass GLTFs. Initial candidates are `Tree_2_A_Color1` (336 triangles), `Tree_4_A_Color1` (404), `Bush_3_A_Color1` (208), and `Grass_1_A_Color1` (44); these are inspection candidates, not final selections. The pack preview is brighter than the current town palette, so approve variants/tint only from real town-lighting captures.
- **Borrow:** the existing glTF-to-GLB converter, asset manifest and validator, `TownGroundDetail` placement geometry, `KitPieceLibrary`, MultiMesh, town material, and camera settings.
- **Reject:** buildings, external plugins, runtime downloads, and a general terrain engine.
- Security router check: this slice changes fixed local presentation data and visuals only; it does not introduce a network, file-input, auth, or user-controlled rendering surface.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/assets/town_presentation.v0.json`, `.schema.json` | Blend, plane extent, landmark groups, variants, scales, density, clearances, and sightline tuning |
| Modify | `shared/assets/surface_material_presentation.v0.json` and schema, only if needed | Reuse the current town grass/soil palette instead of duplicating it |
| Modify | `client/scripts/ground_wall_factory.gd` | Town-only plane sizing and ground material hookup; dungeon path unchanged |
| Create if needed | `client/shaders/town_ground_blend.gdshader`, `client/scripts/town_ground_blend.gd` | World-space smooth soil/grass transition around the existing paved capsules |
| Create | `client/scripts/town_nature_landmarks.gd` | Pure deterministic placement and grouped rendering, separate from the v493 ground planner |
| Modify | `client/scripts/town_dressing.gd`, `client/scripts/town_presentation_loader.gd` | Attach the nature layer and expose catalog data without changing anchor ownership |
| Add | `client/assets/environment/kaykit_forest_nature/*.glb`, `LICENSE.txt` | Only visually selected runtime models and their CC0 license |
| Modify | `assets/manifests/assets.v0.json` | Stable asset IDs, pack/version, source URL, archive hash, source filename, and runtime SHA-256 |
| Create/modify | `client/scripts/showme/showme_town_play_capture.gd`, `client/scripts/showme/visual_capture.gd`, `skills/showme/scripts/render_focus.py`, `tools/showme/screenshot_catalog.py` | Real play-camera positions at normal/max zoom, registered with the town scene suite |
| Create/modify | `client/tests/test_town_nature_landmarks.gd`, `client/tests/test_town_ground_blend.gd`, `client/tests/test_town_ground_detail.gd`, `tools/test_town_dressing.py`, `scripts/client_smoke.sh` | Determinism, footprint exclusion, coverage and build tests |
| Modify | `docs/CODEMAP.md` | Index new presentation files during isolated implementation |
| Coordinator after integration | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle closeout in execution order on the combined result |
| Create | `docs/as-built/v500_town-terrain-landmarks.md`, `docs/as-built/assets/v500/*.png` | Evidence, edge decision, provenance, measurements, limits |

## Maintenance ratchet

Target: new source/test/tool files stay at or below 600 lines. `visual_capture.gd` is 1252 lines now against a 1235-line grandfathered baseline; replace its existing town setup with a smaller delegation so a touch leaves it at or below 1235. `town_ground_detail.gd` is 364 lines; keep nature placement in a new file to avoid growing that planner into another coordinator. Recheck all touched files against `.maintainability/file-size-baseline.tsv`.

Verification: `make maintainability`.

## Task 1 — Capture the real baseline before visual edits

Files: `client/scripts/showme/showme_town_play_capture.gd`, `client/scripts/showme/visual_capture.gd`, `skills/showme/scripts/render_focus.py`, `tools/showme/screenshot_catalog.py`, `docs/as-built/assets/v500/`.

- [x] Add a town play-camera capture that uses the actual isometric angle and `camera_presentations.v0.json` normal and maximum zoom values (currently 12 and 20), with fixed player positions near plaza edge, vendor path, and west/north palisade. Keep the overview `town` focus available.
- [x] Record before images in the real renderer, at Balanced and Performance tiers where relevant. Name images by position and zoom; verify each contains ground and a gameplay reference point.
- [ ] Record matching town-entry, frame p95, draw-call, and memory baselines using the v497/v498 measurement method and fixture after those slices land. Preserve sample count, host, renderer, quality tier, and run order.

Verify: `python3 skills/showme/scripts/render_focus.py --focus town-play` (and its max-zoom option chosen by the capture helper); `make regen-screenshots SUITE="scenes"`; inspect PNGs under `.artifacts/screenshots/latest/`.

## Task 2 — Select, vendor, and register restrained nature assets

Files: `client/assets/environment/kaykit_forest_nature/`, `assets/manifests/assets.v0.json`, `tools/assets/validate_assets.py` only if its current environment checks need extension.

- [x] Inspect the shortlisted GLTFs in town lighting for crown shape, green saturation, transparency, size, and occlusion. Choose the smallest set that yields two distinct silhouettes; drop variants that obscure interactables or exceed ADR-0018 D8 budgets.
- [x] Convert selected source GLTFs with `python3 -m tools.assets.gltf_to_glb <source.gltf> <runtime.glb>`; include no source `.bin`, preview, or unused variants in the change set. Copy the CC0 license beside the selected GLBs.
- [x] Register each selected ID with pack/version, source path and source URL, archive SHA-256, runtime SHA-256, and license. Verify the manifest points only at selected runtime files present in this worktree.

Verify: `make validate-assets`; inspect imported candidates in the real renderer.

## Task 3 — Tune a soft ground transition and complete coverage

Files: `shared/assets/town_presentation.v0.json`, `.schema.json`, optional surface catalog/schema, `client/scripts/ground_wall_factory.gd`, optional `client/shaders/town_ground_blend.gdshader` and `client/scripts/town_ground_blend.gd`, `client/tests/test_town_ground_blend.gd`.

- [x] Add schema-backed terrain controls: enable flag, soil/grass palette reference or colors, transition widths, low-frequency breakup, and plane extent. Compute distance from the v493 plaza and path capsules in town world coordinates; make the outer grass transition smooth and stable under camera movement.
- [x] Review v493 material tuning: it cannot feather opaque tile edges. Use one town-only shader on the existing plane; pass material parameters from data and keep dungeon behavior intact.
- [x] Capture the blend with `dressing.edge.enabled` on and off. Ship the clearer choice and record the reason; preserve the flag and its independent behavior in tests.
- [x] Extend the town plane enough to cover the captured maximum-zoom west/north views.
- [x] Measure added renderer cost after the v497/v498 baseline lands.

Verify: `make validate-shared`; focused client tests through `make client-unit`; normal/max-zoom real-renderer captures.

## Task 4 — Place two nature landmark groups safely

Files: `shared/assets/town_presentation.v0.json`, `.schema.json`, `client/scripts/town_nature_landmarks.gd`, `client/scripts/town_dressing.gd`, `client/scripts/town_presentation_loader.gd`, `client/tests/test_town_nature_landmarks.gd`, `tools/test_town_dressing.py`, `scripts/client_smoke.sh`.

- [x] Define at least two explicit group regions with distinct tree/shrub/grass mixes, candidate weights, scale ranges, seed salts, and occupancy in data. Start outside the palisade on different sides; adjust centers after play-camera inspection.
- [x] Reuse v493's anchors and paved/path geometry for footprint checks. Exclude gate approach and fence crossing, all gameplay anchors and props, service paths, and the configured gate-approach corridor. Use asset bounds plus configured clearance and inspect camera-critical sightlines in play-camera captures.
- [x] Generate placements deterministically from the same town data. Batch repeated small meshes through MultiMesh; avoid a per-blade scene node or per-tree light/shadow cost.
- [x] Test same-input placement equality, weight/scale sensitivity, group distinction, anchor/path/fence/sightline clearance, zero-placement flags, and data/manifest ID consistency. Inspect crowns from the player camera at both zoom values and both quality tiers.

Verify: `make validate-shared`; `make client-unit`; `.venv/bin/pytest tools/test_town_dressing.py`; `make regen-screenshots SUITE="scenes"` and inspect the new town-play PNGs.

## Task 5 — Functional and performance proof

Files: existing client town scenarios; `docs/as-built/assets/v500/`.

- [x] Run vendor, stash, and waypoint client scenarios to verify clicking and approaches remain usable: `make bot-client SCENARIO=15_town_vendor_shop_panel HEADLESS=1`, `make bot-client SCENARIO=23_account_stash_panel HEADLESS=1`, `make bot-client SCENARIO=07_town_teleporter_auto_approach HEADLESS=1`.
- [x] Compare matched town-entry, frame p95, draw calls, and memory runs with the Task 1 baseline on Balanced and Performance tiers. Apply the actual v497/v498 budget; reduce variants, density, shadows, or plane size and rerun if the budget is missed. Do not claim a performance pass from screenshots or headless tests.
- [x] Save before/after play-camera captures at normal and maximum zoom for the plaza edge, vendor/stash approaches, gate, and both landmark silhouettes. Inspect no world-plane void and no misleading blocked walkable route.

No new protocol bot scenario is required: this slice does not change a world preset, movement, interaction contract, or server outcome. Existing client scenarios cover the affected gameplay approaches; pure placement invariants are client/Python tests.

## Task 6 — Closeout

Files: `docs/as-built/v500_town-terrain-landmarks.md`, `docs/CODEMAP.md`; coordinator later updates `docs/progress/slice-lifecycle.md` and `PROGRESS.md`.

- [x] Record selected asset provenance, before/after captures, `edge.enabled` decision, deterministic clearance proof, and limitations.
- [x] Add matched performance measurements after v497/v498 integration.
- [x] Update CODEMAP and prepare the as-built in this worktree. Report changed files, exact focused checks, captures, measurements, and any blocker. The coordinating task updates lifecycle and `PROGRESS.md` when v500 is integrated in order.

Verify in this worktree: `make maintainability`; `make validate-assets`; `make validate-shared`; `make client-unit`; targeted town scenarios and real-renderer captures. The coordinating task runs `make ci` once on the combined result as the final gate.

## Deferred scope and handoff gate

Buildings and plugins remain outside v500. No work here authorizes a branch, commit, merge, cleanup, transfer to `main`, or `/finish` from this isolated worktree. The plan is ready to execute after v497's measured budget and the v498 baseline are available; final performance acceptance remains blocked until then.
