# v399 Plan — WebSocket Reconnect Bot Proof

Spec: [`docs/specs/v399_spec-ws-reconnect-bot-proof.md`](../specs/v399_spec-ws-reconnect-bot-proof.md)

## Task 1 — Bot reconnect proof hook

- [x] Add `_bot_reconnect_proof` flag and `bot_enable_ws_reconnect_proof()` / `bot_simulate_ws_drop()` on `main.gd`
- [x] Pass `bot_reconnect_proof` into `ConnectionRecoveryRuntime.tick`; allow recovery when proof flag is set
- [x] Add `ConnectionOverlay.get_debug_state()` and expose recovery/overlay fields in `get_bot_state()`

## Task 2 — Bot step catalog

- [x] Register new action/wait/assert step types in `bot_step_catalog.gd` + validators
- [x] Implement handlers in `bot_wait_handlers.gd`, `bot_assertion_handlers.gd`, `bot_controller.gd`
- [x] Add `assert_session_unchanged` mirror of existing `assert_session_changed`

## Task 3 — Scenario + tests

- [x] Add `tools/bot/scenarios/client/ws_reconnect_proof.json` (`ci_tier: extended`)
- [x] Extend `test_client_bot.gd` validation coverage for new steps
- [x] Update `test_connection_recovery_runtime.gd` for new `tick` parameter

## Task 4 — Docs

- [x] `docs/as-built/v399_ws-reconnect-bot-proof.md`
- [x] Lifecycle row on `/finish`

## Verification

```bash
make bot-client SCENARIO=ws_reconnect_proof HEADLESS=1
make client-unit
python3 -c "from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()"
```
