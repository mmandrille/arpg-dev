# v442 Spec — Equipped Gear Fit Matrix (All Classes)

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `equipped-gear-fit`

## Purpose

Every equippable item with a 3D model must mount at the correct bone/socket with humanoid-proportional scale on **all five class rigs** (barbarian, paladin, rogue, ranger, sorcerer). Fixes the post–Tier-3 regression where chest/shield appeared at the feet and import double-scale (Poly Pizza `scale:100` nodes) inflated meshes.

## Non-goals

- New art or new item models.
- Protocol/server gameplay changes.
- Full paper-doll UI redesign.
- Exhaustive bot dungeon walks for every item variant.

## Acceptance criteria

1. **Import hardening:** `import_equipment_glb.py` strips non-identity glTF node scales after vertex normalization; `make validate-assets` fails on equipment GLBs with non-identity node scale.
2. **Socket reliability:** `character_visual.gd` does not leave `BoneAttachment3D` nodes at skeleton origin when bone binding fails; sockets refresh after class model swap.
3. **Showme parity:** `visual_capture.gd` gear focus matches `main.gd` class-model + equipment remount path (`set_character_class`, remount after model).
4. **Headless matrix:** New probe runs class × representative item matrix; asserts mounted GLB exists, global Y above slot-specific floor band, global scale within rule-derived maxima.
5. **Visual spot-check:** Showme gear capture for paladin shows helm on head, mail on torso, shield on off-hand, sword at hand — not clustered at feet.
6. `make validate-shared`, `make validate-assets`, `make client-unit`, focused item-visual tests green.

## Scope and files

| Area | Files |
|------|-------|
| Import | `tools/assets/import_equipment_glb.py`, equipment GLBs, `assets/manifests/assets.v0.json` |
| Display tuning | `shared/assets/equipment_display.v0.json`, schema, `equipment_display_loader.gd` |
| Sockets | `client/scripts/character_visual.gd`, `shared/assets/gear_sockets.v0.json` |
| Mounting | `client/scripts/equipment_visuals.gd`, `shared/assets/item_visuals.v0.json` |
| Showme | `client/scripts/showme/visual_capture.gd` |
| Tests | `client/tests/equipped_gear_fit_probe.gd`, `client/tests/test_item_visuals.gd`, `client/scripts/client_smoke.sh` |
| Validation | `tools/assets/validate_assets.py` |
| Docs | `docs/CODEMAP.md`, `PROGRESS.md`, as-built |

## Test proof

- Headless: `equipped_gear_fit_probe` via `make client-unit` / `test_item_visuals.gd`.
- Asset: `make validate-assets`.
- Visual: `python3 skills/showme/scripts/render_focus.py --focus gear --class-id <class>`.
- Bot: deferred (`ci_tier: extended` only if a compact lab scenario is added later).

## Open questions

None — defaults: data-driven transforms in `item_visuals` / `gear_sockets`; code fixes for socket lifecycle and showme parity.
