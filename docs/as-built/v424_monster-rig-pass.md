# v424 As Built — Monster Rig Pass

Date: 2026-07-03
Spec-gate: exempt (client-only asset presentation; no protocol/server/rules/golden coordination)

## Shipped

- Biped monsters use height-normalized rig injection via `MONSTER_TARGET_HEIGHTS` in `rig_monster_glbs.py`.
- `rig_glb_bytes` accepts optional explicit `target_height` for non-hero subjects.

## Proof

```bash
make gen-assets
.venv/bin/python -m pytest tools/assets/test_rig_monster_glbs.py -q
make validate-assets
```
