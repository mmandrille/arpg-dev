## ConnectionRecoveryRuntime — wires ConnectionRecovery into the gameplay coordinator.
class_name ConnectionRecoveryRuntime
extends RefCounted

const ConnectionRecoveryScript := preload("res://scripts/connection_recovery.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var recovery: ConnectionRecovery = ConnectionRecoveryScript.new()
var recovery_count: int = 0


func blocks_input() -> bool:
	return recovery.blocks_input()


func is_active() -> bool:
	return recovery.is_active()


func reset_overlay(overlay: ConnectionOverlay) -> void:
	recovery.cancel()
	if overlay != null:
		overlay.hide_overlay()


static func bot_state_fields(runtime: ConnectionRecoveryRuntime, overlay: ConnectionOverlay, bot_proof_enabled: bool) -> Dictionary:
	return {
		"connection_recovery": {
			"active": runtime.is_active(),
			"blocks_input": runtime.blocks_input(),
			"bot_proof_enabled": bot_proof_enabled,
			"recovery_count": runtime.recovery_count,
		},
		"connection_overlay": overlay.get_debug_state() if overlay != null else {},
	}


func finish_resync(overlay: ConnectionOverlay, debug: Callable) -> void:
	if not recovery.is_active():
		return
	recovery.mark_resynced()
	if overlay != null:
		overlay.hide_overlay()
	if debug.is_valid():
		debug.call("reconnect resync complete")


func tick(
	delta: float,
	ws_state: int,
	gameplay_active: bool,
	bot_mode: bool,
	bot_reconnect_proof: bool,
	visual_replay_enabled: bool,
	intentional_disconnect: bool,
	session_established: bool,
	client,
	overlay: ConnectionOverlay,
	last_server_tick: int,
	handle_message: Callable,
	clear_pending: Callable,
	reconcile_backpressure: Callable,
	debug: Callable,
	return_to_menu: Callable,
	ready_sent_get: Callable,
	ready_sent_set: Callable,
) -> void:
	if not gameplay_active or client == null or (bot_mode and not bot_reconnect_proof) or visual_replay_enabled or intentional_disconnect:
		return
	if not session_established:
		return
	if recovery.phase == ConnectionRecovery.Phase.FAILED:
		return

	if not recovery.is_active():
		if ws_state == WebSocketPeer.STATE_OPEN:
			return
		if ws_state in [WebSocketPeer.STATE_CLOSING, WebSocketPeer.STATE_CLOSED]:
			_start_recovery(client, overlay, clear_pending, reconcile_backpressure, ready_sent_set, debug)
		return

	if recovery.phase == ConnectionRecovery.Phase.AWAITING_SNAPSHOT:
		_poll_messages(client, handle_message)
		if ws_state == WebSocketPeer.STATE_OPEN and not bool(ready_sent_get.call()):
			client.send("client_ready", last_server_tick, {"client_version": "godot", "last_seen_tick": last_server_tick})
			ready_sent_set.call(true)
		if ws_state == WebSocketPeer.STATE_CLOSED:
			ready_sent_set.call(false)
			recovery.note_transport_lost()
		return

	var result := recovery.tick(delta)
	_handle_action(result, client, overlay, ws_state)

	if recovery.phase == ConnectionRecovery.Phase.CONNECTING:
		_poll_messages(client, handle_message)
		ws_state = client.ready_state()
		if ws_state == WebSocketPeer.STATE_OPEN:
			_on_ws_open(client, last_server_tick, ready_sent_get, ready_sent_set, debug)
		elif ws_state == WebSocketPeer.STATE_CLOSED:
			recovery.note_transport_lost()


func on_cancel(return_to_menu: Callable) -> void:
	if return_to_menu.is_valid():
		return_to_menu.call()


func _start_recovery(client, overlay: ConnectionOverlay, clear_pending: Callable, reconcile_backpressure: Callable, ready_sent_set: Callable, debug: Callable) -> void:
	if recovery.is_active():
		return
	recovery.begin()
	if clear_pending.is_valid():
		clear_pending.call()
	if reconcile_backpressure.is_valid():
		reconcile_backpressure.call()
	ready_sent_set.call(false)
	var rules := MainConfigLoaderScript.client_reconnect_rules()
	var max_attempts := int(rules.get("max_attempts", 8))
	if overlay != null:
		overlay.show_reconnecting(0, max_attempts)
	recovery_count += 1
	if debug.is_valid():
		var close_info: Dictionary = client.close_diagnostics()
		debug.call("connection lost; starting reconnect for session %s code=%d reason=%s" % [
			client.session_id,
			int(close_info.get("code", -1)),
			str(close_info.get("reason", "")),
		])


func _handle_action(result: Dictionary, client, overlay: ConnectionOverlay, ws_state: int) -> void:
	var action := str(result.get("action", "none"))
	if action == "none":
		return
	if action == "failed":
		recovery.mark_failed()
		if overlay != null:
			overlay.show_failed()
		return
	var attempt := int(result.get("attempt", recovery.attempt))
	var max_attempts := int(result.get("max_attempts", 8))
	if overlay != null:
		overlay.show_reconnecting(attempt, max_attempts)
	if action == "connect_ws":
		client.reconnect_ws()
	elif action == "http_resume":
		if client.resume_same_session():
			client.reconnect_ws()
		else:
			recovery.note_http_resume_failed()


func _on_ws_open(client, last_server_tick: int, ready_sent_get: Callable, ready_sent_set: Callable, debug: Callable) -> void:
	recovery.mark_connected_awaiting_snapshot()
	ready_sent_set.call(false)
	client.send("client_ready", last_server_tick, {"client_version": "godot", "last_seen_tick": last_server_tick})
	ready_sent_set.call(true)
	if debug.is_valid():
		debug.call("reconnect websocket open; awaiting session snapshot")


func _poll_messages(client, handle_message: Callable) -> void:
	if not handle_message.is_valid():
		return
	for env in client.poll():
		handle_message.call(env)
