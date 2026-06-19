from __future__ import annotations

from typing import Any, Callable


def validate_main_config_gameplay(
    report: Any,
    main_gameplay: dict[str, Any],
    dungeon_generation: dict[str, Any],
    treasure_class_defs: dict[str, Any],
    treasure_class_id_for_table: Callable[[str], str | None],
) -> None:
    if int(main_gameplay.get("base_attack_interval_ticks", 0)) <= 0:
        report.fail("main_config gameplay", "base_attack_interval_ticks must be positive")
    elif float(main_gameplay.get("base_movement_speed", 0)) <= 0:
        report.fail("main_config gameplay", "base_movement_speed must be positive")
    else:
        report.ok("main_config gameplay owns attack cadence and movement speed")

    def treasure_class_at_least_one_drop_rate(class_id: str) -> int | None:
        treasure_class = treasure_class_defs.get(class_id)
        if treasure_class is None:
            return None
        no_drop_chance = 1.0
        for attempt in treasure_class.get("attempts", []):
            total = int(attempt.get("success_weight", 0)) + int(attempt.get("no_drop_weight", 0))
            if total <= 0:
                continue
            no_drop_chance *= int(attempt.get("no_drop_weight", 0)) / total
        return int(round((1.0 - no_drop_chance) * 100))

    monster_drop_sources = {}
    for table_id in ["dungeon_mob_drop", *[band["monster_loot_table"] for band in dungeon_generation.get("loot_bands", [])]]:
        treasure_class_id = treasure_class_id_for_table(table_id)
        if not treasure_class_id:
            continue
        monster_drop_sources[treasure_class_id] = treasure_class_at_least_one_drop_rate(treasure_class_id)
    expected_base_drop_rate = int(main_gameplay.get("base_drop_rate_percent", -1))
    if expected_base_drop_rate < 0 or expected_base_drop_rate > 100:
        report.fail("main_config gameplay", "base_drop_rate_percent must be within [0,100]")
    elif any(rate is None for rate in monster_drop_sources.values()):
        report.fail("main_config gameplay", f"dungeon monster drop profile has unresolved sources {monster_drop_sources}")
    else:
        report.ok("main_config gameplay owns dungeon monster drop rate")

    if int(main_gameplay.get("item_upgrade_cost_gold", -1)) < 0:
        report.fail("main_config gameplay", "item_upgrade_cost_gold must be non-negative")
    elif int(main_gameplay.get("item_upgrade_max_level", 0)) <= 0:
        report.fail("main_config gameplay", "item_upgrade_max_level must be positive")
    else:
        report.ok("main_config gameplay owns starter item upgrade tuning")

    if not str(main_gameplay.get("quest_turn_in_item_def_id", "")):
        report.fail("main_config gameplay", "quest_turn_in_item_def_id must be non-empty")
    elif int(main_gameplay.get("quest_turn_in_reward_gold", -1)) < 0:
        report.fail("main_config gameplay", "quest_turn_in_reward_gold must be non-negative")
    else:
        report.ok("main_config gameplay owns quest turn-in reward tuning")

    if float(main_gameplay.get("companion_assist_radius", 0)) <= 0:
        report.fail("main_config gameplay", "companion_assist_radius must be positive")
    elif float(main_gameplay.get("companion_follow_distance", 0)) <= 0:
        report.fail("main_config gameplay", "companion_follow_distance must be positive")
    elif float(main_gameplay.get("companion_follow_stop_radius", 0)) <= 0:
        report.fail("main_config gameplay", "companion_follow_stop_radius must be positive")
    else:
        report.ok("main_config gameplay owns companion and elite-minion follow tuning")
