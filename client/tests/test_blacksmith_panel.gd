# v391 client blacksmith panel — upgrade + renew recipes

extends SceneTree

const BlacksmithPanelScript := preload("res://scripts/blacksmith_panel.gd")
const BlacksmithMergePanelScript := preload("res://scripts/blacksmith_merge_panel.gd")
const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const BlacksmithRecipesScript := preload("res://scripts/blacksmith_recipes.gd")
const BlacksmithPanelActionsScript := preload("res://scripts/blacksmith_panel_actions.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_rolled_stats_item_level_reads()
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
	_assert_eq("default recipe id", str(state.get("selected_recipe_id", "")), "item_upgrade")
	_assert_eq("preview hidden without resource", (state.get("preview_lines", []) as Array).size(), 0)
	panel.stage_resource_item(stone)
	state = panel.get_debug_state()
	_assert_eq("renew recipe from stone", str(state.get("selected_recipe_id", "")), "item_renew")
	_assert_eq("selected renew recipe label", str(state.get("selected_recipe_label", "")), "Renew Item")
	_assert_true("renew recipe eligibility", _array_contains_text(state.get("preview_lines", []), "Eligible: Equipment (reroll affixes)"))
	_assert_true("renew preview mentions reroll", _array_contains_text(state.get("preview_lines", []), "reroll random affixes"))
	_assert_true("renew preview uses renew stone", _array_contains_text(state.get("preview_lines", []), "Renew Stone"))
	_assert_false("renew preview avoids upgrade shard", _array_contains_text(state.get("preview_lines", []), "Upgrade Shard"))
	var rows: Array = state.get("rows", [])
	_assert_true("renew enables bow with stone", bool((rows[0] as Dictionary).get("upgrade_enabled", false)))
	panel.unstage_resource(false)
	panel.select_recipe("item_renew")
	state = panel.get_debug_state()
	_assert_eq("bot select renew stages stone", str(state.get("staged_resource_id", "")), "stone1")
	panel.bot_select_tab("Merge")
	var merge_view: BlacksmithMergePanel = panel._merge_view
	_assert_true("merge accepts shard drop", merge_view.can_place_item_at(0, shard))
	_assert_true("merge place shard", merge_view.place_item_at(0, shard))
	_assert_false("merge rejects mismatched stone", merge_view.can_place_item_at(1, stone))
	var shard_two := shard.duplicate(true)
	shard_two["item_instance_id"] = "shard2"
	_assert_true("merge accepts matching shard", merge_view.can_place_item_at(1, shard_two))
	panel.queue_free()
	await _test_upgrade_preview_flow()
	print("[gdtest] PASS: test_blacksmith_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_upgrade_preview_flow() -> void:
	var panel := BlacksmithPanelScript.new()
	root.add_child(panel)
	await process_frame
	var emitted: Array = []
	panel.upgrade_inventory_requested.connect(func(item_instance_id: String, _resource_instance_id: String = "") -> void:
		emitted.append(item_instance_id)
	)
	var upgrade_item := {
		"item_instance_id": "2001",
		"item_def_id": "cave_bow",
		"item_template_id": "cave_bow",
		"display_name": "Common Cave Bow",
		"rarity": "common",
		"rolled_stats": {"damage_min": 2, "damage_max": 4},
		"summary_lines": ["Min damage: +2", "Max damage: +4"],
	}
	var shard := {
		"item_instance_id": "shard1",
		"item_def_id": "upgrade_shard",
		"rolled_stats": {"item_level": 2},
	}
	var config := {
		"item_upgrade_cost_gold": 100,
		"item_upgrade_cost_growth_per_level": 50,
		"item_upgrade_max_level": 3,
		"item_upgrade_resource_item_def_id": "upgrade_shard",
		"item_upgrade_resource_count": 1,
		"deepest_dungeon_depth": 30,
		"item_level_levels_per_tier": 10,
	}
	panel.show_blacksmith("smith-1", [upgrade_item, shard], 40, 60, config, "Choose", {})
	panel.stage_inventory_item(upgrade_item)
	panel.stage_resource_item(shard)
	var state := panel.get_debug_state()
	_assert_eq("blacksmith staged item id", str(state.get("staged_item_id", "")), "2001")
	_assert_eq("blacksmith wallet gold", int(state.get("wallet_gold", 0)), 100)
	_assert_eq("blacksmith resource item", str(state.get("resource_item_def_id", "")), "upgrade_shard")
	_assert_eq("blacksmith required shard count", int(state.get("resource_required_count", 0)), 1)
	_assert_eq("blacksmith wallet shard count", int(state.get("resource_wallet_count", 0)), 1)
	var window_size: Dictionary = (state.get("window", {}) as Dictionary).get("minimum_size", {})
	_assert_eq("blacksmith compact window width", int(window_size.get("x", 0)), 320)
	_assert_eq("blacksmith compact window height", int(window_size.get("y", 0)), 294)
	var stage_size: Dictionary = state.get("stage_slot_size", {})
	_assert_eq("blacksmith stage slot width", int(stage_size.get("x", 0)), 84)
	_assert_eq("blacksmith stage slot height", int(stage_size.get("y", 0)), 84)
	_assert_eq("blacksmith stage slot centered", bool(state.get("stage_slot_centered", false)), true)
	_assert_eq("blacksmith stage icon visible", bool(state.get("stage_icon_visible", false)), true)
	_assert_true("blacksmith direct preview includes tier rescale", _array_contains_text(state.get("preview_lines", []), "Stats rescale to the next item tier"))
	_assert_false("blacksmith instruction removed", bool(state.get("instruction_visible", true)))
	var summary_shield: Dictionary = upgrade_item.duplicate(true)
	summary_shield["rolled_stats"] = {"item_level": 0}
	summary_shield["summary_lines"] = ["Armor: +2", "Block: +12%"]
	panel.stage_inventory_item(summary_shield)
	state = panel.get_debug_state()
	_assert_true("blacksmith summary preview includes tier rescale", _array_contains_text(state.get("preview_lines", []), "Stats rescale to the next item tier"))
	var common_bow: Dictionary = upgrade_item.duplicate(true)
	common_bow["item_def_id"] = "cave_bow"
	common_bow["display_name"] = "Common Cave Bow"
	common_bow["rarity"] = "common"
	common_bow["rolled_stats"] = {"item_level": 0}
	common_bow["summary_lines"] = []
	panel.stage_inventory_item(common_bow)
	state = panel.get_debug_state()
	_assert_true("blacksmith template preview includes tier rescale", _array_contains_text(state.get("preview_lines", []), "Stats rescale to the next item tier"))
	panel.unstage_item()
	state = panel.get_debug_state()
	_assert_eq("blacksmith unstage clears item", str(state.get("staged_item_id", "")), "")
	panel.stage_inventory_item(upgrade_item)
	panel.stage_resource_item(shard)
	panel._emit_upgrade(upgrade_item)
	_assert_eq("blacksmith inventory upgrade emitted", str(emitted[0]), "2001")
	panel.unstage_resource(false)
	panel.show_blacksmith("smith-1", [upgrade_item], 100, 0, config, "Choose", {})
	panel.stage_inventory_item(upgrade_item)
	state = panel.get_debug_state()
	_assert_eq("blacksmith missing wallet shard count", int(state.get("resource_wallet_count", -1)), 0)
	var missing_rows: Array = state.get("rows", [])
	_assert_false("blacksmith disables upgrade without shard", bool((missing_rows[0] as Dictionary).get("upgrade_enabled", true)))
	panel._emit_upgrade(upgrade_item)
	_assert_eq("blacksmith missing shard does not emit upgrade", emitted.size(), 1)
	_assert_true("blacksmith missing shard status", str(panel.get_debug_state().get("status", "")).contains("Need 1 Upgrade Shard"))
	panel.hide_display()
	state = panel.get_debug_state()
	_assert_eq("blacksmith close returns staged item", str(state.get("staged_item_id", "")), "")
	panel.queue_free()


func _test_rolled_stats_item_level_reads() -> void:
	var nested := {"rolled_stats": {"stats": {"dex": 5, "item_level": 2}, "requirements": {"level": 2}}}
	_assert_eq("nested rolled_stats item level", BlacksmithUpgradePreviewScript.item_level(nested), 2)
	var flattened := {"rolled_stats": {"dex": 5, "item_level": 1}}
	_assert_eq("flattened rolled_stats item level", BlacksmithUpgradePreviewScript.item_level(flattened), 1)
	_assert_eq("level one item needs level one shard", BlacksmithRecipesScript.required_resource_level(BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE, flattened), 1)
	_assert_eq("depth 11 allows item level 2", BlacksmithPanelActionsScript.max_item_level_for_deepest_depth({
		"deepest_dungeon_depth": 11,
		"item_level_levels_per_tier": 10,
	}), 2)
	_assert_eq("depth 10 caps at item level 1", BlacksmithPanelActionsScript.max_item_level_for_deepest_depth({
		"deepest_dungeon_depth": 10,
		"item_level_levels_per_tier": 10,
	}), 1)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


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
