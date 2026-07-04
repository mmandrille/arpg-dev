# v431 Spec — Rogue Tier-3 Mesh Swap

Status: Approved  
Date: 2026-07-04  
Codename: rogue-tier3-mesh-swap

## Purpose

Swap `assasine.glb` for a CC0 external body mesh sized as the shortest hero (~1.70 m), re-rig, update manifest provenance, and prove equipped gear + animation bones in client.

## Acceptance criteria

1. New source at `assets/characters/rogue/assasine.glb`; prior mesh archived as `assasine_legacy.glb`.
2. `HERO_TARGET_HEIGHTS["rogue"]` ≈ 1.70 m; runtime `rogue.glb` has all `REQUIRED_BONES`.
3. Manifest `character_rogue_v0` provenance updated (CC0, Poly Pizza URL, `sha256`).
4. `make validate-assets`, bone test, `94_rogue_tier3_visual` pass.
5. Showme gear capture with rogue starter gear.

## Asset decision

- **Borrow:** CC0 Poly Pizza **Thief Icon** (Quaternius, `M8kyS2Btfp`) — static hooded thief gesture mesh at ~1.70 m.
- **Reject:** Skinned animated ninjas; runtime network fetch.
