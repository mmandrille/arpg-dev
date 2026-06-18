class_name InteractableRulesLoader
extends RefCounted

static var interactable_rules: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_load_interactable_rules()


static func interactable_definition(def_id: String) -> Dictionary:
	ensure_loaded()
	return interactable_rules.get(def_id, {})


static func has_barrier_when_closed(def_id: String) -> bool:
	return not barrier_size(def_id).is_empty()


static func barrier_size(def_id: String) -> Dictionary:
	var def := interactable_definition(def_id)
	var barrier = def.get("barrier_when_closed", {})
	if typeof(barrier) != TYPE_DICTIONARY:
		return {}
	var size = (barrier as Dictionary).get("size", {})
	if typeof(size) != TYPE_DICTIONARY:
		return {}
	var width := float((size as Dictionary).get("x", 0.0))
	var depth := float((size as Dictionary).get("y", 0.0))
	if width <= 0.0 or depth <= 0.0:
		return {}
	return {"x": width, "y": depth}


static func pick_collider_size(def_id: String) -> Vector3:
	var barrier := barrier_size(def_id)
	if barrier.is_empty():
		return Vector3(1.2, 1.2, 0.45)
	return Vector3(maxf(float(barrier.get("x", 1.2)), 1.2), 1.2, maxf(float(barrier.get("y", 0.45)), 0.45))


static func sync_fog_overlay(fog_overlay, wall_layout: Array, interactable_ids: Array, entities: Dictionary) -> void:
	if fog_overlay == null:
		return
	fog_overlay.set_wall_layout(wall_layout)
	fog_overlay.set_occluder_layout(closed_interactable_occluder_layout(interactable_ids, entities))


static func closed_interactable_occluder_layout(interactable_ids: Array, entities: Dictionary) -> Array:
	var out: Array = []
	for id in interactable_ids:
		var rec: Dictionary = entities.get(id, {})
		if str(rec.get("state", "")) != "closed":
			continue
		var def_id := str(rec.get("interactable_def_id", ""))
		var size := barrier_size(def_id)
		if size.is_empty():
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var pos := node.global_position
		out.append({
			"position": {"x": pos.x, "y": pos.z},
			"size": size,
			"source": "interactable",
			"entity_id": str(id),
			"interactable_def_id": def_id,
		})
	return out


static func _load_interactable_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/interactables.v0.json")
	if not FileAccess.file_exists(path):
		return
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		interactable_rules = parsed.get("interactables", {})
