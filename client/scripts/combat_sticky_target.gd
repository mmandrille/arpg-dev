class_name CombatStickyTarget
extends RefCounted

var target_id: String = ""
var set_count: int = 0
var replaced_count: int = 0
var cleared_count: int = 0


func active() -> bool:
	return target_id != ""


func set_target(next_target_id: String) -> void:
	if next_target_id == "":
		clear()
		return
	if target_id != "" and target_id != next_target_id:
		replaced_count += 1
	target_id = next_target_id
	set_count += 1


func clear() -> void:
	if target_id != "":
		cleared_count += 1
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


func get_debug_state() -> Dictionary:
	return {
		"active": active(),
		"target_id": target_id,
		"set_count": set_count,
		"replaced_count": replaced_count,
		"cleared_count": cleared_count,
	}
