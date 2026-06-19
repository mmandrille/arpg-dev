class_name ClassPresentationsLoader
extends RefCounted

const FALLBACK_ASSET_ID := "character_base_humanoid_v0"

static var _loaded: bool = false
static var _classes: Dictionary = {}
static var _manifest_assets: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_classes = {}
	_manifest_assets = {}
	var root := ProjectSettings.globalize_path("res://")
	var shared_root := root.path_join("../shared")
	var assets_root := root.path_join("../assets")
	var presentations := _read_json(shared_root.path_join("assets/class_presentations.v0.json"))
	var manifest := _read_json(assets_root.path_join("manifests/assets.v0.json"))
	var entries = presentations.get("classes", {})
	if typeof(entries) == TYPE_DICTIONARY:
		_classes = entries
	var assets = manifest.get("assets", {})
	if typeof(assets) == TYPE_DICTIONARY:
		_manifest_assets = assets


static func resolve(class_id: String) -> Dictionary:
	ensure_loaded()
	var entry: Dictionary = _classes.get(class_id, {})
	var model: Dictionary = entry.get("model", {}) if typeof(entry.get("model", {})) == TYPE_DICTIONARY else {}
	var asset_id := str(model.get("asset_id", FALLBACK_ASSET_ID))
	var asset: Dictionary = _manifest_assets.get(asset_id, {})
	if str(asset.get("type", "")) != "character":
		asset_id = FALLBACK_ASSET_ID
		asset = _manifest_assets.get(asset_id, {})
	var runtime_path := str(asset.get("runtime_path", "client/assets/characters/base_humanoid/base_humanoid.glb"))
	return {
		"class_id": class_id,
		"asset_id": asset_id,
		"runtime_path": runtime_path,
		"scene_path": _res_path(runtime_path),
		"scale": _positive_float(model.get("scale", 1.0), 1.0),
		"height_offset": float(model.get("height_offset", 0.0)),
	}


static func packed_scene_for_class(class_id: String) -> PackedScene:
	var resolved := resolve(class_id)
	var scene_path := str(resolved.get("scene_path", ""))
	if scene_path != "" and ResourceLoader.exists(scene_path):
		var packed := load(scene_path) as PackedScene
		if packed != null:
			return packed
	var fallback_path := _res_path("client/assets/characters/base_humanoid/base_humanoid.glb")
	return load(fallback_path) as PackedScene


static func _read_json(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_warning("class presentation data missing: %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("class presentation data malformed: %s" % path)
		return {}
	return parsed


static func _res_path(runtime_path: String) -> String:
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p


static func _positive_float(value, fallback: float) -> float:
	var parsed := float(value)
	if parsed <= 0.0:
		return fallback
	return parsed
