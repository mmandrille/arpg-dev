# v427 Spec — Paladin Tier-3 Hero Body Proof

Status: Approved (autoloop)  
Date: 2026-07-04  
Codename: paladin-tier3-body

## Purpose

Prove the Tier-3 hero pipeline for **paladin** end-to-end: `knight.glb` source → `rig_hero_glbs.py` → manifest → client mount → equipped gear presentation.

## Non-goals

- Replacing `knight.glb` with a new AI mesh (future Tier-3 swap).
- Other classes (follow v428+ / later slices).
- Protocol or combat changes.

## Acceptance criteria

1. `assets/characters/paladin/README.md` documents `knight.glb` Tier-3 source + rig tool.
2. `client/assets/characters/paladin/paladin.glb` passes bone + height contract test.
3. Extended client bot scenario: paladin class, full starter gear equip, `visual_model: character` on local player.
4. `make bot-visual scenario=paladin_tier3_visual` or showme gear capture documents visual proof path.

## Adopt / borrow / reject

- **Adopt** existing `knight.glb` + rig pipeline (borrow Tier-3 workflow from v425 playbook).
