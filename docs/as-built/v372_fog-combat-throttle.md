# v372 As-Built — Fog Combat Throttle

## What shipped

- `shared/assets/fog_presentation.v0.json` adds `combat_crowd_shader_throttle` (`live_monster_threshold`, `shader_update_min_interval_frames`).
- `fog_of_war_overlay.gd` throttles shader uniform updates when live monster count exceeds the threshold.
- `main.gd` forwards perf-status live monster count into the fog overlay each frame.

## Verification

```bash
godot --headless --path client --script res://tests/test_fog_of_war_overlay.gd
make bot-client SCENARIO=fog_of_war_overlay HEADLESS=1
```
