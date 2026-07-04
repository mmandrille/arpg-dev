class_name AttackPresentationLoader
extends RefCounted

static var _loaded: bool = false
static var _clips: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_clips = {}
	var root := ProjectSettings.globalize_path("res://")
	var data := _read_json(root.path_join("../shared/assets/attack_presentation.v0.json"))
	var clips = data.get("clips", {})
	if typeof(clips) == TYPE_DICTIONARY:
		_clips = clips


static func clip_for_weapon(def: Dictionary, weapon_slot: String, item_def_id: String = "") -> String:
	ensure_loaded()
	var attack_mode := str(def.get("attack_mode", "melee"))
	var handedness := str(def.get("handedness", "one_handed"))
	if attack_mode == "ranged":
		if item_def_id.find("staff") >= 0:
			return str(_clips.get("ranged_staff", "attack_staff"))
		return str(_clips.get("ranged_bow", "attack_ranged"))
	if handedness == "two_handed":
		return str(_clips.get("melee_two_handed", "attack_2h"))
	if weapon_slot == "off_hand":
		return str(_clips.get("melee_one_handed_off", "attack_off_hand"))
	return str(_clips.get("melee_one_handed_main", "attack"))


static func _read_json(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}
