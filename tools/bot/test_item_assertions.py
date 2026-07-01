from tools.bot.run import run_assertions


def test_rolled_inventory_item_suffix_is_opt_in():
    inventory = [{
        "item_instance_id": "3001",
        "item_def_id": "long_sword",
        "item_template_id": "long_sword",
        "display_name": "Sword of Executioner's Mark",
        "rarity": "unique",
        "rolled_stats": {"damage_min": 2, "damage_max": 7},
        "requirements": {"level": 1},
        "effect_ids": ["executioners_mark"],
    }]

    run_assertions([
        {"type": "rolled_inventory_item", "item_def_id": "long_sword", "item_template_id": "long_sword", "rarity": "unique"},
    ], [], inventory, {}, None, "test")

    try:
        run_assertions([
            {
                "type": "rolled_inventory_item",
                "item_def_id": "long_sword",
                "item_template_id": "long_sword",
                "display_name_suffix": "Cave Blade",
            },
        ], [], inventory, {}, None, "test")
    except AssertionError as exc:
        assert "display_name missing suffix Cave Blade" in str(exc)
    else:
        raise AssertionError("explicit display_name_suffix mismatch was not rejected")
