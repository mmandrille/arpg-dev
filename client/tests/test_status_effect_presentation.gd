# Unit tests for local status-effect presentation markers.
#
# Run via: godot --headless --path client --script res://tests/test_status_effect_presentation.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_rage_effect_started_drives_world_aura()
	_test_rage_icon_expiry_clears_world_aura()
	_test_holy_shield_started_blinks_models_in_range()
	_test_holy_shield_ended_clears_local_world_effect()

	print("[gdtest] PASS: test_status_effect_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	get_root().add_child(main.player_anchor)
	get_root().add_child(main.entities_root)
	get_root().add_child(main.walls_root)
	return main


func _test_rage_effect_started_drives_world_aura() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{"event_type": "skill_effect_started", "entity_id": "1001", "skill_id": "rage", "amount": 25}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local rage marker active", bool(local_state.get("has_rage_effect", false)))
	_assert_float("local rage visual scale active", float(local_state.get("visual_scale", 0.0)), 1.25)

	main._apply_delta({"events": [{"event_type": "skill_effect_ended", "entity_id": "1001", "skill_id": "rage"}], "changes": []})
	local_state = main._bot_local_player_presentation()
	_assert_true("local rage marker removed", not bool(local_state.get("has_rage_effect", true)))
	_assert_float("local rage visual scale reset", float(local_state.get("visual_scale", 0.0)), 1.0)
	_free_main(main)


func _test_rage_icon_expiry_clears_world_aura() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{"event_type": "skill_effect_started", "entity_id": "1001", "skill_id": "rage", "amount": 25, "remaining_ticks": 1, "total_ticks": 1}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local rage marker active before icon expiry", bool(local_state.get("has_rage_effect", false)))

	main._on_status_effect_expired("rage")
	local_state = main._bot_local_player_presentation()
	_assert_true("local rage marker removed by icon expiry", not bool(local_state.get("has_rage_effect", true)))
	_assert_float("local rage visual scale reset by icon expiry", float(local_state.get("visual_scale", 0.0)), 1.0)
	_free_main(main)


func _test_holy_shield_started_blinks_models_in_range() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3(2.0, 0.0, 3.0)
	main._upsert_entity({
		"id": "1002",
		"type": "player",
		"position": {"x": 5.0, "y": 3.0},
		"hp": 10,
		"max_hp": 10,
		"character_id": "char_near",
	})
	main._upsert_entity({
		"id": "1003",
		"type": "monster",
		"position": {"x": 13.0, "y": 3.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
	})
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "1001",
		"source_entity_id": "1001",
		"target_entity_id": "1001",
		"skill_id": "holy_shield",
		"correlation_id": "corr_holy_shield_blink",
	}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_eq("local holy shield aura pulse", int(local_state.get("holy_shield_aura_pulses", 0)), 1)
	_assert_eq("local holy shield target pulse", int(local_state.get("holy_shield_target_pulses", 0)), 1)
	var remote_state: Array = main._bot_entities_presentation_debug()
	var near_pulses := -1
	var far_pulses := -1
	for row in remote_state:
		var rec: Dictionary = row
		if str(rec.get("id", "")) == "1002":
			near_pulses = int(rec.get("holy_shield_target_pulses", 0))
		if str(rec.get("id", "")) == "1003":
			far_pulses = int(rec.get("holy_shield_target_pulses", 0))
	_assert_eq("near model holy shield target pulse", near_pulses, 1)
	_assert_eq("far model holy shield target pulse", far_pulses, 0)
	_free_main(main)


func _test_holy_shield_ended_clears_local_world_effect() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1001",
		"type": "player",
		"position": {"x": 0.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"effect_ids": ["holy_shield"],
	})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local holy shield marker active before end event", bool(local_state.get("has_holy_shield_effect", false)))

	main._apply_delta({"events": [{"event_type": "skill_effect_ended", "entity_id": "1001", "skill_id": "holy_shield"}], "changes": []})
	local_state = main._bot_local_player_presentation()
	_assert_true("local holy shield marker removed by end event", not bool(local_state.get("has_holy_shield_effect", true)))
	_free_main(main)


func _free_main(main) -> void:
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_float(label: String, got: float, want: float, epsilon: float = 0.001) -> void:
	if abs(got - want) <= epsilon:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
