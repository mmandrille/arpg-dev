class_name CombatInputBuffer
extends RefCounted

const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")

var target_id: String = ""
var remaining_seconds: float = 0.0
var queued_count: int = 0
var replaced_count: int = 0
var expired_count: int = 0
var cleared_count: int = 0


func active() -> bool:
	return target_id != "" and remaining_seconds > 0.0


func queue_attack(target: String, duration_seconds: float = -1.0) -> void:
	if duration_seconds <= 0.0:
		duration_seconds = CombatFeelConfigScript.attack_buffer_seconds()
	if target == "" or duration_seconds <= 0.0:
		clear()
		return
	if active() and target_id != target:
		replaced_count += 1
	target_id = target
	remaining_seconds = duration_seconds
	queued_count += 1


func clear() -> void:
	if target_id != "" or remaining_seconds > 0.0:
		cleared_count += 1
	target_id = ""
	remaining_seconds = 0.0


func tick(delta: float) -> void:
	if not active():
		return
	remaining_seconds = maxf(0.0, remaining_seconds - maxf(0.0, delta))
	if remaining_seconds <= 0.0:
		expired_count += 1
		clear()


func should_clear(player_hp: int, entities: Dictionary) -> bool:
	if not active():
		return false
	if player_hp <= 0:
		return true
	if target_id == "" or not entities.has(target_id):
		return true
	var rec: Dictionary = entities[target_id]
	if str(rec.get("type", "")) != "monster":
		return true
	return int(rec.get("hp", 1)) <= 0


func get_debug_state() -> Dictionary:
	return {
		"active": active(),
		"target_id": target_id,
		"remaining_seconds": remaining_seconds,
		"queued_count": queued_count,
		"replaced_count": replaced_count,
		"expired_count": expired_count,
		"cleared_count": cleared_count,
	}
