# v434 As Built — Humanoid Skeleton v2

Date: 2026-07-04

## Shipped

- Expanded shared humanoid rig from 8 to 17 bones: `chest`, `neck`, `head`, `elbow_l/r`, `knee_l/r`, `foot_l/r` added under existing `spine`, `arm_*`, `leg_*` names.
- Re-rigged all five Tier-3 hero runtime GLBs via `rig_hero_glbs.py`.
- Updated manifest `required_nodes` and provenance sha256 for hero assets.

## Hierarchy

```
root → spine → chest → neck → head
              chest → arm_l → elbow_l → hand_l
              chest → arm_r → elbow_r → hand_r
root → leg_l → knee_l → foot_l
root → leg_r → knee_r → foot_r
```

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py -q
make client-unit
```

Existing `character_anims.tres` clips unchanged; `arm_l`, `arm_r`, `spine`, `leg_*` tracks remain valid.
