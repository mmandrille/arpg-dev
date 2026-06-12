# Unit test for status effect icons above the hotbar.
# Run via: godot --headless --path client --script res://tests/test_status_effects_bar.gd
extends SceneTree

const StatusEffectsBarScript := preload("res://scripts/status_effects_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0
var _expired_effects: Array = []


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var bar := StatusEffectsBarScript.new()
	bar.effect_expired.connect(_on_effect_expired)
	root.add_child(bar)
	await process_frame

	var state := bar.get_debug_state()
	_assert_eq("starts empty", (state.get("effects", []) as Array).size(), 0)
	_assert_false("starts hidden", bool(state.get("visible", true)))

	bar.start_effect({
		"event_type": "skill_effect_started",
		"skill_id": "rage",
		"remaining_ticks": 450,
		"total_ticks": 450,
	})
	state = bar.get_debug_state()
	var effects: Array = state.get("effects", [])
	_assert_eq("rage effect appears", effects.size(), 1)
	_assert_true("bar visible with effect", bool(state.get("visible", false)))
	if effects.size() > 0:
		var effect: Dictionary = effects[0]
		_assert_eq("rage id", str(effect.get("skill_id", "")), "rage")
		_assert_eq("rage icon label", str(effect.get("label", "")), "R")
		_assert_eq("rage total ticks", int(effect.get("total_ticks", -1)), 450)
		_assert_true("rage fraction full", float(effect.get("fraction", 0.0)) > 0.99)

	bar._process(1.0)
	state = bar.get_debug_state()
	effects = state.get("effects", [])
	if effects.size() > 0:
		var decayed: Dictionary = effects[0]
		_assert_eq("rage decays at authoritative tick rate", int(decayed.get("remaining_ticks", 450)), 440)
		_assert_true("rage fraction lowers", float(decayed.get("fraction", 1.0)) < 1.0)

	bar.end_effect("rage")
	state = bar.get_debug_state()
	_assert_eq("rage removed on end", (state.get("effects", []) as Array).size(), 0)
	_assert_false("bar hidden after end", bool(state.get("visible", true)))

	bar.start_effect({
		"event_type": "skill_effect_started",
		"skill_id": "holy_shield",
		"remaining_ticks": 300,
		"total_ticks": 300,
	})
	state = bar.get_debug_state()
	effects = state.get("effects", [])
	_assert_eq("holy shield effect appears", effects.size(), 1)
	if effects.size() > 0:
		var shield: Dictionary = effects[0]
		_assert_eq("holy shield id", str(shield.get("skill_id", "")), "holy_shield")
		_assert_eq("holy shield icon label", str(shield.get("label", "")), "S")
		_assert_eq("holy shield total ticks", int(shield.get("total_ticks", -1)), 300)
	bar.end_effect("holy_shield")
	state = bar.get_debug_state()
	_assert_eq("holy shield removed on end", (state.get("effects", []) as Array).size(), 0)

	bar.start_effect({"skill_id": "rage", "remaining_ticks": 1, "total_ticks": 1})
	bar._process(1.0)
	state = bar.get_debug_state()
	_assert_eq("expired effect removed locally", (state.get("effects", []) as Array).size(), 0)
	_assert_true("expired effect emitted", _expired_effects.has("rage"))

	bar.queue_free()
	print("[gdtest] PASS: test_status_effects_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _on_effect_expired(skill_id: String) -> void:
	_expired_effects.append(skill_id)


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
