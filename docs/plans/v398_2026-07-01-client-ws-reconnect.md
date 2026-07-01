# v398 Plan — Client WebSocket Reconnect

Status: Complete
Goal: Auto-reconnect the interactive client when the gameplay WebSocket drops mid-session.
Architecture: Extract `ConnectionRecovery` state machine + `ConnectionOverlay` UI; wire from `main.gd`
without growing the coordinator. WS redial first, HTTP `resume_session_id` fallback. No server changes.
Tech stack: shared JSON config, Godot 4 GDScript client.

## Baseline and shortcut decision

Reuses `smoke.gd` resume contract and bot `check_persistence` server path. Overlay borrows
`level_loading_overlay.gd` structure (reject external plugins).

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | `client_reconnect` backoff block |
| Modify | `shared/rules/main_config.v0.schema.json` | Schema for `client_reconnect` |
| Create | `client/scripts/connection_recovery.gd` | Backoff state machine + action requests |
| Create | `client/scripts/connection_recovery_runtime.gd` | Coordinator wiring |
| Create | `client/scripts/connection_overlay.gd` | Reconnecting / failed UI |
| Create | `client/scripts/connection_overlay_bridge.gd` | Scene install helper |
| Modify | `client/scripts/main_config_loader.gd` | `client_reconnect` accessors |
| Modify | `client/scripts/net_client.gd` | `reconnect_ws()`, `resume_same_session()` |
| Modify | `client/scripts/main.gd` | Detect drop, drive recovery, block input |
| Create | `client/tests/test_connection_recovery.gd` | Unit tests |
| Modify | `scripts/client_smoke.sh` | Register unit test |

## Maintenance ratchet

- [x] Extract `connection_recovery.gd` + `connection_recovery_runtime.gd` + overlay modules
- [x] Advance `main.gd` baseline with documented v398 wiring comment

## Task 1 — Shared reconnect tuning

- [x] Step 1.1: Add `client_reconnect` to `main_config.v0.json` + schema

## Task 2 — Client recovery modules

- [x] Step 2.1: `ConnectionRecovery` state machine with unit-testable backoff
- [x] Step 2.2: `ConnectionOverlay` UI
- [x] Step 2.3: `MainConfigLoader` + `NetClient` helpers
- [x] Step 2.4: Wire `main.gd`

## Task 3 — Verification and docs

- [x] Step 3.1: Register test in `client_smoke.sh`
- [x] Step 3.2: `make client-unit` green

## Bot scenarios

Deferred — network drop not stable in merge-gate pack.

## Final verification

- [x] `make validate-shared`
- [x] `make client-unit`
- [x] `make maintainability`
