class_name EquipmentDisplayLoader
extends RefCounted

static var _loaded: bool = false
static var _equipped_multiplier: float = 1.0
static var _ground_multiplier: float = 1.0
static var _rig_native_tint_strength: float = 1.0
static var _ground_pose: Dictionary = {}


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
	_rig_native_tint_strength = clampf(float(data.get("rig_native_rarity_tint_strength", 1.0)), 0.0, 1.0)
	_ground_pose = data.get("ground_pose", {}) if typeof(data.get("ground_pose", {})) == TYPE_DICTIONARY else {}


static func equipped_multiplier() -> float:
	ensure_loaded()
	return _equipped_multiplier


static func rig_native_tint_strength() -> float:
	ensure_loaded()
	return _rig_native_tint_strength


static func ground_multiplier() -> float:
	ensure_loaded()
	return _ground_multiplier


## v487 ground pose of a GLB loot model: `default`, then `rig_native` for kit models, then
## `assets[asset_id]`; later layers override individual keys. Returns height, rotation_degrees
## (Vector3), scale, max_extent (0 = uncapped) and rest_on_floor.
static func ground_pose_for(asset_id: String, rig_native: bool) -> Dictionary:
	ensure_loaded()
	var merged := {"height": 0.0, "rotation_degrees": Vector3.ZERO, "scale": 1.0, "max_extent": 0.0, "rest_on_floor": false}
	var layers: Array = [_ground_pose.get("default", {})]
	if rig_native:
		layers.append(_ground_pose.get("rig_native", {}))
	var assets = _ground_pose.get("assets", {})
	if typeof(assets) == TYPE_DICTIONARY:
		layers.append((assets as Dictionary).get(asset_id, {}))
	for layer in layers:
		if typeof(layer) != TYPE_DICTIONARY:
			continue
		var entry := layer as Dictionary
		if entry.has("height"):
			merged["height"] = float(entry["height"])
		if entry.has("scale"):
			merged["scale"] = float(entry["scale"])
		if entry.has("max_extent"):
			merged["max_extent"] = float(entry["max_extent"])
		if entry.has("rest_on_floor"):
			merged["rest_on_floor"] = bool(entry["rest_on_floor"])
		var rot = entry.get("rotation_degrees", null)
		if typeof(rot) == TYPE_DICTIONARY:
			merged["rotation_degrees"] = Vector3(float(rot.get("x", 0.0)), float(rot.get("y", 0.0)), float(rot.get("z", 0.0)))
	return merged


static func _read_json(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_error("EquipmentDisplayLoader: cannot read %s" % path)
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	return parsed if typeof(parsed) == TYPE_DICTIONARY else {}
