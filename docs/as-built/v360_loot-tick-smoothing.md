# v360 As-Built — Loot Tick Smoothing

Date: 2026-06-28

## Shipped

- `movement_presentation.v0.json` adds `loot_enabled` and `loot_snap_distance`.
- `EntityTickSmoothingRuntime.apply_loot_authoritative` gates loot nodes separately from combat entities.
- Bot debug exposes `loot_tick_smoothing`; extended scenario `86_loot_tick_smoothing`.

## Verification

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-client SCENARIO=86_loot_tick_smoothing HEADLESS=1
```
