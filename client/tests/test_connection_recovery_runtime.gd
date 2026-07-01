# Headless integration tests for ConnectionRecoveryRuntime wiring.
extends SceneTree

const ConnectionRecoveryRuntimeScript := preload("res://scripts/connection_recovery_runtime.gd")
const ConnectionOverlayScript := preload("res://scripts/connection_overlay.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass: int = 0
var _fail: int = 0
var _cleared: bool = false
var _reconciled: bool = false
var _ready_sent: bool = false


func _initialize() -> void:
	MainConfigLoaderScript.ensure_loaded()
	call_deferred("_run")


func _run() -> void:
	await _test_transport_loss_starts_recovery_and_blocks_ready()
	await _test_resync_finishes_recovery()
	if _fail == 0:
		print("[gdtest] PASS: test_connection_recovery_runtime (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_connection_recovery_runtime (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _overlay() -> ConnectionOverlay:
	var overlay := ConnectionOverlayScript.new()
	get_root().add_child(overlay)
	await process_frame
	return overlay


func _test_transport_loss_starts_recovery_and_blocks_ready() -> void:
	var runtime := ConnectionRecoveryRuntimeScript.new()
	var overlay := await _overlay()
	var client := _MockNetClient.new()
	_cleared = false
	_reconciled = false
	_ready_sent = false
	runtime.tick(
		0.0,
		WebSocketPeer.STATE_CLOSED,
		true,
		false,
		false,
		false,
		false,
		true,
		client,
		overlay,
		12,
		Callable(),
		Callable(self, "_mark_cleared"),
		Callable(self, "_mark_reconciled"),
		Callable(),
		Callable(),
		Callable(self, "_ready_sent_get"),
		Callable(self, "_ready_sent_set"),
	)
	_assert_true("recovery active after transport loss", runtime.is_active())
	_assert_true("recovery blocks input", runtime.blocks_input())
	_assert_true("pending state cleared", _cleared)
	_assert_true("reconcile backpressure invoked", _reconciled)
	_assert_true("client_ready reset", not _ready_sent)
	_assert_true("overlay visible", overlay.visible)


func _test_resync_finishes_recovery() -> void:
	var runtime := ConnectionRecoveryRuntimeScript.new()
	var overlay := await _overlay()
	var client := _MockNetClient.new()
	runtime.recovery.begin()
	runtime.recovery.mark_connected_awaiting_snapshot()
	runtime.finish_resync(overlay, Callable(self, "_noop_debug"))
	_assert_true("resync clears active recovery", not runtime.is_active())
	_assert_false("overlay hidden after resync", overlay.visible)


func _noop_debug(_message: String) -> void:
	pass


func _mark_cleared() -> void:
	_cleared = true


func _mark_reconciled() -> void:
	_reconciled = true


func _ready_sent_get() -> bool:
	return _ready_sent


func _ready_sent_set(value: bool) -> void:
	_ready_sent = value


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
		return
	_fail += 1
	print("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	_assert_eq(label, value, true)


func _assert_false(label: String, value: bool) -> void:
	_assert_eq(label, value, false)


class _MockNetClient:
	var session_id: String = "sess_test"
	var reconnect_calls: int = 0
	var resume_calls: int = 0
	var sent_messages: Array = []


	func reconnect_ws() -> void:
		reconnect_calls += 1


	func resume_same_session() -> bool:
		resume_calls += 1
		return true


	func send(_kind: String, _tick: int, _payload: Dictionary = {}) -> void:
		sent_messages.append(_kind)


	func poll() -> Array:
		return []


	func ready_state() -> int:
		return WebSocketPeer.STATE_CLOSED
