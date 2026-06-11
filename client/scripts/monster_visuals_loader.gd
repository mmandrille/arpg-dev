class_name MonsterVisualsLoader
extends RefCounted

static var _loaded: bool = false
static var _visuals: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_visuals = {}
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/monster_visuals.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("monster visuals missing: %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("monster visuals malformed: %s" % path)
		return
	var entries = parsed.get("monster_visuals", {})
	if typeof(entries) == TYPE_DICTIONARY:
		_visuals = entries


static func resolve(monster_def_id: String, visual_model: String = "") -> Dictionary:
	ensure_loaded()
	if visual_model in ["monster_dummy", "monster_quadruped", "monster_tiny_flyer"]:
		return _entry_for_scene(visual_model)
	if _visuals.has(monster_def_id):
		var entry := (_visuals[monster_def_id] as Dictionary).duplicate()
		entry["visual_model"] = str(entry.get("scene", "monster_dummy"))
		return entry
	return _entry_for_scene("monster_dummy")


static func scene_for_monster(monster_def_id: String, visual_model: String = "") -> String:
	return str(resolve(monster_def_id, visual_model).get("scene", "monster_dummy"))


static func _entry_for_scene(scene: String) -> Dictionary:
	for key in _visuals.keys():
		var entry := _visuals[key] as Dictionary
		if str(entry.get("scene", "")) == scene:
			var out := entry.duplicate()
			out["visual_model"] = scene
			return out
	return {
		"asset_id": "monster_dummy_v0",
		"scene": "monster_dummy",
		"scale": 1.0,
		"height_offset": 0.0,
		"animation_profile": "ground_biped",
		"visual_model": "monster_dummy",
	}
