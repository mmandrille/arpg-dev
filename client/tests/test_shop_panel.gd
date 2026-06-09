# Unit test for the town vendor shop panel.
# Run via: godot --headless --path client --script res://tests/test_shop_panel.gd
extends SceneTree

const ShopPanelScript := preload("res://scripts/shop_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := ShopPanelScript.new()
	root.add_child(panel)
	await process_frame

	var emitted: Array = []
	panel.intent_requested.connect(func(intent_type: String, payload: Dictionary) -> void:
		emitted.append({"type": intent_type, "payload": payload.duplicate(true)})
	)

	var offers := [
		{"offer_id": "fixed:red_potion", "kind": "fixed", "item_def_id": "red_potion", "display_name": "Red Potion", "category": "consumable", "buy_price": 20, "summary_lines": ["Kind: consumable", "Restores 5 HP"]},
		{"offer_id": "fixed:blue_potion", "kind": "fixed", "item_def_id": "blue_potion", "display_name": "Blue Potion", "category": "consumable", "buy_price": 20, "summary_lines": ["Kind: consumable", "Restores 5 mana"]},
		{"offer_id": "generated:depth3:000", "kind": "generated", "item_template_id": "cave_bow", "item_def_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "rolled_stats": {"damage_min": 2, "damage_max": 5}, "buy_price": 50, "summary_lines": ["Slot: Main hand", "Damage 2-5", "Requires level 1"], "comparison": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "offered": 5, "equipped": 4, "delta": 1}]}},
		{"offer_id": "generated:depth3:001", "kind": "generated", "item_template_id": "cave_gloves", "item_def_id": "cave_gloves", "display_name": "Magic Cave Gloves", "rarity": "magic", "slot": "gloves", "category": "equipment", "rolled_stats": {"armor": 3}, "buy_price": 90, "summary_lines": ["Slot: gloves", "Armor +3", "Requires level 1"], "comparison": {"slot": "gloves", "deltas": [{"stat": "armor", "offered": 3, "equipped": 0, "delta": 3}]}},
	]
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common"},
		{"item_instance_id": "2002", "item_def_id": "red_potion"},
	]
	var sell_appraisals := [
		{"item_instance_id": "2001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "sell_price": 27, "summary_lines": ["Slot: Main hand", "Damage 2-5", "Requires level 1"], "comparison": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "offered": 5, "equipped": 4, "delta": 1}]}},
		{"item_instance_id": "2002", "item_def_id": "red_potion", "display_name": "Red Potion", "category": "consumable", "sell_price": 5, "summary_lines": ["Kind: consumable", "Restores 5 HP"]},
	]
	panel.show_shop("1004", "town_vendor", offers, 60, inventory, {}, "Town Vendor", sell_appraisals)
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("offer count", int(state.get("offer_count", 0)), 4)
	_assert_eq("fixed count", int(state.get("fixed_offer_count", 0)), 2)
	_assert_eq("generated count", int(state.get("generated_offer_count", 0)), 2)
	_assert_true("red potion buy enabled", bool(state.get("buy_buttons", {}).get("fixed:red_potion", {}).get("enabled", false)))
	_assert_false("expensive generated disabled", bool(state.get("buy_buttons", {}).get("generated:depth3:001", {}).get("enabled", true)))
	_assert_eq("sell rows", int(state.get("sell_row_count", 0)), 2)
	_assert_true("offer rows include summaries", _rows_have_summary(state.get("offer_rows", [])))
	var sell_rows: Array = state.get("sell_rows", [])
	_assert_true("sell rows include price", not sell_rows.is_empty() and int((sell_rows[0] as Dictionary).get("sell_price", 0)) > 0)
	_assert_true("comparisons rendered", int(state.get("comparison_row_count", 0)) >= 2)

	panel.bot_click_buy_offer("fixed:red_potion")
	_assert_eq("buy emitted count", emitted.size(), 1)
	_assert_eq("buy emitted type", str(emitted[0]["type"]), "shop_buy_intent")
	_assert_eq("buy emitted entity", str(emitted[0]["payload"].get("shop_entity_id", "")), "1004")
	_assert_eq("buy emitted offer", str(emitted[0]["payload"].get("offer_id", "")), "fixed:red_potion")

	panel.bot_click_sell_item("", true, 0)
	_assert_eq("sell emitted count", emitted.size(), 2)
	_assert_eq("sell emitted type", str(emitted[1]["type"]), "shop_sell_intent")
	_assert_eq("sell emitted item", str(emitted[1]["payload"].get("item_instance_id", "")), "2001")

	panel.show_status("insufficient gold", true)
	_assert_eq("status text", str(panel.get_debug_state().get("status", "")), "insufficient gold")

	panel.queue_free()
	print("[gdtest] PASS: test_shop_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
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


func _rows_have_summary(rows: Variant) -> bool:
	if typeof(rows) != TYPE_ARRAY:
		return false
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			return false
		if (row as Dictionary).get("summary_lines", []).is_empty():
			return false
	return true
