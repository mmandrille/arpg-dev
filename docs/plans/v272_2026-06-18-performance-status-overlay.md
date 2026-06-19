# v272 Plan - Performance Status Overlay

## Context

This slice follows v268-v271. The performance data is already measured and logged by the backend;
the user wants it visible during gameplay in the existing top-right status text, renamed
**Performance Status**, with old debug contents removed.

## Tasks

- [x] Extend the `state_delta` schema/example with optional server-owned `performance`.
- [x] Add a backend payload builder that reuses crowded-combat timing, counters, room shape, and
  guardrail state.
- [x] Fan out a throttled performance-only state delta from `sessionLoop` without making clients
  authoritative.
- [x] Add client ping tracking to `NetClient` and capture accepted/rejected round trips in
  `main.gd`.
- [x] Add a focused client formatter and move the status label to the top-right Performance Status
  presentation.
- [x] Add focused backend/client tests, then run:
  - `make validate-shared`
  - `cd server && go test -count=1 ./internal/realtime`
  - `make client-unit`

## Notes

- `state_delta.performance` is optional. Empty `changes` and `events` are valid when the message
  exists only to refresh status telemetry.
- FPS and ping are client-side presentation values; backend timing fields stay authoritative.
- No external assets or plugins are involved.
