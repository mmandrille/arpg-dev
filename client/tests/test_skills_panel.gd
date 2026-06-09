# Unit test for the v44 skills panel.
# Run via: godot --headless --path client --script res://tests/test_skills_panel.gd
extends SceneTree

const SkillsPanelScript := preload("res://scripts/skills_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := SkillsPanelScript.new()
	root.add_child(panel)
	await process_frame

	var emitted: Array = []
	panel.allocate_skill_point_requested.connect(func(skill_id: String) -> void:
		emitted.append(skill_id)
	)

	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": [{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": true}],
	})
	panel.set_interactive(true)
	panel.ensure_display_visible()
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("unspent skill points", int(state.get("unspent_skill_points", -1)), 1)
	_assert_eq("magic bolt rank", int(state.get("rank", -1)), 0)
	_assert_eq("magic bolt max rank", int(state.get("max_rank", -1)), 5)
	_assert_true("spend button enabled", bool(state.get("spend_button_enabled", false)))

	panel.bot_click_skill_button("magic_bolt")
	_assert_eq("spend signal count", emitted.size(), 1)
	_assert_eq("spend signal skill", str(emitted[0]), "magic_bolt")

	panel.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": [{"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": false}],
	})
	state = panel.get_debug_state()
	_assert_eq("rank updates", int(state.get("rank", -1)), 1)
	_assert_false("spend disabled with no points", bool(state.get("spend_button_enabled", true)))

	panel.queue_free()
	print("[gdtest] PASS: test_skills_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
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
