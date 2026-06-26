class_name DungeonRoomPresentationLoader
extends RefCounted

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/dungeon_room_presentation.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("[dungeon_room_presentation] missing config: %s" % path)
		_config = _defaults()
		_loaded = true
		return
	var parsed = JSON.parse_string(file.get_as_text())
	_config = parsed if typeof(parsed) == TYPE_DICTIONARY else _defaults()
	_loaded = true


static func config() -> Dictionary:
	ensure_loaded()
	return _config


static func archetype(name: String) -> Dictionary:
	var archetypes: Dictionary = config().get("room_archetypes", {})
	return archetypes.get(name, {}) as Dictionary


static func treasure_radius() -> float:
	return float(config().get("treasure_radius", 6.0))


static func corridor_max_span() -> float:
	return float(config().get("corridor_max_span", 8.0))


static func ambient_motes() -> Dictionary:
	var value: Dictionary = config().get("ambient_motes", {})
	if value.is_empty():
		return _defaults().get("ambient_motes", {}) as Dictionary
	return value


static func _defaults() -> Dictionary:
	return {
		"version": 0,
		"room_archetypes": {
			"combat": {"floor_tint": "#6b6058", "alpha": 0.14},
			"corridor": {"floor_tint": "#5a6670", "alpha": 0.10},
			"rest": {"floor_tint": "#5f6a58", "alpha": 0.12},
			"treasure": {"floor_tint": "#8a7340", "alpha": 0.20},
		},
		"treasure_radius": 6.0,
		"corridor_max_span": 8.0,
		"ambient_motes": {
			"count_min": 6,
			"count_max": 18,
			"depth_per_level": 2,
			"color": "#c8b8a0",
			"alpha": 0.35,
			"height": 0.55,
			"radius": 0.06,
		},
	}
