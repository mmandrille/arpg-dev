# v430 Spec — Barbarian Tier-3 Mesh Swap

Status: Approved  
Date: 2026-07-04  
Codename: barbarian-tier3-mesh-swap

## Purpose

Swap `goliath_barbarian.glb` for a CC0 external body mesh sized as the tallest hero (~1.97 m), re-rig, update manifest provenance, and prove equipped gear + animation bones in client (v429 paladin template).

## Non-goals

- Other class hero meshes (rogue, sorcerer, ranger).
- Protocol, combat, or server changes.
- CI pack promotion (extended client scenario only).

## Acceptance criteria

1. New source at `assets/characters/barbarian/goliath_barbarian.glb`; prior mesh archived as `goliath_barbarian_legacy.glb`.
2. `HERO_TARGET_HEIGHTS["barbarian"]` ≈ 1.97 m; `rig_hero_glbs.py` produces runtime `barbarian.glb` with all `REQUIRED_BONES`.
3. Manifest `character_barbarian_v0` provenance records Poly Pizza origin URL + CC0-1.0 + runtime `sha256`.
4. `make validate-assets`, bone contract test, and `93_barbarian_tier3_visual` client scenario pass.
5. Showme gear capture documents swapped body with barbarian starter gear.

## Asset decision

- **Borrow:** CC0 Poly Pizza **Male Fighter** (mastjie, `GorWw41SFf`) — bulky static mesh, scaled to ~1.97 m for widest/tallest class silhouette.
- **Reject:** Skinned Quaternius animated giants; runtime network fetch; unmanifested GLBs.
