extends SceneTree

const PathRejectBackoffScript := preload("res://scripts/path_reject_backoff.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_target_backoff_blocks_resend()
	_test_goal_backoff_blocks_resend()
	_test_backoff_expires()
	print("[gdtest] PASS: test_path_reject_backoff (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_target_backoff_blocks_resend() -> void:
	var backoff: PathRejectBackoff = PathRejectBackoffScript.new()
	backoff.note_target_reject("2001", 1000, 750)
	_assert_true("target blocked during backoff", backoff.blocks_target("2001", 1500))
	_assert_true("other target not blocked", not backoff.blocks_target("2002", 1500))


func _test_goal_backoff_blocks_resend() -> void:
	var backoff: PathRejectBackoff = PathRejectBackoffScript.new()
	var goal := Vector2(8.0, 5.0)
	backoff.note_goal_reject(goal, 2000, 750)
	_assert_true("goal blocked during backoff", backoff.blocks_goal(goal, 2500))
	_assert_true("different goal not blocked", not backoff.blocks_goal(Vector2(9.0, 5.0), 2500))


func _test_backoff_expires() -> void:
	var backoff: PathRejectBackoff = PathRejectBackoffScript.new()
	backoff.note_target_reject("2001", 1000, 500)
	_assert_true("target blocked before expiry", backoff.blocks_target("2001", 1400))
	_assert_true("target unblocked after expiry", not backoff.blocks_target("2001", 1600))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)
