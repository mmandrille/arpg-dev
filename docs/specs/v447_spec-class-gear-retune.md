# v447 Spec — Class Gear Retune

**Status:** Complete  
**Date:** 2026-07-06  
**Codename:** `class-gear-retune`

## Purpose

Retune equipped-gear transforms and socket offsets so representative weapons, shields, armor, and
boots read correctly on **all five class rigs** after v444 class body forks. Fixes sword grip
alignment, shield orientation, undersized helm/mail, and invisible boots reported on showme and
in-game captures.

## Non-goals

- New production item meshes or Poly Pizza replacements.
- Protocol, server, combat, or inventory rule changes.
- Exhaustive retune of every `3d_model` item def (representative + starter + showme set first).
- CI pack promotion unless budget-neutral.
- Class body morphology or skeleton changes (v444 owns that).

## Asset decision (adopt / borrow / reject)

- **Adopt:** existing `item_visuals` + `gear_sockets` + `class_transforms` pipeline (v439/v442).
- **Borrow:** v442 `equipped_gear_fit_probe` and showme gear focus for visual proof.
- **Reject:** client-only `ModelRoot.scale` cheats or Blender per-class mesh edits.

## Acceptance criteria

1. **Sword grip:** `long_sword` and class starter main-hand weapons mount with hilt at `hand_r`, not
   mid-blade (showme paladin + probe matrix).
2. **Shield orientation:** `shield` / `starter_paladin_shield` stand vertical on `hand_l` (90° CW from
   current flat pose).
3. **Helm / mail scale:** `helm` and `mail` are head/torso-proportional on all five classes — not
   sub-pixel specks.
4. **Boots visible:** boots readable on **both** feet at gameplay camera distance.
5. **Class matrix:** `equipped_gear_fit_probe` covers all five classes; adds minimum global-scale
   assertions for head/chest/boots slots where headless-safe.
6. **Showme parity:** gear focus matches gameplay mount path for paladin and rogue.
7. `make validate-shared`, `make validate-assets`, `make client-unit` green.

## Scope and files

| Area | Files |
|------|-------|
| Item transforms | `shared/assets/item_visuals.v0.json` |
| Socket offsets | `shared/assets/gear_sockets.v0.json` |
| Dual boots | `client/scripts/equipment_visuals.gd` |
| Tests | `client/tests/equipped_gear_fit_probe.gd`, `client/tests/test_item_visuals.gd` |
| Docs | `PROGRESS.md`, `docs/as-built/v447_class-gear-retune.md` |

## Test proof

- Headless: `equipped_gear_fit_probe` via `make client-unit`.
- Visual: `python3 skills/showme/scripts/render_focus.py --focus gear --class-id <class>`.
- Bot: deferred (`ci_tier: extended` only if added later).

## Open questions

None — defaults: representative items + starters + showme set; code mirror for right boot;
shared base retune + per-class `class_transforms` only on failing class×item pairs.
