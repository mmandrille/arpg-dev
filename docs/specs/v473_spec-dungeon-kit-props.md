# v473 — Dungeon kit torches and treasure chests (ADR-0018 P2b)

- **Status:** Implemented (v473). This spec was written during implementation, right after the
  owner asked for "P2b straight away". It records the scope as built.
- **Date:** 2026-09-29
- **Codename:** `dungeon-kit-props`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) P2b
- **Baseline:** v472 (`3e98532e`), on top of v471 kit walls

## Purpose

This slice replaces the procedural stand-ins for two gameplay-backed dungeon props with KayKit
Dungeon Remastered 1.0 models:

- **Wall torches.** They now use the kit `torch_mounted` model. It is rotated so the bracket faces
  away from its wall, toward the room. Two things stay as before:
  - The existing emissive flame and `OmniLight3D` are kept, re-anchored at the kit flame and scaled
    by the catalog.
  - Fog-of-war torch holes stay driven by `dungeon_torch_presentation`, so the torch placement
    count, positions and sampling are unchanged.
- **Treasure chests** (`treasure_chest`). Normal chests use the kit `chest`; elite-objective chests
  use `chest_gold`. The presentation contract is kept, so `InteractableStatePresentation`, the
  markers and bot debug work unchanged:
  - root name `TreasureChest`
  - a `ChestLidPivot` (the kit lid node, which already sits on the hinge), so the existing −68° X
    tween opens it
  - a hidden emissive `ChestInnerGlow`
  - the objective and quest markers

The server, protocol, interactable state and collision do not change.

## Non-goals

- **Doors.** The kit has no door model, only a `wall_doorway` frame.
- **Stairs.** The kit stairs are 5 m ascending flights that don't fit the current stair
  interactable footprint.
- **Town stash and unique chests.**
- **Free-standing dressing** (barrels, crates, candles). The server doesn't know about it, so
  players would walk through it.

## Acceptance criteria

1. The `torch_mounted`, `chest` and `chest_gold` kit GLBs are vendored as `environment` manifest
   assets with provenance, within ADR-0018 D8 budgets.
2. `dungeon_kit_presentation.v0.json` gains two blocks, and validator step [9] checks their ids:
   - `torch`: enable flag, asset, flame offset, flame scale
   - `chest`: enable flag, asset and elite asset, lid node per variant, scale
3. `DungeonTorchPlacement.mounts_from_walls` returns `{position, facing}`, where facing points
   toward the room. `placements_from_walls` returns exactly those positions.
4. On kit levels each torch has a `KitTorch` body plus a flame and no procedural bracket. With the
   kit disabled, torches keep the legacy bracket.
5. The kit chest keeps the contract in Purpose. Rotating `ChestLidPivot` the way the open tween
   does raises the lid. Elite chests use the gold variant and still expose their marker and lid.
6. `TownNodeFactory.make_chest_node` routes only `treasure_chest` to the kit.
7. `make ci` passes.

## Files

| Area | Files |
|------|-------|
| Assets | `client/assets/environment/kaykit_dungeon/{torch_mounted,chest,chest_gold}.glb`; `assets/manifests/assets.v0.json` |
| Shared | `shared/assets/dungeon_kit_presentation.v0.json` (+ schema) |
| Tools | `tools/assets/validate_assets.py` (kit id check) |
| Client | new `client/scripts/dungeon_kit_props.gd`; `dungeon_kit_presentation_loader.gd`, `dungeon_torch_placement.gd`, `dungeon_torch_lights.gd`, `town_node_factory.gd` |
| Tests | new `client/tests/test_dungeon_kit_props.gd` (registered); `client/tests/item_visual_interactable_probe.gd` (accepts the kit chest; the lock-plate check stays for the fallback) |

**Asset/plugin decision:** adopt the vendored KayKit pack (v469/v471). Reject door and stair
substitutes that don't fit the gameplay footprints.
