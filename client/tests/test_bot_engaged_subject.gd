extends SceneTree

const WaitHandlersScript := preload("res://scripts/bot_wait_handlers.gd")


class Runner:
	extends RefCounted
	var _memory := {}
	var _controller = null
	var _step_elapsed := 0.0
	var _last_retry_at := 0.0
	var pending_action := {}
	var failure := ""

	func _event_matches(_step: Dictionary, _event: Dictionary) -> bool:
		return true

	func _presentation_row_matches(step: Dictionary, rec: Dictionary) -> bool:
		return str(rec.get("type", "")) == str(step.get("entity_type", "")) \
			and str(rec.get("monster_def_id", "")) == str(step.get("monster_def_id", ""))

	func _fail(message: String) -> void:
		failure = message


func _initialize() -> void:
	var runner := Runner.new()
	var aggro_step := {
		"type": "click_entity_until_event", "entity_type": "monster",
		"monster_def_id": "dungeon_mob", "event_type": "monster_aggro",
		"remember_event_entity_id": true,
	}
	var state := {"pending_events": [{"event_type": "monster_aggro", "entity_id": "m1"}]}
	if not WaitHandlersScript.evaluate(runner, aggro_step, "click_entity_until_event", state):
		_fail("aggro event was not accepted")
		return
	if str(runner._memory.get("remembered_event_entity_id", "")) != "m1":
		_fail("aggro subject was not remembered")
		return
	var near_step := {
		"type": "wait_entity_near_player", "entity_type": "monster",
		"monster_def_id": "dungeon_mob", "distance": 4.0,
		"remembered_event_entity": true, "require_alive": true, "in_view": true,
	}
	state = {
		"player_pos": {"x": 0.0, "z": 0.0},
		"entities_presentation_debug": [
			{"id": "m1", "type": "monster", "monster_def_id": "dungeon_mob", "hp": 7, "position": {"x": 8.0, "z": 0.0}},
			{"id": "m2", "type": "monster", "monster_def_id": "dungeon_mob", "hp": 7, "position": {"x": 2.0, "z": 0.0}},
		],
	}
	if WaitHandlersScript.evaluate(runner, near_step, "wait_entity_near_player", state):
		_fail("unrelated nearby monster satisfied engaged-subject wait")
		return
	var rows: Array = state["entities_presentation_debug"]
	var engaged: Dictionary = rows[0]
	engaged["position"] = {"x": 3.0, "z": 0.0}
	if not WaitHandlersScript.evaluate(runner, near_step, "wait_entity_near_player", state):
		_fail("remembered engaged monster was not accepted nearby")
		return
	engaged["hp"] = 0
	if WaitHandlersScript.evaluate(runner, near_step, "wait_entity_near_player", state):
		_fail("dead engaged monster satisfied live view")
		return
	print("[gdtest] PASS: test_bot_engaged_subject")
	quit(0)


func _fail(message: String) -> void:
	push_error(message)
	print("[gdtest] FAIL: test_bot_engaged_subject — %s" % message)
	quit(1)
