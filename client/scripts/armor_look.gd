## Armor look on kit heroes (ADR-0018 D5 / P3b, v483). Data: shared/assets/armor_look.v0.json.
##
## Equipped armor recolours the hero's own body-part meshes instead of mounting meshes:
## chest/belt -> *_Body, gloves -> arms, boots -> legs, head -> shows the class headgear.
## The colour goes in the material DETAIL layer (a 1x1 texture, Mix blend, alpha = strength), which
## lerps the coloured kit atlas toward the armor colour. albedo_color stays owned by the white base
## tint, rarity tint and ModelReactionController (hit flash / death darken); both edit the mesh's
## current material_override in place, so neither erases the other.
class_name ArmorLook
extends RefCounted

const DEFAULT_PATH := "../shared/assets/armor_look.v0.json"
const MODE_TINT := "tint"
const MODE_HEADGEAR := "headgear"
const MODE_NONE := "none"

static var _loaded: bool = false
static var _config: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("ArmorLook: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_config = parsed as Dictionary
	else:
		push_warning("ArmorLook: malformed JSON: %s" % path)


## "tint" / "headgear" for armor slots, "none" for UI-only slots, "" for mesh slots (weapons).
static func slot_mode(slot: String) -> String:
	ensure_loaded()
	var entry = (_config.get("slots", {}) as Dictionary).get(slot, null)
	if typeof(entry) == TYPE_DICTIONARY:
		return str((entry as Dictionary).get("mode", ""))
	if (_config.get("no_world_visual_slots", []) as Array).has(slot):
		return MODE_NONE
	return ""


## {color: Color, strength: float} for an armor item, or {} when the catalog has no entry.
static func item_look(item_def_id: String) -> Dictionary:
	ensure_loaded()
	var entry = (_config.get("items", {}) as Dictionary).get(item_def_id, null)
	if typeof(entry) != TYPE_DICTIONARY:
		return {}
	var look := entry as Dictionary
	return {
		"color": Color(str(look.get("color", "#ffffff"))),
		"strength": clampf(float(look.get("strength", _config.get("default_strength", 0.6))), 0.0, 1.0),
	}


## Recompute the whole armor look from `equipped` (slot -> item_def_id; empty/missing = no item).
## Idempotent: regions without an item are cleared and headgear hides when the head slot is empty.
static func apply(model_root: Node, equipped: Dictionary) -> Dictionary:
	ensure_loaded()
	var winners := _region_winners(equipped)
	var head_def := str(equipped.get("head", ""))
	var state := {"regions": {}, "headgear_visible": head_def != "", "headgear_meshes": 0}
	if model_root == null:
		return state
	for node in model_root.find_children("*", "MeshInstance3D", true, false):
		var mesh := node as MeshInstance3D
		var region := _region_for_mesh(str(mesh.name))
		if region == "":
			continue
		var winner: Dictionary = winners.get(region, {})
		if _region_is_headgear(region):
			state["headgear_meshes"] = int(state["headgear_meshes"]) + 1
			mesh.visible = head_def != ""
		if winner.is_empty():
			ModelDetailTint.clear_detail(mesh)
			continue
		ModelDetailTint.set_detail(mesh, winner["color"], float(winner["strength"]))
		(state["regions"] as Dictionary)[region] = {
			"slot": winner["slot"],
			"item_def_id": winner["item_def_id"],
			"color": (winner["color"] as Color).to_html(false),
		}
	return state


## Region -> {slot, item_def_id, color, strength}; body_precedence decides shared regions.
static func _region_winners(equipped: Dictionary) -> Dictionary:
	var slots: Dictionary = _config.get("slots", {})
	var ordered: Array = (_config.get("body_precedence", []) as Array).duplicate()
	for slot in slots.keys():
		if not ordered.has(slot):
			ordered.append(slot)
	var winners := {}
	for slot in ordered:
		var def_id := str(equipped.get(str(slot), ""))
		if def_id == "" or not slots.has(slot):
			continue
		var region := str((slots[slot] as Dictionary).get("region", ""))
		if region == "" or winners.has(region):
			continue
		var look := item_look(def_id)
		if look.is_empty():
			continue
		winners[region] = {"slot": str(slot), "item_def_id": def_id, "color": look["color"], "strength": look["strength"]}
	return winners


static func _region_for_mesh(mesh_name: String) -> String:
	var regions: Dictionary = _config.get("regions", {})
	for region in regions.keys():
		for suffix in regions[region]:
			if mesh_name.ends_with(str(suffix)):
				return str(region)
	return ""


static func _region_is_headgear(region: String) -> bool:
	for entry in (_config.get("slots", {}) as Dictionary).values():
		if str((entry as Dictionary).get("region", "")) == region and str((entry as Dictionary).get("mode", "")) == MODE_HEADGEAR:
			return true
	return false
