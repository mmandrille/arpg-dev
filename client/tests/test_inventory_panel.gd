# Unit test for the inventory panel.
# Run via: godot --headless --path client --script res://tests/test_inventory_panel.gd
extends SceneTree

const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_identical_state_refresh_keeps_slots_alive()
	print("[gdtest] PASS: test_inventory_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_identical_state_refresh_keeps_slots_alive() -> void:
	var panel := InventoryPanelScript.new()
	root.add_child(panel)
	await process_frame

	var blade := {
		"item_instance_id": "item-1",
		"item_def_id": "cave_blade",
		"item_template_id": "cave_blade",
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
		"item_def_id": "cave_ring",
		"item_template_id": "cave_ring",
		"display_name": "Training Ring",
		"rarity": "magic",
		"summary_lines": ["Magic +1"],
	}
	panel.set_inventory_state([blade.duplicate(true), ring], {}, 3, 15, 12)
	await process_frame
	var changed_slot_id := int(panel._bag_grid.get_child(0).get_instance_id())
	_assert_true("changed inventory refresh rebuilds slots", changed_slot_id != stable_slot_id)
	panel.free()


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
