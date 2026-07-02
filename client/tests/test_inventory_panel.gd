# Unit test for the inventory panel.
# Run via: godot --headless --path client --script res://tests/test_inventory_panel.gd
extends SceneTree

const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ItemRequirementViewsScript := preload("res://scripts/item_requirement_views.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_identical_state_refresh_keeps_slots_alive()
	_test_requirement_warning_helpers()
	await _test_inventory_bag_marks_unmet_requirements()
	print("[gdtest] PASS: test_inventory_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_identical_state_refresh_keeps_slots_alive() -> void:
	var panel := InventoryPanelScript.new()
	root.add_child(panel)
	await process_frame

	var blade := {
		"item_instance_id": "item-1",
		"item_def_id": "long_sword",
		"item_template_id": "long_sword",
		"display_name": "Training Blade",
		"rarity": "common",
		"summary_lines": ["Damage 1-2"],
	}
	panel.set_inventory_state([blade], {}, 3, 15, 12)
	await process_frame
	var first_slot_id := int(panel._bag_grid.get_child(0).get_instance_id())

	panel.set_inventory_state([blade.duplicate(true)], {}, 3, 15, 12)
	await process_frame
	var stable_slot_id := int(panel._bag_grid.get_child(0).get_instance_id())
	_assert_eq("identical inventory refresh keeps hovered slot", stable_slot_id, first_slot_id)

	var ring := {
		"item_instance_id": "item-2",
		"item_def_id": "ring",
		"item_template_id": "ring",
		"display_name": "Training Ring",
		"rarity": "magic",
		"summary_lines": ["Magic +1"],
	}
	panel.set_inventory_state([blade.duplicate(true), ring], {}, 3, 15, 12)
	await process_frame
	var changed_slot_id := int(panel._bag_grid.get_child(0).get_instance_id())
	_assert_true("changed inventory refresh rebuilds slots", changed_slot_id != stable_slot_id)
	panel.free()


func _test_requirement_warning_helpers() -> void:
	var met_item := {
		"item_def_id": "boots",
		"requirements_met": true,
		"requirement_status": [{"stat": "str", "required": 10, "current": 12, "met": true}],
	}
	var unmet_item := {
		"item_def_id": "boots",
		"requirements_met": false,
		"requirement_status": [{"stat": "str", "required": 14, "current": 12, "met": false}],
	}
	var consumable := {"item_def_id": "health_potion", "requirements_met": false}
	_assert_true("met requirements pass", ItemRequirementViewsScript.requirements_met(met_item))
	_assert_false("unmet requirements fail", ItemRequirementViewsScript.requirements_met(unmet_item))
	_assert_true("equippable unmet shows warning", ItemRequirementViewsScript.shows_invalid_requirement_warning(unmet_item, true))
	_assert_false("equippable met hides warning", ItemRequirementViewsScript.shows_invalid_requirement_warning(met_item, true))
	_assert_false("consumable never shows warning", ItemRequirementViewsScript.shows_invalid_requirement_warning(consumable, false))


func _test_inventory_bag_marks_unmet_requirements() -> void:
	var panel := InventoryPanelScript.new()
	root.add_child(panel)
	var boots := {
		"item_instance_id": "boots-1",
		"item_def_id": "boots",
		"item_template_id": "boots",
		"display_name": "Vigorous Boots",
		"rarity": "rare",
		"slot": "boots",
		"requirements_met": false,
		"requirement_status": [{"stat": "str", "required": 14, "current": 12, "met": false}],
	}
	panel.set_inventory_state([boots], {}, 3, 15, 0)
	await process_frame
	_assert_true("bag item flagged as invalid requirements", panel._item_shows_requirement_warning(boots))
	var slot: InventoryPanelScript.InventorySlotButton = panel._bag_grid.get_child(0)
	var normal_style: StyleBoxFlat = slot.get_theme_stylebox("normal")
	_assert_true("invalid requirement slot border is reddish", normal_style.border_color.r > 0.55 and normal_style.border_color.g < 0.45)
	panel.free()


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


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
