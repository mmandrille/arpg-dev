## Static loader for shared/assets/item_visuals.v0.json (agent rule 7): the canonical
## item_def_id -> asset_id + slot + socket link. EquipmentVisuals (equipped models) and
## LootNodeFactory (v487 ground models) both read it here, so hand items show one model everywhere.
class_name ItemVisualsLoader
extends RefCounted

const HAND_SLOTS := ["main_hand", "off_hand"]

static var _loaded: bool = false
static var _visuals: Dictionary = {}


static func invalidate() -> void:
	_loaded = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_visuals = {}
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_visuals.v0.json")
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("ItemVisualsLoader: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY and typeof(parsed.get("item_visuals", {})) == TYPE_DICTIONARY:
		_visuals = parsed.get("item_visuals", {})


static func all() -> Dictionary:
	ensure_loaded()
	return _visuals


static func visual_for(item_def_id: String) -> Dictionary:
	ensure_loaded()
	var entry = _visuals.get(item_def_id, {})
	return entry if typeof(entry) == TYPE_DICTIONARY else {}


## The kit model a hand item (main_hand/off_hand) is wielded with, or "" for other slots.
static func hand_asset_id(item_def_id: String) -> String:
	var entry := visual_for(item_def_id)
	if not (str(entry.get("slot", "")) in HAND_SLOTS):
		return ""
	return str(entry.get("asset_id", ""))


static func is_rig_native(item_def_id: String) -> bool:
	return bool(visual_for(item_def_id).get("rig_native", false))
