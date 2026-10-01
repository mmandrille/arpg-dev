# v494 — Dungeon room dressing

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `dungeon-room-dressing`
- **Area:** Graphics
- **Baseline:** v493 complete; second slice of the two-slice town/dungeon dressing sequence recorded in the v493 spec.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1, D2, D6, D8, D9.

## Purpose

Every generated dungeon floor should have restrained, varied KayKit dressing so rooms no longer read as bare walls and tiles. Place free-standing props deterministically from the floor layout, with data-owned density, weights, clearances, orientation, and scale. Re-entering the same floor must produce the same visible arrangement. Dressing is presentation only and must not pretend to block a route that the server considers walkable.

## Non-goals

- Server obstacles, collision, loot, destructible props, or new interactable behavior.
- Changes to room generation, paths, torches, chests, stairs, or combat rules.
- A new art family, runtime asset download, or Godot plugin.

## Acceptance criteria

- [ ] Every generated dungeon floor with at least one safe candidate area renders dressing, including samples from multiple biomes/depths; town and non-dungeon views do not retain it after transitions. A floor with no safe candidate area remains clear and reports that case in the placement test.
- [ ] Given the same floor identity, geometry, and catalog, placements are identical across builds; different floor identities produce visible variation without runtime randomness.
- [ ] Candidate positions respect catalog clearances from walk routes, narrow passages, stairs, doors, chests, torches, spawn and interaction anchors, and existing wall/column meshes. Decoration does not make an apparent blockade across a traversable passage.
- [ ] Asset selection, density, scale, and placement limits are schema-backed presentation data; invalid or missing asset IDs fail validation.
- [ ] Instance count and draw-call delta stay within an explicit budget set from a pre-change live capture; the implementation records how it avoids recreating assets every frame.
- [ ] Before/after real-renderer captures from the play camera show at least a sparse and a dense dungeon floor. Visual review confirms enemy silhouettes, telegraphs, loot, and navigation remain legible.

## Scope and likely files

- **Presentation data:** extend `shared/assets/dungeon_kit_presentation.v0.json` and schema with a `dressing` block; use `assets/manifests/assets.v0.json` IDs.
- **Client:** add a focused planner/builder beside `client/scripts/dungeon_kit_floor.gd`; integrate at the existing floor build/teardown boundary in `client/scripts/wall_renderer.gd`; reuse `client/scripts/kit_piece_library.gd`.
- **Tests/tools:** focused GDScript placement and teardown tests, `tools/validate_shared.py` / `tools/assets/validate_assets.py` checks as needed, CODEMAP and as-built capture links.
- **Contracts:** no server, protocol, gameplay-rule, or golden change expected.

## Test and bot proof

- Unit tests cover repeatability, varied floor keys, catalog toggles and weights, exclusion zones, maximum placements, and removal on level transition.
- `make validate-shared`, `make validate-assets`, `make client-unit`, and dungeon navigation/combat client scenarios stay green.
- Run the relevant `make regen-screenshots SUITE=...` jobs and inspect real-renderer PNGs; record before/after and draw-call measurements in the as-built. Select an existing navigation/combat scenario for `make bot-visual scenario=...` in the plan.

## Asset/plugin decision

- **Borrow** the already vendored CC0 KayKit Dungeon Remastered props used by town dressing and existing torch/chest art. Check their scale and silhouette in the real dungeon camera before selecting the final set.
- **Reject** new packs and plugins for this slice. The plan may add a vetted, provenance-recorded piece from the already staged KayKit pack only if the existing assets cannot fill a clear visual role.

## Open questions and risks

- **Default:** decoration is nonblocking. A blocking prop would require a separate server-authored obstacle slice and a new route proof.
- The floor layout must expose enough stable, client-visible geometry and anchors for clearance. The plan must map those inputs before selecting a placement algorithm; if an anchor is unavailable, omit nearby dressing instead of guessing.
- A dense arrangement may worsen the current draw-call and first-spawn costs. v495 addresses the measured hitch; v494 must still stay within its own recorded budget.
