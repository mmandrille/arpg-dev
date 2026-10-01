"""Cross-check consumable golden fixtures against shared gameplay rules."""


def validate_use_consumable_golden(report, golden, items, main_gameplay):
    item_id = golden["item_def_id"]
    heal = golden["heal"]
    item = items["items"].get(item_id)
    if item is None:
        report.fail("use_consumable item", f"unknown item_def_id {item_id}")
    elif item.get("category") != "consumable":
        report.fail("use_consumable item", f"{item_id} is not consumable")
    elif item.get("heal") != heal:
        report.fail("use_consumable item", "heal range mismatch with item rules")
    else:
        report.ok("use_consumable golden matches consumable item rules")

    potion_rules = main_gameplay.get("potion_rules", {})
    if int(heal["max"]) < int(heal["min"]):
        report.fail("use_consumable heal", "max must be >= min")
    else:
        restore_per_level = int(potion_rules.get("restore_multiplier_per_level", 3))
        bad_health_cases = []
        for case in golden["cases"]:
            level = max(1, int(case.get("item_level", 1)))
            restore = restore_per_level * level
            capped = min(restore, int(case["player_max_hp"]) - int(case["player_hp"]))
            if capped != int(case["expected_heal"]) or int(case["player_hp"]) + capped != int(case["expected_player_hp"]):
                bad_health_cases.append(case["name"])
        if bad_health_cases:
            report.fail("use_consumable cases", f"{len(bad_health_cases)} case(s) violate leveled restore + HP cap: {bad_health_cases}")
        else:
            report.ok("use_consumable cases satisfy configured potion-level restore + HP cap")

    rejuvenation_item = items["items"].get("rejuv_potion")
    rejuvenation_cases = golden.get("rejuvenation_cases", [])
    if not rejuvenation_cases:
        report.fail("use_consumable rejuvenation", "golden has no rejuvenation cases")
    elif rejuvenation_item is None or rejuvenation_item.get("category") != "consumable" or not rejuvenation_item.get("leveled_consumable"):
        report.fail("use_consumable rejuvenation", "rejuv_potion must be a leveled consumable in item rules")
    else:
        min_percent = int(potion_rules.get("rejuv_min_restore_percent", 33))
        bad_rejuvenation_cases = []
        for case in rejuvenation_cases:
            percent = max(min_percent, max(1, int(case["item_level"])))
            hp_restore = (int(case["player_max_hp"]) * percent + 50) // 100
            mana_restore = (int(case["player_max_mana"]) * percent + 50) // 100
            hp_restore = min(hp_restore, int(case["player_max_hp"]) - int(case["player_hp"]))
            mana_restore = min(mana_restore, int(case["player_max_mana"]) - int(case["player_mana"]))
            if (
                hp_restore != int(case["expected_heal"])
                or int(case["player_hp"]) + hp_restore != int(case["expected_player_hp"])
                or mana_restore != int(case["expected_mana"])
                or int(case["player_mana"]) + mana_restore != int(case["expected_player_mana"])
            ):
                bad_rejuvenation_cases.append(case["name"])
        if bad_rejuvenation_cases:
            report.fail("use_consumable rejuvenation", f"{len(bad_rejuvenation_cases)} case(s) violate configured restore/cap formula: {bad_rejuvenation_cases}")
        else:
            report.ok("use_consumable rejuvenation cases satisfy configured percentage restore + resource caps")
