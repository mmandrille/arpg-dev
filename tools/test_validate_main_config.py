from __future__ import annotations

from tools.validate_main_config import validate_main_config_gameplay


class CapturingReport:
    def __init__(self) -> None:
        self.ok_labels: list[str] = []
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.ok_labels.append(label)

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def _treasure_class_id_for_table(table_id: str) -> str | None:
    mapping = {
        "dungeon_mob_drop": "dungeon_mob_drop",
        "dungeon_depth_1_mob": "dungeon_depth_1_mob",
        "dungeon_depth_2_mob": "dungeon_depth_2_mob",
        "dungeon_depth_3_mob": "dungeon_depth_3_mob",
    }
    return mapping.get(table_id)


def _valid_main_gameplay() -> dict:
    return {
        "base_attack_interval_ticks": 10,
        "base_movement_speed": 4.5,
        "minimum_monster_aggro_radius": 8.0,
        "base_drop_rate_percent": 35,
        "item_upgrade_cost_gold": 100,
        "item_upgrade_max_level": 10,
        "bishop_respec_resource_item_def_id": "badge_respec",
        "bishop_respec_resource_count": 1,
        "bishop_revive_resource_item_def_id": "badge_revive",
        "bishop_revive_resource_count": 1,
        "badge_reward_rules": [
            {
                "resource_item_def_id": "badge_shard",
                "unlock_depth": 1,
                "base_chance_percent": 5,
                "chance_per_depth_percent": 1,
            }
        ],
        "quest_turn_in_item_def_id": "quest_token",
        "quest_turn_in_reward_gold": 50,
        "companion_assist_radius": 12.0,
        "companion_follow_distance": 4.0,
        "companion_follow_stop_radius": 2.0,
    }


def _valid_dungeon_generation() -> dict:
    return {
        "loot_bands": [
            {"monster_loot_table": "dungeon_depth_1_mob", "chest_loot_table": "dungeon_depth_1_chest"},
            {"monster_loot_table": "dungeon_depth_2_mob", "chest_loot_table": "dungeon_depth_2_chest"},
            {"monster_loot_table": "dungeon_depth_3_mob", "chest_loot_table": "dungeon_depth_3_chest"},
        ]
    }


def _valid_treasure_classes() -> dict:
    def _class_with_drop() -> dict:
        return {"attempts": [{"success_weight": 1, "no_drop_weight": 0}]}

    return {
        "dungeon_mob_drop": _class_with_drop(),
        "dungeon_depth_1_mob": _class_with_drop(),
        "dungeon_depth_2_mob": _class_with_drop(),
        "dungeon_depth_3_mob": _class_with_drop(),
    }


def test_validate_main_config_gameplay_passes_with_valid_fixtures() -> None:
    report = CapturingReport()
    validate_main_config_gameplay(
        report,
        _valid_main_gameplay(),
        _valid_dungeon_generation(),
        _valid_treasure_classes(),
        _treasure_class_id_for_table,
    )

    assert report.failures == []
    assert any("attack cadence" in label for label in report.ok_labels)
    assert any("companion" in label for label in report.ok_labels)


def test_validate_main_config_gameplay_fails_on_non_positive_movement_speed() -> None:
    report = CapturingReport()
    gameplay = _valid_main_gameplay()
    gameplay["base_movement_speed"] = 0

    validate_main_config_gameplay(
        report,
        gameplay,
        _valid_dungeon_generation(),
        _valid_treasure_classes(),
        _treasure_class_id_for_table,
    )

    assert any("base_movement_speed" in failure for failure in report.failures)


def test_validate_main_config_gameplay_fails_on_empty_badge_reward_rules() -> None:
    report = CapturingReport()
    gameplay = _valid_main_gameplay()
    gameplay["badge_reward_rules"] = []

    validate_main_config_gameplay(
        report,
        gameplay,
        _valid_dungeon_generation(),
        _valid_treasure_classes(),
        _treasure_class_id_for_table,
    )

    assert any("badge_reward_rules" in failure for failure in report.failures)
