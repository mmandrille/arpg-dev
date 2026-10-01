class_name BotTransportDelay
extends RefCounted

# A client-bot-only, ordered transport fixture. The fixed sequence gives
# repeatable jitter without changing any wire envelope or simulation rule.
const PROFILE_LOCAL := "local"
const PROFILE_BOUNDED := "bounded_80_20"
const BASE_MS := 80
const JITTER_MS := 20
const OFFSETS := [-20, 0, 20, -10, 10]

var profile: String = PROFILE_LOCAL
var _outbound: Array[Dictionary] = []
var _inbound: Array[Dictionary] = []
var _outbound_count: int = 0
var _inbound_count: int = 0
var _outbound_last_due: int = 0
var _inbound_last_due: int = 0


func configure(next_profile: String) -> bool:
	if next_profile not in [PROFILE_LOCAL, PROFILE_BOUNDED]:
		return false
	profile = next_profile
	clear()
	return true


func enabled() -> bool:
	return profile == PROFILE_BOUNDED


func clear() -> void:
	_outbound.clear()
	_inbound.clear()
	_outbound_count = 0
	_inbound_count = 0
	_outbound_last_due = 0
	_inbound_last_due = 0


func queue_outbound(value: Variant, now_ms: int) -> void:
	_outbound_count += 1
	_outbound_last_due = _queue(_outbound, value, now_ms, _outbound_count, _outbound_last_due)


func queue_inbound(value: Variant, now_ms: int) -> void:
	_inbound_count += 1
	_inbound_last_due = _queue(_inbound, value, now_ms, _inbound_count, _inbound_last_due)


func release_outbound(now_ms: int) -> Array:
	return _release(_outbound, now_ms)


func release_inbound(now_ms: int) -> Array:
	return _release(_inbound, now_ms)


func _queue(entries: Array[Dictionary], value: Variant, now_ms: int, count: int, last_due: int) -> int:
	var offset: int = OFFSETS[(count - 1) % OFFSETS.size()]
	var due: int = maxi(now_ms + BASE_MS + offset, last_due)
	entries.append({"due": due, "value": value})
	return due


func _release(entries: Array[Dictionary], now_ms: int) -> Array:
	var ready: Array = []
	while not entries.is_empty() and int(entries[0]["due"]) <= now_ms:
		ready.append(entries.pop_front()["value"])
	return ready
