# Unit test for the town vendor shop panel.
# Run via: godot --headless --path client --script res://tests/test_shop_panel.gd
extends SceneTree

const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const StatLabels := preload("res://scripts/stat_labels.gd")

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
		{"offer_id": "generated:depth3:000", "kind": "generated", "item_template_id": "cave_bow", "item_def_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "rolled_stats": {"damage_min": 2, "damage_max": 2}, "buy_price": 50, "summary_lines": ["Slot: Main hand", "Damage 2-2", "Requires level 1"], "requirements": {"level": 2, "str": 15}, "requirement_status": [{"stat": "level", "required": 2, "current": 2, "met": true}, {"stat": "str", "required": 15, "current": 12, "met": false}], "equip_preview": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "current": 4, "preview": 6, "delta": 2}]}, "comparison": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "offered": 2, "equipped": 4, "delta": -2}]}},
		{"offer_id": "generated:depth3:001", "kind": "generated", "item_template_id": "cave_gloves", "item_def_id": "cave_gloves", "display_name": "Magic Cave Gloves", "rarity": "magic", "slot": "gloves", "category": "equipment", "rolled_stats": {"armor": 3}, "buy_price": 90, "summary_lines": ["Slot: gloves", "Armor +3", "Requires level 1"], "comparison": {"slot": "gloves", "deltas": [{"stat": "armor", "offered": 3, "equipped": 0, "delta": 3}]}},
	]
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common"},
		{"item_instance_id": "2002", "item_def_id": "red_potion"},
	]
	var sell_appraisals := [
		{"item_instance_id": "2001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "sell_price": 27, "summary_lines": ["Slot: Main hand", "Damage 2-2", "Requires level 1"], "requirements": {"level": 2, "str": 15}, "requirement_status": [{"stat": "level", "required": 2, "current": 2, "met": true}, {"stat": "str", "required": 15, "current": 12, "met": false}], "equip_preview": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "current": 4, "preview": 6, "delta": 2}]}, "comparison": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "offered": 2, "equipped": 4, "delta": -2}]}},
		{"item_instance_id": "2002", "item_def_id": "red_potion", "display_name": "Red Potion", "category": "consumable", "sell_price": 5, "summary_lines": ["Kind: consumable", "Restores 5 HP"]},
	]
	panel.show_shop("1004", "town_vendor", offers, 60, inventory, {}, "Town Vendor", sell_appraisals)
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("offer count", int(state.get("offer_count", 0)), 4)
	_assert_eq("fixed count", int(state.get("fixed_offer_count", 0)), 2)
	_assert_eq("generated count", int(state.get("generated_offer_count", 0)), 2)
	_assert_eq("buyback count", int(state.get("buyback_offer_count", 0)), 0)
	_assert_true("red potion buy enabled", bool(state.get("buy_buttons", {}).get("fixed:red_potion", {}).get("enabled", false)))
	_assert_false("expensive generated disabled", bool(state.get("buy_buttons", {}).get("generated:depth3:001", {}).get("enabled", true)))
	_assert_eq("sell rows", int(state.get("sell_row_count", 0)), 2)
	_assert_true("offer rows include summaries", _rows_have_summary(state.get("offer_rows", [])))
	_assert_eq("source depth debug defaults", int((state.get("offer_rows", [])[2] as Dictionary).get("source_depth", 0)), 0)
	var sell_rows: Array = state.get("sell_rows", [])
	_assert_true("sell rows include price", not sell_rows.is_empty() and int((sell_rows[0] as Dictionary).get("sell_price", 0)) > 0)
	_assert_true("comparisons rendered", int(state.get("comparison_row_count", 0)) >= 2)
	_assert_true("requirements rendered", int(state.get("requirement_row_count", 0)) >= 4)
	_assert_true("equip previews rendered", int(state.get("equip_preview_row_count", 0)) >= 2)
	_assert_eq("vendor grid columns", int(state.get("vendor_columns", 0)), 5)
	_assert_eq("vendor grid rows", int(state.get("vendor_rows", 0)), 10)
	_assert_eq("vendor slot count", int(state.get("vendor_slot_count", 0)), 50)
	_assert_eq("occupied vendor slots", int(state.get("occupied_vendor_slot_count", 0)), 4)
	_assert_false("header gold hidden", bool(state.get("header_gold_visible", true)))
	var tooltip_lines: Array = panel._tooltip_lines(offers[2])
	_assert_false("tooltip stats exclude requirements", _array_contains_text(tooltip_lines, "Requires"))
	_assert_false("tooltip stats exclude comparison", _array_contains_text(tooltip_lines, "vs equipped"))
	_assert_true("tooltip requirements extracted", _array_contains_text(panel._requirement_lines(offers[2]), "Level 2"))
	_assert_true("tooltip stat requirements extracted", _array_contains_text(panel._requirement_lines(offers[2]), "%s 15(-3)" % StatLabels.display_name("str")))
	var comparison_entries: Array = panel._comparison_entries(offers[2])
	_assert_true("tooltip preview extracted", _array_contains_text(comparison_entries, "preview"))
	_assert_true("tooltip comparison extracted", _array_contains_text(comparison_entries, "vs equipped"))
	var offer_tooltip := panel._make_offer_tooltip(offers[2])
	_assert_eq("vendor tooltip uses shared panel", offer_tooltip.get_script(), ItemTooltipPanelScript)
	offer_tooltip.queue_free()

	panel.bot_click_buy_offer("fixed:red_potion")
	_assert_eq("buy emitted count", emitted.size(), 1)
	_assert_eq("buy emitted type", str(emitted[0]["type"]), "shop_buy_intent")
	_assert_eq("buy emitted entity", str(emitted[0]["payload"].get("shop_entity_id", "")), "1004")
	_assert_eq("buy emitted offer", str(emitted[0]["payload"].get("offer_id", "")), "fixed:red_potion")

	panel.bot_click_sell_item("", true, 0)
	_assert_eq("sell emitted count", emitted.size(), 2)
	_assert_eq("sell emitted type", str(emitted[1]["type"]), "shop_sell_intent")
	_assert_eq("sell emitted item", str(emitted[1]["payload"].get("item_instance_id", "")), "2001")

	panel._handle_inventory_drop({"source": "bag", "item": inventory[0]})
	_assert_eq("drop sell emitted count", emitted.size(), 3)
	_assert_eq("drop sell emitted type", str(emitted[2]["type"]), "shop_sell_intent")
	_assert_eq("drop sell emitted item", str(emitted[2]["payload"].get("item_instance_id", "")), "2001")

	var refreshed_offers := [
		offers[0],
		offers[1],
		{"offer_id": "generated:depth3:001", "kind": "generated", "item_template_id": "cave_gloves", "item_def_id": "cave_gloves", "display_name": "Magic Cave Gloves", "rarity": "magic", "slot": "gloves", "category": "equipment", "rolled_stats": {"armor": 3}, "buy_price": 90, "source_depth": 3, "summary_lines": ["Slot: gloves", "Armor +3", "Requires level 1"]},
		{"offer_id": "buyback:2001", "kind": "buyback", "item_template_id": "cave_bow", "item_def_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "buy_price": 27, "summary_lines": ["Slot: Main hand", "Damage 2-2", "Requires level 1"]},
	]
	panel.apply_shop_refresh(refreshed_offers, [sell_appraisals[1]])
	state = panel.get_debug_state()
	_assert_eq("refreshed offer count", int(state.get("offer_count", 0)), 4)
	_assert_eq("refreshed fixed count", int(state.get("fixed_offer_count", 0)), 2)
	_assert_eq("refreshed generated count", int(state.get("generated_offer_count", 0)), 1)
	_assert_eq("refreshed buyback count", int(state.get("buyback_offer_count", 0)), 1)
	_assert_true("generated removal applied", not _rows_contain_offer_id(state.get("offer_rows", []), "generated:depth3:000"))
	_assert_true("buyback row applied", _rows_contain_offer_id(state.get("offer_rows", []), "buyback:2001"))
	_assert_eq("sell appraisals refreshed", int(state.get("sell_row_count", 0)), 1)
	_assert_eq("source depth debug refreshed", int(_row_for_offer(state.get("offer_rows", []), "generated:depth3:001").get("source_depth", 0)), 3)

	var mystery_offer := {
		"offer_id": "mystery:wp:-3:ring:000",
		"kind": "mystery",
		"concealed": true,
		"mystery_label": "Unidentified ring",
		"slot": "ring",
		"category": "equipment",
		"buy_price": 75,
		"source_depth_min": 1,
		"source_depth_max": 3,
	}
	panel.show_shop("1005", "town_mystery_seller", [mystery_offer], 100, inventory, {}, "Mystery Seller", [])
	state = panel.get_debug_state()
	_assert_eq("mystery offer count", int(state.get("mystery_offer_count", 0)), 1)
	_assert_eq("mystery fixed count", int(state.get("fixed_offer_count", 0)), 0)
	_assert_true("mystery buy enabled", bool(state.get("buy_buttons", {}).get("mystery:wp:-3:ring:000", {}).get("enabled", false)))
	var mystery_row := _row_for_offer(state.get("offer_rows", []), "mystery:wp:-3:ring:000")
	_assert_true("mystery row concealed", bool(mystery_row.get("concealed", false)))
	_assert_eq("mystery row label", str(mystery_row.get("mystery_label", "")), "Unidentified ring")
	_assert_eq("mystery identity hidden", int(mystery_row.get("identity_field_count", -1)), 0)
	_assert_eq("mystery item def hidden", str(mystery_row.get("item_def_id", "")), "")
	_assert_eq("mystery rarity hidden", str(mystery_row.get("rarity", "")), "")
	_assert_eq("mystery source min", int(mystery_row.get("source_depth_min", 0)), 1)
	_assert_eq("mystery source max", int(mystery_row.get("source_depth_max", 0)), 3)
	_assert_true("mystery summary source window", _array_contains_text(mystery_row.get("summary_lines", []), "Source depths: 1-3"))
	_assert_eq("mystery no comparison", int(mystery_row.get("comparison_count", -1)), 0)
	_assert_eq("mystery no requirements", int(mystery_row.get("requirement_count", -1)), 0)
	_assert_eq("mystery no equip preview", int(mystery_row.get("equip_preview_count", -1)), 0)
	var mystery_tooltip_lines: Array = panel._tooltip_lines(mystery_offer)
	_assert_true("mystery tooltip label", _array_contains_text(mystery_tooltip_lines, "Unidentified ring"))
	_assert_false("mystery tooltip hides rarity", _array_contains_text(mystery_tooltip_lines, "Rarity:"))
	_assert_false("mystery tooltip hides requirements", _array_contains_text(mystery_tooltip_lines, "Requires"))
	_assert_false("mystery tooltip hides comparison", _array_contains_text(mystery_tooltip_lines, "vs equipped"))
	var mystery_tooltip := panel._make_offer_tooltip(mystery_offer)
	_assert_eq("mystery tooltip uses shared panel", mystery_tooltip.get_script(), ItemTooltipPanelScript)
	mystery_tooltip.queue_free()
	panel.bot_click_buy_offer("", "mystery", 0)
	_assert_eq("mystery buy emitted count", emitted.size(), 4)
	_assert_eq("mystery buy emitted entity", str(emitted[3]["payload"].get("shop_entity_id", "")), "1005")
	_assert_eq("mystery buy emitted offer", str(emitted[3]["payload"].get("offer_id", "")), "mystery:wp:-3:ring:000")

	var inventory_panel := InventoryPanelScript.new()
	root.add_child(inventory_panel)
	await process_frame
	var inventory_emitted: Array = []
	inventory_panel.intent_requested.connect(func(intent_type: String, payload: Dictionary) -> void:
		inventory_emitted.append({"type": intent_type, "payload": payload.duplicate(true)})
	)
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60)
	_assert_true("inventory red potion tooltip effect", _array_contains_text(inventory_panel._tooltip_lines(inventory[1]), "Restores 5 HP"))
	_assert_true("inventory blue potion tooltip effect", _array_contains_text(inventory_panel._tooltip_lines({"item_def_id": "blue_potion"}), "Restores 5 mana"))
	inventory_panel.set_inventory_state([sell_appraisals[0]], {}, 3, 15, 60)
	var inventory_state := inventory_panel.get_debug_state()
	_assert_true("inventory requirements rendered", int(inventory_state.get("requirement_row_count", 0)) >= 2)
	_assert_true("inventory equip previews rendered", int(inventory_state.get("equip_preview_row_count", 0)) >= 1)
	var inventory_tooltip := inventory_panel._make_item_tooltip(sell_appraisals[0])
	_assert_eq("inventory tooltip uses shared panel", inventory_tooltip.get_script(), ItemTooltipPanelScript)
	_assert_false("inventory tooltip stats exclude requirements", _array_contains_text(inventory_panel._tooltip_lines(sell_appraisals[0]), "Requires"))
	_assert_true("inventory tooltip requirements extracted", _array_contains_text(inventory_panel._requirement_lines(sell_appraisals[0]), "Level 2"))
	_assert_true("inventory tooltip stat requirements extracted", _array_contains_text(inventory_panel._requirement_lines(sell_appraisals[0]), "%s 15(-3)" % StatLabels.display_name("str")))
	_assert_true("inventory tooltip preview extracted", _array_contains_text(inventory_panel._comparison_entries(sell_appraisals[0]), "preview"))
	_assert_true("inventory tooltip comparison extracted", _array_contains_text(inventory_panel._comparison_entries(sell_appraisals[0]), "vs equipped"))
	inventory_tooltip.queue_free()
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60)
	inventory_panel._handle_drop_on_slot("bag", {
		"source": "shop_offer",
		"shop_entity_id": "1004",
		"offer_id": "fixed:red_potion",
		"item": offers[0],
	})
	_assert_eq("inventory drop buy emitted count", inventory_emitted.size(), 1)
	_assert_eq("inventory drop buy emitted type", str(inventory_emitted[0]["type"]), "shop_buy_intent")
	_assert_eq("inventory drop buy emitted offer", str(inventory_emitted[0]["payload"].get("offer_id", "")), "fixed:red_potion")
	inventory_panel.set_shop_sell_context("1004")
	inventory_panel._handle_double_click(inventory[0])
	_assert_eq("inventory double click sell emitted count", inventory_emitted.size(), 2)
	_assert_eq("inventory double click sell emitted type", str(inventory_emitted[1]["type"]), "shop_sell_intent")
	_assert_eq("inventory double click sell item", str(inventory_emitted[1]["payload"].get("item_instance_id", "")), "2001")
	inventory_panel.queue_free()

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


func _rows_contain_offer_id(rows: Variant, offer_id: String) -> bool:
	return not _row_for_offer(rows, offer_id).is_empty()


func _row_for_offer(rows: Variant, offer_id: String) -> Dictionary:
	if typeof(rows) != TYPE_ARRAY:
		return {}
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("offer_id", "")) == offer_id:
			return row as Dictionary
	return {}


func _array_contains_text(rows: Array, needle: String) -> bool:
	for row in rows:
		if str(row).contains(needle):
			return true
	return false
