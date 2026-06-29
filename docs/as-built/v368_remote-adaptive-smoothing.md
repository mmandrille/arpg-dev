# v368 As-Built — Remote Adaptive Smoothing

## What shipped

- `movement_presentation.v0.json` `remote_adaptive` block (enabled, min/max duration, distance_per_second).
- `EntityTickSmoothing.begin_segment` accepts duration override; runtime scales remote entity segments.
- Non-local `player`/`monster`/`companion` entities use adaptive duration in `main.gd`.

## Proof

```bash
make validate-shared
/opt/homebrew/bin/godot --headless --path client --script res://tests/test_entity_tick_smoothing.gd
```
