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
- Run multiple Godot instances in the same CI step.
- Test every client scenario — v14 scopes to the core interaction surfaces: movement,
  combat click, loot pickup, and inventory panel (open/equip/unequip/drop).

---

## 3. Files to create / modify

```
client/
  scripts/
    bot_controller.gd          NEW — GDScript bot: scenario runner, synthetic input, assertions
    bot_scenario_runner.gd     NEW — frame-tick step executor (one step per process frame)
    main.gd                    MOD — expose thin read interface for bot (entities, inventory, equipped, player_hp)
  tests/
    test_client_bot.gd         NEW — headless Godot smoke: runs all client scenarios, exits non-zero on failure

tools/
  bot/
    scenarios/client/          NEW — client-side scenario JSON files
      01_click_to_kill.json
      02_inventory_open_close.json
      03_inventory_equip_unequip.json
      04_click_to_move.json

Makefile                       MOD — add `bot-client` target
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
  └─ launches Godot headless --headless --resolution 1280x720
       └─ ARPG_BOT_CLIENT=1  ARPG_BOT_SCENARIO=all (or specific file)
            └─ main.gd._ready()
                 ├─ normal session create (same auth + WS path)
                 └─ if ARPG_BOT_CLIENT:
                      add_child(BotController)
                      BotController.load_scenarios(scenario_dir)
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
  └─ on all scenarios complete:
       get_tree().quit(exit_code)
```

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

---

## 6. Integration with existing machinery

### 6.1 Input locking

`_input_locked()` in `main.gd` currently returns `true` when `visual_replay_enabled or
autoplay_enabled`. The bot must NOT set `autoplay_enabled` or `visual_replay_enabled`.
Instead, `BotController` is the only active agent when `ARPG_BOT_CLIENT=1`. The
`_input_locked()` check must be extended to also return `true` when a bot is running, so
human input does not interfere during windowed runs.

### 6.2 Session lifecycle

The bot reuses the same login + session-create path as the normal client. No new auth or
REST endpoints are needed. The bot reads `client.session_id` and `client.world_id` from
`NetClient` after creation.

### 6.3 `make bot-client` target

```makefile
bot-client:
    ARPG_BOT_CLIENT=1 ARPG_BASE_URL=http://localhost:8080 \
    ARPG_DEV_TOKEN=local-dev-token \
    $(GODOT) --headless --resolution 1280x720 \
             --path client res://main.tscn; \
    test $$? -eq 0
```

Requires `make db-up && make server` to be running (same as `make bot`). Does not launch
the server itself — CI composition handles that.

### 6.4 `make ci` integration

`make ci` runs: `validate-shared → test-go → bot → bot-client → replay`. The new
`bot-client` step runs after the Python bot confirms the server is healthy.

---

## 7. Acceptance criteria

1. `make bot-client` exits 0 with all four client scenarios passing (kill, open/close
   inventory, equip/unequip, click-to-move) against a live server.
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
7. `make ci` is green end-to-end including the new `bot-client` step.
8. Python bot and visual replay are unchanged and still pass.

---

## 8. Open questions / deferred

| # | Question | Status |
|---|----------|--------|
| D-1 | Does headless ray-pick work in Godot 4.6.3 with `--resolution 1280x720`? | Must spike (Task 0 in plan) |
| D-2 | Should client scenarios reuse `tools/bot/scenarios/*.json` format or get a separate dir? | Separate `client/` subdir chosen — avoids Python bot accidentally loading them |
| D-3 | Screenshot diff / pixel-level visual assertion | Deferred: out of scope for v14 |
| D-4 | Multi-bot / competing bots | Deferred: requires multiplayer sessions |
| D-5 | Bot AI complexity (pathfinding goals, behavior trees) | Deferred: v14 uses simple sequential state machine |

---

## 9. Testing plan

1. **Spike (manual):** Run a minimal headless Godot script that creates a session, renders
   one frame, and confirms `camera.unproject_position` + `intersect_ray` returns a hit.
   Command: see Task 0 in plan.
2. **Unit (GDScript):** `make client-smoke` — existing `test_golden.gd` suite must still
   pass; new `test_client_bot.gd` runs all four client scenarios and exits non-zero on
   failure.
3. **Integration:** `make bot-client` against live server — four scenarios, all green.
4. **Regression:** `make ci` end-to-end — Python bot + client bot + replay all green.
