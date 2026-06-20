class_name CombatInputBuffer
extends RefCounted

const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")
const DEFAULT_ATTACK_BUFFER_SECONDS := CombatFeelConfigScript.ATTACK_BUFFER_SECONDS

var target_id: String = ""
var remaining_seconds: float = 0.0


func active() -> bool:
	return target_id != "" and remaining_seconds > 0.0


func queue_attack(target: String, duration_seconds: float = DEFAULT_ATTACK_BUFFER_SECONDS) -> void:
	if target == "" or duration_seconds <= 0.0:
		clear()
		return
	target_id = target
	remaining_seconds = duration_seconds


func clear() -> void:
	target_id = ""
	remaining_seconds = 0.0


func tick(delta: float) -> void:
	if not active():
		return
	remaining_seconds = maxf(0.0, remaining_seconds - maxf(0.0, delta))
	if remaining_seconds <= 0.0:
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
