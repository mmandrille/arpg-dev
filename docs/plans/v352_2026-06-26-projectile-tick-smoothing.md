# v352 Plan: Projectile Tick Smoothing

Date: 2026-06-26  
Spec: [`docs/specs/v352_spec-projectile-tick-smoothing.md`](../specs/v352_spec-projectile-tick-smoothing.md)

## Tasks

- [x] 1.1 Extend movement_presentation JSON + schema with projectile smoothing fields
- [x] 1.2 Extend runtime: projectile apply/tick/facing + debug state
- [x] 1.3 Wire `main.gd` projectile upsert; remove tween path
- [x] 1.4 Bot scenario 85 + wait/assert handlers + unit tests
- [x] 1.5 Docs: as-built, lifecycle, PROGRESS

## Verification

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=85_projectile_tick_smoothing
```
