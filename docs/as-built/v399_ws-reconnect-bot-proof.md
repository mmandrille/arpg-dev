# v399 as-built — WebSocket Reconnect Bot Proof

## What it proved

- Extended client bot scenario `ws_reconnect_proof` simulates an unintended WS drop during
  gameplay and verifies v398 reconnect behavior end-to-end in headless Godot.
- Opt-in `enable_ws_reconnect_proof` allows `ConnectionRecoveryRuntime` to run while
  `bot_mode` is true; `simulate_ws_drop` closes the socket without intentional disconnect.
- Scenario asserts reconnect overlay visibility, input block, same `session_id`, and gameplay
  resync without main-menu navigation.

## Key files

- `tools/bot/scenarios/client/ws_reconnect_proof.json` — extended bot proof
- `client/scripts/bot_step_catalog.gd` — new reconnect step types
- `client/scripts/main.gd` — bot proof hooks + `get_bot_state()` recovery fields
- `client/scripts/connection_recovery_runtime.gd` — bot-proof gate

## Verification

```bash
make bot-client SCENARIO=ws_reconnect_proof HEADLESS=1
make client-unit
```

## Non-goals honored

- No server/protocol changes; no merge-gate pack promotion.
- `main.gd` / panel coordinator extractions deferred to `$refactor`.
