# v327 As Built - Chest Open Burst

Date: 2026-06-23

## What Shipped

- `ChestPresentation.sync_open_burst` spawns a short-lived golden torus on chest open.
- `main.gd` interactable state transitions call the burst when treasure chests open.

## Proof

```bash
godot --headless --path client --script res://tests/test_look_and_feel_polish.gd
```
