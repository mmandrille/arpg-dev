# Unit test for class affinity tooltip lines.
# Run via: godot --headless --path client --script res://tests/test_class_affinity_tooltip.gd
extends SceneTree

const ClassAffinityTooltipScript := preload("res://scripts/class_affinity_tooltip.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_status_lines_use_active_color()
	_test_fallback_affinities_for_wrong_class()
	print("[gdtest] PASS: test_class_affinity_tooltip (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_status_lines_use_active_color() -> void:
	var item := {
		"class_affinity_status": [
			{"display": "+10% attack speed (Rogue)", "active": true},
			{"display": "-15% attack speed (non-Paladin)", "active": false},
		],
	}
	var lines := ClassAffinityTooltipScript.lines_for_item(item)
	_assert_eq("status line count", lines.size(), 2)
	_assert_eq("active color", (lines[0] as Dictionary).get("color"), ClassAffinityTooltipScript.ACTIVE_COLOR)
	_assert_eq("inactive color", (lines[1] as Dictionary).get("color"), ClassAffinityTooltipScript.INACTIVE_COLOR)


func _test_fallback_affinities_for_wrong_class() -> void:
	var item := {
		"class_affinities": [
			{"class": "barbarian", "stat": "damage_percent", "value": 10},
		],
	}
	var lines := ClassAffinityTooltipScript.lines_for_item(item, "rogue")
	_assert_eq("fallback line count", lines.size(), 1)
	_assert_eq("wrong class inactive", (lines[0] as Dictionary).get("color"), ClassAffinityTooltipScript.INACTIVE_COLOR)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
