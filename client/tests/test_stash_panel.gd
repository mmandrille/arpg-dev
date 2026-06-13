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
		{"stash_item_id": "9001", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Common Cave Bow", "rarity": "common", "slot": "main_hand", "rolled_stats": {"damage_min": 2, "damage_max": 2}, "summary_lines": ["Slot: Main hand", "Damage 2-2"]},
		{"stash_item_id": "9002", "item_def_id": "cave_ring", "item_template_id": "cave_ring", "display_name": "Rare Cave Ring", "rarity": "rare", "slot": "ring", "rolled_stats": {"max_hp": 4}, "summary_lines": ["Slot: Ring", "Maximum Health +4"]},
		{"stash_item_id": "9003", "item_def_id": "red_potion", "display_name": "Red Potion", "category": "consumable", "summary_lines": ["Kind: consumable", "Restores 5 HP"]},
	]
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Magic Cave Blade", "rarity": "magic", "slot": "main_hand", "rolled_stats": {"damage_min": 3, "damage_max": 7}, "summary_lines": ["Slot: Main hand", "Damage 3-7"]},
		{"item_instance_id": "2002", "item_def_id": "red_potion"},
	]

	panel.show_stash("1005", "account_stash", stash_items, 3, 50, inventory, {}, 7, [], "Account Stash")
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	var window: Dictionary = state.get("window", {})
	_assert_eq("stash window title", str(window.get("title", "")), "Account Stash")
	_assert_true("stash window has close button", bool(window.get("close_visible", false)))
	_assert_true("stash window is draggable", bool(window.get("draggable", false)))
	_assert_eq("stash item count", int(state.get("stash_item_count", 0)), 3)
	_assert_eq("filtered stash item count", int(state.get("filtered_stash_item_count", 0)), 3)
	_assert_eq("stash gold", int(state.get("stash_gold", 0)), 3)
	_assert_eq("gold", int(state.get("gold", 0)), 7)
	_assert_eq("capacity", int(state.get("stash_capacity", 0)), 50)
	_assert_false("stash panel does not expose embedded bag rows", state.has("bag_rows"))
	_assert_false("stash panel does not expose embedded deposit buttons", state.has("deposit_buttons"))
	_assert_true("withdraw enabled", bool(state.get("withdraw_buttons", {}).get("9001", {}).get("enabled", false)))
	_assert_true("gold deposit enabled", bool(state.get("deposit_gold_enabled", false)))
	_assert_true("gold withdraw enabled", bool(state.get("withdraw_gold_enabled", false)))
	_assert_false("gold amount input initially hidden", bool(state.get("gold_amount_visible", false)))
	_assert_true("stash rows include summary", _rows_have_summary(state.get("stash_rows", [])))
	_assert_true("stash panel opens on left side", float((window.get("position", {}) as Dictionary).get("x", 999.0)) <= 24.0)

	panel.bot_open_deposit_gold()
	state = panel.get_debug_state()
	_assert_true("deposit opens gold amount input", bool(state.get("gold_amount_visible", false)))
	_assert_eq("deposit amount mode", str(state.get("gold_amount_mode", "")), "deposit")
	_assert_eq("deposit amount defaults to all carried gold", str(state.get("gold_amount_text", "")), "7")
	_assert_true("deposit amount ok enabled", bool(state.get("gold_amount_ok_enabled", false)))
	panel.bot_set_gold_amount_text("abc")
	state = panel.get_debug_state()
	_assert_false("non-integer gold amount disables ok", bool(state.get("gold_amount_ok_enabled", true)))
	panel.bot_open_withdraw_gold()
	state = panel.get_debug_state()
	_assert_eq("withdraw amount mode", str(state.get("gold_amount_mode", "")), "withdraw")
	_assert_eq("withdraw amount defaults to all stash gold", str(state.get("gold_amount_text", "")), "3")
	_assert_true("withdraw amount ok enabled", bool(state.get("gold_amount_ok_enabled", false)))

	panel.bot_set_search_text(" ring ")
	state = panel.get_debug_state()
	_assert_eq("search text trimmed", str(state.get("stash_search_text", "")), "ring")
	_assert_eq("search filters ring", int(state.get("filtered_stash_item_count", 0)), 1)
	_assert_eq("filtered row is ring", str((state.get("stash_rows", [])[0] as Dictionary).get("item_def_id", "")), "cave_ring")

	panel.bot_select_sort_mode("rarity")
	state = panel.get_debug_state()
	_assert_eq("sort mode rarity", str(state.get("stash_sort_mode", "")), "rarity")
	_assert_eq("filtered sort preserves ring", str((state.get("stash_rows", [])[0] as Dictionary).get("stash_item_id", "")), "9002")

	panel.bot_set_search_text("")
	panel.bot_select_sort_mode("name")
	state = panel.get_debug_state()
	_assert_eq("search cleared count", int(state.get("filtered_stash_item_count", 0)), 3)
	_assert_eq("name sort first bow", str((state.get("stash_rows", [])[0] as Dictionary).get("item_def_id", "")), "cave_bow")

	panel.bot_drag_bag_to_stash("", true, 0)
	_assert_eq("deposit emitted count", emitted.size(), 1)
	_assert_eq("deposit emitted type", str(emitted[0]["type"]), "stash_deposit_item_intent")
	_assert_eq("deposit emitted entity", str(emitted[0]["payload"].get("stash_entity_id", "")), "1005")
	_assert_eq("deposit emitted item", str(emitted[0]["payload"].get("item_instance_id", "")), "2001")

	panel.bot_drag_stash_to_bag("", "", true, 0)
	_assert_eq("withdraw emitted count", emitted.size(), 2)
	_assert_eq("withdraw emitted type", str(emitted[1]["type"]), "stash_withdraw_item_intent")
	_assert_eq("withdraw emitted item", str(emitted[1]["payload"].get("stash_item_id", "")), "9001")

	panel.bot_set_search_text("ring")
	panel.bot_drag_stash_to_bag("", "", true, 0)
	_assert_eq("filtered withdraw emitted count", emitted.size(), 3)
	_assert_eq("filtered withdraw emitted item", str(emitted[2]["payload"].get("stash_item_id", "")), "9002")
	panel.bot_set_search_text("")

	panel._handle_drop_on_stash({"source": "bag", "item": inventory[0]})
	_assert_eq("bag drag to stash emitted count", emitted.size(), 4)
	_assert_eq("bag drag to stash type", str(emitted[3]["type"]), "stash_deposit_item_intent")
	_assert_eq("bag drag to stash item", str(emitted[3]["payload"].get("item_instance_id", "")), "2001")

	var inventory_panel := InventoryPanelScript.new()
	root.add_child(inventory_panel)
	await process_frame
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
	inventory_panel._handle_drop_on_slot("bag", {
		"source": "corpse",
		"corpse_entity_id": "corpse_entity_1",
		"item_instance_id": "dead_item_1",
		"item": {"item_instance_id": "dead_item_1", "item_def_id": "red_potion"},
	})
	_assert_eq("corpse drag to bag emitted count", inventory_emitted.size(), 2)
	_assert_eq("corpse drag to bag type", str(inventory_emitted[1]["type"]), "corpse_withdraw_item_intent")
	_assert_eq("corpse drag to bag entity", str(inventory_emitted[1]["payload"].get("corpse_entity_id", "")), "corpse_entity_1")
	_assert_eq("corpse drag to bag item", str(inventory_emitted[1]["payload"].get("item_instance_id", "")), "dead_item_1")
	inventory_panel._handle_drop_on_slot("bag", {
		"source": "unique_chest",
		"stash_entity_id": "unique_chest_entity_1",
		"stash_item_id": "unique_item_1",
		"item": {"stash_item_id": "unique_item_1", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Embercall Blade", "rarity": "unique"},
	})
	_assert_eq("unique chest drag to bag emitted count", inventory_emitted.size(), 3)
	_assert_eq("unique chest drag to bag type", str(inventory_emitted[2]["type"]), "unique_chest_take_item_intent")
	_assert_eq("unique chest drag to bag entity", str(inventory_emitted[2]["payload"].get("chest_entity_id", "")), "unique_chest_entity_1")
	_assert_eq("unique chest drag to bag item", str(inventory_emitted[2]["payload"].get("chest_item_id", "")), "unique_item_1")
	_assert_true("test equip slot kind recognized", inventory_panel._slot_kind_is_equipment("equip:main_hand"))
	_assert_true("test stash item can equip to main hand", inventory_panel._item_can_equip_to(stash_items[0], "main_hand"))
	_assert_false("non-rogue cannot equip main-hand weapon to off hand", inventory_panel._item_can_equip_to(inventory[0], "off_hand"))
	inventory_panel.set_inventory_state([
		{"item_instance_id": "2003", "item_def_id": "training_bow", "slot": "main_hand", "equipped": true, "rarity": "common"},
	], {"main_hand": "2003"})
	var paper_slots: Dictionary = inventory_panel.get_debug_state().get("paper_doll_slots", {})
	var off_hand_slot: Dictionary = paper_slots.get("off_hand", {})
	_assert_true("two-handed weapon greys off hand", bool(off_hand_slot.get("blocked_by_two_handed", false)))
	_assert_eq("two-handed off hand shows weapon icon", str(off_hand_slot.get("item_def_id", "")), "training_bow")
	inventory_panel.set_character_progression({"character_class": "rogue"})
	_assert_true("rogue can equip one-handed weapon to off hand", inventory_panel._item_can_equip_to(inventory[0], "off_hand"))
	inventory_panel._handle_drop_on_slot("equip:main_hand", {
		"source": "stash",
		"stash_entity_id": "1005",
		"stash_item_id": "9001",
		"item": stash_items[0],
	})
	_assert_eq("stash drag to equip emitted count", inventory_emitted.size(), 4)
	if inventory_emitted.size() >= 4:
		_assert_eq("stash drag to equip type", str(inventory_emitted[3]["type"]), "stash_equip_item_intent")
		_assert_eq("stash drag to equip stash item", str(inventory_emitted[3]["payload"].get("stash_item_id", "")), "9001")
		_assert_eq("stash drag to equip slot", str(inventory_emitted[3]["payload"].get("slot", "")), "main_hand")
	inventory_panel.queue_free()

	panel.bot_click_deposit_gold(7)
	_assert_eq("deposit gold emitted count", emitted.size(), 5)
	_assert_eq("deposit gold type", str(emitted[4]["type"]), "stash_deposit_gold_intent")
	_assert_eq("deposit gold amount", int(emitted[4]["payload"].get("amount", 0)), 7)

	panel.bot_click_withdraw_gold(3)
	_assert_eq("withdraw gold emitted count", emitted.size(), 6)
	_assert_eq("withdraw gold type", str(emitted[5]["type"]), "stash_withdraw_gold_intent")
	_assert_eq("withdraw gold amount", int(emitted[5]["payload"].get("amount", 0)), 3)

	panel.set_stash_state([], 0, 50)
	panel.set_inventory_state([inventory[0]], {"main_hand": "2001"}, 0, [])
	state = panel.get_debug_state()
	_assert_false("empty withdraw gold disabled", bool(state.get("withdraw_gold_enabled", true)))
	_assert_false("empty deposit gold disabled", bool(state.get("deposit_gold_enabled", true)))

	panel.show_status("stored", false)
	_assert_eq("status text", str(panel.get_debug_state().get("status", "")), "stored")
	panel.show_corpse(
		"corpse_entity_1",
		"p2",
		[
			{"item_instance_id": "dead_item_1", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Lost Cave Blade", "rarity": "common", "slot": "main_hand", "summary_lines": ["Slot: Main hand"]},
		],
		inventory,
		{},
		67,
		[]
	)
	state = panel.get_debug_state()
	_assert_eq("corpse panel mode", str(state.get("container_mode", "")), "corpse")
	_assert_eq("corpse entity id retained", str(state.get("stash_entity_id", "")), "corpse_entity_1")
	panel.bot_drag_stash_to_bag("dead_item_1")
	_assert_eq("corpse withdraw emitted count", emitted.size(), 7)
	_assert_eq("corpse withdraw emitted type", str(emitted[6]["type"]), "corpse_withdraw_item_intent")
	_assert_eq("corpse withdraw emitted entity", str(emitted[6]["payload"].get("corpse_entity_id", "")), "corpse_entity_1")
	_assert_eq("corpse withdraw emitted item", str(emitted[6]["payload"].get("item_instance_id", "")), "dead_item_1")
	panel._emit_withdraw({
		"item_instance_id": "dead_item_2",
		"stash_item_id": "dead_item_2",
		"item_def_id": "red_potion",
	})
	_assert_eq("corpse double-click emitted count", emitted.size(), 8)
	_assert_eq("corpse double-click emitted type", str(emitted[7]["type"]), "corpse_withdraw_item_intent")
	_assert_eq("corpse double-click emitted entity", str(emitted[7]["payload"].get("corpse_entity_id", "")), "corpse_entity_1")
	_assert_eq("corpse double-click emitted item", str(emitted[7]["payload"].get("item_instance_id", "")), "dead_item_2")
	panel.show_unique_chest(
		"unique_chest_entity_1",
		[
			{"stash_item_id": "unique_item_1", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Embercall Blade", "rarity": "unique", "slot": "main_hand", "summary_lines": ["Slot: Main hand"]},
		],
		inventory,
		{},
		67,
		[]
	)
	state = panel.get_debug_state()
	_assert_eq("unique chest panel mode", str(state.get("container_mode", "")), "unique_chest")
	_assert_eq("unique chest entity id retained", str(state.get("stash_entity_id", "")), "unique_chest_entity_1")
	_assert_false("unique chest deposit gold hidden", bool(state.get("deposit_gold_enabled", true)))
	_assert_false("unique chest withdraw gold hidden", bool(state.get("withdraw_gold_enabled", true)))
	panel.bot_drag_stash_to_bag("unique_item_1")
	_assert_eq("unique chest take emitted count", emitted.size(), 9)
	_assert_eq("unique chest take emitted type", str(emitted[8]["type"]), "unique_chest_take_item_intent")
	_assert_eq("unique chest take emitted entity", str(emitted[8]["payload"].get("chest_entity_id", "")), "unique_chest_entity_1")
	_assert_eq("unique chest take emitted item", str(emitted[8]["payload"].get("chest_item_id", "")), "unique_item_1")
	var drag_start_position: Dictionary = (panel.get_debug_state().get("window", {}) as Dictionary).get("position", {})
	panel.bot_drag_window_by(Vector2(35, 20))
	state = panel.get_debug_state()
	var moved_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("stash drag moved x", int(moved_position.get("x", 0)), int(drag_start_position.get("x", 0)) + 35)
	_assert_eq("stash drag moved y", int(moved_position.get("y", 0)), int(drag_start_position.get("y", 0)) + 20)
	panel.bot_drag_window_by(Vector2(-10000, -10000))
	state = panel.get_debug_state()
	var clamped_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("stash drag clamps x", int(clamped_position.get("x", -1)), 0)
	_assert_eq("stash drag clamps y", int(clamped_position.get("y", -1)), 0)
	panel.bot_click_close()
	_assert_false("stash close button hides panel", panel.visible)

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
