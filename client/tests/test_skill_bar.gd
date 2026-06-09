# Unit test for the v44 skill hotbar slot.
# Run via: godot --headless --path client --script res://tests/test_skill_bar.gd
extends SceneTree

const SkillBarScript := preload("res://scripts/skill_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var bar := SkillBarScript.new()
	root.add_child(bar)
	await process_frame

	var emitted: Array = []
	bar.cast_skill_requested.connect(func(skill_id: String) -> void:
		emitted.append(skill_id)
	)

	bar.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": [{"skill_id": "magic_bolt", "rank": 0, "max_rank": 5, "can_spend": false}],
	})
	bar.set_interactive(true)
	var state := bar.get_debug_state()
	_assert_eq("rank starts zero", int(state.get("rank", -1)), 0)
	_assert_false("unranked skill disabled", bool(state.get("enabled", true)))

	bar.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": [{"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": false}],
	})
	state = bar.get_debug_state()
	_assert_true("ranked skill enabled", bool(state.get("enabled", false)))
	bar.use_slot()
	_assert_eq("cast signal count", emitted.size(), 1)
	_assert_eq("cast signal skill", str(emitted[0]), "magic_bolt")

	bar.set_skill_cooldowns([{"skill_id": "magic_bolt", "remaining_ticks": 40, "total_ticks": 40}])
	state = bar.get_debug_state()
	_assert_false("cooldown disables slot", bool(state.get("enabled", true)))
	_assert_eq("cooldown remaining", int(state.get("remaining_ticks", -1)), 40)
	_assert_eq("cooldown total", int(state.get("total_ticks", -1)), 40)
	_assert_true("cooldown fraction full", float(state.get("cooldown_fraction", 0.0)) > 0.99)

	bar._process(0.10)
	state = bar.get_debug_state()
	_assert_true("cooldown locally decays", int(state.get("remaining_ticks", 40)) < 40)

	bar.set_skill_cooldowns([])
	state = bar.get_debug_state()
	_assert_true("slot re-enables without cooldown", bool(state.get("enabled", false)))

	bar.queue_free()
	print("[gdtest] PASS: test_skill_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
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
