extends SceneTree

const BotScenarioRunnerScript := preload("res://scripts/bot_scenario_runner.gd")

var _failed := false


func _initialize() -> void:
	_test_wait_entity_near_player()
	if not _failed:
		print("[gdtest] PASS: test_bot_entity_distance")
	quit(1 if _failed else 0)


func _test_wait_entity_near_player() -> void:
	var data := {
		"id": "entity_near_player_wait_test",
		"runner": "godot_client",
		"world_id": "ranged_monster_ai_lab",
		"client_steps": [
			{"type": "wait_entity_near_player", "entity_type": "monster", "monster_def_id": "dungeon_archer", "distance": 3.0, "timeout_s": 5.0},
		],
	}
	_assert_eq("wait entity near player scenario valid", BotScenarioRunnerScript.validate_scenario(data), "")
	_assert_ne("distance is required", BotScenarioRunnerScript.validate_step({"type": "wait_entity_near_player", "entity_type": "monster", "timeout_s": 5.0}, 0), "")
	var runner := BotScenarioRunnerScript.new()
	runner.load_scenario(data)
	runner.tick(0.016, {
		"player_pos": {"x": 6.0, "z": 5.0},
		"entities_presentation_debug": [{"type": "monster", "monster_def_id": "dungeon_archer", "position": {"x": 10.0, "z": 5.0}}],
	})
	_assert_true("far archer does not pass", not runner.is_done())
	runner.tick(0.016, {
		"player_pos": {"x": 6.0, "z": 5.0},
		"entities_presentation_debug": [{"type": "monster", "monster_def_id": "dungeon_archer", "position": {"x": 8.5, "z": 5.0}}],
	})
	_assert_true("near archer passes", runner.is_done() and runner.passed())


func _assert_eq(label: String, got, expected) -> void:
	if got != expected:
		printerr("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
		_failed = true


func _assert_ne(label: String, got, not_expected) -> void:
	if got == not_expected:
		printerr("[gdtest] FAIL %s: expected something other than %s" % [label, str(not_expected)])
		_failed = true


func _assert_true(label: String, condition: bool) -> void:
	if not condition:
		printerr("[gdtest] FAIL %s" % label)
		_failed = true
