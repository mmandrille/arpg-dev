# v399 Spec — WebSocket Reconnect Bot Proof

Status: Complete  
Date: 2026-07-01  
Codename: ws-reconnect-bot-proof  
Baseline: v398 `client-ws-reconnect` complete

## Purpose

Add an **extended client bot scenario** that simulates an unintended WebSocket drop during
gameplay, asserts the v398 reconnect overlay and recovery state machine activate, and verifies
same-session resync without returning to the main menu.

## Decisions

| # | Decision |
|---|----------|
| Q-1 | Opt-in `enable_ws_reconnect_proof` bot step enables recovery while `bot_mode` is true |
| Q-2 | `simulate_ws_drop` closes the gameplay socket without `_intentional_disconnect` |
| Q-3 | Extended-only scenario (`ci_tier: extended`); not merge-gate pack |
| Q-4 | Compact `vertical_slice` lab; no navigation setup |

## Non-goals

- `main.gd` or `inventory_panel.gd` coordinator extractions (use `$refactor`)
- Server/protocol changes
- HTTP-resume failure or give-up overlay paths (unit tests remain owner)
- Co-op reconnect proof

## Acceptance criteria

- [ ] Bot step catalog includes `enable_ws_reconnect_proof`, `simulate_ws_drop`, `wait_connection_recovery`, `wait_connection_resync`, `assert_connection_recovery`, `assert_session_unchanged`
- [ ] `ConnectionRecoveryRuntime.tick` runs in bot mode when reconnect proof is enabled
- [ ] `get_bot_state()` exposes connection recovery + overlay debug fields
- [ ] Extended scenario `ws_reconnect_proof` passes headless: drop → overlay → resync → same session
- [ ] `make client-unit` green; scenario registered with `ci_tier: extended`

## Scope and files

| Area | Files |
|------|-------|
| Client | `connection_recovery_runtime.gd`, `connection_overlay.gd`, `main.gd`, `bot_step_catalog.gd`, `bot_wait_handlers.gd`, `bot_assertion_handlers.gd`, `bot_controller.gd`, `bot_action_step_validator.gd` |
| Bot | `tools/bot/scenarios/client/ws_reconnect_proof.json` |
| Tests | `client/tests/test_client_bot.gd`, `client/tests/test_connection_recovery_runtime.gd` |
| Docs | plan, as-built, lifecycle |

## Test and bot proof

```bash
make bot-client SCENARIO=ws_reconnect_proof HEADLESS=1
make client-unit
```

## Asset decision

Reject external assets; reuse existing connection overlay presentation.
