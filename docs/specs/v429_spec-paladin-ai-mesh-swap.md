# v429 Spec — Paladin AI Mesh Swap (Evaluation)

Status: Approved  
Date: 2026-07-04  
Codename: paladin-ai-mesh-swap

## Purpose

Execute the v425 Tier-3 playbook for **paladin only**: swap `knight.glb` for an external body mesh, re-rig, update manifest provenance, and produce before/after visual proof so the user can decide whether to repeat for other classes.

## Non-goals

- Swapping other class hero meshes.
- Protocol, combat, or server changes.
- CI pack promotion (extended client scenario only).

## Acceptance criteria

1. New source at `assets/characters/paladin/knight.glb`; prior mesh archived as `knight_legacy.glb`.
2. `rig_hero_glbs.py` produces runtime `paladin.glb` with all `REQUIRED_BONES`.
3. Manifest `character_paladin_v0` provenance records origin URL + license + runtime `sha256`.
4. `make validate-assets`, bone contract test, and `92_paladin_tier3_visual` client scenario pass.
5. Showme gear capture documents the swapped body with starter equipment.

## Asset decision

- **Borrow:** CC0 Poly Pizza warrior (mastjie) as Tier-3 evaluation stand-in when no Meshy/Tripo export is in-repo. Same pipeline applies to a true AI GLB dropped at `knight.glb`.
- **Reject:** Runtime network fetch; unmanifested GLBs; skipping provenance.
