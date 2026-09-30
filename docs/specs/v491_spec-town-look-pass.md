# v491 — Town look pass: paved plaza, road to the gate, kit dressing (ADR-0018 follow-up)

- **Status:** Implemented (v491)
- **Date:** 2026-09-30
- **Codename:** `town-look-pass`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D6/D10
- **Baseline:** v490 (`545bb59e`)
- **Spec gate:** client presentation + presentation data only; exempt, written anyway.

## Purpose

The town is the hub players see between every run, and it reads as a debug map: a flat
140 × 90 m pixel-noise grass plane with the eleven services scattered as boxes. Worse, what we
captured was not what players saw:

- The `scenes/town` capture used a preview-only scene with cabins and a campfire that the live town
  never builds.
- `TownAmbientLife` adds its silhouettes under the live ground node at raw town coordinates, but the
  ground node sits at (50, 0, 25). The silhouettes render ~50 m outside the fence, and they stay
  attached after descending into the dungeon.

After this slice, the live town (and the capture, through the same code) gets:

1. **A paved plaza** of KayKit Dungeon Remastered floor tiles (the kit floor variants, deterministic
   pick) inside `plaza.radius_m` of the town centre, plus **a road** of `path_width_m` from the
   plaza to the gate. MultiMesh, like the dungeon floor.
2. **Kit dressing**: ten CC0 Dungeon Remastered props (banners, barrels, crates, keg, a stall table,
   the sword-and-shield rack, a trunk, candles) grouped behind the services, listed with position,
   yaw and scale in `town_presentation.v0.json` → `dressing.props`. Presentation only: the server
   does not know them, so placement keeps clear of services, spawn, plaza road and the fence (a
   pytest enforces clearance against the world preset).
3. **`TownDressing.sync(ground_node, level)`**: one world-aligned root (offset by the ground node's
   position) that holds plaza, props and the ambient silhouettes in town and is removed on every
   other level. It fixes both bugs above.
4. The town preview capture uses `TownDressing` too; the preview-only cabins and campfire are deleted.

## Non-goals

- Trees, buildings, terrain shaders: they need a nature/village kit (see "Follow-up packs").
- Lighting changes, NPC characters, collision for dressing, any server/world-preset change.

## Acceptance criteria

- [ ] In town, the ground node has one `TownDressing` child whose props sit at their catalog world
  positions (ground offset compensated); on a dungeon level it has none.
- [ ] Plaza tiles cover the centre and the road reaches the gate; no tile centre lies outside the
  plaza disc ∪ road (unit test from catalog data).
- [ ] Ambient silhouettes render inside the fence (world positions from the catalog-derived root).
- [ ] `pytest tools/test_town_dressing.py`: every prop is ≥ `min_clearance_m` from each world-preset
  interactable/monster/spawn, off the road, and inside the fence; the plaza centre/gate match
  `town_presentation` center/gate_position.
- [ ] `make validate-shared` / `validate-assets` accept the catalog and ten new CC0 manifest entries.
- [ ] Visual gate: `scenes/town` before/after in the as-built.
- [ ] `make client-unit` and town client scenarios green; `make ci` green.

## Follow-up packs (owner download)

For trees, cottages and a real market square, download into `.artifacts/kaykit/itch/` (CC0,
kaylousberg.itch.io): **KayKit Forest Nature Pack** (trees, bushes, rocks, grass), **KayKit Medieval
Hexagon Pack** (village buildings, market stalls, wells, fences, paths). Then a v49x slice can replace
the flat grass edge and the service boxes.

## Files

manifest + `client/assets/environment/kaykit_dungeon/<10 props>.glb`; `shared/assets/town_presentation.v0.json`
+ schema; `tools/assets/validate_assets.py`; new `tools/test_town_dressing.py`; new
`client/scripts/town_dressing.gd` (+ `client/tests/test_town_dressing.gd`); `client/scripts/ground_wall_factory.gd`,
`client/scripts/town_node_factory.gd`, `client/scripts/town_ambient_life.gd`, `client/scripts/town_presentation_loader.gd`; docs.
