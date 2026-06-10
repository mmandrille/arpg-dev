# Unit test for the account stash panel.
# Run via: godot --headless --path client --script res://tests/test_stash_panel.gd
extends SceneTree

const StashPanelScript := preload("res://scripts/stash_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := StashPanelScript.new()
	root.add_child(panel)
	await process_frame

	var emitted: Array = []
	panel.intent_requested.connect(func(intent_type: String, payload: Dictionary) -> void:
		emitted.append({"type": intent_type, "payload": payload.duplicate(true)})
	)

	var stash_items := [
		{"stash_item_id": "9001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "rolled_stats": {"damage_min": 2, "damage_max": 5}, "summary_lines": ["Slot: Main hand", "Damage 2-5"]},
	]
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Magic Cave Blade", "rarity": "magic", "slot": "main_hand", "rolled_stats": {"damage_min": 3, "damage_max": 7}, "summary_lines": ["Slot: Main hand", "Damage 3-7"]},
		{"item_instance_id": "2002", "item_def_id": "red_potion"},
	]

	panel.show_stash("1005", "account_stash", stash_items, 3, 50, inventory, {}, 7, [], "Account Stash")
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("stash item count", int(state.get("stash_item_count", 0)), 1)
	_assert_eq("stash gold", int(state.get("stash_gold", 0)), 3)
	_assert_eq("gold", int(state.get("gold", 0)), 7)
	_assert_eq("capacity", int(state.get("stash_capacity", 0)), 50)
	_assert_false("stash panel does not expose embedded bag rows", state.has("bag_rows"))
	_assert_false("stash panel does not expose embedded deposit buttons", state.has("deposit_buttons"))
	_assert_true("withdraw enabled", bool(state.get("withdraw_buttons", {}).get("9001", {}).get("enabled", false)))
	_assert_true("gold deposit enabled", bool(state.get("deposit_gold_enabled", false)))
	_assert_true("gold withdraw enabled", bool(state.get("withdraw_gold_enabled", false)))
	_assert_true("stash rows include summary", _rows_have_summary(state.get("stash_rows", [])))
	_assert_true("stash panel opens on left side", panel._panel.offset_left <= 24.0)

	panel.bot_drag_bag_to_stash("", true, 0)
	_assert_eq("deposit emitted count", emitted.size(), 1)
	_assert_eq("deposit emitted type", str(emitted[0]["type"]), "stash_deposit_item_intent")
	_assert_eq("deposit emitted entity", str(emitted[0]["payload"].get("stash_entity_id", "")), "1005")
	_assert_eq("deposit emitted item", str(emitted[0]["payload"].get("item_instance_id", "")), "2001")

	panel.bot_drag_stash_to_bag("", "", true, 0)
	_assert_eq("withdraw emitted count", emitted.size(), 2)
	_assert_eq("withdraw emitted type", str(emitted[1]["type"]), "stash_withdraw_item_intent")
	_assert_eq("withdraw emitted item", str(emitted[1]["payload"].get("stash_item_id", "")), "9001")

	panel._handle_drop_on_stash({"source": "bag", "item": inventory[0]})
	_assert_eq("bag drag to stash emitted count", emitted.size(), 3)
	_assert_eq("bag drag to stash type", str(emitted[2]["type"]), "stash_deposit_item_intent")
	_assert_eq("bag drag to stash item", str(emitted[2]["payload"].get("item_instance_id", "")), "2001")

	var inventory_panel := InventoryPanelScript.new()
	root.add_child(inventory_panel)
	var inventory_emitted: Array = []
	inventory_panel.intent_requested.connect(func(intent_type: String, payload: Dictionary) -> void:
		inventory_emitted.append({"type": intent_type, "payload": payload.duplicate(true)})
	)
	inventory_panel._handle_drop_on_slot("bag", {
		"source": "stash",
		"stash_entity_id": "1005",
		"stash_item_id": "9001",
		"item": stash_items[0],
	})
	_assert_eq("stash drag to bag emitted count", inventory_emitted.size(), 1)
	_assert_eq("stash drag to bag type", str(inventory_emitted[0]["type"]), "stash_withdraw_item_intent")
	_assert_eq("stash drag to bag item", str(inventory_emitted[0]["payload"].get("stash_item_id", "")), "9001")
	inventory_panel.queue_free()

	panel.bot_click_deposit_gold(1)
	_assert_eq("deposit gold emitted count", emitted.size(), 4)
	_assert_eq("deposit gold type", str(emitted[3]["type"]), "stash_deposit_gold_intent")
	_assert_eq("deposit gold amount", int(emitted[3]["payload"].get("amount", 0)), 1)

	panel.bot_click_withdraw_gold(1)
	_assert_eq("withdraw gold emitted count", emitted.size(), 5)
	_assert_eq("withdraw gold type", str(emitted[4]["type"]), "stash_withdraw_gold_intent")
	_assert_eq("withdraw gold amount", int(emitted[4]["payload"].get("amount", 0)), 1)

	panel.set_stash_state([], 0, 50)
	panel.set_inventory_state([inventory[0]], {"main_hand": "2001"}, 0, [])
	state = panel.get_debug_state()
	_assert_false("empty withdraw gold disabled", bool(state.get("withdraw_gold_enabled", true)))
	_assert_false("empty deposit gold disabled", bool(state.get("deposit_gold_enabled", true)))

	panel.show_status("stored", false)
	_assert_eq("status text", str(panel.get_debug_state().get("status", "")), "stored")

	panel.queue_free()
	print("[gdtest] PASS: test_stash_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
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
