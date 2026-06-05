from tools.bot.protocol import make_envelope, next_message_id, to_ws_url
from tools.bot.run import load_scenarios, select_scenarios


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
