class_name SurfaceMaterialLoader
extends RefCounted

const STYLE_TOWN_GROUND := "town_ground"
const STYLE_DUNGEON_FLOOR := "dungeon_floor"
const STYLE_DUNGEON_WALL := "dungeon_wall"
const STYLE_DUNGEON_COLUMN := "dungeon_column"
const STYLE_WATER := "water"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/surface_material_presentation.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("[surface_material] missing config: %s" % path)
		_config = _defaults()
		_loaded = true
		return
	var parsed = JSON.parse_string(file.get_as_text())
	_config = parsed if typeof(parsed) == TYPE_DICTIONARY else _defaults()
	_loaded = true


static func config() -> Dictionary:
	ensure_loaded()
	return _config


static func material(style_id: String) -> Dictionary:
	var materials: Dictionary = config().get("materials", {}) as Dictionary
	var style: Dictionary = materials.get(style_id, {}) as Dictionary
	if style.is_empty():
		style = (_defaults().get("materials", {}) as Dictionary).get(style_id, {}) as Dictionary
	return style


static func color(style_id: String, key: String, fallback: Color) -> Color:
	var style := material(style_id)
	if not style.has(key):
		return fallback
	return Color(str(style.get(key, "#" + fallback.to_html(false))))


static func scalar(style_id: String, key: String, fallback: float) -> float:
	return float(material(style_id).get(key, fallback))


static func integer(style_id: String, key: String, fallback: int) -> int:
	return int(material(style_id).get(key, fallback))


static func uv_scale(style_id: String, fallback: Vector2) -> Vector2:
	var raw = material(style_id).get("uv_scale", [])
	if typeof(raw) == TYPE_ARRAY and raw.size() >= 2:
		return Vector2(float(raw[0]), float(raw[1]))
	return fallback


static func style_for_ground_level(level: int) -> String:
	return STYLE_TOWN_GROUND if level == 0 else STYLE_DUNGEON_FLOOR


static func _defaults() -> Dictionary:
	return {
		"version": 0,
		"materials": {
			STYLE_TOWN_GROUND: {
				"low": "#2f6136",
				"high": "#79aa58",
				"accent": "#b7b56c",
				"dark": "#274c2b",
				"soil": "#8f7447",
				"soil_dark": "#6f5f39",
				"contrast": 1.0,
				"uv_scale": [28.0, 18.0],
				"roughness": 0.92,
				"normal_scale": 0.0,
			},
			STYLE_DUNGEON_FLOOR: {
				"low": "#3c3f43",
				"high": "#73706b",
				"accent": "#a09a8e",
				"dark": "#25282c",
				"soil": "#4b4842",
				"soil_dark": "#302d2a",
				"contrast": 1.08,
				"uv_scale": [28.0, 18.0],
				"roughness": 0.88,
				"normal_scale": 0.18,
			},
			STYLE_DUNGEON_WALL: {
				"low": "#34363a",
				"high": "#6b6255",
				"accent": "#8a8173",
				"dark": "#17191c",
				"soil": "#4d4740",
				"soil_dark": "#26231f",
				"contrast": 1.12,
				"uv_scale": [2.0, 2.0],
				"roughness": 0.94,
				"normal_scale": 0.24,
			},
			STYLE_DUNGEON_COLUMN: {
				"low": "#4c4942",
				"high": "#8d846f",
				"accent": "#b2aa94",
				"dark": "#24231f",
				"soil": "#5d5549",
				"soil_dark": "#38342d",
				"contrast": 1.16,
				"uv_scale": [2.0, 2.0],
				"roughness": 0.98,
				"normal_scale": 0.22,
			},
			STYLE_WATER: {
				"low": "#1f5f7a",
				"high": "#5fa8bb",
				"accent": "#c2eef1",
				"dark": "#153949",
				"soil": "#2c748d",
				"soil_dark": "#194b60",
				"contrast": 1.1,
				"uv_scale": [3.0, 3.0],
				"roughness": 0.7,
				"normal_scale": 0.0,
				"overlay_alpha": 0.28,
				"overlay_y": 0.026,
				"overlay_bands": 5,
			},
		},
	}
