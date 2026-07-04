# v435 Spec — Bone Gear Sockets

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `bone-gear-sockets`

## Purpose

Replace hardcoded root-relative `FALLBACK_SOCKETS` with `BoneAttachment3D` mounts on the v434 skeleton, with per-class offset tuning in shared data.

## Non-goals

- New item defs or protocol changes
- Tier-3 weapon/armor meshes (v436–v438)

## Acceptance criteria

1. All gear mount sockets (`head`, `chest`, `gloves`, `belt`, `boots`, rings, amulet, hands) attach via `BoneAttachment3D` when a skeleton is present.
2. Socket bone + offset data lives in `shared/assets/gear_sockets.v0.json` with per-class overrides.
3. `character_visual.gd` reads class id and applies merged socket table.
4. Animation tests assert `head_socket` → `head`, `chest_socket` → `chest` on all five classes.
5. Extended client bot scenario equips starter loadout on at least one class.

## Test proof

- `make validate-shared`
- `make client-unit`
- `SCENARIO=97_bone_gear_sockets HEADLESS=1` client bot (extended)
