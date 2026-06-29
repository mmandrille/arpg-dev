# v367 As-Built — Camera Follow Damping

## What shipped

- `follow_damping_seconds` in `camera_presentations.v0.json` (isometric **0.12**, chest_view **0**).
- `PlayerCameraController.tick_follow(delta)` exponential isometric follow; `sync_to_player` snaps on mode change.
- `main.gd` calls `tick_follow` each gameplay frame.

## Proof

```bash
make validate-shared
/opt/homebrew/bin/godot --headless --path client --script res://tests/test_camera_mode_settings.gd
```
