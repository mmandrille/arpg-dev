## WallOcclusionPresentationLoader — static singleton for wall fade tuning data.
class_name WallOcclusionPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/wall_occlusion_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = _default_config()
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("WallOcclusionPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("WallOcclusionPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("WallOcclusionPresentationLoader: malformed JSON: %s" % path)
		return
	_config = _merge_defaults(parsed as Dictionary)


static func config() -> Dictionary:
	ensure_loaded()

	return _config.duplicate(true)


static func faded_alpha() -> float:
	return float(config().get("faded_alpha", 0.34))


static func opaque_alpha() -> float:
	return float(config().get("opaque_alpha", 1.0))


static func segment_inflate() -> float:
	return float(config().get("segment_inflate", 0.02))


static func move_epsilon() -> float:
	return float(config().get("move_epsilon", 0.04))


static func min_rebuild_interval_frames() -> int:
	return int(config().get("min_rebuild_interval_frames", 2))


static func _default_config() -> Dictionary:
	return {
		"version": 0,
		"faded_alpha": 0.34,
		"opaque_alpha": 1.0,
		"segment_inflate": 0.02,
		"move_epsilon": 0.04,
		"min_rebuild_interval_frames": 2,
	}


static func _merge_defaults(parsed: Dictionary) -> Dictionary:
	var out := _default_config()
	for key in out.keys():
		if parsed.has(key):
			out[key] = parsed[key]

	return out
