class_name DungeonSurfaceDetailLoader
extends RefCounted

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/dungeon_surface_detail_presentation.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("[dungeon_surface_detail] missing config: %s" % path)
		_config = _defaults()
		_loaded = true
		return
	var parsed = JSON.parse_string(file.get_as_text())
	_config = parsed if typeof(parsed) == TYPE_DICTIONARY else _defaults()
	_loaded = true


static func config() -> Dictionary:
	ensure_loaded()
	return _config


static func floor_detail_y() -> float:
	return float(config().get("floor_detail_y", 0.024))


static func wall_detail_inset() -> float:
	return float(config().get("wall_detail_inset", 0.06))


static func wall_detail_height_ratio() -> float:
	return float(config().get("wall_detail_height_ratio", 0.58))


static func corridor_large_detail_every() -> int:
	return int(config().get("corridor_large_detail_every", 3))


static func room_large_detail_every() -> int:
	return int(config().get("room_large_detail_every", 2))


static func floor_motifs() -> Dictionary:
	return config().get("floor_motifs", {}) as Dictionary


static func wall_motifs() -> Dictionary:
	return config().get("wall_motifs", {}) as Dictionary


static func _defaults() -> Dictionary:
	return {
		"version": 0,
		"floor_detail_y": 0.024,
		"wall_detail_inset": 0.06,
		"wall_detail_height_ratio": 0.58,
		"corridor_large_detail_every": 3,
		"room_large_detail_every": 2,
		"floor_motifs": {
			"ritual_square": {
				"color": "#d2c0a0",
				"alpha": 0.2,
				"scale_min": 0.18,
				"scale_max": 0.42,
				"band_width": 0.12,
			},
			"scratched_cross": {
				"color": "#8c7a66",
				"alpha": 0.18,
				"scale_min": 0.16,
				"scale_max": 0.36,
				"band_width": 0.08,
			},
			"broken_tile": {
				"color": "#564d45",
				"alpha": 0.16,
				"scale_min": 0.2,
				"scale_max": 0.48,
				"band_width": 0.18,
			},
		},
		"wall_motifs": {
			"plaque": {
				"color": "#73624d",
				"alpha": 0.9,
				"width_ratio": 0.34,
				"height_ratio": 0.2,
				"depth": 0.08,
			},
			"crack_cluster": {
				"color": "#2a2522",
				"alpha": 0.75,
				"width_ratio": 0.22,
				"height_ratio": 0.3,
				"depth": 0.04,
			},
			"stone_inset": {
				"color": "#8a7963",
				"alpha": 0.86,
				"width_ratio": 0.42,
				"height_ratio": 0.18,
				"depth": 0.06,
			},
		},
	}
