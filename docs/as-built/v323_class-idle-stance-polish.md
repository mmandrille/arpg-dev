# v323 As Built - Class Idle Stance Polish

Date: 2026-06-23

## What Shipped

- Added optional per-class `idle_stance` data (`scale`, `lean_degrees`) in `class_presentations.v0.json`.
- `ClassIdleStance` applies stance offsets when character models mount.

## Proof

```bash
make validate-shared
godot --headless --path client --script res://tests/test_look_and_feel_polish.gd
```
