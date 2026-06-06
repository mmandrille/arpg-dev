# Spec: `godot-client-bot`

**Status:** Draft
**Branch:** `feature/godot-client-bot`
**Slice:** v14
**Related:** v6 (visual-bot-scenario-runner), v13 (inventory-ui), ADR-0001 §D8.5

---

## 1. Purpose

All existing test tiers exercise the server or passively replay recorded sessions through
the client. None of them drive the actual client input pipeline: ray-pick targeting,
inventory panel drag-drop, keyboard shortcuts, or animation triggers. A human watching
`make bot-visual` is the only current verification that these client code paths work.

This slice adds a **Godot-native bot** (`BotController`) that runs inside the Godot
process and drives `main.gd` through synthetic `InputEvent` injection — the same path a
real human uses. It proves client-side behavior automatically, in CI, without a human
watching.

**Division of responsibility after this slice:**

| Tier | What it owns | When it runs |
|---|---|---|
| Python protocol bot | Server correctness — authoritative state, persistence, replay determinism | `make ci` (always) |
| Visual replay | Human QA — "does this look right?" | `make bot-visual` (on demand) |
| Godot client bot | Client correctness — input handlers, UI widgets, rendering reactions | `make bot-client` (CI + dev) |

Features that touch both layers get tests in both. The Godot bot does not replace the
Python bot; it covers the surface the Python bot structurally cannot reach.

---

## 2. What it does NOT do

- Replace or remove the Python protocol bot or visual replay.
- Test server-side Go logic (that remains the Python bot's job).
- Assert pixel-perfect rendering or screenshot diff.
- Implement competing or multiplayer bots (requires multi-player sessions, currently deferred).
- Run multiple Godot instances concurrently in the same CI step.
- Test every client scenario — v14 scopes to the core interaction surfaces: movement,
  combat click, loot pickup, and inventory panel (open/equip/unequip/drop).

---

## 3. Files to create / modify

```
client/
  scripts/
    bot_controller.gd          NEW — GDScript bot: scenario runner, synthetic input, assertions
    bot_scenario_runner.gd     NEW — frame-tick step executor (one step per process frame)
    main.gd                    MOD — owns bot runtime; exposes thin read interface for bot (entities, inventory, equipped, player_hp)
  tests/
    test_client_bot.gd         NEW — unit-style runner tests only; `main.tscn` owns integration execution

tools/
  bot/
    scenarios/client/          NEW — client-side scenario JSON files
      01_click_to_kill.json
      02_inventory_open_close.json
      03_inventory_equip_unequip.json
      04_inventory_lab_drop_item.json
      05_click_to_move.json

Makefile                       MOD — add `bot-client` target
make/agents.mk                 MOD — expose `bot-client`
scripts/ci.sh                  MOD — run `bot-client` after Python bot
docs/specs/v14_spec-godot-client-bot.md   THIS FILE
```

---

## 4. Data shapes

### 4.1 Client bot scenario JSON

Stored in `tools/bot/scenarios/client/*.json`. Reuses the same scenario envelope as the
Python bot but with a `"runner": "godot_client"` field and a `client_steps` key instead
of `steps`.

```json
{
  "id": "click_to_kill",
  "runner": "godot_client",
  "world_id": "vertical_slice",
  "title": "Click-to-kill: synthetic LMB on monster",
  "description": "Bot clicks the first visible monster, waits for monster_killed, asserts entity removed from scene.",
  "client_steps": [
    { "type": "wait_ws_open",   "timeout_s": 5.0 },
    { "type": "wait_entity",    "entity_type": "monster", "timeout_s": 5.0 },
    { "type": "click_entity",   "entity_type": "monster" },
    { "type": "wait_event",     "event_type": "monster_killed", "timeout_s": 10.0 },
    { "type": "assert_entity_removed", "entity_type": "monster" }
  ]
}
```

### 4.2 Step types (v14 scope)

| Step type | Action |
|---|---|
| `wait_ws_open` | Spin until `client.ready_state() == STATE_OPEN` |
| `wait_entity` | Spin until at least one entity of `entity_type` exists in scene |
| `click_entity` | Project entity world pos → screen; inject `InputEventMouseButton LEFT pressed` |
| `wait_event` | Spin until a matching `event_type` appears in `state_delta.events` |
| `assert_entity_removed` | Assert no entity of that type remains in `entities` dict |
| `press_key` | Inject `InputEventKey` for a named key (`KEY_I`, etc.) |
| `assert_panel_visible` | Assert named UI panel `.visible == true` |
| `wait_inventory_item` | Spin until `inventory` has at least one item |
| `drag_bag_to_weapon_slot` | Project bag item screen pos → weapon slot pos; inject press+motion+release events |
| `assert_equipped` | Assert `equipped["weapon"] != null` |
| `drag_weapon_to_bag` | Reverse: weapon slot → bag area |
| `assert_unequipped` | Assert `equipped["weapon"] == null` |
| `drag_bag_to_outside` | Drag a bag item outside the inventory panel to trigger `drop_intent` |
| `assert_inventory_missing` | Assert inventory no longer contains an item matching `item_def_id` or `item_instance_id` |
| `wait_loot_item` | Spin until a loot entity matching `item_def_id` exists in scene |
| `click_floor` | Project a floor coordinate → screen; inject left click for `move_to_intent` |
| `wait_player_near` | Spin until `predicted_pos` is within `distance` of target world position |

### 4.3 Bot state exposed from main.gd

A new `get_bot_state() -> Dictionary` method on `main.gd` returns a snapshot the bot
reads without coupling to internal variables:

```gdscript
{
  "ws_open": bool,
  "player_hp": int,
  "player_pos": Vector2,       # x, z flat
  "entities": Dictionary,      # same as main.entities
  "inventory": Array,
  "equipped": Dictionary,
  "loot_ids": Array,
  "monster_ids": Array,
  "pending_events": Array,     # events seen since last bot poll (cleared on read)
}
```

---

## 5. Architecture / flow

```
Makefile: make bot-client
  └─ for each selected client scenario, launches Godot headless --headless --resolution 1280x720
       └─ ARPG_BOT_CLIENT=1  ARPG_BOT_SCENARIO=<scenario file>  ARPG_WORLD_ID=<scenario.world_id>
            └─ main.gd._ready()
                 ├─ normal session create (same auth + WS path)
                 └─ if ARPG_BOT_CLIENT:
                      add_child(BotController)
                      BotController.load_scenario(ARPG_BOT_SCENARIO)
                      BotController.start()

BotController._process(delta):
  ├─ BotScenarioRunner.tick(delta)   # one step at a time, timeout guarded
  ├─ on step "click_entity":
  │    pos = camera.unproject_position(entity_world_pos)
  │    get_viewport().push_input(InputEventMouseButton{LEFT, pressed, pos})
  │    get_viewport().push_input(InputEventMouseButton{LEFT, released, pos})
  │    → flows into main.gd._unhandled_input()
  │         → _try_action_at_mouse() → ray-pick → client.send("action_intent")
  │              → server → state_delta → _apply_delta() → entities updated
  ├─ on step "press_key":
  │    get_viewport().push_input(InputEventKey{keycode, pressed})
  │    → flows into main.gd._unhandled_input()
  ├─ on assertion failure or timeout:
  │    print error, set exit_code = 1
  └─ on scenario complete:
       get_tree().quit(exit_code)
```

### Scenario lifecycle

`main.tscn` is the integration owner. It creates the normal client session, connects the
WebSocket, mounts `BotController`, and exits the process when the scenario finishes.

`make bot-client` may accept `SCENARIO=all` or a specific scenario id/file. For `all`, the
shell target runs one Godot process per client scenario rather than trying to reset scene
state inside a single process. Each process reads the scenario JSON before launch, sets
`ARPG_WORLD_ID` from `scenario.world_id`, and starts with a fresh server session. This keeps
scenario state isolated and avoids adding a client-side session reset path.

### Headless ray-pick validation (must spike before full implementation)

`_try_action_at_mouse()` uses physics ray queries that depend on the viewport having a
non-zero size. With `--headless --resolution 1280x720`, Godot 4 creates a virtual
viewport of that size, and physics queries operate on the physics world (not the display).
The spike (Task 0 in the plan) must confirm:

1. `camera.unproject_position(world_pos)` returns a valid screen coord
2. `direct_space_state.intersect_ray()` hits the expected `StaticBody3D` pick collider
3. `collider.get_meta("entity_id")` returns the correct id

If the spike fails, the fallback is a `click_entity_direct` step variant that bypasses
ray-pick and calls `client.send("action_intent", tick, {"target_id": id})` directly,
still exercising the WebSocket + server + delta path (but not the ray-pick client path).
In that case, a separate non-CI manual test documents that ray-pick works in windowed mode.

### Headless GUI drag validation (must spike before full implementation)

Inventory equip, unequip, and drop must use the real `InventoryPanel` Control path, not a
direct `client.send(...)` shortcut. The spike (Task 0 in the plan) must confirm that
synthetic mouse press + motion + release events in `--headless --resolution 1280x720`
trigger:

1. `InventorySlotButton._get_drag_data()`
2. `_drop_data()` on the weapon slot and bag area
3. `NOTIFICATION_DRAG_END` with `gui_is_drag_successful() == false` when dropping outside
   the panel, causing `drop_intent`

If Godot headless does not deliver Control drag events reliably, the fallback is to add a
small test-only input adapter on `InventoryPanel` that invokes the same internal handlers
(`_handle_drop_on_slot` / outside-drop path) after validating screen coordinates. That
fallback still belongs inside the client UI layer and must not bypass `main.gd` intent
routing or the WebSocket.

---

## 6. Integration with existing machinery

### 6.1 Bot mode and input locking

`_input_locked()` in `main.gd` currently returns `true` when `visual_replay_enabled or
autoplay_enabled`. The bot must NOT set `autoplay_enabled` or `visual_replay_enabled`, and
bot mode must not be modeled as a normal `_input_locked()` state. `BotController` is the
only active agent when `ARPG_BOT_CLIENT=1`; if windowed debugging needs physical-input
suppression, implement that as a separate bot-mode guard that still allows bot-dispatched
events through the normal handlers.

### 6.2 Session lifecycle

The bot reuses the same login + session-create path as the normal client. No new auth or
REST endpoints are needed. The bot reads `client.session_id` and `client.world_id` from
`NetClient` after creation.

### 6.3 `make bot-client` target

```makefile
bot-client:
    GODOT="$(GODOT)" BASE_URL="$(BASE_URL)" DEV_TOKEN="$(DEV_TOKEN)" \
    SCENARIO="$(or $(SCENARIO),$(scenario),all)" ./scripts/bot_client.sh
```

`scripts/bot_client.sh` discovers `tools/bot/scenarios/client/*.json`, filters by
`SCENARIO`, validates each scenario has `"runner": "godot_client"` and `world_id`, then
launches one headless `main.tscn` process per selected scenario:

```bash
ARPG_BOT_CLIENT=1 \
ARPG_BOT_SCENARIO="$scenario_path" \
ARPG_WORLD_ID="$world_id" \
ARPG_BASE_URL="$BASE_URL" \
ARPG_DEV_TOKEN="$DEV_TOKEN" \
"$GODOT" --headless --resolution 1280x720 --path client res://main.tscn
```

Requires `make db-up && make server` to be running (same as `make bot`). Does not launch
the server itself — CI composition handles that.

### 6.4 `make ci` integration

`make ci` is implemented by `scripts/ci.sh`, so the slice must update that script, not only
the Makefile docs. CI order becomes:

`validate-shared → test-go → client-unit → bot → bot-client → replay`

The new `bot-client` step runs after the Python bot confirms the server is healthy and
before replay verifies the final recorded session.

### 6.5 Bot input gating

The bot must exercise the same input handlers as a human. Therefore bot mode must not make
`_input_locked()` return `true` for bot-dispatched events, because `_unhandled_input()`,
`_handle_input()`, and inventory intent routing return early when input is locked.

Implementation requirement:

- Physical/manual input may be ignored during bot mode if needed for windowed debugging.
- Synthetic bot events must pass through `_unhandled_input()` and `InventoryPanel` GUI input.
- Add an explicit bot-dispatch guard if necessary, for example `bot_input_active`, instead
  of treating bot mode as equivalent to `visual_replay_enabled` or `autoplay_enabled`.
- Existing `visual_replay_enabled` and `autoplay_enabled` locking behavior remains unchanged.

---

## 7. Acceptance criteria

1. `make bot-client` exits 0 with all five client scenarios passing (kill, open/close
   inventory, equip/unequip, inventory-lab drop item, click-to-move) against a live server.
2. `make bot-client` exits non-zero and prints a clear error when a step assertion fails
   (e.g., `[bot-client] FAIL click_to_kill: assert_entity_removed timed out after 10s`).
3. `make bot-client` exits non-zero when a step times out.
4. The `click_entity` step exercises `_try_action_at_mouse()` → ray-pick → `action_intent`
   (confirmed by server receiving and processing the intent, visible in `state_delta`).
   If the headless spike fails, the fallback step type is used and documented.
5. The `press_key KEY_I` step exercises `_unhandled_input()` → `inventory_panel.toggle()`
   and the `assert_panel_visible` step confirms `inventory_panel.visible == true`.
6. The `drag_bag_to_weapon_slot` step exercises the inventory panel's drag-equip path and
   `assert_equipped` confirms `equipped["weapon"] != null` in scene state.
7. The inventory-lab drop scenario exercises the panel's outside-drop path and confirms the
   item leaves inventory, appears as world loot, can be picked up again, and remains server
   authoritative.
8. `make ci` is green end-to-end including the new `bot-client` step.
9. Python bot and visual replay are unchanged and still pass.

---

## 8. Open questions / deferred

| # | Question | Status |
|---|----------|--------|
| D-1 | Does headless ray-pick work in Godot 4.6.3 with `--resolution 1280x720`? | Must spike (Task 0 in plan) |
| D-2 | Does headless Control drag/drop work in Godot 4.6.3 with `--resolution 1280x720`? | Must spike (Task 0 in plan) |
| D-3 | Should client scenarios reuse `tools/bot/scenarios/*.json` format or get a separate dir? | Separate `client/` subdir chosen — avoids Python bot accidentally loading them |
| D-4 | Screenshot diff / pixel-level visual assertion | Deferred: out of scope for v14 |
| D-5 | Multi-bot / competing bots | Deferred: requires multiplayer sessions |
| D-6 | Bot AI complexity (pathfinding goals, behavior trees) | Deferred: v14 uses simple sequential state machine |

---

## 9. Testing plan

1. **Spike (manual):** Run a minimal headless Godot script that creates a session, renders
   one frame, and confirms `camera.unproject_position` + `intersect_ray` returns a hit.
   Command: see Task 0 in plan.
2. **Spike (manual):** Run a minimal headless Godot script that opens the inventory panel
   and confirms synthetic drag/drop reaches `InventoryPanel` equip, unequip, and outside
   drop paths. Command: see Task 0 in plan.
3. **Unit (GDScript):** `make client-unit` — existing `test_golden.gd` suite must still
   pass; `test_client_bot.gd` validates scenario loading/step validation without requiring
   a live server.
4. **Integration:** `make bot-client` against live server — five scenarios, all green,
   owned by `main.tscn`.
5. **Regression:** `make ci` end-to-end — Python bot + client bot + replay all green.
