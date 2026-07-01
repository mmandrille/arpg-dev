# Headless tests for ConnectionRecovery backoff behavior.
extends SceneTree

const ConnectionRecoveryScript := preload("res://scripts/connection_recovery.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	MainConfigLoaderScript.ensure_loaded()
	_test_delay_grows_with_attempt()
	_test_begin_starts_active_recovery()
	_test_tick_requests_connect_then_waits_after_transport_loss()
	_test_mark_resynced_clears_active_state()
	_test_give_up_marks_failed()
	if _fail == 0:
		print("[gdtest] PASS: test_connection_recovery (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_connection_recovery (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_delay_grows_with_attempt() -> void:
	var first := ConnectionRecoveryScript.delay_for_attempt(1)
	var second := ConnectionRecoveryScript.delay_for_attempt(2)
	_assert_true("second delay >= first", second >= first)
	_assert_true("first delay positive", first > 0.0)


func _test_begin_starts_active_recovery() -> void:
	var recovery := ConnectionRecoveryScript.new()
	recovery.begin()
	_assert_true("begin active", recovery.is_active())
	_assert_true("begin blocks input", recovery.blocks_input())


func _test_tick_requests_connect_then_waits_after_transport_loss() -> void:
	var recovery := ConnectionRecoveryScript.new()
	recovery.begin()
	var first := recovery.tick(0.0)
	_assert_eq("first action connect", str(first.get("action", "")), "connect_ws")
	recovery.note_transport_lost()
	_assert_eq("phase waiting after loss", recovery.phase, ConnectionRecovery.Phase.WAITING)
	var waiting := recovery.tick(0.0)
	_assert_eq("waiting does not reconnect immediately", str(waiting.get("action", "")), "none")


func _test_mark_resynced_clears_active_state() -> void:
	var recovery := ConnectionRecoveryScript.new()
	recovery.begin()
	recovery.mark_connected_awaiting_snapshot()
	recovery.mark_resynced()
	_assert_true("resync clears active", not recovery.is_active())


func _test_give_up_marks_failed() -> void:
	var recovery := ConnectionRecoveryScript.new()
	recovery.begin()
	recovery.elapsed_s = 999.0
	var result := recovery.tick(0.0)
	_assert_eq("give up action", str(result.get("action", "")), "failed")
	_assert_eq("failed phase", recovery.phase, ConnectionRecovery.Phase.FAILED)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
		return
	_fail += 1
	print("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	_assert_eq(label, value, true)
