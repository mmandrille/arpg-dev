# v436 Spec — Weapons Tier-3 Melee

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `weapons-tier3-melee`

## Purpose

Replace procedural melee weapon and shield GLBs with CC0 external meshes (Poly Pizza / Quaternius), normalized via `tools/assets/import_equipment_glb.py`.

## Non-goals

- New item defs or balance changes
- Ranged weapons (v437) or armor (v438)

## Acceptance criteria

1. `rusty_sword`, `long_sword`, `rapier`, `starter_axe`, `kite_shield`, `tower_shield` runtime GLBs are Tier-3 imports with manifest provenance.
2. Legacy procedural GLBs archived under `assets/equipment/**/**/*_legacy.glb`.
3. `make validate-assets` passes for updated manifest entries.

## Asset decision

- **Borrow:** Quaternius CC0 swords/axe/shield from Poly Pizza (Referer-gated download).
- **Reject:** itch.io browser-only flows; keeping procedural meshes as runtime.
