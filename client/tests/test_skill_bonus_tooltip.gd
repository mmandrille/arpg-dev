extends SceneTree

const SkillBonusTooltipScript := preload("res://scripts/skill_bonus_tooltip.gd")


func _init() -> void:
	_test_active_line_is_green()
	_test_wrong_class_line_is_red()
	_test_unlearned_line_is_gray()
	print("[gdtest] PASS: test_skill_bonus_tooltip")
	quit()


func _test_active_line_is_green() -> void:
	var item := {
		"skill_bonus_status": [{
			"skill_id": "magic_bolt",
			"skill_class": "sorcerer",
			"value": 2,
			"active": true,
			"display": "+2 Magic Bolt",
		}],
	}
	var lines := SkillBonusTooltipScript.lines_for_item(item, "sorcerer")
	_assert_eq(lines.size(), 1, "one active line")
	_assert_eq(lines[0].get("text", ""), "+2 Magic Bolt", "display text")
	_assert_true("active line is green", _colors_close(lines[0].get("color", Color.WHITE), SkillBonusTooltipScript.ACTIVE_COLOR))


func _test_wrong_class_line_is_red() -> void:
	var item := {
		"skill_bonus_status": [{
			"skill_id": "heal",
			"skill_class": "paladin",
			"value": 2,
			"active": false,
			"display": "+2 Heal",
		}],
	}
	var lines := SkillBonusTooltipScript.lines_for_item(item, "sorcerer")
	_assert_true("wrong class line is red", _colors_close(lines[0].get("color", Color.WHITE), SkillBonusTooltipScript.INACTIVE_WRONG_CLASS_COLOR))


func _test_unlearned_line_is_gray() -> void:
	var item := {
		"skill_bonus_status": [{
			"skill_id": "magic_bolt",
			"skill_class": "sorcerer",
			"value": 1,
			"active": false,
			"display": "+1 Magic Bolt",
		}],
	}
	var lines := SkillBonusTooltipScript.lines_for_item(item, "sorcerer")
	_assert_true("unlearned line is gray", _colors_close(lines[0].get("color", Color.WHITE), SkillBonusTooltipScript.INACTIVE_UNLEARNED_COLOR))


func _colors_close(a: Color, b: Color) -> bool:
	return abs(a.r - b.r) < 0.01 and abs(a.g - b.g) < 0.01 and abs(a.b - b.b) < 0.01


func _assert_eq(got: Variant, want: Variant, label: String) -> void:
	if got != want:
		push_error("%s: got %s want %s" % [label, str(got), str(want)])
		quit(1)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		push_error(label)
		quit(1)
