# v391 client blacksmith panel — upgrade + renew recipes

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
		"deepest_dungeon_depth": 30,
		"item_level_levels_per_tier": 10,
	}
	var item := {
		"item_instance_id": "2001",
		"item_def_id": "cave_bow",
		"item_template_id": "cave_bow",
		"display_name": "Common Cave Bow",
		"rarity": "common",
		"rolled_stats": {"damage_min": 1, "damage_max": 2, "upgrade_pity": {"failures": 1}},
	}
	var shard := {
		"item_instance_id": "shard1",
		"item_def_id": "upgrade_shard",
		"rolled_stats": {"item_level": 2},
	}
	var stone := {
		"item_instance_id": "stone1",
		"item_def_id": "renew_stone",
		"rolled_stats": {"item_level": 1},
	}
	panel.show_blacksmith("smith-1", [item, shard, stone], 100, 0, config, "Choose", {})
	panel.stage_inventory_item(item)
	var state := panel.get_debug_state()
	_assert_eq("recipe option count", (state.get("recipe_options", []) as Array).size(), 2)
	_assert_eq("selected recipe id", str(state.get("selected_recipe_id", "")), "item_upgrade")
	panel.select_recipe("item_renew")
	state = panel.get_debug_state()
	_assert_eq("selected renew recipe id", str(state.get("selected_recipe_id", "")), "item_renew")
	_assert_eq("selected renew recipe label", str(state.get("selected_recipe_label", "")), "Renew Item")
	_assert_true("renew recipe eligibility", _array_contains_text(state.get("preview_lines", []), "Eligible: Equipment (reroll affixes)"))
	var rows: Array = state.get("rows", [])
	_assert_true("renew enables bow with stone", bool((rows[0] as Dictionary).get("upgrade_enabled", false)))
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


func _array_contains_text(rows: Array, needle: String) -> bool:
	for row in rows:
		if str(row).contains(needle):
			return true
	return false
