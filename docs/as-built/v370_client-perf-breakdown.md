# v370 As-Built — Client Perf Breakdown

## What shipped

- `PerfPhaseTimer` static helper records per-frame subsystem milliseconds when `ARPG_PERF_DEBUG=1`.
- `[client-perf]` log lines now append `net_poll`, `delta`, `entities`, and `fog` phase totals.
- Hooks in `main.gd` (poll, delta apply, entity smoothing) and `fog_of_war_overlay.gd`.

## Verification

```bash
GODOT=godot godot --headless --path client --script res://tests/test_perf_phase_timer.gd
make client-unit
```
