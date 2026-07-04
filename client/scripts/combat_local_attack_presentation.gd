class_name CombatLocalAttackPresentation
extends RefCounted

const ClientAudioBridgeScript := preload("res://scripts/client_audio_bridge.gd")
const AttackAnimationScalingScript := preload("res://scripts/attack_animation_scaling.gd")
const AttackPresentationLoaderScript := preload("res://scripts/attack_presentation_loader.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
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


static func present_local_start(tracker: CombatLocalAttackPresentation, target: String, audio_controller, player_anim, weapon_slot: String = "main_hand", attack_mode: String = "", attack_speed: float = 1.0, inventory: Array = [], equipped: Dictionary = {}) -> void:
	if tracker != null:
		tracker.start(target)
	ClientAudioBridgeScript.attack(audio_controller)
	_play_animation(player_anim, weapon_slot, attack_mode, attack_speed, inventory, equipped)


static func present_result(tracker: CombatLocalAttackPresentation, ev: Dictionary, local_player_id: String, audio_controller, player_anim, attack_mode: String = "", attack_speed: float = 1.0, inventory: Array = [], equipped: Dictionary = {}) -> void:
	if str(ev.get("source_entity_id", "")) != local_player_id:
		return
	if tracker != null and tracker.consume_if_matches(ev, local_player_id):
		return
	ClientAudioBridgeScript.attack(audio_controller)
	_play_animation(player_anim, str(ev.get("weapon_slot", "main_hand")), attack_mode, attack_speed, inventory, equipped)


static func _play_animation(player_anim, weapon_slot: String = "main_hand", attack_mode: String = "", attack_speed: float = 1.0, inventory: Array = [], equipped: Dictionary = {}) -> void:
	if player_anim != null:
		var clip := _attack_clip_for(weapon_slot, attack_mode, inventory, equipped)
		player_anim.play_one_shot(clip, attack_mode, AttackAnimationScalingScript.speed_scale_for(attack_speed))


static func _attack_clip_for(weapon_slot: String, attack_mode: String, inventory: Array, equipped: Dictionary) -> String:
	var item := _equipped_item_for_slot(inventory, equipped, weapon_slot)
	if item.is_empty():
		return "attack_off_hand" if weapon_slot == "off_hand" else "attack"
	var def_id := str(item.get("item_def_id", ""))
	ItemRulesLoaderScript.ensure_loaded()
	var def: Dictionary = ItemRulesLoaderScript.item_definition(def_id)
	if def.is_empty():
		return "attack_off_hand" if weapon_slot == "off_hand" else "attack"
	if attack_mode == "":
		attack_mode = str(def.get("attack_mode", "melee"))
	return AttackPresentationLoaderScript.clip_for_weapon(def, weapon_slot, def_id)


static func _equipped_item_for_slot(inventory: Array, equipped: Dictionary, slot: String) -> Dictionary:
	var raw_id = equipped.get(slot, null)
	if raw_id == null:
		return {}
	var item_id := str(raw_id)
	if item_id == "":
		return {}
	for row in inventory:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		if str((row as Dictionary).get("item_instance_id", "")) == item_id:
			return (row as Dictionary).duplicate(true)
	return {}


func _target_for_event(ev: Dictionary) -> String:
	var result := str(ev.get("target_entity_id", ""))
	if result == "":
		result = str(ev.get("entity_id", ""))
	return result
