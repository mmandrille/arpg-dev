from tools.bot.bot_types import RuntimeState
from tools.bot.stash_assertions import (
    assert_stash_event,
    filtered_stash_items,
    find_stash_item_by_id,
    select_stash_item,
)


def test_stash_assertion_helpers_filter_select_and_find():
    state = RuntimeState()
    state.stash_items = [
        {"stash_item_id": "9002", "item_def_id": "long_sword", "item_template_id": "long_sword", "display_name": "Long Sword"},
        {"stash_item_id": "9001", "item_def_id": "gold", "display_name": "Gold", "amount": 10},
        {"stash_item_id": "9003", "item_def_id": "bow", "item_template_id": "bow", "display_name": "Magic Cave Bow"},
    ]

    assert [item["stash_item_id"] for item in filtered_stash_items(state.stash_items, {"rolled": True})] == ["9002", "9003"]
    assert [item["stash_item_id"] for item in filtered_stash_items(state.stash_items, {"rolled": False})] == ["9001"]
    assert select_stash_item(state, {"item_template_id": "bow"})["stash_item_id"] == "9003"
    assert find_stash_item_by_id(state.stash_items, "9002")["item_def_id"] == "long_sword"
    assert find_stash_item_by_id(state.stash_items, None) is None


def test_stash_assertion_helpers_reject_missing_and_bad_index():
    state = RuntimeState()
    state.stash_items = [{"stash_item_id": "9001", "item_def_id": "long_sword"}]

    try:
        select_stash_item(state, {"action": "withdraw_stash_item", "item_def_id": "bow"})
    except AssertionError as exc:
        assert "no matching stash item" in str(exc)
    else:
        raise AssertionError("missing stash item was not rejected")

    try:
        select_stash_item(state, {"action": "withdraw_stash_item", "stash_index": 3})
    except AssertionError as exc:
        assert "stash_index 3 out of range" in str(exc)
    else:
        raise AssertionError("bad stash index was not rejected")


def test_stash_event_assertion_helper_filters_by_stash_and_item():
    events = [
        {"event_type": "stash_item_withdrawn", "stash_id": "account_stash", "stash_item_id": "9001"},
        {"event_type": "stash_item_withdrawn", "stash_id": "unique_test_chest", "stash_item_id": "9001"},
        {"event_type": "stash_item_withdrawn", "stash_id": "account_stash", "stash_item_id": "9002"},
    ]
    calls = []

    def record_count(got, assertion, label, suffix=""):
        calls.append((got, label, suffix))

    assert_stash_event(
        events,
        {"event_type": "stash_item_withdrawn", "stash_id": "account_stash", "stash_item_id": "9001", "equals": 1},
        "unit",
        record_count,
    )

    assert calls == [(1, "unit: stash_event", ": [{'event_type': 'stash_item_withdrawn', 'stash_id': 'account_stash', 'stash_item_id': '9001'}]")]
