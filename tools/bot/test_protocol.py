import json

from tools.bot.protocol import make_envelope, next_message_id, to_ws_url
from tools.bot.run import (
    RuntimeState,
    find_inventory_item,
    find_interactable,
    find_loot,
    find_monster,
    find_player,
    ingest_message,
    load_scenarios,
    run_assertions,
    run_runtime_assertions,
    select_scenarios,
)


def test_to_ws_url_http():
    assert to_ws_url("http://localhost:8080", "/v0/ws?session_id=s") == "ws://localhost:8080/v0/ws?session_id=s"


def test_to_ws_url_https():
    assert to_ws_url("https://example.test", "/v0/ws?session_id=s") == "wss://example.test/v0/ws?session_id=s"


def test_make_envelope_required_fields():
    env = make_envelope("move_intent", "sess_x", 7, {"direction": {"x": 1, "y": 0}, "duration_ticks": 5})
    assert env["type"] == "move_intent"
    assert env["session_id"] == "sess_x"
    assert env["tick"] == 7
    assert env["message_id"]  # non-empty
    assert env["payload"]["duration_ticks"] == 5
    assert "correlation_id" not in env


def test_message_ids_unique():
    assert next_message_id() != next_message_id()


def test_load_scenarios_discovers_vertical_slice():
    scenarios = load_scenarios()
    ids = {scenario.id for scenario in scenarios}

    assert "vertical_slice" in ids
    vertical = next(s for s in scenarios if s.id == "vertical_slice")
    assert vertical.world_id == "vertical_slice"


def test_load_scenarios_catalog_order():
    scenarios = load_scenarios()

    assert [s.path.name for s in scenarios[:2]] == ["01_vertical_slice.json", "02_gear_before_combat.json"]
    assert [s.id for s in scenarios[:2]] == ["vertical_slice", "gear_before_combat"]


def test_load_scenarios_gear_asserts_outcome_not_timing():
    scenarios = load_scenarios()
    gear = next(s for s in scenarios if s.id == "gear_before_combat")

    assert {
        "type": "monster_dead",
        "monster_def_id": "training_dummy_reward",
    } in gear.assertions


def test_load_scenarios_discovers_door_lab():
    scenarios = load_scenarios()
    door = next(s for s in scenarios if s.id == "door_lab")

    assert door.world_id == "door_lab"
    assert not any("expect_reject" in step for step in door.steps)
    assert {
        "type": "interactable_state",
        "interactable_def_id": "wooden_door",
        "state": "open",
    } in door.assertions


def test_load_scenarios_discovers_path_maze():
    scenarios = load_scenarios()
    maze = next(s for s in scenarios if s.id == "path_maze")

    assert maze.world_id == "path_maze"
    assert maze.steps == [{
        "action": "action_once_until_event",
        "monster_def_id": "training_dummy_path",
        "event_type": "monster_killed",
    }]


def test_load_scenarios_dungeon_levels_loots_coin_before_returning():
    scenarios = load_scenarios()
    dungeon = next(s for s in scenarios if s.id == "dungeon_levels")

    assert dungeon.world_id == "dungeon_levels"
    assert {"action": "pick_up_loot", "item_def_id": "training_badge"} in dungeon.steps
    assert dungeon.steps[-1] == {
        "action": "assert_player_at_used_stair",
        "direction": "down",
        "tolerance": 0.001,
    }
    assert {
        "type": "inventory_contains",
        "item_def_id": "training_badge",
        "equipped": False,
    } in dungeon.assertions


def test_select_scenarios_all_returns_catalog_order():
    scenarios = load_scenarios()

    assert select_scenarios(scenarios, "all") == scenarios


def test_select_scenarios_accepts_file_name_and_stem():
    scenarios = load_scenarios()

    by_file = select_scenarios(scenarios, "06_ranged_lab.json")
    assert [s.id for s in by_file] == ["ranged_lab"]

    by_stem = select_scenarios(scenarios, "06_ranged_lab")
    assert [s.id for s in by_stem] == ["ranged_lab"]


def test_load_scenarios_discovers_inventory_lab():
    scenarios = load_scenarios()
    inventory = next(s for s in scenarios if s.id == "inventory_lab")
    raw = json.loads(inventory.path.read_text())

    assert inventory.world_id == "inventory_lab"
    assert {"action": "unequip_slot", "slot": "main_hand"} in inventory.steps
    assert {"action": "drop_inventory_item", "item_def_id": "rusty_sword"} in inventory.steps
    assert {
        "type": "equipped_weapon_def",
        "item_def_id": "rusty_sword",
    } in inventory.assertions
    assert raw.get("visual", {}).get("inventory_panel") is True


def test_select_scenarios_rejects_unknown_id():
    scenarios = load_scenarios()

    try:
        select_scenarios(scenarios, "missing")
    except ValueError as exc:
        assert "unknown scenario" in str(exc)
    else:
        raise AssertionError("expected ValueError")


def test_load_scenarios_rejects_unknown_world(tmp_path):
    (tmp_path / "01_bad.json").write_text(json.dumps({
        "id": "bad",
        "world_id": "missing",
        "steps": [],
        "assertions": [],
    }))

    try:
        load_scenarios(tmp_path)
    except ValueError as exc:
        assert "unknown world_id" in str(exc)
    else:
        raise AssertionError("expected ValueError")


def test_runtime_state_selectors_from_snapshot_and_delta():
    state = RuntimeState()
    ingest_message({
        "type": "session_snapshot",
        "tick": 0,
        "payload": {
            "server_tick": 0,
            "entities": [
                {"id": "1001", "type": "player", "position": {"x": 0, "y": 5}, "hp": 10, "max_hp": 10},
                {"id": "1002", "type": "loot", "item_def_id": "rusty_sword", "position": {"x": 6, "y": 5}},
                {"id": "1003", "type": "monster", "monster_def_id": "training_dummy_reward", "position": {"x": 12, "y": 5}, "hp": 3, "max_hp": 3},
                {"id": "1004", "type": "interactable", "interactable_def_id": "wooden_door", "state": "closed", "position": {"x": 4, "y": 5}},
            ],
            "inventory": [],
            "equipped": {"main_hand": None},
            "hotbar_capacity": 2,
            "hotbar": [{"slot_index": i, "item_instance_id": None} for i in range(10)],
        },
    }, state)

    assert find_player(state)["id"] == "1001"
    assert find_loot(state, "rusty_sword")["id"] == "1002"
    assert find_monster(state, "training_dummy_reward")["id"] == "1003"
    assert find_interactable(state, "wooden_door")["id"] == "1004"

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "changes": [
                {"op": "entity_remove", "entity_id": "1002"},
                {"op": "inventory_add", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": False}},
                {"op": "inventory_update", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": True}},
                {"op": "equipped_update", "slot": "main_hand", "item_instance_id": "1004"},
                {"op": "inventory_remove", "item_instance_id": "1004"},
            ],
            "events": [],
        },
    }, state)

    assert find_loot(state, "rusty_sword") is None
    assert find_inventory_item(state.inventory, "rusty_sword") is None
    assert state.equipped["main_hand"] == "1004"


def test_runtime_state_waits_for_destination_level_delta():
    state = RuntimeState()
    ingest_message({
        "type": "session_snapshot",
        "tick": 0,
        "payload": {
            "server_tick": 0,
            "current_level": -1,
            "entities": [
                {"id": "1001", "type": "player", "position": {"x": 14, "y": 18}, "hp": 10, "max_hp": 10},
                {"id": "1002", "type": "interactable", "interactable_def_id": "stairs_down", "state": "ready", "position": {"x": 14, "y": 18}},
            ],
            "inventory": [],
            "equipped": {"main_hand": None},
            "hotbar_capacity": 2,
            "hotbar": [{"slot_index": i, "item_instance_id": None} for i in range(10)],
        },
    }, state)

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "level": -1,
            "changes": [{"op": "entity_remove", "entity_id": "1001"}],
            "events": [{"event_type": "level_changed", "from_level": -1, "to_level": -2}],
        },
    }, state)

    assert state.current_level == -2
    assert state.pending_level_load == -2
    assert find_interactable(state, "stairs_down") is None
    assert find_interactable(state, "stairs_up") is None

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "level": -2,
            "changes": [
                {"op": "entity_spawn", "entity": {"id": "1001", "type": "player", "position": {"x": 9, "y": 11}, "hp": 10, "max_hp": 10}},
                {"op": "entity_spawn", "entity": {"id": "1003", "type": "interactable", "interactable_def_id": "stairs_up", "state": "ready", "position": {"x": 9, "y": 11}}},
                {"op": "entity_spawn", "entity": {"id": "1004", "type": "interactable", "interactable_def_id": "stairs_down", "state": "ready", "position": {"x": 28, "y": 14}}},
            ],
            "events": [],
        },
    }, state)

    assert state.pending_level_load is None
    assert find_player(state)["id"] == "1001"
    assert find_interactable(state, "stairs_up")["id"] == "1003"


def test_intent_accepted_increments_pending_attack_count():
    state = RuntimeState()
    state.pending_attack_monsters["msg-attack"] = "training_dummy_reward"

    ingest_message({
        "type": "intent_accepted",
        "tick": 4,
        "payload": {"accepted_message_id": "msg-attack", "server_tick": 4},
    }, state)

    assert state.accepted_attack_counts["training_dummy_reward"] == 1
    assert "msg-attack" not in state.pending_attack_monsters


def test_runtime_assertion_monster_killed_in_attacks_passes():
    state = RuntimeState(
        accepted_attack_counts={"training_dummy_reward": 1},
        killed_monster_def_ids={"training_dummy_reward"},
    )

    run_runtime_assertions([
        {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1}
    ], state, "test")


def test_runtime_assertion_monster_killed_in_attacks_fails_over_max():
    state = RuntimeState(
        accepted_attack_counts={"training_dummy_reward": 2},
        killed_monster_def_ids={"training_dummy_reward"},
    )

    try:
        run_runtime_assertions([
            {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1}
        ], state, "test")
    except AssertionError as exc:
        assert "killed in 2 accepted attacks, max 1" in str(exc)
    else:
        raise AssertionError("expected AssertionError")


def test_structured_assertions():
    entities = [
        {"id": "1001", "type": "player", "hp": 9},
        {"id": "1003", "type": "monster", "monster_def_id": "training_dummy_reward", "hp": 0},
        {"id": "1007", "type": "interactable", "interactable_def_id": "wooden_door", "state": "open"},
    ]
    inventory = [
        {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": True},
        {"item_instance_id": "1006", "item_def_id": "training_badge", "slot": "", "equipped": False},
    ]

    run_assertions([
        {"type": "inventory_count", "equals": 2},
        {"type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": True},
        {"type": "inventory_contains", "item_def_id": "training_badge", "equipped": False},
        {"type": "monster_dead", "monster_def_id": "training_dummy_reward"},
        {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1},
        {"type": "interactable_state", "interactable_def_id": "wooden_door", "state": "open"},
        {"type": "equipped_weapon_def", "item_def_id": "rusty_sword"},
    ], entities, inventory, {"main_hand": "1004"}, None, "test")


def test_structured_assertions_reject_unknown_type():
    try:
        run_assertions([{"type": "nope"}], [], [], {}, None, "test")
    except AssertionError as exc:
        assert "unknown assertion type" in str(exc)
    else:
        raise AssertionError("expected AssertionError")
