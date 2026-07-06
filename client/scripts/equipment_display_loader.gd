class_name EquipmentDisplayLoader
extends RefCounted

static var _loaded: bool = false
static var _equipped_multiplier: float = 1.0
static var _ground_multiplier: float = 1.0


static func invalidate() -> void:
	_loaded = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_equipped_multiplier = 1.0
	_ground_multiplier = 1.0
	var root := ProjectSettings.globalize_path("res://")
	var data := _read_json(root.path_join("../shared/assets/equipment_display.v0.json"))
	var multipliers: Dictionary = data.get("glb_mesh_multipliers", {}) if typeof(data.get("glb_mesh_multipliers", {})) == TYPE_DICTIONARY else {}
	_equipped_multiplier = float(multipliers.get("equipped", 1.0))
	_ground_multiplier = float(multipliers.get("ground", 1.0))


static func equipped_multiplier() -> float:
	ensure_loaded()
	return _equipped_multiplier


static func ground_multiplier() -> float:
	ensure_loaded()
	return _ground_multiplier


static func _read_json(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_error("EquipmentDisplayLoader: cannot read %s" % path)
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}
