# v272 As-Built - Performance Status Overlay

Spec: [`docs/specs/v272_spec-performance-status-overlay.md`](../specs/v272_spec-performance-status-overlay.md)
Plan: [`docs/plans/v272_2026-06-18-performance-status-overlay.md`](../plans/v272_2026-06-18-performance-status-overlay.md)

## Shipped

- `state_delta` now has an optional strict `performance` payload with backend tick timing,
  path/crowd counters, tick-budget status, loop counts, and room shape.
- The authoritative `sessionLoop` builds that payload from the same profiler/counters used by the
  crowded-combat logs and fanouts a throttled performance-only state delta with empty changes/events.
- The client status toggle now controls a top-right **Performance Status** panel instead of the old
  generic debug text. The panel shows FPS, best-effort intent ping, backend timing, path counters,
  monster/wall/entity counts, loop counts, and overrun/degradation status.
- The settings label was renamed from `Status text` to `Performance status` in English and Spanish.

## Proof

- `make validate-shared`
- `cd server && go test -count=1 ./internal/realtime`
- `godot --headless --path client --script res://tests/test_net_client.gd`
- `godot --headless --path client --script res://tests/test_coop_client.gd`
- `make client-unit`
- `make maintainability`
- `make ci` (green on 2026-06-18)

## Notes

- Ping is client-estimated from `intent_accepted`/`intent_rejected` round trips. Backend timing
  remains server-owned.
- The performance delta is presentation telemetry only; clients still do not own AI, navigation,
  combat, loot, or persistence truth.
