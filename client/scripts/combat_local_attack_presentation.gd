class_name CombatLocalAttackPresentation
extends RefCounted

const ClientAudioBridgeScript := preload("res://scripts/client_audio_bridge.gd")
const RESULT_EVENTS := ["monster_damaged", "monster_killed", "attack_missed", "attack_blocked"]

var target_id: String = ""


func start(target: String) -> void:
	target_id = target


func active() -> bool:
	return target_id != ""


func clear() -> void:
	target_id = ""


func consume_if_matches(ev: Dictionary, local_player_id: String) -> bool:
	if target_id == "" or local_player_id == "":
		return false
	if str(ev.get("source_entity_id", "")) != local_player_id:
		return false
	if not (str(ev.get("event_type", "")) in RESULT_EVENTS):
		return false
	if _target_for_event(ev) != target_id:
		return false
	clear()
	return true


static func present_local_start(tracker: CombatLocalAttackPresentation, target: String, audio_controller, player_anim, weapon_slot: String = "main_hand", attack_mode: String = "") -> void:
	if tracker != null:
		tracker.start(target)
	ClientAudioBridgeScript.attack(audio_controller)
	_play_animation(player_anim, weapon_slot, attack_mode)


static func present_result(tracker: CombatLocalAttackPresentation, ev: Dictionary, local_player_id: String, audio_controller, player_anim, attack_mode: String = "") -> void:
	if str(ev.get("source_entity_id", "")) != local_player_id:
		return
	if tracker != null and tracker.consume_if_matches(ev, local_player_id):
		return
	ClientAudioBridgeScript.attack(audio_controller)
	_play_animation(player_anim, str(ev.get("weapon_slot", "main_hand")), attack_mode)


static func _play_animation(player_anim, weapon_slot: String = "main_hand", attack_mode: String = "") -> void:
	if player_anim != null:
		player_anim.play_one_shot("attack_off_hand" if weapon_slot == "off_hand" else "attack", attack_mode)


func _target_for_event(ev: Dictionary) -> String:
	var result := str(ev.get("target_entity_id", ""))
	if result == "":
		result = str(ev.get("entity_id", ""))
	return result
