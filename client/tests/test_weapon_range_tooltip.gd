extends SceneTree

const WeaponRangeTooltipScript := preload("res://scripts/weapon_range_tooltip.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	ItemRulesLoader.ensure_loaded()
	_test_bow_range_line()
	_test_bow_projectile_speed_line()
	_test_melee_range_line()
	_test_fractional_range_line()
	_test_ensure_after_slot()
	_finish()


func _test_bow_range_line() -> void:
	var line := WeaponRangeTooltipScript.line_for_item({
		"item_def_id": "bow",
		"item_template_id": "bow",
	})
	_assert_eq("bow range line", line, "Range: 12 tiles")


func _test_bow_projectile_speed_line() -> void:
	var line := WeaponRangeTooltipScript.projectile_speed_line_for_item({
		"item_def_id": "bow",
		"item_template_id": "bow",
	})
	_assert_eq("bow projectile speed line", line, "Projectile speed: 25 tiles/s")


func _test_melee_range_line() -> void:
	var line := WeaponRangeTooltipScript.format_range_line(1.0)
	_assert_eq("melee whole tile uses singular", line, "Range: 1 tile")


func _test_fractional_range_line() -> void:
	var line := WeaponRangeTooltipScript.format_range_line(1.5)
	_assert_eq("fractional reach keeps decimal", line, "Range: 1.5 tiles")


func _test_ensure_after_slot() -> void:
	var lines: Array = ["Slot: Both hands", "Damage 1-3"]
	WeaponRangeTooltipScript.ensure_after_slot(lines, {
		"item_def_id": "bow",
		"item_template_id": "bow",
	})
	_assert_eq("range inserted after slot", str(lines[1]), "Range: 12 tiles")
	_assert_eq("projectile speed inserted after range", str(lines[2]), "Projectile speed: 25 tiles/s")


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_weapon_range_tooltip (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_weapon_range_tooltip (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
