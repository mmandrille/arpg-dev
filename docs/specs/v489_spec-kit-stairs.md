# v489 — Kit stairs: dungeon stairs and the town descent use KayKit pieces (ADR-0018 P2 follow-up)

- **Status:** Implemented (v489)
- **Date:** 2026-09-30
- **Codename:** `kit-stairs`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D6, P2 row ("no kit door; stairs deferred")
- **Baseline:** v488 (`788fd050`)
- **Spec gate:** client presentation only (no protocol, server, rules or golden change); exempt, written anyway.

## Purpose

`stairs_up` / `stairs_down` still render as `TownNodeFactory.make_stair_node` box primitives
(flat grey steps, a black box pit): the last primitive interactable in kit dungeon rooms and the
town's way down. After this slice:

1. `stairs_up` is the KayKit Dungeon Remastered `stairs_narrow` flight, scaled from data.
2. `stairs_down` is the KayKit `floor_tile_big_grate_open` hatch framing a dark pit face (the kit
   has no descending stair and the kit floor covers anything below floor level, so a lit frame +
   dark opening is the honest read of "way down").
3. Interactable state (ready vs locked/disabled) tints the kit meshes through `ModelTint`
   (texture kept), colours from the catalog, replacing the base-box recolour.
4. Everything is data in `shared/assets/dungeon_kit_presentation.v0.json` → `stairs` (asset ids,
   scales, pit colour, state tints), schema-backed; validator step [9] resolves the asset ids.
5. The legacy primitive stays as the fallback when the catalog block is disabled or an asset fails
   to load.

## Non-goals

- Teleporter, waypoint or door kit art (no kit door model).
- Floor cut-outs under stairs, real descending geometry.
- Any stair gameplay, position, footprint, pick-collider or minimap change.

## Acceptance criteria

- [ ] `make_stair_node("stairs_up"|"stairs_down")` returns a node containing the catalog kit piece
  (`KitStairsModel`) scaled per catalog; the down variant also has the `KitStairsPit` face.
- [ ] With `stairs.enabled = false` (or a missing asset) the legacy primitive is returned.
- [ ] `apply` of a locked/disabled state tints the kit meshes with the catalog locked tint and keeps
  their albedo texture; ready restores the ready tint.
- [ ] `make validate-shared` / `make validate-assets` accept the catalog and the two new manifest
  entries (CC0 provenance, sha256, within D8 environment budgets) and reject an unknown stair id.
- [ ] Visual gate: `make regen-screenshots SUITE="scenes"` `stairs` and dungeon-room captures
  reviewed before/after; committed in the as-built.
- [ ] `make client-unit` green; existing stair client scenarios still pass.

## Files

`assets/manifests/assets.v0.json`; `client/assets/environment/kaykit_dungeon/{stairs_narrow,floor_tile_big_grate_open}.glb` (+ import + texture);
`shared/assets/dungeon_kit_presentation.v0.json` + schema; `tools/assets/validate_assets.py` (step [9]);
new `client/scripts/kit_stairs.gd`; `client/scripts/town_node_factory.gd`,
`client/scripts/interactable_state_presentation.gd`, `client/scripts/dungeon_kit_presentation_loader.gd`;
new `client/tests/test_kit_stairs.gd` (registered); docs.

**Asset decision:** adopt two more CC0 KayKit Dungeon Remastered 1.0 pieces (already staged);
reject the 5 m `stairs` and `stairs_wide` (too wide for the interactable footprint); no plugin.

## Open questions and risks

- Exact scales and pit size are tuned from captures (visual only).
