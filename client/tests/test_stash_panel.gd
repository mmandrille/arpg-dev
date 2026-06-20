# Unit test for the account stash panel.
# Run via: godot --headless --path client --script res://tests/test_stash_panel.gd
extends SceneTree

const StashPanelScript := preload("res://scripts/stash_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const InventoryTransferRouterScript := preload("res://scripts/inventory_transfer_router.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const MainScript := preload("res://scripts/main.gd")

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
	await process_frame
	var first_stash_slot_id := int(panel._stash_grid.get_child(0).get_instance_id())
	panel.show_stash("1005", "account_stash", _dup_array(stash_items), 3, 50, _dup_array(inventory), {}, 7, [], "Account Stash")
	await process_frame
	_assert_eq("identical account stash refresh keeps hovered slot", int(panel._stash_grid.get_child(0).get_instance_id()), first_stash_slot_id)

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
	_assert_true("test equip slot kind recognized", InventoryTransferRouterScript.is_equipment_slot("equip:main_hand"))
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
			{"stash_item_id": "unique_item_2", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "display_name": "Stormstring Bow", "rarity": "unique", "slot": "main_hand", "summary_lines": ["Slot: Main hand"]},
			{"stash_item_id": "unique_item_3", "item_def_id": "cave_ring", "item_template_id": "cave_ring", "display_name": "Bloodbound Sigil", "rarity": "unique", "slot": "ring", "summary_lines": ["Slot: Ring"]},
			{"stash_item_id": "unique_item_4", "item_def_id": "cave_amulet", "item_template_id": "cave_amulet", "display_name": "Lantern Amulet", "rarity": "unique", "slot": "amulet", "summary_lines": ["Slot: Amulet"]},
			{"stash_item_id": "set_item_1", "item_def_id": "verdant_vanguard_blade", "item_template_id": "verdant_vanguard_blade", "display_name": "Verdant Vanguard Blade", "rarity": "set", "slot": "main_hand", "summary_lines": ["Slot: Main hand", "Set: Verdant Vanguard (2/5 equipped)", "2-piece set bonus: Armor +3 (active)", "3-piece set bonus: Max HP +8 (inactive)"]},
			{"stash_item_id": "set_item_2", "item_def_id": "verdant_vanguard_helm", "item_template_id": "verdant_vanguard_helm", "display_name": "Verdant Vanguard Helm", "rarity": "set", "slot": "head", "summary_lines": ["Slot: Head"]},
			{"stash_item_id": "set_item_3", "item_def_id": "stormrunner_covenant_bow", "item_template_id": "stormrunner_covenant_bow", "display_name": "Stormrunner Covenant Bow", "rarity": "set", "slot": "main_hand", "summary_lines": ["Slot: Main hand"]},
			{"stash_item_id": "set_item_4", "item_def_id": "stormrunner_covenant_mask", "item_template_id": "stormrunner_covenant_mask", "display_name": "Stormrunner Covenant Mask", "rarity": "set", "slot": "head", "summary_lines": ["Slot: Head"]},
			{"stash_item_id": "set_item_5", "item_def_id": "stormrunner_covenant_loop", "item_template_id": "stormrunner_covenant_loop", "display_name": "Stormrunner Covenant Loop", "rarity": "set", "slot": "ring", "summary_lines": ["Slot: Ring"]},
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
	_assert_true("unique chest tabs visible", bool(state.get("unique_chest_tabs_visible", false)))
	_assert_eq("unique chest starts on uniques tab", str(state.get("unique_chest_active_tab", "")), "uniques")
	var unique_chest_counts: Dictionary = state.get("unique_chest_tab_counts", {})
	_assert_eq("unique chest unique tab count", int(unique_chest_counts.get("uniques", 0)), 4)
	_assert_eq("unique chest set tab count", int(unique_chest_counts.get("sets", 0)), 5)
	_assert_eq("unique chest uniques tab filters items", int(state.get("filtered_stash_item_count", 0)), 4)
	_assert_true("unique chest uniques tab includes first unique", _row_for_stash_id(state.get("stash_rows", []), "unique_item_1").size() > 0)
	var main := MainScript.new()
	main.stash_panel = panel
	main.inventory = _dup_array(inventory)
	main.equipped = {}
	main.gold = 67
	main.hotbar = []
	main.stash_items = [
		{"stash_item_id": "account_unique_1", "item_def_id": "starter_sorcerer_staff", "item_template_id": "starter_sorcerer_staff", "display_name": "Account Staff", "rarity": "unique"},
		{"stash_item_id": "account_unique_2", "item_def_id": "starter_sorcerer_staff", "item_template_id": "starter_sorcerer_staff", "display_name": "Account Staff", "rarity": "unique"},
		{"stash_item_id": "account_unique_3", "item_def_id": "starter_sorcerer_staff", "item_template_id": "starter_sorcerer_staff", "display_name": "Account Staff", "rarity": "unique"},
		{"stash_item_id": "account_set_1", "item_def_id": "cave_helm", "item_template_id": "cave_helm", "display_name": "Account Set", "rarity": "set"},
		{"stash_item_id": "account_set_2", "item_def_id": "cave_helm", "item_template_id": "cave_helm", "display_name": "Account Set", "rarity": "set"},
		{"stash_item_id": "account_set_3", "item_def_id": "cave_helm", "item_template_id": "cave_helm", "display_name": "Account Set", "rarity": "set"},
		{"stash_item_id": "account_set_4", "item_def_id": "cave_helm", "item_template_id": "cave_helm", "display_name": "Account Set", "rarity": "set"},
	]
	main._refresh_inventory_ui()
	state = panel.get_debug_state()
	unique_chest_counts = state.get("unique_chest_tab_counts", {})
	_assert_eq("inventory refresh preserves unique chest unique count", int(unique_chest_counts.get("uniques", 0)), 4)
	_assert_eq("inventory refresh preserves unique chest set count", int(unique_chest_counts.get("sets", 0)), 5)
	main.free()
	panel.bot_select_unique_chest_tab("sets")
	state = panel.get_debug_state()
	_assert_eq("unique chest sets tab selected", str(state.get("unique_chest_active_tab", "")), "sets")
	_assert_eq("unique chest sets tab filters items", int(state.get("filtered_stash_item_count", 0)), 5)
	var set_item_row := _row_for_stash_id(state.get("stash_rows", []), "set_item_1")
	_assert_true("unique chest sets tab includes first set", set_item_row.size() > 0)
	var set_chest_tooltip := panel._make_item_tooltip(set_item_row)
	_assert_eq("unique chest set tooltip name is rarity green", set_chest_tooltip.debug_first_main_line_color(), "55e66f")
	_assert_true("unique chest tooltip ignores mouse recursively", _all_controls_ignore_mouse(set_chest_tooltip))
	var set_chest_texts: Array = set_chest_tooltip.debug_main_line_font_sizes()
	_assert_true("unique chest set tooltip separates equipped count", _array_contains_text(set_chest_texts, "(2/5 equipped)"))
	_assert_true("unique chest set tooltip separates bonus label", _array_contains_text(set_chest_texts, "2-piece set bonus:"))
	_assert_true("unique chest set tooltip separates bonus value", _array_contains_text(set_chest_texts, "Armor +3 (active)"))
	set_chest_tooltip.queue_free()
	var unique_chest_item := {
		"stash_item_id": "unique_item_1",
		"item_def_id": "cave_blade",
		"item_template_id": "cave_blade",
		"display_name": "Embercall Blade",
		"rarity": "unique",
		"slot": "main_hand",
		"summary_lines": ["Slot: Main hand"],
		"effect_ids": ["everburning_wound"],
	}
	var unique_chest_tooltip := panel._make_item_tooltip(unique_chest_item)
	_assert_eq("unique chest tooltip uses shared panel", unique_chest_tooltip.get_script(), ItemTooltipPanelScript)
	_assert_eq("unique chest unique tooltip name is rarity orange", unique_chest_tooltip.debug_first_main_line_color(), "ffb26b")
	var unique_chest_texts: Array = unique_chest_tooltip.debug_main_line_font_sizes()
	_assert_true("unique chest tooltip names effect", _array_contains_text(unique_chest_texts, "Unique effect: Everburning Wound"))
	_assert_true("unique chest tooltip summarizes effect", _array_contains_text(unique_chest_texts, "All hero damage applies burn"))
	unique_chest_tooltip.queue_free()
	await process_frame
	var first_unique_chest_slot_id := int(panel._stash_grid.get_child(0).get_instance_id())
	panel.show_unique_chest("unique_chest_entity_1", _dup_array(panel.stash_items), inventory, {}, 67, [])
	await process_frame
	_assert_eq("identical unique chest refresh keeps hovered slot", int(panel._stash_grid.get_child(0).get_instance_id()), first_unique_chest_slot_id)
	panel.bot_drag_stash_to_bag("set_item_1")
	_assert_eq("unique chest take emitted count", emitted.size(), 9)
	_assert_eq("unique chest take emitted type", str(emitted[8]["type"]), "unique_chest_take_item_intent")
	_assert_eq("unique chest take emitted entity", str(emitted[8]["payload"].get("chest_entity_id", "")), "unique_chest_entity_1")
	_assert_eq("unique chest take emitted item", str(emitted[8]["payload"].get("chest_item_id", "")), "set_item_1")
	panel.show_unique_chest(
		"unique_chest_entity_1",
		[
			{"stash_item_id": "unique_item_1", "item_def_id": "cave_blade", "item_template_id": "cave_blade", "display_name": "Embercall Blade", "rarity": "unique", "slot": "main_hand", "summary_lines": ["Slot: Main hand"]},
			{"stash_item_id": "set_item_2", "item_def_id": "verdant_vanguard_helm", "item_template_id": "verdant_vanguard_helm", "display_name": "Verdant Vanguard Helm", "rarity": "set", "slot": "head", "summary_lines": ["Slot: head"]},
		],
		inventory,
		{},
		77
	)
	state = panel.get_debug_state()
	_assert_eq("unique chest refresh keeps sets tab selected", str(state.get("unique_chest_active_tab", "")), "sets")
	_assert_eq("unique chest refresh still filters set rows", str((state.get("stash_rows", [])[0] as Dictionary).get("stash_item_id", "")), "set_item_2")
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


func _array_contains_text(rows: Array, needle: String) -> bool:
	for row in rows:
		if str(row).contains(needle):
			return true
	return false


func _row_for_stash_id(rows: Variant, stash_item_id: String) -> Dictionary:
	if typeof(rows) != TYPE_ARRAY:
		return {}
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("stash_item_id", "")) == stash_item_id:
			return row as Dictionary
	return {}


func _all_controls_ignore_mouse(control: Control) -> bool:
	if control.mouse_filter != Control.MOUSE_FILTER_IGNORE:
		return false
	for child in control.get_children():
		var child_control := child as Control
		if child_control != null and not _all_controls_ignore_mouse(child_control):
			return false
	return true


func _dup_array(values: Array) -> Array:
	var out := []
	for value in values:
		if typeof(value) == TYPE_DICTIONARY:
			out.append((value as Dictionary).duplicate(true))
		else:
			out.append(value)
	return out
