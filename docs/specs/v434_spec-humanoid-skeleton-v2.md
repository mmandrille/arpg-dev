# v434 Spec — Humanoid Skeleton v2

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `humanoid-skeleton-v2`

## Purpose

Expand the shared eight-bone humanoid rig to a standard 17-bone game skeleton so future gear sockets, attack clips, and armor mounts attach to credible anatomy on Tier-3 hero meshes.

## Non-goals

- Finger bones, facial rig, runtime IK
- Server or protocol changes
- New animation clips (existing `character_anims.tres` paths must keep working)

## Bone hierarchy

```
root
├── spine
│   └── chest
│       ├── neck
│       │   └── head
│       ├── arm_l
│       │   └── elbow_l
│       │       └── hand_l
│       └── arm_r
│           └── elbow_r
│               └── hand_r
├── leg_l
│   └── knee_l
│       └── foot_l
└── leg_r
    └── knee_r
        └── foot_r
```

**17 bones:** `root`, `spine`, `chest`, `neck`, `head`, `arm_l`, `elbow_l`, `hand_l`, `arm_r`, `elbow_r`, `hand_r`, `leg_l`, `knee_l`, `foot_l`, `leg_r`, `knee_r`, `foot_r`.

Existing names (`spine`, `arm_l`, `hand_l`, `arm_r`, `hand_r`, `leg_l`, `leg_r`) are preserved so `character_anims.tres` tracks remain valid; new bones are children in the extended hierarchy.

## Acceptance criteria

1. `REQUIRED_BONES` in `rig_hero_glbs.py` lists all 17 bones in skin joint order.
2. All five Tier-3 hero runtime GLBs re-rigged with the expanded skeleton and pass `make validate-assets`.
3. `assets/manifests/assets.v0.json` `required_nodes` updated for every hero asset entry.
4. `client/tests/test_animation.gd` asserts all 17 bones on each class model; idle/walk/attack clips still rotate `arm_r`, `arm_l`, `spine`, `leg_*`.
5. `tools/assets/test_rig_hero_glbs.py` green for all heroes.
6. Showme `--focus classes` captures all five classes without visible mesh tearing.

## Scope and files

| Area | Files |
|------|-------|
| Rig tool | `tools/assets/rig_hero_glbs.py`, `tools/assets/test_rig_hero_glbs.py` |
| Manifest | `assets/manifests/assets.v0.json` |
| Runtime GLBs | `client/assets/characters/*/*.glb` (regenerated) |
| Client tests | `client/tests/test_animation.gd` |
| Monster rig | inherits via `rig_monster_glbs.py` import of `REQUIRED_BONES` |

## Test proof

- `make validate-assets`
- `.venv/bin/pytest tools/assets/test_rig_hero_glbs.py -q`
- `make client-unit` (animation tests)
- `python3 skills/showme/scripts/render_focus.py --focus classes`

## Risks

- Vertex weight heuristics must map head/chest/limb regions to new joint indices without stretching Tier-3 meshes.
- Monster bipeds re-rigged with the same tool inherit the expanded bone set.
