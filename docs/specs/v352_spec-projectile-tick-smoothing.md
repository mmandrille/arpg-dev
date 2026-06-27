# v352 Spec: Projectile Tick Smoothing

Status: Complete  
Date: 2026-06-26  
Codename: `projectile-tick-smoothing`  
Baseline: v351 `window-display-mode`

## Purpose

Replace tween-based projectile position updates with the same tick-smoothing model used for
entities (v349), driven from `shared/assets/movement_presentation.v0.json`. Authoritative
projectile nodes interpolate between 10 Hz snapshots at render rate instead of restarting
SceneTree tweens each tick.

## Non-goals

- No server, protocol, shared golden, or sim changes.
- No skill-authored preview projectiles (`_track_skill_authored_projectile` path unchanged).
- No leap/charge/teleport smoothing (v353).
- No interactable/loot smoothing.

## Acceptance criteria

- `movement_presentation.v0.json` adds projectile smoothing fields under `tick_smoothing`:
  `projectiles_enabled`, `projectile_snap_distance`.
- Wire projectiles through `EntityTickSmoothingRuntime` (reuse `EntityTickSmoothing`).
- Remove tween-based `_move_projectile_node` path for authoritative projectile entities.
- Projectiles face travel direction on authoritative segments and during interpolation.
- Bot debug exposes `projectile_tick_smoothing` with active segment state.
- Extended client bot scenario casts a projectile and waits for active-then-settled smoothing.
- Unit tests cover runtime projectile apply/tick/facing helpers.

## Scope and likely files

| Area | Files |
|------|-------|
| Shared tuning | `movement_presentation.v0.json`, schema |
| Runtime | `entity_tick_smoothing_runtime.gd` |
| Integration | `main.gd` |
| Bot | `85_projectile_tick_smoothing.json`, assertion handlers |
| Tests | `test_projectile_tick_smoothing.gd`, `client_smoke.sh` |

## Test and bot proof

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=85_projectile_tick_smoothing
```

## Asset decision

- Adopt: v349 `EntityTickSmoothing` + runtime.
- Reject: separate tween path; external plugins.
