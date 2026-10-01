from tools.bot.runtime_queries import event_matches
from tools.bot.boss_lane_actions import boss_lane_target
from tools.bot.bot_types import RuntimeState


def test_boss_lane_event_match_requires_authoritative_descriptor():
    expected = {
        "event_type": "boss_phase_started",
        "pattern_id": "shifting_bulwark",
        "phase_index": 1,
        "lane_present": True,
        "lane_stage_index": 1,
        "lane_safe_index": 1,
        "lane_count": 3,
    }
    event = {
        "event_type": "boss_phase_started",
        "pattern_id": "shifting_bulwark",
        "phase_index": 1,
        "lane": {"stage_index": 1, "safe_index": 1, "count": 3},
    }
    assert event_matches(event, expected)
    assert not event_matches({key: value for key, value in event.items() if key != "lane"}, expected)
    assert not event_matches({**event, "lane": {**event["lane"], "safe_index": 0}}, expected)
    assert not event_matches({**event, "lane": {**event["lane"], "stage_index": 0}}, expected)


def test_boss_lane_target_follows_latest_authoritative_frame():
    state = RuntimeState(events=[
        {"event_type": "boss_phase_started", "pattern_id": "shifting_bulwark", "lane": {
            "origin": {"x": 10, "y": 10}, "forward": {"x": 1, "y": 0}, "right": {"x": 0, "y": -1},
            "count": 3, "width": 2, "length": 6, "safe_index": 1, "danger_lanes": [0, 2],
        }},
    ])
    assert boss_lane_target(state, "shifting_bulwark", "safe") == {"x": 12.0, "y": 10.0}
    assert boss_lane_target(state, "shifting_bulwark", "danger") == {"x": 12.0, "y": 8.0}
