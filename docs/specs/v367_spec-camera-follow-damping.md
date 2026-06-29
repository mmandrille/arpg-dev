# v367 Spec: Camera Follow Damping

Status: Draft  
Date: 2026-06-28  
Codename: `camera-follow-damping`  
Baseline: v366 `path-cache-effectiveness`

## Purpose

Data-driven isometric camera follow damping so prediction/reconcile anchor moves do not jerk the view.

## Non-goals

- No chest_view rig changes (camera parented to socket).
- No server/protocol changes.

## Acceptance criteria

- `camera_presentations.v0.json` adds `follow_damping_seconds` per mode.
- `PlayerCameraController.tick_follow(delta)` smooths isometric follow; 0 disables damping.
- `main.gd` calls `tick_follow` during active gameplay.
- Unit test proves damped follow lags then catches up.

## Test proof

```bash
make validate-shared
make client-unit
```
