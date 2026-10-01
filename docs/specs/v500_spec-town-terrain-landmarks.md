# v500 — Town terrain and nature landmarks

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `town-terrain-landmarks`
- **Area:** Graphics
- **Depends on:** v493 town-ground data, v497 frame budget, and the Forest Nature source pack staged for this slice.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1–D3, D8–D9.

## Purpose

Give the town a convincing transition from its stone service area into surrounding ground and a recognizable nature silhouette. v493 added paths and detail but did not achieve a soft edge: the dirt tiles read almost like more stone. Use a restrained in-repo terrain treatment plus selected KayKit Forest Nature trees, shrubs, and grass around the palisade and service approaches. The play camera, not the overview preview, determines the result.

## Non-goals

- New buildings, cottages, stall structures, or the Medieval Hexagon pack, which is not present.
- Server collision, NPC behavior, interactable relocation, or changes to town gameplay anchors.
- A general terrain engine or replacement of the existing dungeon ground renderer.

## Acceptance criteria

- [ ] From the normal and maximum play-camera zoom, stone/path to ground reads as a continuous transition rather than a second hard polygonal band. The `dressing.edge.enabled` flag remains meaningful; the plan decides whether v493's taupe dirt edge stays enabled after the new blend is inspected.
- [ ] At least two visibly distinct, data-placed nature groupings give orientation around the town without covering the vendor, stash, quest giver, waypoint, gate, or player/monster silhouettes.
- [ ] Nature placement is stable for the same town data and excludes gameplay anchors, service paths, palisade crossing, and camera-critical sightlines. It is presentation only and creates no deceptive obstacle on a walkable path.
- [ ] At maximum zoom from playable town positions, ground coverage does not expose the v493 world-plane void. Any plane-extension cost is included in the performance comparison.
- [ ] Asset IDs, placement, weights, scales, and shader/blend tuning are schema-backed in `shared/assets/town_presentation.v0.json` or the existing surface catalog. Only selected runtime GLBs are vendored with CC0 license, pack/version, archive SHA-256, and per-file provenance recorded in the asset manifest.
- [ ] Real-renderer before/after captures from the play camera show plaza edge, service paths, palisade, and nature silhouettes at normal and maximum zoom. Balanced and Performance tiers stay readable and within the v497 frame budget.

## Scope and likely files

- **Source assets:** the owner-supplied `KayKit_Forest_Nature_Pack_1.0_FREE` (CC0) is staged under `.artifacts/kaykit/itch/` with GLTF models, their `.bin` files, texture, license, and preview. Original archive SHA-256: `2ee83e63bb7695f2d884ec27ddf6fce020789a452e7d5c5b0bbdfc4f6ea1fc8c`.
- **Runtime assets:** select a small set of trees, bushes, and grass after a visual/import check; convert with the existing `tools/assets/gltf_to_glb.py`, vendor only the chosen GLBs, update `assets/manifests/assets.v0.json`, and validate ADR-0018 budgets.
- **Client/data:** `client/scripts/town_ground_detail.gd`, `client/scripts/town_dressing.gd`, `client/scripts/ground_wall_factory.gd`, `shared/assets/town_presentation.v0.json` and schema; a focused in-repo `.gdshader` only if catalog/material tuning cannot create the blend.
- **Tests/docs:** town clearance and determinism tests, asset validator, CODEMAP, and as-built play-camera captures. No server or protocol change expected.

## Test and bot proof

- `make validate-assets`, `make validate-shared`, `make client-unit`, and town vendor/stash/waypoint client scenarios.
- Run the town visual-regression suite in the real renderer and inspect normal/max-zoom captures. The existing `showme town` overview is insufficient until it uses the play camera.
- Compare live town-entry cost, frame p95, draw calls, and memory against the v497/v498 baseline; inspect leaves/crowns over interactables and combat targets.

## Asset/plugin decision

- **Adopt** selected GLTF models from the owner-supplied CC0 KayKit Forest Nature Pack 1.0. Its bright green preview must be checked under the game's town lighting before choosing variants or tint.
- **Borrow** the existing town planner, KayKit asset pipeline, MultiMesh use, and current ground material. An in-repo shader is allowed by ADR-0018 if necessary for the edge.
- **Reject** external Godot plugins, runtime downloads, and a different art family. Defer buildings until a compatible source pack is available and separately specified.

## Open questions and risks

- Forest Nature supplies trees, bushes, rocks, and grass but no town buildings. This slice improves nature landmarks; a separate built-environment slice would be needed for cottages or stalls.
- Tree crowns can hide gameplay. The plan must choose outside-fence placement, camera/occlusion handling, or a strict height/clearance limit from real captures.
- The staged source directory is ignored by Git; a future worktree or machine needs the original owner-supplied pack or the selected vendored runtime GLBs. The implementation must not assume `.artifacts/` is present in CI.
