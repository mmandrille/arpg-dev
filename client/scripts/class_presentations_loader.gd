class_name ClassPresentationsLoader
extends RefCounted

## Entities with no or an unknown class render the `fallback_class` kit hero (ADR-0018 P3c).
static var _loaded: bool = false
static var _classes: Dictionary = {}
static var _fallback_class: String = ""
static var _manifest_assets: Dictionary = {}


static func invalidate() -> void:
	_loaded = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_classes = {}
	_fallback_class = ""
	_manifest_assets = {}
	var root := ProjectSettings.globalize_path("res://")
	var shared_root := root.path_join("../shared")
	var assets_root := root.path_join("../assets")
	var presentations := _read_json(shared_root.path_join("assets/class_presentations.v0.json"))
	var manifest := _read_json(assets_root.path_join("manifests/assets.v0.json"))
	var entries = presentations.get("classes", {})
	if typeof(entries) == TYPE_DICTIONARY:
		_classes = entries
	_fallback_class = str(presentations.get("fallback_class", ""))
	var assets = manifest.get("assets", {})
	if typeof(assets) == TYPE_DICTIONARY:
		_manifest_assets = assets


static func resolve(class_id: String) -> Dictionary:
	ensure_loaded()
	var entry := _presented_entry(class_id)
	var model: Dictionary = entry.get("model", {}) if typeof(entry.get("model", {})) == TYPE_DICTIONARY else {}
	var asset_id := str(model.get("asset_id", ""))
	var asset: Dictionary = _manifest_assets.get(asset_id, {})
	if str(asset.get("type", "")) != "character":
		entry = _classes.get(_fallback_class, {})
		model = entry.get("model", {}) if typeof(entry.get("model", {})) == TYPE_DICTIONARY else {}
		asset_id = str(model.get("asset_id", ""))
		asset = _manifest_assets.get(asset_id, {})
	var runtime_path := str(asset.get("runtime_path", ""))
	return {
		"class_id": class_id,
		"asset_id": asset_id,
		"runtime_path": runtime_path,
		"scene_path": _res_path(runtime_path) if runtime_path != "" else "",
		"scale": _positive_float(model.get("scale", 1.0), 1.0),
		"height_offset": float(model.get("height_offset", 0.0)),
		"clip_profile": str(model.get("clip_profile", "")),
		"idle_stance": entry.get("idle_stance", {}) if typeof(entry.get("idle_stance", {})) == TYPE_DICTIONARY else {},
	}


## The class's own entry, or the fallback class entry when the class is empty or unknown.
static func _presented_entry(class_id: String) -> Dictionary:
	var entry = _classes.get(class_id, null)
	if typeof(entry) == TYPE_DICTIONARY:
		return entry as Dictionary
	return _classes.get(_fallback_class, {})


static func idle_stance_for_class(class_id: String) -> Dictionary:
	var entry: Dictionary = _classes.get(class_id, {})
	if typeof(entry.get("idle_stance", {})) == TYPE_DICTIONARY:
		return (entry.get("idle_stance", {}) as Dictionary).duplicate(true)
	return {}


static func packed_scene_for_class(class_id: String) -> PackedScene:
	var resolved := resolve(class_id)
	var scene_path := str(resolved.get("scene_path", ""))
	if scene_path != "" and ResourceLoader.exists(scene_path):
		var packed := load(scene_path) as PackedScene
		if packed != null:
			return packed
	push_warning("class presentation model missing for %s: %s" % [class_id, scene_path])
	return null


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
