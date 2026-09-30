extends RefCounted

## Keeps a remote player's hero model in step with the authoritative entity
## `character_class` (v485). Spawn already builds the right model; this covers a
## later snapshot/delta carrying a different class for an existing record.

const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")


static func class_changed(rec: Dictionary, e: Dictionary) -> bool:
	if str(rec.get("type", "")) != "player" or not e.has("character_class"):
		return false
	return str(e.get("character_class", "")) != str(rec.get("character_class", ""))


## Swaps the model in place via `apply_model(root, class_id)` (main's
## `_apply_character_class_model`), then rebinds the animation and reaction
## controllers, which hold references into the old model.
static func swap_model(rec: Dictionary, class_id: String, base_tint: Color, apply_model: Callable) -> void:
	rec["character_class"] = class_id
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var reaction = rec.get("reaction", null)
	if reaction != null:
		reaction.dispose()
	apply_model.call(node, class_id)
	var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
	rec["controller"] = AnimationControllerScript.new(ap) if ap != null else null
	rec["reaction"] = ModelReactionControllerScript.new(node, base_tint) if reaction != null else null
	rec["base_tint"] = base_tint.to_html(false)


## Sorted entity ids of remote players (bot state).
static func remote_player_ids(entities: Dictionary) -> Array:
	var out: Array = []
	for id in entities.keys():
		if str((entities[id] as Dictionary).get("type", "")) == "player":
			out.append(str(id))
	out.sort()
	return out


## Class id applied to each remote player's hero model, by entity id (bot state).
static func rendered_classes(entities: Dictionary) -> Dictionary:
	var out := {}
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		if str(rec.get("type", "")) == "player" and node != null:
			out[str(id)] = str(node.get("class_id")) if "class_id" in node else ""
	return out
