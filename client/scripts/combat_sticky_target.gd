class_name CombatStickyTarget
extends RefCounted

var target_id: String = ""


func active() -> bool:
	return target_id != ""


func set_target(next_target_id: String) -> void:
	if next_target_id == "":
		clear()
		return
	target_id = next_target_id


func clear() -> void:
	target_id = ""


func should_clear(player_hp: int, entities: Dictionary) -> bool:
	if not active():
		return false
	if player_hp <= 0:
		return true
	if not entities.has(target_id):
		return true
	var rec: Dictionary = entities[target_id]
	if str(rec.get("type", "")) != "monster":
		return true
	return int(rec.get("hp", 1)) <= 0
