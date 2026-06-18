# v267 Spec - Perf Debug Mode

Status: Complete
Date: 2026-06-18
Codename: perf-debug-mode

## Purpose

Make local lag investigations measurable by adding an opt-in perf debug mode that prints sampled
backend and client performance lines during `make play`.

## Non-goals

- No gameplay tuning or enemy-count balance changes.
- No production metrics dashboard, trace backend, or profiler integration.
- No protocol/schema change.
- No always-on logging; perf output must remain opt-in.

## Acceptance Criteria

- `ARPG_PERF_DEBUG=1 make play` enables both backend and Godot client perf sampling.
- Backend logs include `backend_perf` with tick, total tick-loop ms, sim ms, input/result/change/event
  counts, client count, game level, and active-level entity counts.
- Client logs include `[client-perf]` with FPS, average frame ms, process/physics ms, server tick,
  reconciliation delta, entity breakdown, node count, render object count, draw calls, and primitive
  count.
- Perf logs are sampled around once per second to avoid creating the lag being diagnosed.
- Default `make play` behavior remains unchanged when `ARPG_PERF_DEBUG` is unset.
- File-size ratchets remain green.

## Scope and Likely Files

- Server:
  - `server/internal/game/perf_debug.go`
  - `server/internal/realtime/perf_debug.go`
  - `server/internal/realtime/session_loop.go`
- Client:
  - `client/scripts/perf_debug_sampler.gd`
  - `client/scripts/main.gd`
- Tooling:
  - `scripts/play.sh`
- Docs:
  - `docs/plans/v267_2026-06-18-perf-debug-mode.md`
  - `docs/as-built/v267_perf-debug-mode.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Test and Bot Proof

```bash
cd server && go test ./internal/game ./internal/realtime
make client-unit
make maintainability
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
```
