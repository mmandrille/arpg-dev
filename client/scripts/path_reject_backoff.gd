class_name PathRejectBackoff
extends RefCounted

var _target_until_ms: Dictionary = {}
var _goal_until_ms: Dictionary = {}


func clear() -> void:
	_target_until_ms.clear()
	_goal_until_ms.clear()


func clear_target(target_id: String) -> void:
	if target_id != "":
		_target_until_ms.erase(target_id)


func note_target_reject(target_id: String, now_ms: int, duration_ms: int) -> void:
	if target_id == "" or duration_ms <= 0:
		return
	_target_until_ms[target_id] = now_ms + duration_ms


func note_goal_reject(goal: Vector2, now_ms: int, duration_ms: int) -> void:
	if duration_ms <= 0:
		return
	_goal_until_ms[_goal_key(goal)] = now_ms + duration_ms


func blocks_target(target_id: String, now_ms: int) -> bool:
	if target_id == "":
		return false
	var until := int(_target_until_ms.get(target_id, 0))
	return until > now_ms


func blocks_goal(goal: Vector2, now_ms: int) -> bool:
	var until := int(_goal_until_ms.get(_goal_key(goal), 0))
	return until > now_ms


func _goal_key(goal: Vector2) -> String:
	return "%.2f:%.2f" % [goal.x, goal.y]
