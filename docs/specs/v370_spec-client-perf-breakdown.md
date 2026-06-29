# v370 Spec: Client Perf Breakdown

Status: Approved
Date: 2026-06-29
Codename: client-perf-breakdown

## Purpose

Extend `ARPG_PERF_DEBUG` client logs with per-subsystem millisecond breakdowns
(`net_poll`, `delta`, `entities`, `fog`) so crowded combat regressions can be
attributed without guessing from aggregate `process_ms`.

## Non-goals

- No gameplay, protocol, or server changes.
- No always-on metrics export or UI overlay.

## Acceptance criteria

- With `ARPG_PERF_DEBUG=1`, `[client-perf]` lines append phase timings when samples run.
- `PerfPhaseTimer` accumulates `delta`, `net_poll`, `entities`, and `fog` buckets.
- Headless unit test covers timer accumulation and formatting.

## Scope

- `client/scripts/perf_phase_timer.gd` (new)
- `client/scripts/perf_debug_sampler.gd`
- `client/scripts/main.gd`, `fog_of_war_overlay.gd`
- `client/tests/test_perf_phase_timer.gd`
- `scripts/client_smoke.sh`

## Verification

```bash
make client-unit
```
