# Focused regression test for shop tooltip hover stability.
# Run via: godot --headless --path client --script res://tests/test_shop_tooltip_stability.gd
extends SceneTree

const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := ShopPanelScript.new()
	root.add_child(panel)
	await process_frame

	var offers := [
		{"offer_id": "fixed:red_potion", "kind": "fixed", "item_def_id": "red_potion", "display_name": "Red Potion", "category": "consumable", "buy_price": 20, "summary_lines": ["Kind: consumable"]},
		{"offer_id": "generated:depth3:000", "kind": "generated", "item_template_id": "bow", "item_def_id": "bow", "display_name": "Bow", "rarity": "common", "slot": "main_hand", "category": "equipment", "item_level": 3, "buy_price": 50, "summary_lines": ["Slot: Main hand"]},
	]
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "bow", "item_template_id": "bow", "display_name": "Bow", "rarity": "common"},
	]
	var sell_appraisals := [
		{"item_instance_id": "2001", "item_def_id": "bow", "item_template_id": "bow", "display_name": "Bow", "rarity": "common", "sell_price": 27},
	]

	panel.show_shop("1004", "town_vendor", offers, 60, inventory, {}, "Town Vendor", sell_appraisals)
	await process_frame
	var first_vendor_slot_id := int(panel._vendor_grid.get_child(0).get_instance_id())
	panel.show_shop("1004", "town_vendor", _dup_array(offers), 60, _dup_array(inventory), {}, "Town Vendor", _dup_array(sell_appraisals))
	await process_frame
	_assert_eq("identical vendor refresh keeps hovered slot", int(panel._vendor_grid.get_child(0).get_instance_id()), first_vendor_slot_id)
	var vendor_tooltip := panel._make_offer_tooltip(offers[1])
	_assert_eq("vendor tooltip uses shared panel", vendor_tooltip.get_script(), ItemTooltipPanelScript)
	_assert_true("vendor tooltip ignores mouse recursively", _all_controls_ignore_mouse(vendor_tooltip))
	vendor_tooltip.queue_free()

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
	await process_frame
	var first_mystery_slot_id := int(panel._vendor_grid.get_child(0).get_instance_id())
	panel.show_shop("1005", "town_mystery_seller", [mystery_offer.duplicate(true)], 100, _dup_array(inventory), {}, "Mystery Seller", [])
	await process_frame
	_assert_eq("identical mystery refresh keeps hovered slot", int(panel._vendor_grid.get_child(0).get_instance_id()), first_mystery_slot_id)
	var mystery_tooltip := panel._make_offer_tooltip(mystery_offer)
	_assert_eq("mystery tooltip uses shared panel", mystery_tooltip.get_script(), ItemTooltipPanelScript)
	_assert_true("mystery tooltip ignores mouse recursively", _all_controls_ignore_mouse(mystery_tooltip))
	mystery_tooltip.queue_free()

	panel.queue_free()
	print("[gdtest] PASS: test_shop_tooltip_stability (%d passed, %d failed)" % [_pass_count, _fail_count])
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
