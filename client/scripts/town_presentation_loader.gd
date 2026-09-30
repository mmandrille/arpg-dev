## TownPresentationLoader — static singleton for town perimeter and night mood tuning.
class_name TownPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/town_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = _default_config()
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("TownPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("TownPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("TownPresentationLoader: malformed JSON: %s" % path)
		return
	_config = _merge_defaults(parsed as Dictionary)


static func night_lighting() -> Dictionary:
	ensure_loaded()
	var night: Dictionary = _config.get("night_lighting", {})
	return {
		"directional_color": str(night.get("directional_color", "#948b7c")),
		"directional_energy": float(night.get("directional_energy", 1.0)),
		"ambient_color": str(night.get("ambient_color", "#393b3e")),
		"ambient_energy": float(night.get("ambient_energy", 0.30)),
	}


static func _default_config() -> Dictionary:
	return {
		"night_lighting": {
			"directional_color": "#948b7c",
			"directional_energy": 1.0,
			"ambient_color": "#393b3e",
			"ambient_energy": 0.30,
		},
	}


## Town layout + v491 dressing (presentation only). Empty dressing when the catalog has none.
static func center() -> Vector2:
	ensure_loaded()
	return _vec2(_config.get("center", {}))


static func gate_position() -> Vector2:
	ensure_loaded()
	return _vec2(_config.get("gate_position", {}))


static func radius_m() -> float:
	ensure_loaded()
	return float(_config.get("radius_m", 0.0))


static func dressing() -> Dictionary:
	ensure_loaded()
	var raw = _config.get("dressing", {})
	return (raw as Dictionary).duplicate(true) if typeof(raw) == TYPE_DICTIONARY else {}


static func _vec2(raw) -> Vector2:
	if typeof(raw) != TYPE_DICTIONARY:
		return Vector2.ZERO
	return Vector2(float((raw as Dictionary).get("x", 0.0)), float((raw as Dictionary).get("y", 0.0)))


static func _merge_defaults(parsed: Dictionary) -> Dictionary:
	var out := parsed.duplicate(true)
	out["night_lighting"] = _default_config()["night_lighting"]
	if typeof(parsed.get("night_lighting", {})) == TYPE_DICTIONARY:
		out["night_lighting"] = (_default_config()["night_lighting"] as Dictionary).merged(parsed["night_lighting"] as Dictionary)
	return out
