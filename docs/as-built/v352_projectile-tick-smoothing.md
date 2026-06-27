# v352 As-Built — Projectile Tick Smoothing

Date: 2026-06-26  
Spec: [`docs/specs/v352_spec-projectile-tick-smoothing.md`](../specs/v352_spec-projectile-tick-smoothing.md)  
Plan: [`docs/plans/v352_2026-06-26-projectile-tick-smoothing.md`](../plans/v352_2026-06-26-projectile-tick-smoothing.md)

## Shipped behavior

- **`movement_presentation.v0.json`**: `projectiles_enabled`, `projectile_snap_distance` under
  `tick_smoothing`.
- **Authoritative projectiles** use `EntityTickSmoothingRuntime.apply_projectile_authoritative`
  instead of SceneTree tweens; facing updates on segments and during interpolation.
- **Bot debug** exposes `projectile_tick_smoothing`.
- **Extended bot proof**: `85_projectile_tick_smoothing` in archer lab.

## Boundaries

- Skill-authored preview projectiles unchanged.
- No leap/charge/teleport smoothing (v353).

## Verification

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_projectile_tick_smoothing.gd
HEADLESS=1 make bot-visual scenario=85_projectile_tick_smoothing
```
