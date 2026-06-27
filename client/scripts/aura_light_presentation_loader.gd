## Static loader for aura soft-light presentation tuning.
class_name AuraLightPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/aura_light_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = _default_config()
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("AuraLightPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("AuraLightPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("AuraLightPresentationLoader: malformed JSON: %s" % path)
		return
	_config = _merge_defaults(parsed as Dictionary)


static func config() -> Dictionary:
	ensure_loaded()

	return _config.duplicate(true)


static func aura_entry(aura_id: String) -> Dictionary:
	ensure_loaded()
	var auras: Dictionary = _config.get("auras", {})
	return (auras.get(aura_id, {}) as Dictionary).duplicate(true)


static func priority_list() -> Array:
	ensure_loaded()
	var priority = _config.get("priority", [])
	if priority is Array:
		return priority.duplicate()
	return []


static func presentation_personal_radius() -> float:
	ensure_loaded()

	return float(_config.get("presentation_personal_radius", 1.15))


static func cast_pulse(aura_id: String) -> Dictionary:
	ensure_loaded()
	var pulses: Dictionary = _config.get("cast_pulse", {})
	return (pulses.get(aura_id, {}) as Dictionary).duplicate(true)


static func _default_config() -> Dictionary:
	return {
		"version": 0,
		"priority": [
			"sanctuary",
			"holy_shield",
			"rage",
			"elite_command_radius_preview",
			"elite_command",
		],
		"presentation_personal_radius": 1.15,
		"auras": {},
		"cast_pulse": {},
	}


static func _merge_defaults(raw: Dictionary) -> Dictionary:
	var out := _default_config()
	for key in raw.keys():
		out[key] = raw[key]

	return out
