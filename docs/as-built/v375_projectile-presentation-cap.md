# v375 As-Built — Projectile Presentation Cap

## What shipped

- `shared/rules/main_config.v0.json` adds `client_perf.projectile_visible_cap`.
- `projectile_presentation_cap.gd` keeps only the nearest N projectile nodes visible each frame.
- Presentation-only cull; authoritative projectiles unchanged on the wire.

## Verification

```bash
godot --headless --path client --script res://tests/test_projectile_presentation_cap.gd
make bot scenario=crowded_melee_perf_probe
```
