extends SceneTree

const ClassBodyTintScript := preload("res://scripts/class_body_tint.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_all_classes_define_body_tint()
	_test_barbarian_tint_shifts_red()
	_test_rogue_tint_darkens()
	print("[gdtest] PASS: test_class_body_tint (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_all_classes_define_body_tint() -> void:
	for class_id in ["barbarian", "paladin", "ranger", "rogue", "sorcerer"]:
		var cfg := ClassPresentationsLoaderScript.body_tint_for_class(class_id)
		_assert(not cfg.is_empty(), "%s body_tint missing" % class_id)
		_assert(float(cfg.get("strength", 0.0)) > 0.0, "%s body_tint strength should be positive" % class_id)


func _test_barbarian_tint_shifts_red() -> void:
	var before := ClassBodyTintScript.DEFAULT_SKIN
	var after := ClassBodyTintScript.representative_color("barbarian")
	_assert(after.r > before.r, "barbarian tint should increase red channel")
	_assert(after.g < before.g + 0.02, "barbarian tint should not oversaturate green")


func _test_rogue_tint_darkens() -> void:
	var before := ClassBodyTintScript.DEFAULT_SKIN
	var after := ClassBodyTintScript.representative_color("rogue")
	_assert(after.r < before.r, "rogue tint should darken red channel")
	_assert(after.g < before.g, "rogue tint should darken green channel")
	_assert(after.b < before.b + 0.02, "rogue tint should stay neutral-dark")


func _assert(condition: bool, message: String) -> void:
	if condition:
		_pass_count += 1
	else:
		_fail_count += 1
		printerr("[gdtest] FAIL: %s" % message)
