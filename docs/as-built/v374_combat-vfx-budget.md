# v374 As-Built — Combat VFX Budget

## What shipped

- `shared/rules/main_config.v0.json` adds `client_perf.windup_marker_max_concurrent`.
- Windup marker pool caps concurrent telegraph decals; excess markers skip spawn instead of stacking GPU cost.
- Tween cleanup fixed for recycled windup nodes (no orphaned tweens under crowd load).

## Verification

```bash
make validate-shared
make bot scenario=crowded_melee_perf_probe
ARPG_PERF_DEBUG=1 make bot-visual scenario=crowded_melee_perf_probe
```
