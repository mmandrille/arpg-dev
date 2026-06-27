## Static loader for dungeon torch presentation tuning.
class_name DungeonTorchPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/dungeon_torch_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = _default_config()
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("DungeonTorchPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("DungeonTorchPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("DungeonTorchPresentationLoader: malformed JSON: %s" % path)
		return
	_config = _merge_defaults(parsed as Dictionary)


static func config() -> Dictionary:
	ensure_loaded()

	return _config.duplicate(true)


static func _default_config() -> Dictionary:
	return {
		"enabled": true,
		"wall_segment_tiles": 10.0,
		"torches_per_segment_min": 0,
		"torches_per_segment_max": 2,
		"wall_inset": 0.55,
		"wall_end_margin": 0.75,
		"max_shader_torches": 32,
		"mount_height_fraction": 0.72,
		"fog_light_radius": 5.0,
		"torch_feather_world": 0.35,
		"omni_light_enabled": false,
		"omni_range": 5.0,
		"omni_energy": 0.85,
		"light_color": "#ff9b3d",
		"flame_color": "#ff7a1a",
		"flame_emission_color": "#ffd45a",
		"flame_emission_energy": 3.2,
	}


static func _merge_defaults(raw: Dictionary) -> Dictionary:
	var out := _default_config()
	for key in raw.keys():
		out[key] = raw[key]

	return out
