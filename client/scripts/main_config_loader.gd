## MainConfigLoader — lazy singleton for shared/rules/main_config.v0.json gameplay values.
class_name MainConfigLoader
extends RefCounted

static var gameplay: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/main_config.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		gameplay = (parsed as Dictionary).get("gameplay", {})


static func base_movement_speed() -> float:
	ensure_loaded()
	return float(gameplay.get("base_movement_speed", 1.0))


static func movement_acceleration_seconds() -> float:
	ensure_loaded()
	return float(gameplay.get("movement_acceleration_seconds", 2.0))


static func movement_min_speed_factor() -> float:
	ensure_loaded()
	return float(gameplay.get("movement_min_speed_factor", 0.2))


static func _read_json(path: String):
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("[main-config] cannot open %s" % path)
		return null
	return JSON.parse_string(file.get_as_text())
