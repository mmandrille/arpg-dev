# Godot Client Bot Implementation Plan

> **For agentic workers:** implement task-by-task and keep checkbox status current. Do not accept a bot that sends intents directly unless the required headless spike proves the real input path is unavailable and the documented fallback is applied only for that specific path.

**Goal:** Add a Godot-native client bot that runs inside `main.tscn`, drives the same input handlers as a human through synthetic events, and proves click targeting, click-to-move, inventory toggle, inventory equip/unequip, and inventory drop/pickup flows in CI.

**Architecture:** The Go server remains authoritative for every gameplay outcome. The Godot bot is client-side automation only: it injects input, waits for authoritative snapshots/deltas, and asserts reconciled client state. It does not replace the Python protocol bot or visual replay.

**Tech stack:** Godot 4.6.x GDScript runner, existing `main.gd` / `InventoryPanel` input paths, client scenario JSON, Makefile/script integration, and the existing live server used by bot gates.

**Spec:** [`docs/specs/v14_spec-godot-client-bot.md`](../specs/v14_spec-godot-client-bot.md)

**Branch:** `feature/godot-client-bot`.

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/bot_controller.gd` | Load one client scenario, mount the runner, print PASS/FAIL sentinels, and quit with the right exit code |
| Create | `client/scripts/bot_scenario_runner.gd` | Execute `client_steps` one at a time with per-step timeouts and assertion errors |
| Modify | `client/scripts/main.gd` | Mount bot mode, expose `get_bot_state()`, track pending events, and keep bot input flowing through normal handlers |
| Create | `client/tests/test_client_bot.gd` | Unit-style tests for scenario parsing, step validation, timeout handling, and sentinels |
| Create | `tools/bot/scenarios/client/01_click_to_kill.json` | Synthetic entity click through ray-pick and `action_intent` |
| Create | `tools/bot/scenarios/client/02_inventory_open_close.json` | `KEY_I` toggle through `_unhandled_input()` |
| Create | `tools/bot/scenarios/client/03_inventory_equip_unequip.json` | Inventory drag from bag to weapon slot and back |
| Create | `tools/bot/scenarios/client/04_inventory_lab_drop_item.json` | Inventory outside-drop, loot respawn, and pickup again |
| Create | `tools/bot/scenarios/client/05_click_to_move.json` | Synthetic floor click through `move_to_intent` |
| Create | `scripts/bot_client.sh` | Discover/filter client scenarios and launch one headless Godot process per scenario |
| Modify | `Makefile` | Add/help expose `bot-client` target if top-level help is maintained there |
| Modify | `make/agents.mk` | Add `bot-client` target |
| Modify | `scripts/ci.sh` | Run `bot-client` in the CI sequence |
| Modify | `scripts/client_smoke.sh` | Include `test_client_bot.gd` in client unit/smoke gates if needed |
| Modify | `PROGRESS.md` | Update only when v14 ships |

---

---

## Task 0: Headless Input Feasibility Spikes

- [ ] **Step 0.1:** Confirm headless ray-pick before implementing all click scenarios. Run a minimal headless path at `--resolution 1280x720` that:
  - creates or joins a normal session,
  - waits until a monster pick body exists,
  - calls `camera.unproject_position(entity_world_pos)`,
  - performs the same physics ray query as `_try_action_at_mouse()`,
  - verifies the hit collider exposes the expected `entity_id` metadata.

- [ ] **Step 0.2:** If Step 0.1 passes, keep `click_entity` as the default entity-click implementation and assert it flows through `_try_action_at_mouse()` to `action_intent`.

- [ ] **Step 0.3:** If Step 0.1 fails because Godot headless cannot ray-pick reliably, implement a documented `click_entity_direct` fallback that sends the same `action_intent` through `NetClient`. Also add a manual windowed ray-pick check note in this plan or the slice completion notes. Do not use direct sending unless the spike fails.

- [ ] **Step 0.4:** Confirm headless Control drag/drop before implementing inventory drag scenarios. In `--headless --resolution 1280x720`, synthetic mouse press + motion + release must trigger:
  - `InventorySlotButton._get_drag_data()`,
  - `_drop_data()` on the weapon slot and bag area,
  - `NOTIFICATION_DRAG_END` with `gui_is_drag_successful() == false` for outside drop.

- [ ] **Step 0.5:** If Step 0.4 fails, add a small test-only input adapter inside the inventory UI layer. The adapter must validate screen coordinates and invoke the same internal inventory drop/equip handlers. It must still route intents through `main.gd` and WebSocket, not mutate inventory state directly.

---

## Task 1: Bot Scenario Contract

- [ ] **Step 1.1:** Create `tools/bot/scenarios/client/` and keep these scenarios separate from Python protocol scenarios so `tools/bot/run.py` does not discover them accidentally.

- [ ] **Step 1.2:** Implement scenario validation for:
  - `runner == "godot_client"`,
  - non-empty `id`,
  - valid `world_id`,
  - non-empty `client_steps`,
  - known step `type`,
  - required fields per step,
  - positive `timeout_s` where timeout is supported.

- [ ] **Step 1.3:** Create `01_click_to_kill.json`:
  - `world_id: "vertical_slice"`,
  - wait for WebSocket open,
  - wait for a monster,
  - click the monster,
  - wait for `monster_killed`,
  - assert the monster is removed or dead according to current client entity behavior.

- [ ] **Step 1.4:** Create `02_inventory_open_close.json`:
  - use a world that supports inventory assertions, preferably `inventory_lab`,
  - press `KEY_I`,
  - assert inventory panel visible,
  - press `KEY_I`,
  - assert inventory panel hidden.

- [ ] **Step 1.5:** Create `03_inventory_equip_unequip.json`:
  - `world_id: "inventory_lab"`,
  - pick up `rusty_sword` through the client click path,
  - open the inventory panel,
  - drag bag item to weapon slot,
  - assert `equipped.weapon != null`,
  - drag weapon slot back to bag,
  - assert `equipped.weapon == null`.

- [ ] **Step 1.6:** Create `04_inventory_lab_drop_item.json`:
  - `world_id: "inventory_lab"`,
  - pick up `rusty_sword`,
  - open panel,
  - drag bag item outside the panel,
  - assert the inventory item is missing,
  - wait for matching loot in the scene,
  - click the loot,
  - assert the inventory contains the item again.

- [ ] **Step 1.7:** Create `05_click_to_move.json`:
  - choose an unobstructed target in an existing world,
  - click the floor coordinate,
  - wait until the predicted/player position is within the configured distance.

---

## Task 2: Bot Runtime

- [ ] **Step 2.1:** Create `client/scripts/bot_controller.gd` with:
  - scenario loading from `ARPG_BOT_SCENARIO`,
  - startup validation,
  - clear `[bot-client] RUN/PASS/FAIL <scenario_id>` logs,
  - process exit `0` on success and non-zero on validation, timeout, or assertion failure.

- [ ] **Step 2.2:** Create `client/scripts/bot_scenario_runner.gd` with a frame-tick executor. It should run one step at a time, track elapsed time, and produce failures with scenario id, step index, step type, expected state, observed state, and timeout.

- [ ] **Step 2.3:** Implement wait/assertion steps:
  - `wait_ws_open`,
  - `wait_entity`,
  - `wait_event`,
  - `assert_entity_removed`,
  - `assert_panel_visible`,
  - `wait_inventory_item`,
  - `assert_equipped`,
  - `assert_unequipped`,
  - `assert_inventory_missing`,
  - `wait_loot_item`,
  - `wait_player_near`.

- [ ] **Step 2.4:** Implement synthetic input steps:
  - `press_key`,
  - `click_entity`,
  - `click_floor`,
  - `drag_bag_to_weapon_slot`,
  - `drag_weapon_to_bag`,
  - `drag_bag_to_outside`.

- [ ] **Step 2.5:** Make synthetic mouse events include press and release at stable viewport positions. Drag steps should include enough motion frames to satisfy Godot Control drag thresholds in headless mode.

- [ ] **Step 2.6:** Keep bot assertions based on reconciled state from `main.gd`, not local optimistic predictions.

---

## Task 3: `main.gd` Integration

- [ ] **Step 3.1:** In `main.gd._ready()`, when `ARPG_BOT_CLIENT=1`, mount `BotController` after normal startup/session setup has begun. Bot mode must use the same auth, session create, and WebSocket path as a normal client.

- [ ] **Step 3.2:** Pass these environment values through the existing runtime config path:
  - `ARPG_BOT_CLIENT=1`,
  - `ARPG_BOT_SCENARIO=<scenario file>`,
  - `ARPG_WORLD_ID=<scenario.world_id>`,
  - `ARPG_BASE_URL`,
  - `ARPG_DEV_TOKEN`.

- [ ] **Step 3.3:** Add `get_bot_state() -> Dictionary` returning:
  - `ws_open`,
  - `player_hp`,
  - `player_pos`,
  - `entities`,
  - `inventory`,
  - `equipped`,
  - `loot_ids`,
  - `monster_ids`,
  - `pending_events`.

- [ ] **Step 3.4:** Track authoritative `state_delta.events` in a bot pending-events buffer. Clear that buffer only after `get_bot_state()` returns it to the bot.

- [ ] **Step 3.5:** Do not model bot mode as input-locked. `_input_locked()` should continue locking visual replay/autoplay paths, but bot-dispatched events must pass through `_unhandled_input()` and `InventoryPanel` GUI input.

- [ ] **Step 3.6:** If windowed bot debugging needs manual input suppression, add an explicit bot-dispatch guard such as `bot_input_active`. Do not block the bot's own synthetic events.

- [ ] **Step 3.7:** Expose only read-oriented helpers/state to the bot. The bot must not mutate gameplay state or inventory directly.

---

## Task 4: Runner Script And Make Targets

- [ ] **Step 4.1:** Add `scripts/bot_client.sh`. It should:
  - fail if `GODOT` is missing or not executable,
  - discover `tools/bot/scenarios/client/*.json`,
  - support `SCENARIO=all`, scenario id, or explicit scenario path,
  - validate `runner`, `world_id`, and `client_steps` before launch,
  - run scenarios sequentially,
  - fail if any scenario exits non-zero or omits the PASS sentinel.

- [ ] **Step 4.2:** Launch one fresh Godot process per scenario:

```bash
ARPG_BOT_CLIENT=1 \
ARPG_BOT_SCENARIO="$scenario_path" \
ARPG_WORLD_ID="$world_id" \
ARPG_BASE_URL="$BASE_URL" \
ARPG_DEV_TOKEN="$DEV_TOKEN" \
"$GODOT" --headless --resolution 1280x720 --path client res://main.tscn
```

- [ ] **Step 4.3:** Add `bot-client` to `make/agents.mk`:

```makefile
bot-client:
 GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" \
 SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_client.sh
```

- [ ] **Step 4.4:** Expose `bot-client` in top-level Makefile help if help text is maintained outside `make/agents.mk`.

- [ ] **Step 4.5:** Document that `make bot-client` requires a live DB/server, like `make bot`; it should not start the server itself.

---

## Task 5: Unit And Smoke Coverage

- [ ] **Step 5.1:** Add `client/tests/test_client_bot.gd` for runner behavior that does not require a live server:
  - valid scenario loads,
  - invalid runner is rejected,
  - unknown step type is rejected,
  - missing required step field is rejected,
  - timeout failure includes scenario id and step type,
  - success/failure sentinel formatting is stable.

- [ ] **Step 5.2:** Wire `test_client_bot.gd` into the existing Godot unit path used by `make client-unit`.

- [ ] **Step 5.3:** Keep `make client-smoke` green. Do not make the existing smoke path require a live server unless it already does for that mode.

- [ ] **Step 5.4:** Add focused tests around fallback adapters only if a headless spike forces a fallback. Test that fallback adapters still call the same intent-routing methods used by GUI handlers.

---

## Task 6: CI Integration

- [ ] **Step 6.1:** Update `scripts/ci.sh` labels/counts and run order to include `bot-client`:

```text
validate-shared -> test-go -> client-unit -> bot -> bot-client -> replay
```

- [ ] **Step 6.2:** Ensure the `bot-client` CI step uses the already-running CI server and fails hard if Godot is unavailable.

- [ ] **Step 6.3:** Keep Python protocol bot and replay verification unchanged. The new client bot is an additional gate, not a replacement.

- [ ] **Step 6.4:** Confirm `make bot-visual` behavior is unchanged; bot mode must not affect visual replay or autoplay input locking.

---

## Task 7: Verification

- [ ] **Step 7.1:** Run client unit tests:

```bash
make client-unit
```

- [ ] **Step 7.2:** With DB and server running, run all client scenarios:

```bash
make bot-client
```

- [ ] **Step 7.3:** Run the existing Python protocol bot:

```bash
make bot
```

- [ ] **Step 7.4:** Run the full gate:

```bash
make ci
```

- [ ] **Step 7.5:** If the ray-pick fallback was used, run and record a manual windowed check showing real ray-pick works outside headless mode. Use the implemented bot launch path, but remove `--headless` from the Godot invocation and point `ARPG_BOT_SCENARIO` at `tools/bot/scenarios/client/01_click_to_kill.json`.

```bash
ARPG_BOT_CLIENT=1 \
ARPG_BOT_SCENARIO=tools/bot/scenarios/client/01_click_to_kill.json \
ARPG_WORLD_ID=vertical_slice \
"$GODOT" --resolution 1280x720 --path client res://main.tscn
```

---

## Task 8: Completion Docs

- [ ] **Step 8.1:** Update `PROGRESS.md` only after v14 is implemented and `make ci` is green.

- [ ] **Step 8.2:** Mark the v14 spec status complete only after acceptance criteria are met.

- [ ] **Step 8.3:** If any fallback path was required, document exactly which headless behavior failed, which fallback was adopted, and what follow-up would remove it.

---

## Acceptance Checklist

- [ ] `make bot-client` exits `0` with all five client scenarios passing against a live server.
- [ ] `make bot-client` exits non-zero with a clear `[bot-client] FAIL ...` line on assertion failure.
- [ ] Step timeouts exit non-zero and include scenario id, step type, and timeout.
- [ ] Entity click uses `_try_action_at_mouse()` -> ray-pick -> `action_intent`, unless the documented headless spike failed and fallback was adopted.
- [ ] `press_key KEY_I` uses `_unhandled_input()` and toggles the real inventory panel.
- [ ] Inventory drag-equip and drag-unequip use `InventoryPanel` UI behavior and reconcile from authoritative state.
- [ ] Inventory outside-drop removes the item from inventory, spawns world loot, and allows pickup again.
- [ ] `make ci` is green end to end including `bot-client`.
- [ ] Python bot and visual replay remain unchanged and green.

---

## Assumptions And Constraints

- Godot is installed in CI; `bot-client` must not silently skip when Godot is missing.
- No server, shared protocol, or schema change is required for v14.
- Each scenario gets one fresh Godot process and one fresh server session.
- Bot scenarios assert client behavior; Python scenarios continue to own server correctness, persistence, and replay determinism.
- Fallbacks are allowed only for proven headless limitations and must stay inside the client input/UI layer.
