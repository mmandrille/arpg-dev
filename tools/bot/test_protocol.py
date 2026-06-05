import json

from tools.bot.protocol import make_envelope, next_message_id, to_ws_url
from tools.bot.run import (
    RuntimeState,
    find_inventory_item,
    find_loot,
    find_monster,
    find_player,
    ingest_message,
    load_scenarios,
    run_assertions,
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


def test_select_scenarios_all_returns_catalog_order():
    scenarios = load_scenarios()

    assert select_scenarios(scenarios, "all") == scenarios


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
            ],
            "inventory": [],
            "equipped": {"weapon": None},
        },
    }, state)

    assert find_player(state)["id"] == "1001"
    assert find_loot(state, "rusty_sword")["id"] == "1002"
    assert find_monster(state, "training_dummy_reward")["id"] == "1003"

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "changes": [
                {"op": "entity_remove", "entity_id": "1002"},
                {"op": "inventory_add", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "weapon", "equipped": False}},
                {"op": "inventory_update", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "weapon", "equipped": True}},
                {"op": "equipped_update", "slot": "weapon", "item_instance_id": "1004"},
            ],
            "events": [],
        },
    }, state)

    assert find_loot(state, "rusty_sword") is None
    assert find_inventory_item(state.inventory, "rusty_sword")["equipped"] is True
    assert state.equipped["weapon"] == "1004"


def test_structured_assertions():
    entities = [
        {"id": "1001", "type": "player", "hp": 9},
        {"id": "1003", "type": "monster", "monster_def_id": "training_dummy_reward", "hp": 0},
    ]
    inventory = [
        {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "weapon", "equipped": True},
        {"item_instance_id": "1006", "item_def_id": "training_badge", "slot": "", "equipped": False},
    ]

    run_assertions([
        {"type": "inventory_count", "equals": 2},
        {"type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": True},
        {"type": "inventory_contains", "item_def_id": "training_badge", "equipped": False},
        {"type": "monster_dead", "monster_def_id": "training_dummy_reward"},
    ], entities, inventory, {"weapon": "1004"}, None, "test")


def test_structured_assertions_reject_unknown_type():
    try:
        run_assertions([{"type": "nope"}], [], [], {}, None, "test")
    except AssertionError as exc:
        assert "unknown assertion type" in str(exc)
    else:
        raise AssertionError("expected AssertionError")
