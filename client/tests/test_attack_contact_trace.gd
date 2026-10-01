extends SceneTree

const AttackContactTraceScript := preload("res://scripts/attack_contact_trace.gd")
const BotWaitHandlersScript := preload("res://scripts/bot_wait_handlers.gd")
const BotScenarioRunnerScript := preload("res://scripts/bot_scenario_runner.gd")
const BotCombatContactAssertionsScript := preload("res://scripts/bot_combat_contact_assertions.gd")

var _failures := 0


func _initialize() -> void:
	_test_trace_sequence()
	_test_equipped_item_wait()
	_test_timeline_scenario()
	_test_combat_contact_and_buffer_assertions()
	if _failures > 0:
		printerr("[gdtest] FAIL: test_attack_contact_trace (%d failed)" % _failures)
		quit(1)
	else:
		print("[gdtest] PASS: test_attack_contact_trace")
		quit(0)


func _test_trace_sequence() -> void:
	AttackContactTraceScript.begin_test()
	AttackContactTraceScript.record("input_attack", {"target_id": "1007"})
	AttackContactTraceScript.record("swing_start", {"clip": "attack", "target_id": "1007"})
	AttackContactTraceScript.sample({"last_tick": 11, "local_player_presentation": {"animation": {"current_clip": "attack", "clip_position_s": 0.1, "is_moving": false}}})
	AttackContactTraceScript.record_result({"event_type": "player_damaged", "entity_id": "1001"}, 12)
	AttackContactTraceScript.record_result({"event_type": "monster_damaged", "entity_id": "1007", "outcome": "hit"}, 12)
	AttackContactTraceScript.record_text({"event_type": "monster_damaged", "outcome": "hit"}, "1007")
	AttackContactTraceScript.sample({"last_tick": 12, "local_player_presentation": {"animation": {"current_clip": "hit", "clip_position_s": 0.0, "is_moving": false}}})
	AttackContactTraceScript.sample({"last_tick": 12, "local_player_presentation": {"animation": {"current_clip": "idle", "clip_position_s": 0.0, "is_moving": false}}})
	var rows := AttackContactTraceScript.rows_for_test()
	var stages: Array = []
	for row in rows:
		stages.append(str(row.get("stage", "")))
	_assert_eq("trace stage order", stages, ["input_attack", "swing_start", "frame_sample", "result_received", "feedback_text_requested", "frame_sample", "frame_sample", "locomotion_return"])
	_assert_eq("clip position sampled in milliseconds", (rows[2] as Dictionary).get("clip_position_ms"), 100.0)
	_assert_eq("server tick retained at result", (rows[3] as Dictionary).get("server_tick"), 12)
	_assert_eq("authoritative outcome retained", (rows[3] as Dictionary).get("outcome"), "hit")


func _test_equipped_item_wait() -> void:
	var step := {"type": "wait_equipped", "slot": "main_hand", "item_def_id": "bow"}
	var state := {"equipped": {"main_hand": null}, "inventory": [{"item_instance_id": "item-bow", "item_def_id": "bow"}]}
	_assert_false("null equipment is absent", BotWaitHandlersScript.evaluate(null, step, "wait_equipped", state))
	state["equipped"] = {"main_hand": "item-sword"}
	state["inventory"].append({"item_instance_id": "item-sword", "item_def_id": "long_sword"})
	_assert_false("old weapon does not satisfy bow wait", BotWaitHandlersScript.evaluate(null, step, "wait_equipped", state))
	state["equipped"] = {"main_hand": "item-bow"}
	_assert_true("new bow satisfies wait", BotWaitHandlersScript.evaluate(null, step, "wait_equipped", state))


func _test_timeline_scenario() -> void:
	var path := ProjectSettings.globalize_path("res://") + "../tools/bot/scenarios/client/v496_attack_contact_timeline.json"
	var scenario = JSON.parse_string(FileAccess.get_file_as_string(path))
	_assert_true("timeline scenario parses", scenario is Dictionary)
	if scenario is Dictionary:
		_assert_eq("timeline scenario validates", BotScenarioRunnerScript.validate_scenario(scenario as Dictionary), "")


func _test_combat_contact_and_buffer_assertions() -> void:
	var contact_step := {"outcome": "block", "blocked": true, "target_monster_def_id": "combat_lab_blocking_target", "impact_feedback_delta_max": 0}
	var state := {
		"last_monster_damage_feedback": {"outcome": "block", "blocked": true, "target_monster_def_id": "combat_lab_blocking_target", "impact_feedback_delta": 0},
		"attack_buffer": {"active": true, "target_id": "1007", "queued_count": 2, "replaced_count": 0, "cleared_count": 0},
	}
	_assert_true("blocked event without contact passes", BotCombatContactAssertionsScript.matches("assert_combat_contact", contact_step, state))
	state["last_monster_damage_feedback"]["impact_feedback_delta"] = 1
	_assert_false("blocked event with hit reaction fails", BotCombatContactAssertionsScript.matches("assert_combat_contact", contact_step, state))
	var buffer_step := {"active": true, "queued_count_min": 1, "queued_count_max": 2, "replaced_count_max": 0}
	_assert_true("bounded rapid attack queue passes", BotCombatContactAssertionsScript.matches("assert_attack_buffer", buffer_step, state))
	state["attack_buffer"]["queued_count"] = 3
	_assert_false("unbounded rapid attack queue fails", BotCombatContactAssertionsScript.matches("assert_attack_buffer", buffer_step, state))
	_assert_eq("contact assertion step validates", BotScenarioRunnerScript.validate_step(contact_step.merged({"type": "assert_combat_contact"}), 0), "")
	_assert_eq("buffer assertion step validates", BotScenarioRunnerScript.validate_step(buffer_step.merged({"type": "assert_attack_buffer"}), 0), "")


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got != want:
		_failures += 1
		push_error("[gdtest] %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_false(label: String, value: bool) -> void:
	_assert_eq(label, value, false)


func _assert_true(label: String, value: bool) -> void:
	_assert_eq(label, value, true)
