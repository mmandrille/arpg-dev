# Unit test for focused blacksmith panel upgrade state.
# Run via: godot --headless --path client --script res://tests/test_blacksmith_panel.gd
extends SceneTree

const BlacksmithPanelScript := preload("res://scripts/blacksmith_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := BlacksmithPanelScript.new()
	root.add_child(panel)
	await process_frame

	var config := {
		"item_upgrade_cost_gold": 100,
		"item_upgrade_cost_growth_per_level": 50,
		"item_upgrade_max_level": 3,
		"item_upgrade_success_chance_percent": 25,
		"item_upgrade_pity_failure_threshold": 2,
		"item_upgrade_resource_item_def_id": "upgrade_shard",
		"item_upgrade_resource_count": 1,
	}
	var item := {
		"item_instance_id": "2001",
		"item_def_id": "cave_bow",
		"item_template_id": "cave_bow",
		"display_name": "Common Cave Bow",
		"rarity": "common",
		"rolled_stats": {"damage_min": 1, "damage_max": 2, "upgrade_pity": {"failures": 1}},
	}
	panel.show_blacksmith("smith-1", [item], 100, 0, config, "Choose", {"upgrade_shard": 1})
	panel.stage_inventory_item(item)
	var state := panel.get_debug_state()
	_assert_true("recipe selector visible", bool(state.get("recipe_selector_visible", false)))
	_assert_false("empty upgrade history hidden", bool((state.get("upgrade_history", {}) as Dictionary).get("visible", true)))
	_assert_eq("selected recipe id", str(state.get("selected_recipe_id", "")), "item_upgrade")
	_assert_eq("selected recipe label", str(state.get("selected_recipe_label", "")), "Upgrade Item")
	_assert_eq("recipe option count", (state.get("recipe_options", []) as Array).size(), 3)
	_assert_eq("pity threshold", int(state.get("pity_threshold", 0)), 2)
	_assert_eq("pity failure count", int(state.get("pity_failure_count", 0)), 1)
	_assert_false("pity not guaranteed before threshold", bool(state.get("pity_guaranteed", true)))
	_assert_true("preview shows selected recipe", _array_contains_text(state.get("preview_lines", []), "Recipe: Upgrade Item"))
	_assert_true("pity preview shows progress", _array_contains_text(state.get("preview_lines", []), "Pity: 1/2 failures"))
	_assert_true("preview shows success result", _array_contains_text(state.get("preview_lines", []), "On success: Level 0 -> 1"))
	_assert_true("preview shows failure result", _array_contains_text(state.get("preview_lines", []), "On failure: item unchanged; pity 1 -> 2 failures"))
	_assert_true("preview shows attempt spend", _array_contains_text(state.get("preview_lines", []), "Spend on attempt: 100 gold, 1 Upgrade Shard"))
	_assert_true("preview shows after attempt balance", _array_contains_text(state.get("preview_lines", []), "After attempt: 0 gold, 0 Upgrade Shard"))
	panel.select_recipe("weapon_honing")
	state = panel.get_debug_state()
	_assert_eq("selected weapon recipe id", str(state.get("selected_recipe_id", "")), "weapon_honing")
	_assert_eq("selected weapon recipe label", str(state.get("selected_recipe_label", "")), "Hone Weapon")
	_assert_true("weapon recipe eligibility", _array_contains_text(state.get("preview_lines", []), "Eligible: Weapons only"))
	var rows: Array = state.get("rows", [])
	_assert_true("weapon recipe keeps bow enabled", bool((rows[0] as Dictionary).get("upgrade_enabled", false)))
	var shield := item.duplicate(true)
	shield["item_def_id"] = "cave_shield"
	shield["item_template_id"] = "cave_shield"
	shield["rolled_stats"] = {"armor": 2, "block_percent": 5}
	panel.stage_inventory_item(shield)
	state = panel.get_debug_state()
	rows = state.get("rows", [])
	_assert_false("weapon recipe disables shield", bool((rows[0] as Dictionary).get("upgrade_enabled", true)))
	_assert_true("weapon recipe rejects shield preview", _array_contains_text(state.get("preview_lines", []), "Recipe cannot modify this item"))
	panel.select_recipe("armor_reinforcement")
	state = panel.get_debug_state()
	rows = state.get("rows", [])
	_assert_eq("selected armor recipe id", str(state.get("selected_recipe_id", "")), "armor_reinforcement")
	_assert_eq("selected armor recipe label", str(state.get("selected_recipe_label", "")), "Reinforce Armor")
	_assert_true("armor recipe eligibility", _array_contains_text(state.get("preview_lines", []), "Eligible: Armor pieces only"))
	_assert_true("armor recipe enables shield", bool((rows[0] as Dictionary).get("upgrade_enabled", false)))
	panel.stage_inventory_item(item)
	state = panel.get_debug_state()
	rows = state.get("rows", [])
	_assert_false("armor recipe disables bow", bool((rows[0] as Dictionary).get("upgrade_enabled", true)))
	_assert_true("armor recipe rejects bow preview", _array_contains_text(state.get("preview_lines", []), "Recipe cannot modify this item"))
	panel.select_recipe("item_upgrade")

	item["rolled_stats"] = {"item_level": 0, "damage_min": 1, "damage_max": 2, "upgrade_pity": {"failures": 2}}
	panel.stage_inventory_item(item)
	state = panel.get_debug_state()
	_assert_eq("guaranteed failure count", int(state.get("pity_failure_count", 0)), 2)
	_assert_true("pity guaranteed at threshold", bool(state.get("pity_guaranteed", false)))
	_assert_true("pity preview shows guarantee", _array_contains_text(state.get("preview_lines", []), "Next upgrade guaranteed"))
	_assert_false("guaranteed preview hides failure result", _array_contains_text(state.get("preview_lines", []), "On failure"))
	panel.select_recipe("weapon_honing")
	panel.update_after_upgrade(item, 0, 0, 100, true, {"upgrade_shard": 0})
	state = panel.get_debug_state()
	var history: Dictionary = state.get("upgrade_history", {})
	_assert_true("upgrade history visible", bool(history.get("visible", false)))
	_assert_eq("upgrade history first count", int(history.get("row_count", 0)), 1)
	_assert_true("upgrade history recipe", str(history.get("text", "")).contains("Hone Weapon: Success"))
	for i in range(5):
		panel.update_after_upgrade(item, 0, 0, 101 + i, false, {"upgrade_shard": 0})
	history = panel.get_debug_state().get("upgrade_history", {})
	_assert_eq("upgrade history capped", int(history.get("row_count", 0)), int(history.get("max_entries", 0)))
	var history_rows: Array = history.get("rows", [])
	_assert_true("upgrade history newest first", str((history_rows[0] as Dictionary).get("text", "")).contains("105 gold"))
	_assert_false("upgrade history drops oldest", str(history.get("text", "")).contains("100 gold"))

	panel.queue_free()
	print("[gdtest] PASS: test_blacksmith_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _array_contains_text(rows: Array, needle: String) -> bool:
	for row in rows:
		if str(row).contains(needle):
			return true
	return false
