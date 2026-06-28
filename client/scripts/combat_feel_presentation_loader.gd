## CombatFeelPresentationLoader — static singleton for client-only feel tuning.
class_name CombatFeelPresentationLoader
extends RefCounted

static var _loaded: bool = false
static var _input_buffer_seconds: float = 0.45
static var _command_retarget_grace_seconds: float = 0.22
static var _movement_smoothing: Dictionary = {}
static var _melee_lunge: Dictionary = {}
static var _level_loading_min_display_seconds: float = 0.55
static var _attack_animation: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/combat_feel_presentation.v0.json")
	if not FileAccess.file_exists(path):
		push_warning("CombatFeelPresentationLoader: data file missing: %s" % path)
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("CombatFeelPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("CombatFeelPresentationLoader: malformed JSON: %s" % path)
		return
	_input_buffer_seconds = float(parsed.get("input_buffer_seconds", _input_buffer_seconds))
	_command_retarget_grace_seconds = float(parsed.get("command_retarget_grace_seconds", _command_retarget_grace_seconds))
	var movement_smoothing = parsed.get("movement_smoothing", {})
	if typeof(movement_smoothing) == TYPE_DICTIONARY:
		_movement_smoothing = movement_smoothing
	var melee_lunge = parsed.get("melee_lunge", {})
	if typeof(melee_lunge) == TYPE_DICTIONARY:
		_melee_lunge = melee_lunge
	_level_loading_min_display_seconds = float(parsed.get("level_loading_min_display_seconds", _level_loading_min_display_seconds))
	var attack_animation = parsed.get("attack_animation", {})
	if typeof(attack_animation) == TYPE_DICTIONARY:
		_attack_animation = attack_animation


static func input_buffer_seconds() -> float:
	ensure_loaded()
	return _input_buffer_seconds


static func command_retarget_grace_seconds() -> float:
	ensure_loaded()
	return _command_retarget_grace_seconds


static func movement_smoothing() -> Dictionary:
	ensure_loaded()
	return _movement_smoothing


static func melee_lunge() -> Dictionary:
	ensure_loaded()
	return _melee_lunge


static func level_loading_min_display_seconds() -> float:
	ensure_loaded()
	return _level_loading_min_display_seconds


static func attack_animation() -> Dictionary:
	ensure_loaded()
	return _attack_animation


static func reset_for_tests() -> void:
	_loaded = false
	_input_buffer_seconds = 0.45
	_command_retarget_grace_seconds = 0.22
	_movement_smoothing = {}
	_melee_lunge = {}
	_level_loading_min_display_seconds = 0.55
	_attack_animation = {}
