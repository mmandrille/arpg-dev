# Unit test for the town vendor shop panel.
# Run via: godot --headless --path client --script res://tests/test_shop_panel.gd
extends SceneTree

const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const MarketPanelScript := preload("res://scripts/market_panel.gd")
const BlacksmithPanelScript := preload("res://scripts/blacksmith_panel.gd")
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
	var window: Dictionary = state.get("window", {})
	_assert_eq("shop window title", str(window.get("title", "")), "Town Vendor")
	_assert_true("shop window has close button", bool(window.get("close_visible", false)))
	_assert_true("shop window is draggable", bool(window.get("draggable", false)))
	_assert_eq("offer count", int(state.get("offer_count", 0)), 4)
	_assert_eq("fixed count", int(state.get("fixed_offer_count", 0)), 2)
	_assert_eq("generated count", int(state.get("generated_offer_count", 0)), 2)
	_assert_eq("buyback count", int(state.get("buyback_offer_count", 0)), 0)
	_assert_false("vendor reroll hidden", bool(state.get("reroll_visible", false)))
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
	_assert_eq("vendor tooltip gold value", offer_tooltip.debug_gold_value_text(), "50 gold")
	offer_tooltip.queue_free()
	var magic_offer_tooltip := panel._make_offer_tooltip(offers[3])
	_assert_eq("vendor magic tooltip name is rarity blue", magic_offer_tooltip.debug_first_main_line_color(), "93c5fd")
	var magic_offer_fonts: Array = magic_offer_tooltip.debug_main_line_font_sizes()
	_assert_eq("vendor tooltip rarity uses smaller font", _font_size_for_text(magic_offer_fonts, "Rarity: Magic"), 19)
	_assert_eq("vendor tooltip slot uses smaller font", _font_size_for_text(magic_offer_fonts, "Slot: gloves"), 19)
	magic_offer_tooltip.queue_free()
	var set_offer: Dictionary = offers[3].duplicate(true)
	set_offer["display_name"] = "Verdant Vanguard Gloves"
	set_offer["rarity"] = "set"
	var set_offer_tooltip := panel._make_offer_tooltip(set_offer)
	_assert_eq("vendor set tooltip name is rarity green", set_offer_tooltip.debug_first_main_line_color(), "55e66f")
	set_offer_tooltip.queue_free()

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
	_assert_eq("shop window title updates", str((state.get("window", {}) as Dictionary).get("title", "")), "Mystery Seller")
	_assert_eq("mystery offer count", int(state.get("mystery_offer_count", 0)), 1)
	_assert_eq("mystery fixed count", int(state.get("fixed_offer_count", 0)), 0)
	_assert_true("mystery reroll visible", bool(state.get("reroll_visible", false)))
	_assert_true("mystery reroll enabled", bool(state.get("reroll_enabled", false)))
	_assert_eq("mystery reroll cost", int(state.get("reroll_cost", 0)), 50)
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
	_assert_eq("mystery tooltip gold value", mystery_tooltip.debug_gold_value_text(), "75 gold")
	mystery_tooltip.queue_free()
	panel.bot_click_reroll()
	_assert_eq("mystery reroll emitted count", emitted.size(), 4)
	_assert_eq("mystery reroll emitted type", str(emitted[3]["type"]), "shop_reroll_intent")
	_assert_eq("mystery reroll emitted entity", str(emitted[3]["payload"].get("shop_entity_id", "")), "1005")
	panel.bot_click_buy_offer("", "mystery", 0)
	_assert_eq("mystery buy emitted count", emitted.size(), 5)
	_assert_eq("mystery buy emitted entity", str(emitted[4]["payload"].get("shop_entity_id", "")), "1005")
	_assert_eq("mystery buy emitted offer", str(emitted[4]["payload"].get("offer_id", "")), "mystery:wp:-3:ring:000")
	var drag_start_position: Dictionary = (panel.get_debug_state().get("window", {}) as Dictionary).get("position", {})
	panel.bot_drag_window_by(Vector2(35, 20))
	state = panel.get_debug_state()
	var moved_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("shop drag moved x", int(moved_position.get("x", 0)), int(drag_start_position.get("x", 0)) + 35)
	_assert_eq("shop drag moved y", int(moved_position.get("y", 0)), int(drag_start_position.get("y", 0)) + 20)
	panel.bot_drag_window_by(Vector2(-10000, -10000))
	state = panel.get_debug_state()
	var clamped_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("shop drag clamps x", int(clamped_position.get("x", -1)), 0)
	_assert_eq("shop drag clamps y", int(clamped_position.get("y", -1)), 0)
	panel.bot_click_close()
	_assert_false("shop close button hides panel", panel.visible)

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
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [{"slot_index": 0, "item_instance_id": "2003"}], 2)
	inventory_panel._handle_shift_click(inventory[1])
	_assert_eq("inventory shift click hotbar emitted count", inventory_emitted.size(), 1)
	_assert_eq("inventory shift click hotbar emitted type", str(inventory_emitted[0]["type"]), "assign_hotbar_intent")
	_assert_eq("inventory shift click hotbar slot", int(inventory_emitted[0]["payload"].get("slot_index", -1)), 1)
	_assert_eq("inventory shift click hotbar item", str(inventory_emitted[0]["payload"].get("item_instance_id", "")), "2002")
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [{"slot_index": 0, "item_instance_id": "2003"}, {"slot_index": 1, "item_instance_id": "2002"}], 2)
	inventory_panel._handle_shift_click(inventory[1])
	_assert_eq("inventory shift click full belt emits no intent", inventory_emitted.size(), 1)
	_assert_eq("inventory shift click full belt hint", inventory_panel._gesture_hint.text, "belt full")
	var amulet := {
		"item_instance_id": "2003",
		"item_def_id": "cave_amulet",
		"item_template_id": "cave_amulet",
		"display_name": "Common Cave Amulet",
		"rarity": "common",
		"rolled_stats": {"max_hp": 4, "mana_regen_per_10_seconds": 2},
	}
	var amulet_lines: Array = inventory_panel._tooltip_lines(amulet)
	_assert_false("inventory amulet no base title", _array_contains_text(amulet_lines, "Base stats"))
	_assert_false("inventory amulet no random title", _array_contains_text(amulet_lines, "Random stats"))
	_assert_true("inventory amulet base max hp", _array_contains_text(amulet_lines, "Max HP: +2"))
	_assert_true("inventory amulet random mana regen", _array_contains_text(amulet_lines, "Mana regen / 10s: +2"))
	var first_separator := amulet_lines.find(InventoryPanelScript.TOOLTIP_STAT_SEPARATOR)
	var last_separator := amulet_lines.rfind(InventoryPanelScript.TOOLTIP_STAT_SEPARATOR)
	_assert_true("inventory amulet two separators", first_separator >= 0 and last_separator > first_separator)
	_assert_true("inventory amulet separator before base", first_separator < amulet_lines.find("Max HP: +2"))
	_assert_true("inventory amulet separator before random", last_separator < amulet_lines.find("Mana regen / 10s: +2"))
	var blade := {
		"item_instance_id": "2004",
		"item_def_id": "cave_blade",
		"item_template_id": "cave_blade",
		"display_name": "Magic Cave Blade",
		"rarity": "magic",
		"rolled_stats": {"damage_min": 3, "damage_max": 4, "max_hp": 3},
	}
	var blade_lines: Array = inventory_panel._tooltip_lines(blade)
	_assert_true("inventory blade type reach before stats", blade_lines.find("Reach: 1.5") < blade_lines.find(InventoryPanelScript.TOOLTIP_STAT_SEPARATOR))
	_assert_true("inventory blade type mode before stats", blade_lines.find("Mode: melee") < blade_lines.find(InventoryPanelScript.TOOLTIP_STAT_SEPARATOR))
	_assert_true("inventory blade random min delta", _array_contains_text(blade_lines, "Min damage: +1"))
	_assert_false("inventory blade hides zero max delta", _array_contains_text(blade_lines, "Max damage:"))
	_assert_true("inventory blade random hp delta", _array_contains_text(blade_lines, "Max HP: +3"))
	inventory_panel.set_inventory_state([sell_appraisals[0]], {}, 3, 15, 60)
	var inventory_state := inventory_panel.get_debug_state()
	_assert_true("inventory requirements rendered", int(inventory_state.get("requirement_row_count", 0)) >= 2)
	_assert_true("inventory equip previews rendered", int(inventory_state.get("equip_preview_row_count", 0)) >= 1)
	var inventory_tooltip := inventory_panel._make_item_tooltip(sell_appraisals[0])
	_assert_eq("inventory tooltip uses shared panel", inventory_tooltip.get_script(), ItemTooltipPanelScript)
	_assert_eq("inventory tooltip gold value", inventory_tooltip.debug_gold_value_text(), "27 gold")
	_assert_eq("inventory tooltip item level footer", inventory_tooltip.debug_item_level_text(), "Level 2")
	_assert_false("inventory tooltip hides level from requirements block", _array_contains_text(inventory_tooltip.debug_requirement_texts(), "Level 2"))
	_assert_false("inventory tooltip stats exclude requirements", _array_contains_text(inventory_panel._tooltip_lines(sell_appraisals[0]), "Requires"))
	_assert_true("inventory tooltip requirements extracted", _array_contains_text(inventory_panel._requirement_lines(sell_appraisals[0]), "Level 2"))
	_assert_true("inventory tooltip stat requirements extracted", _array_contains_text(inventory_panel._requirement_lines(sell_appraisals[0]), "%s 15(-3)" % StatLabels.display_name("str")))
	_assert_true("inventory tooltip preview extracted", _array_contains_text(inventory_panel._comparison_entries(sell_appraisals[0]), "preview"))
	_assert_true("inventory tooltip comparison extracted", _array_contains_text(inventory_panel._comparison_entries(sell_appraisals[0]), "vs equipped"))
	inventory_tooltip.queue_free()
	var blade_tooltip := inventory_panel._make_item_tooltip(blade)
	_assert_eq("inventory magic tooltip name is rarity blue", blade_tooltip.debug_first_main_line_color(), "93c5fd")
	var blade_fonts: Array = blade_tooltip.debug_main_line_font_sizes()
	_assert_eq("inventory tooltip rarity uses smaller font", _font_size_for_text(blade_fonts, "Rarity: Magic"), 19)
	_assert_eq("inventory tooltip slot uses smaller font", _font_size_for_text(blade_fonts, "Slot: main_hand"), 19)
	blade_tooltip.queue_free()
	var ring := {
		"item_instance_id": "2005",
		"item_def_id": "cave_ring",
		"item_template_id": "cave_ring",
		"display_name": "Common Cave Ring",
		"rarity": "common",
		"slot": "ring",
		"rolled_stats": {"max_hp": 2, "mana_regen_per_10_seconds": 2},
	}
	var ring_tooltip := inventory_panel._make_item_tooltip(ring)
	_assert_eq("inventory generated tooltip computed gold value", ring_tooltip.debug_gold_value_text(), "21 gold")
	ring_tooltip.queue_free()
	var unique_blade := {
		"item_instance_id": "2006",
		"item_def_id": "cave_blade",
		"item_template_id": "cave_blade",
		"display_name": "Embercall Blade",
		"rarity": "unique",
		"slot": "main_hand",
		"summary_lines": ["Slot: Main hand", "Damage 5-9", "Requires level 2"],
		"requirements": {"level": 2},
		"comparison": {"slot": "main_hand", "deltas": [{"stat": "damage_max", "offered": 9, "equipped": 4, "delta": 5}]},
		"effect_ids": ["everburning_wound"],
	}
	var unique_plain_lines := Array(inventory_panel._tooltip(unique_blade).split("\n"))
	_assert_true("inventory unique tooltip names effect", _array_contains_text(unique_plain_lines, "Unique effect: Everburning Wound"))
	_assert_eq("inventory unique tooltip effect summary at bottom", str(unique_plain_lines[unique_plain_lines.size() - 1]), "All hero damage applies burn for 10 seconds, ticking once per second for 10% of the original hit damage.")
	var unique_tooltip := inventory_panel._make_item_tooltip(unique_blade)
	var unique_tooltip_texts: Array = unique_tooltip.debug_main_line_font_sizes()
	_assert_true("inventory unique rendered tooltip names effect", _array_contains_text(unique_tooltip_texts, "Unique effect: Everburning Wound"))
	_assert_true("inventory unique rendered tooltip summarizes effect", _array_contains_text(unique_tooltip_texts, "All hero damage applies burn"))
	unique_tooltip.queue_free()
	var potion_tooltip := inventory_panel._make_item_tooltip(inventory[1])
	_assert_eq("inventory fixed tooltip computed gold value", potion_tooltip.debug_gold_value_text(), "10 gold")
	potion_tooltip.queue_free()
	inventory_emitted.clear()
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
	inventory_panel.clear_shop_sell_context()
	inventory_panel.set_market_context("offer")
	inventory_panel._handle_double_click(inventory[0])
	_assert_eq("inventory double click market emitted count", inventory_emitted.size(), 3)
	_assert_eq("inventory double click market type", str(inventory_emitted[2]["type"]), "market_stage_inventory_item")
	_assert_eq("inventory double click market context", str(inventory_emitted[2]["payload"].get("context", "")), "offer")
	inventory_panel.set_market_hidden_item_ids(["2001"])
	var hidden_state := inventory_panel.get_debug_state()
	_assert_eq("inventory hides staged market offer item", int(hidden_state.get("visible_bag_count", 0)), 1)
	inventory_panel.clear_market_context()
	hidden_state = inventory_panel.get_debug_state()
	_assert_eq("inventory restores hidden market item on context clear", int(hidden_state.get("visible_bag_count", 0)), 2)
	inventory_panel.set_blacksmith_context(true)
	inventory_panel._handle_double_click(inventory[0])
	_assert_eq("inventory double click blacksmith emitted count", inventory_emitted.size(), 4)
	_assert_eq("inventory double click blacksmith type", str(inventory_emitted[3]["type"]), "blacksmith_stage_inventory_item")
	var weapon_sets := [
		{"index": 0, "main_hand": "2001", "off_hand": null},
		{"index": 1, "main_hand": null, "off_hand": null},
	]
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [], 2, 0, weapon_sets)
	inventory_panel.viewed_weapon_set = 1
	inventory_panel._handle_drop_on_slot("equip:main_hand", {"source": "bag", "item": inventory[0]})
	_assert_eq("inventory hand tab equip emitted count", inventory_emitted.size(), 5)
	_assert_eq("inventory hand tab equip type", str(inventory_emitted[4]["type"]), "equip_intent")
	_assert_eq("inventory hand tab equip weapon set", int(inventory_emitted[4]["payload"].get("weapon_set", -1)), 1)
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [], 2, 0, weapon_sets)
	_assert_eq("inventory preserves manually viewed inactive hand tab", int(inventory_panel.get_debug_state().get("viewed_weapon_set", -1)), 1)
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [], 2, 1, weapon_sets)
	_assert_eq("inventory follows active weapon set swap to set 2", int(inventory_panel.get_debug_state().get("viewed_weapon_set", -1)), 1)
	inventory_panel.set_inventory_state(inventory, {}, 3, 15, 60, [], 2, 0, weapon_sets)
	_assert_eq("inventory follows active weapon set swap back to set 1", int(inventory_panel.get_debug_state().get("viewed_weapon_set", -1)), 0)
	inventory_panel.queue_free()

	var market_panel := MarketPanelScript.new()
	root.add_child(market_panel)
	await process_frame
	var market_emitted: Array = []
	market_panel.market_action_requested.connect(func(action: String, payload: Dictionary) -> void:
		market_emitted.append({"action": action, "payload": payload.duplicate(true)})
	)
	market_panel.show_market("market-1", [{
		"listing_id": "listing-1",
		"item_def_id": "cave_mail",
		"item_template_id": "cave_mail",
		"display_name": "Magic Cave Mail",
		"rarity": "magic",
		"seller_account_id": "other",
		"price_gold": 25,
		"requirements": {"level": 1},
		"rolled_stats": {"armor": 6},
	}, {
		"listing_id": "listing-owned",
		"item_def_id": "cave_bow",
		"item_template_id": "cave_bow",
		"display_name": "Common Cave Bow",
		"rarity": "common",
		"seller_account_id": "me",
		"price_gold": 25,
		"requirements": {"level": 1},
		"rolled_stats": {"damage_min": 1, "damage_max": 2},
	}], inventory, "me", "Active listings")
	var initial_market_state := market_panel.get_debug_state()
	var initial_tab_titles: Array = initial_market_state.get("visible_tab_titles", [])
	_assert_true("market offer tab hidden by default", not initial_tab_titles.has("Offer"))
	_assert_eq("market browse hides owned listings", (initial_market_state.get("listing_rows", []) as Array).size(), 1)
	_assert_eq("market publish shows owned listings", (initial_market_state.get("owned_listing_rows", []) as Array).size(), 1)
	market_panel.stage_inventory_item("publish", inventory[0])
	market_panel.bot_set_publish_price(33)
	market_panel._emit_publish_action()
	_assert_eq("market publish inventory action", str(market_emitted[0]["action"]), "publish_inventory")
	_assert_eq("market publish inventory item", str(market_emitted[0]["payload"].get("item_instance_id", "")), "2001")
	_assert_eq("market publish inventory price", int(market_emitted[0]["payload"].get("price_gold", 0)), 33)
	market_panel.selected_listing_id = "listing-1"
	market_panel.stage_inventory_item("offer", inventory[0])
	market_panel.stage_inventory_item("offer", inventory[1])
	market_panel._emit_offer_action()
	_assert_eq("market offer inventory action", str(market_emitted[1]["action"]), "offer_inventory")
	_assert_eq("market offer inventory count", (market_emitted[1]["payload"].get("item_instance_ids", []) as Array).size(), 2)
	market_panel.return_to_browse_after_offer()
	var market_state := market_panel.get_debug_state()
	_assert_eq("market returns to browse after offer", int(market_state.get("tab", -1)), 0)
	_assert_eq("market clears staged offer after return", int(market_state.get("staged_offer_count", -1)), 0)
	market_panel.selected_listing_id = "listing-1"
	market_panel.stage_inventory_item("offer", inventory[0])
	market_panel.stage_inventory_item("offer", inventory[1])
	market_state = market_panel.get_debug_state()
	_assert_eq("market staged offer count", int(market_state.get("staged_offer_count", 0)), 2)
	var staged_offer_slots: Array = market_state.get("staged_offer_slots", [])
	_assert_eq("market staged offer slot count", staged_offer_slots.size(), 10)
	_assert_true("market staged offer slot uses icon", bool((staged_offer_slots[0] as Dictionary).get("has_icon", false)))
	_assert_true("market staged offer slot uses shared tooltip", bool((staged_offer_slots[0] as Dictionary).get("uses_shared_tooltip", false)))
	var staged_slot_size: Dictionary = (staged_offer_slots[0] as Dictionary).get("slot_size", {})
	_assert_eq("market staged offer slot inventory width", int(staged_slot_size.get("x", 0)), 68)
	var staged_tooltip := market_panel._make_item_tooltip(inventory[0])
	_assert_eq("market staged tooltip uses shared panel", staged_tooltip.get_script(), ItemTooltipPanelScript)
	staged_tooltip.queue_free()
	var market_unique_tooltip := market_panel._make_item_tooltip(unique_blade)
	var market_unique_texts: Array = market_unique_tooltip.debug_main_line_font_sizes()
	_assert_true("market unique tooltip names effect", _array_contains_text(market_unique_texts, "Unique effect: Everburning Wound"))
	_assert_true("market unique tooltip summarizes effect", _array_contains_text(market_unique_texts, "All hero damage applies burn"))
	market_unique_tooltip.queue_free()
	_assert_true("market publish rows centered", bool(market_state.get("publish_rows_centered", false)))
	_assert_true("market offer rows top aligned", bool(market_state.get("offer_rows_top_aligned", false)))
	_assert_true("market publish price and button share row", bool(market_state.get("publish_button_same_row", false)))
	_assert_eq("market publish price half width", int(market_state.get("publish_price_width", 0)), 180)
	_assert_eq("market publish button half width", int(market_state.get("publish_button_width", 0)), 180)
	var listing_rows: Array = market_state.get("listing_rows", [])
	_assert_true("market listing debug row exists", not listing_rows.is_empty())
	var listing_row: Dictionary = listing_rows[0] if not listing_rows.is_empty() else {}
	_assert_eq("market browse listing is foreign", str(listing_row.get("seller_account_id", "")), "other")
	_assert_true("market listing row has item icon", bool(listing_row.get("has_icon", false)))
	_assert_false("market visible detail hides listing id", str(listing_row.get("visible_detail", "")).contains("listing-1"))
	_assert_true("market listing shows level", _array_contains_text(listing_row.get("stat_lines", []), "Level 1"))
	_assert_true("market listing shows base armor", _array_contains_text(listing_row.get("stat_lines", []), "Base Armor: +3"))
	_assert_true("market listing shows rolled armor", _array_contains_text(listing_row.get("stat_lines", []), "Rolled Armor: +3"))
	market_panel.show_offers("listing-owned", [{
		"offer_id": "offer-1",
		"listing_id": "listing-owned",
		"bidder_account_id": "bidder",
		"status": "active",
		"items": [inventory[0], inventory[1]],
	}])
	market_state = market_panel.get_debug_state()
	_assert_true("market offer tab visible while viewing offers", (market_state.get("visible_tab_titles", []) as Array).has("Offer"))
	_assert_eq("market offer view selected tab", int(market_state.get("tab", -1)), 2)
	var offer_rows: Array = market_state.get("offer_rows", [])
	_assert_eq("market viewed offer row count", offer_rows.size(), 1)
	var viewed_offer_slots: Array = (offer_rows[0] as Dictionary).get("item_slots", [])
	_assert_eq("market viewed offer slot count", viewed_offer_slots.size(), 2)
	_assert_true("market viewed offer slot uses icon", bool((viewed_offer_slots[0] as Dictionary).get("has_icon", false)))
	_assert_true("market viewed offer slot uses shared tooltip", bool((viewed_offer_slots[0] as Dictionary).get("uses_shared_tooltip", false)))
	market_panel.queue_free()

	var blacksmith_panel := BlacksmithPanelScript.new()
	root.add_child(blacksmith_panel)
	await process_frame
	var blacksmith_emitted: Array = []
	blacksmith_panel.upgrade_inventory_requested.connect(func(item_instance_id: String) -> void:
		blacksmith_emitted.append(item_instance_id)
	)
	var upgrade_item: Dictionary = (inventory[0] as Dictionary).duplicate(true)
	upgrade_item["rolled_stats"] = {"damage_min": 2, "damage_max": 4}
	upgrade_item["summary_lines"] = ["Min damage: +2", "Max damage: +4"]
	blacksmith_panel.show_blacksmith("smith-1", [upgrade_item], 40, 60, {"item_upgrade_cost_gold": 100, "item_upgrade_cost_growth_per_level": 50, "item_upgrade_max_level": 3}, "Choose")
	blacksmith_panel.stage_inventory_item(upgrade_item)
	var blacksmith_state := blacksmith_panel.get_debug_state()
	_assert_eq("blacksmith staged item id", str(blacksmith_state.get("staged_item_id", "")), "2001")
	_assert_eq("blacksmith wallet gold", int(blacksmith_state.get("wallet_gold", 0)), 100)
	var blacksmith_window_size: Dictionary = (blacksmith_state.get("window", {}) as Dictionary).get("minimum_size", {})
	_assert_eq("blacksmith compact window width", int(blacksmith_window_size.get("x", 0)), 320)
	_assert_eq("blacksmith compact window height", int(blacksmith_window_size.get("y", 0)), 254)
	var stage_size: Dictionary = blacksmith_state.get("stage_slot_size", {})
	_assert_eq("blacksmith stage slot width", int(stage_size.get("x", 0)), 84)
	_assert_eq("blacksmith stage slot height", int(stage_size.get("y", 0)), 84)
	_assert_eq("blacksmith stage slot centered", bool(blacksmith_state.get("stage_slot_centered", false)), true)
	_assert_eq("blacksmith stage icon visible", bool(blacksmith_state.get("stage_icon_visible", false)), true)
	_assert_true("blacksmith direct preview includes min damage", _array_contains_text(blacksmith_state.get("preview_lines", []), "Min damage: 2 -> 3"))
	_assert_false("blacksmith instruction removed", bool(blacksmith_state.get("instruction_visible", true)))
	var summary_shield: Dictionary = upgrade_item.duplicate(true)
	summary_shield["rolled_stats"] = {"item_level": 0}
	summary_shield["summary_lines"] = ["Armor: +2", "Block: +12%"]
	blacksmith_panel.stage_inventory_item(summary_shield)
	blacksmith_state = blacksmith_panel.get_debug_state()
	_assert_true("blacksmith summary preview includes armor", _array_contains_text(blacksmith_state.get("preview_lines", []), "Armor: 2 -> 3"))
	_assert_true("blacksmith summary preview includes block", _array_contains_text(blacksmith_state.get("preview_lines", []), "Block: 12 -> 13"))
	var common_bow: Dictionary = upgrade_item.duplicate(true)
	common_bow["item_def_id"] = "cave_bow"
	common_bow["display_name"] = "Common Cave Bow"
	common_bow["rarity"] = "common"
	common_bow["rolled_stats"] = {"item_level": 0}
	common_bow["summary_lines"] = []
	blacksmith_panel.stage_inventory_item(common_bow)
	blacksmith_state = blacksmith_panel.get_debug_state()
	_assert_true("blacksmith template preview includes bow min damage", _array_contains_text(blacksmith_state.get("preview_lines", []), "Min damage: 1 -> 2"))
	_assert_true("blacksmith template preview includes bow max damage", _array_contains_text(blacksmith_state.get("preview_lines", []), "Max damage: 2 -> 3"))
	blacksmith_panel.unstage_item()
	blacksmith_state = blacksmith_panel.get_debug_state()
	_assert_eq("blacksmith unstage clears item", str(blacksmith_state.get("staged_item_id", "")), "")
	blacksmith_panel.stage_inventory_item(upgrade_item)
	blacksmith_panel._emit_upgrade(upgrade_item)
	_assert_eq("blacksmith inventory upgrade emitted", str(blacksmith_emitted[0]), "2001")
	blacksmith_panel.hide_display()
	blacksmith_state = blacksmith_panel.get_debug_state()
	_assert_eq("blacksmith close returns staged item", str(blacksmith_state.get("staged_item_id", "")), "")
	blacksmith_panel.queue_free()

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


func _font_size_for_text(rows: Array, text: String) -> int:
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("text", "")) == text:
			return int((row as Dictionary).get("font_size", -1))
	return -1
