# v349 As-Built — Movement Tick Smoothing

Date: 2026-06-26  
Spec: [`docs/specs/v349_spec-movement-tick-smoothing.md`](../specs/v349_spec-movement-tick-smoothing.md)  
Plan: [`docs/plans/v349_2026-06-26-movement-tick-smoothing.md`](../plans/v349_2026-06-26-movement-tick-smoothing.md)

## Shipped behavior

- **Data-driven tuning** (`shared/assets/movement_presentation.v0.json`): `tick_smoothing.enabled`,
  `snapshot_interval_seconds` (default 0.1), `snap_distance` (default 2.0).
- **`EntityTickSmoothing`**: per-entity linear interpolation between authoritative positions; large
  deltas snap immediately.
- **`EntityTickSmoothingRuntime`**: wires local player anchor + entity map; advanced each frame in
  `main.gd` `_process`.
- **Orthogonal to v299**: anchor-offset visual smoothing on `CharacterVisual` remains unchanged.
- **Bot proof**: `84_entity_tick_smoothing` waits for active-then-settled tick smoothing during a
  floor move; `80_movement_visual_smoothing` regression passes.

## Boundaries

- No server/protocol/golden changes.
- Projectiles, interactables, and loot nodes unchanged.
- Leap/charge visuals still bypass tick smoothing on player authoritative apply.

## Verification

```bash
make validate-shared
make maintainability
make client-unit
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
```
