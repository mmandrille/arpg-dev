# v434 Plan — Humanoid Skeleton v2

Spec: [`docs/specs/v434_spec-humanoid-skeleton-v2.md`](../specs/v434_spec-humanoid-skeleton-v2.md)

## Tasks

- [x] Expand `REQUIRED_BONES`, `_joint_globals`, `_joint_nodes`, `_joint_for_vertex` in `rig_hero_glbs.py`
- [x] Update hero `required_nodes` in `assets/manifests/assets.v0.json`
- [x] Re-rig all five heroes: `python3 tools/assets/rig_hero_glbs.py`
- [x] Update `client/tests/test_animation.gd` bone list (17 bones)
- [x] Update `tools/assets/test_rig_hero_glbs.py` if needed
- [x] Run focused verification

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py -q
make client-unit
python3 skills/showme/scripts/render_focus.py --focus classes
```
