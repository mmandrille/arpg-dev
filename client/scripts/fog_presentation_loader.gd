## FogPresentationLoader — static singleton for fog compositor tuning data.
class_name FogPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/fog_presentation.v0.json"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = _default_config()
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	if not FileAccess.file_exists(path):
		push_warning("FogPresentationLoader: data file missing: %s" % path)
		return
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("FogPresentationLoader: could not open: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("FogPresentationLoader: malformed JSON: %s" % path)
		return
	_config = _merge_defaults(parsed as Dictionary)


static func config() -> Dictionary:
	ensure_loaded()

	return _config.duplicate(true)


static func shadow_reach_multiplier() -> float:
	return float(config().get("shadow_reach_multiplier", 1.25))


static func falloff_power() -> float:
	return float(config().get("falloff_power", 2.0))


static func edge_feather_world() -> float:
	return float(config().get("edge_feather_world", 1.15))


static func darkness_alpha() -> float:
	return float(config().get("darkness_alpha", 1.0))


static func perspective_ambient_suppression() -> Dictionary:
	var value: Dictionary = config().get("perspective_ambient_suppression", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {"directional_scale": 0.02, "ambient_scale": 0.02}

	return value


static func ambient_suppression() -> Dictionary:
	var value: Dictionary = config().get("ambient_suppression", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {"directional_scale": 0.35, "ambient_scale": 0.12}

	return value


static func organic_edge() -> Dictionary:
	var value: Dictionary = config().get("organic_edge", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {}

	return value


static func shadow_cache() -> Dictionary:
	var value: Dictionary = config().get("shadow_cache", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {
			"move_epsilon": 0.006,
			"viewport_size_epsilon_px": 1.0,
			"performance_min_rebuild_interval_frames": 3,
		}

	return value


static func shadow() -> Dictionary:
	var value: Dictionary = config().get("shadow", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {}

	return value


static func point_light() -> Dictionary:
	var value: Dictionary = config().get("point_light", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {}

	return value


static func perspective() -> Dictionary:
	var value: Dictionary = config().get("perspective", {})
	if typeof(value) != TYPE_DICTIONARY:
		return {
			"sample_heights": [0.0, 1.25, 2.5, 3.75],
			"height_sample_max_ground_scale": 2.0,
		}

	return value


static func height_sample_max_ground_scale() -> float:
	return maxf(1.0, float(perspective().get("height_sample_max_ground_scale", 2.0)))


static func reset_for_tests() -> void:
	_loaded = false
	_config = {}


static func _default_config() -> Dictionary:
	return {
		"version": 0,
		"falloff_power": 2.0,
		"edge_feather_world": 1.15,
		"shadow_reach_multiplier": 1.25,
		"darkness_alpha": 1.0,
		"ambient_suppression": {"directional_scale": 0.35, "ambient_scale": 0.12},
		"perspective_ambient_suppression": {"directional_scale": 0.0, "ambient_scale": 0.0},
		"shadow_cache": {
			"move_epsilon": 0.006,
			"viewport_size_epsilon_px": 1.0,
			"performance_min_rebuild_interval_frames": 3,
		},
		"organic_edge": {
			"world_amplitude": 0.65,
			"segments": 18.0,
			"seed": 41.0,
			"rotation_cycles_per_second": 0.12,
			"rotation_move_epsilon": 0.006,
			"enabled_isometric": true,
			"enabled_perspective": false,
		},
		"shadow": {
			"edge_epsilon": 0.08,
			"start_offset": 0.16,
			"wall_height": 1.0,
			"gloom_color": "#1a1c21",
			"gloom_alpha": 0.42,
			"core_color": "#000000",
			"core_alpha": 0.82,
			"gloom_scale": 1.035,
		},
		"perspective": {
			"sample_heights": [0.0, 1.25, 2.5, 3.75],
			"height_sample_max_ground_scale": 2.0,
		},
		"point_light": {
			"energy": 3.0,
			"attenuation": 2.0,
			"color": "#ffffff",
			"range_multiplier": 1.0,
			"height_fraction": 0.55,
			"min_height": 0.75,
			"height_offset": 2.0,
			"shadow_enabled": true,
			"shadow_bias": 0.08,
			"shadow_normal_bias": 1.2,
		},
	}


static func _merge_defaults(parsed: Dictionary) -> Dictionary:
	var merged := _default_config()
	for key in parsed.keys():
		if typeof(parsed[key]) == TYPE_DICTIONARY and typeof(merged.get(key, null)) == TYPE_DICTIONARY:
			var nested: Dictionary = merged[key].duplicate(true)
			for nested_key in (parsed[key] as Dictionary).keys():
				nested[nested_key] = parsed[key][nested_key]
			merged[key] = nested
		else:
			merged[key] = parsed[key]

	return merged
