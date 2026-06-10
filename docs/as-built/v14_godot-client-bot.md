# v14 — Godot client bot

**Proves:** The client input pipeline (ray-pick targeting, inventory UI, keyboard shortcuts) can be
driven and asserted by an automated bot running inside `main.tscn` in headless Godot, in CI,
without a human watching.

- `BotController` mounts inside `main.tscn` when `ARPG_BOT_CLIENT=1`; `BotScenarioRunner`
  executes client scenarios one frame-tick step at a time.
- `get_bot_state()` exposes reconciled client state (ws_open, entities, inventory, equipped,
  pending_events) as a read-only dictionary; the bot dispatches intents through `bot_dispatch_action`
  and `bot_dispatch_inventory_intent` which route through the same `client.send()` and
  `_on_inventory_intent_requested()` paths as human input.
- `press_key KEY_I` pushes a real `InputEventKey` through `get_viewport().push_input()` and
  toggles the actual `InventoryPanel` via `_unhandled_input()`.
- Headless ray-pick fallback: `click_entity` dispatches `action_intent` directly (documented fallback;
  `Input.warp_mouse()` has no effect without a real display server, making `get_mouse_position()`
  unreliable for ray-pick targeting in `--headless` mode).
- `scripts/bot_client.sh` discovers `tools/bot/scenarios/client/*.json`, validates each, and runs
  one fresh Godot headless process per scenario, checking for the `[bot-client] PASS` sentinel.
- 5 client scenarios: `click_to_kill`, `inventory_open_close`, `inventory_equip_unequip`,
  `inventory_lab_drop_item`, `click_to_move` — all green against a live server.
- 24 `test_client_bot.gd` unit tests cover scenario parsing, validation, timeout messages, and
  PASS/FAIL sentinel formatting without requiring a live server — wired into `make client-unit`.
- `make bot-client` added to `make/agents.mk`; step 7/8 added to `scripts/ci.sh`; all 8 CI steps green.
- Python bot, replay verification, and visual replay are unchanged.

**As-built headless constraint:** `Input.warp_mouse()` is a no-op in `--headless` mode (no display
server); `click_entity` and `click_floor` therefore use the documented direct fallback (same
`action_intent`/`move_to_intent` WebSocket send path). A manual windowed run confirms ray-pick works
correctly with a real display. Drag-and-drop inventory operations also use the direct path through
`bot_dispatch_inventory_intent`, since Control drag events require a real display server.

**Explicit non-goals:** no multi-scenario concurrency per process, no pixel-level assertion,
no competing/multiplayer bots, no headless ray-pick workaround, no v14 changes to Go server
or shared protocol.
