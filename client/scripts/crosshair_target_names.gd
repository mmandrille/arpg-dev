## CrosshairTargetNames — display labels for crosshair-locked entities.
class_name CrosshairTargetNames
extends RefCounted

const InteractableRulesLoaderScript := preload("res://scripts/interactable_rules_loader.gd")


static func display_name(rec: Dictionary) -> String:
	var typ := str(rec.get("type", ""))
	match typ:
		"interactable":
			return _interactable_name(rec)
		"loot":
			return _loot_name(rec)
		"monster":
			return _monster_name(rec)
		_:
			return ""


static func _interactable_name(rec: Dictionary) -> String:
	var def_id := str(rec.get("interactable_def_id", ""))
	var def := InteractableRulesLoaderScript.interactable_definition(def_id)
	var name := str(def.get("name", "")).strip_edges()
	if name != "":
		return name
	return _titleize(def_id)


static func _loot_name(rec: Dictionary) -> String:
	var name := str(rec.get("display_name", "")).strip_edges()
	if name != "":
		return name
	var item_def_id := str(rec.get("item_def_id", ""))
	if item_def_id != "":
		return _titleize(item_def_id)
	return "Loot"


static func _monster_name(rec: Dictionary) -> String:
	if int(rec.get("hp", 1)) <= 0:
		return _corpse_name(rec)
	var name := str(rec.get("display_name", "")).strip_edges()
	if name != "":
		return name
	var monster_def_id := str(rec.get("monster_def_id", ""))
	if monster_def_id != "":
		return _titleize(monster_def_id)
	return "Monster"


static func _corpse_name(rec: Dictionary) -> String:
	var monster_def_id := str(rec.get("monster_def_id", ""))
	if monster_def_id == "":
		return "Corpse"
	return "%s Corpse" % _titleize(monster_def_id)


static func _titleize(raw: String) -> String:
	var text := raw.strip_edges()
	if text == "":
		return ""
	return text.replace("_", " ").capitalize()
