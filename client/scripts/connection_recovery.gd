## ConnectionRecovery — client-side WebSocket reconnect backoff state machine.
class_name ConnectionRecovery
extends RefCounted

const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

enum Phase { IDLE, WAITING, CONNECTING, AWAITING_SNAPSHOT, FAILED }

var phase: int = Phase.IDLE
var attempt: int = 0
var wait_remaining_s: float = 0.0
var use_http_resume: bool = false
var elapsed_s: float = 0.0
var ws_failures_since_http: int = 0


func begin() -> void:
	phase = Phase.WAITING
	attempt = 0
	wait_remaining_s = 0.0
	use_http_resume = false
	elapsed_s = 0.0
	ws_failures_since_http = 0


func cancel() -> void:
	phase = Phase.IDLE
	attempt = 0
	wait_remaining_s = 0.0
	use_http_resume = false
	elapsed_s = 0.0
	ws_failures_since_http = 0


func mark_connected_awaiting_snapshot() -> void:
	if phase == Phase.IDLE or phase == Phase.FAILED:
		return
	phase = Phase.AWAITING_SNAPSHOT


func mark_resynced() -> void:
	cancel()


func mark_failed() -> void:
	phase = Phase.FAILED


func is_active() -> bool:
	return phase == Phase.WAITING or phase == Phase.CONNECTING or phase == Phase.AWAITING_SNAPSHOT


func blocks_input() -> bool:
	return is_active()


func show_cancel() -> bool:
	return attempt >= 1 and phase != Phase.FAILED


func tick(delta: float) -> Dictionary:
	if phase == Phase.IDLE or phase == Phase.FAILED:
		return {"action": "none"}

	elapsed_s += delta
	var rules := _rules()
	var give_up_s := float(rules.get("give_up_s", 45.0))
	var max_attempts := int(rules.get("max_attempts", 8))

	if elapsed_s >= give_up_s or attempt >= max_attempts:
		phase = Phase.FAILED
		return {"action": "failed", "reason": "give_up"}

	if phase == Phase.AWAITING_SNAPSHOT:
		return {"action": "none"}

	if phase == Phase.WAITING:
		wait_remaining_s -= delta
		if wait_remaining_s > 0.0:
			return {"action": "none"}
		phase = Phase.CONNECTING
		attempt += 1
		var http_after := int(rules.get("http_resume_after_ws_failures", 2))
		if ws_failures_since_http >= http_after:
			use_http_resume = true
		return {
			"action": "http_resume" if use_http_resume else "connect_ws",
			"attempt": attempt,
			"max_attempts": max_attempts,
		}

	if phase == Phase.CONNECTING:
		return {"action": "none"}

	return {"action": "none"}


func note_ws_closed() -> void:
	note_transport_lost()


func note_transport_lost() -> void:
	if phase == Phase.IDLE or phase == Phase.FAILED:
		return
	ws_failures_since_http += 1
	_schedule_wait()


func note_http_resume_failed() -> void:
	if phase == Phase.IDLE or phase == Phase.FAILED:
		return
	_schedule_wait()


func note_connect_started() -> void:
	if phase != Phase.CONNECTING:
		return
	# CONNECTING phase waits for main to observe OPEN/CLOSED via poll.


static func delay_for_attempt(attempt_number: int) -> float:
	var rules := _rules()
	var initial := float(rules.get("initial_delay_s", 0.5))
	var max_delay := float(rules.get("max_delay_s", 8.0))
	var exponent := int(rules.get("backoff_exponent", 2))
	var scaled := initial * pow(float(exponent), float(max(0, attempt_number - 1)))
	return minf(max_delay, scaled)


func _schedule_wait() -> void:
	phase = Phase.WAITING
	wait_remaining_s = delay_for_attempt(attempt)


static func _rules() -> Dictionary:
	MainConfigLoaderScript.ensure_loaded()
	return MainConfigLoaderScript.client_reconnect_rules()
