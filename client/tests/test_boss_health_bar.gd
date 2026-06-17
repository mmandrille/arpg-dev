# Unit test for the v53 boss health bar.
# Run via: godot --headless --path client --script res://tests/test_boss_health_bar.gd
extends SceneTree

const BossHealthBarScript := preload("res://scripts/boss_health_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var bar: BossHealthBar = BossHealthBarScript.new()
	root.add_child(bar)
	await process_frame

	var state := bar.get_debug_state()
	_assert_false("starts hidden", bool(state.get("visible", true)))

	bar.show_boss("3001", "cave_warden", "Cave Warden", 18, 24)
	state = bar.get_debug_state()
	_assert_true("shows live boss", bool(state.get("visible", false)))
	_assert_eq("boss id", str(state.get("boss_id", "")), "3001")
	_assert_eq("template id", str(state.get("boss_template_id", "")), "cave_warden")
	_assert_eq("title", str(state.get("title", "")), "Cave Warden")
	_assert_eq("hp", int(state.get("hp", 0)), 18)
	_assert_eq("max hp", int(state.get("max_hp", 0)), 24)
	_assert_true("ratio is three quarters", absf(float(state.get("ratio", 0.0)) - 0.75) < 0.001)
	_assert_true("portrait visible", bool(state.get("portrait_visible", false)))
	_assert_eq("portrait kind", str(state.get("portrait_kind", "")), "cave_warden")
	_assert_eq("portrait label", str(state.get("portrait_label", "")), "CW")

	bar.show_boss("3001", "cave_warden", "Cave Warden", 30, 24)
	state = bar.get_debug_state()
	_assert_eq("hp clamps to max", int(state.get("hp", 0)), 24)
	_assert_true("ratio clamps to one", absf(float(state.get("ratio", 0.0)) - 1.0) < 0.001)

	bar.set_phase_state({
		"phase_kind": "telegraph",
		"pattern_id": "charged_melee",
		"phase_index": 0,
		"duration_ticks": 30,
		"remaining_ticks": 15,
	})
	state = bar.get_debug_state()
	_assert_eq("phase kind", str(state.get("phase_kind", "")), "telegraph")
	_assert_eq("pattern id", str(state.get("pattern_id", "")), "charged_melee")
	_assert_eq("phase index", int(state.get("phase_index", -2)), 0)
	_assert_eq("duration ticks", int(state.get("duration_ticks", 0)), 30)
	_assert_eq("remaining ticks", int(state.get("remaining_ticks", 0)), 15)
	_assert_true("phase ratio half", absf(float(state.get("phase_ratio", 0.0)) - 0.5) < 0.001)

	bar.set_phase_state({
		"phase_kind": "active",
		"duration_ticks": 4,
		"remaining_ticks": 9,
	})
	state = bar.get_debug_state()
	_assert_eq("remaining clamps to duration", int(state.get("remaining_ticks", 0)), 4)
	_assert_true("phase ratio clamps one", absf(float(state.get("phase_ratio", 0.0)) - 1.0) < 0.001)

	bar.clear_phase_state()
	state = bar.get_debug_state()
	_assert_eq("phase kind clears", str(state.get("phase_kind", "x")), "")
	_assert_eq("phase index clears", int(state.get("phase_index", 0)), -1)
	_assert_eq("phase ratio clears", float(state.get("phase_ratio", 1.0)), 0.0)

	bar.show_boss("3001", "cave_warden", "Cave Warden", -5, 0)
	state = bar.get_debug_state()
	_assert_false("dead boss hides", bool(state.get("visible", true)))
	_assert_false("dead hides portrait", bool(state.get("portrait_visible", true)))
	_assert_eq("dead hp clamps zero", int(state.get("hp", -1)), 0)
	_assert_eq("invalid max hp clamps one", int(state.get("max_hp", 0)), 1)
	_assert_eq("dead clears phase", str(state.get("phase_kind", "x")), "")

	bar.hide_boss()
	state = bar.get_debug_state()
	_assert_false("hide clears visibility", bool(state.get("visible", true)))
	_assert_eq("hide clears boss id", str(state.get("boss_id", "x")), "")
	_assert_eq("hide clears portrait kind", str(state.get("portrait_kind", "x")), "")

	bar.queue_free()
	print("[gdtest] PASS: test_boss_health_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
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
