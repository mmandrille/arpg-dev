# v398 as-built — Client WebSocket reconnect

## What it proved

- Interactive gameplay auto-detects unintended WebSocket `CLOSING`/`CLOSED` after the first
  `session_snapshot` and enters a blocking **Reconnecting…** overlay.
- Retry uses data-driven exponential backoff from `shared/rules/main_config.v0.json` →
  `client_reconnect` (WS redial first; HTTP `resume_session_id` after repeated WS failures).
- Successful resync sends `client_ready` with `last_seen_tick`, applies `session_snapshot`, and
  resumes input without main-menu navigation.
- Give-up shows **Connection lost** + **Return to menu**; cancel (after first attempt) returns to
  menu via existing teardown.

## Key files

- `client/scripts/connection_recovery.gd` — backoff state machine
- `client/scripts/connection_recovery_runtime.gd` — coordinator wiring via callables
- `client/scripts/connection_overlay.gd` + `connection_overlay_bridge.gd` — UI
- `client/scripts/net_client.gd` — `reconnect_ws()`, `resume_same_session()`
- `client/tests/test_connection_recovery.gd`

## Verification

```bash
make validate-shared
make client-unit
make maintainability
```

## Non-goals honored

- No server/protocol changes; no merge-gate bot scenario for network drop simulation.
