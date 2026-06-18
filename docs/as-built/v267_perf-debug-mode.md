# v267 As-Built - Perf Debug Mode

Date: 2026-06-18
Spec: [`docs/specs/v267_spec-perf-debug-mode.md`](../specs/v267_spec-perf-debug-mode.md)
Plan: [`docs/plans/v267_2026-06-18-perf-debug-mode.md`](../plans/v267_2026-06-18-perf-debug-mode.md)

## Shipped Behavior

- `ARPG_PERF_DEBUG=1 make play` now passes perf debug mode to both the backend server and the Godot
  client.
- Backend realtime sessions emit sampled structured `backend_perf` logs with tick, total tick-loop
  ms, sim ms, input/result/change/event/ack/reject counts, client count, active game level, entity
  breakdown, and wall count.
- The client emits sampled `[client-perf]` lines with FPS, average frame ms, process/physics ms,
  server tick, WebSocket state, reconciliation delta, entity breakdown, node count, render object
  count, draw calls, and primitive count.
- Sampling is opt-in and roughly once per second, keeping default local play unchanged.

## Boundaries

- No gameplay tuning, enemy-count changes, protocol changes, dashboard, or profiler integration
  shipped.
- Backend perf logs currently cover the active session loop used by local/co-op play.

## Usage

```bash
ARPG_PERF_DEBUG=1 make play
```

Look for:

```text
[backend] {"message":"backend_perf", ...}
[client1] [client-perf] fps=...
```

If backend `total_ms` or `sim_ms` approaches the tick budget while client frame time is fine, focus
on simulation/persistence/fanout. If backend timings stay low while `[client-perf]` FPS/frame ms is
bad, focus on Godot rendering, node counts, or client-side entity updates.

## Verification

```bash
cd server && go test ./internal/game ./internal/realtime
make client-unit
make maintainability
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
```

All focused commands passed on 2026-06-18.
