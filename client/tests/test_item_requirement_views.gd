# Unit test for item requirement view formatting.
# Run via: godot --headless --path client --script res://tests/test_item_requirement_views.gd
extends SceneTree

const ItemRequirementViewsScript := preload("res://scripts/item_requirement_views.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	TextCatalogScript.ensure_loaded()
	call_deferred("_run")


func _run() -> void:
	_test_class_requirement_line()
	_test_stat_requirement_line()
	print("[gdtest] PASS: test_item_requirement_views (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_class_requirement_line() -> void:
	var line := ItemRequirementViewsScript.format_requirement_status({
		"stat": "class",
		"required": 1,
		"current": 0,
		"met": false,
		"class_id": "paladin",
	})
	_assert_eq("class line text", str(line.get("text", "")), TextCatalogScript.get_text("character.class.paladin", "Paladin"))
	var met_line := ItemRequirementViewsScript.format_requirement_status({
		"stat": "class",
		"required": 1,
		"current": 1,
		"met": true,
		"class_id": "barbarian",
	})
	_assert_eq("met class line text", str(met_line.get("text", "")), TextCatalogScript.get_text("character.class.barbarian", "Barbarian"))


func _test_stat_requirement_line() -> void:
	var line := ItemRequirementViewsScript.format_requirement_status({
		"stat": "str",
		"required": 10,
		"current": 8,
		"met": false,
	})
	_assert_true("stat suffix present", str(line.get("text", "")).ends_with("(-2)"))


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: got %s want %s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: expected true" % label)
