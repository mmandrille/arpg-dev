class_name PotionIconLabel
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")


static func is_leveled_potion(item_def_id: String) -> bool:
	ItemRulesLoaderScript.ensure_loaded()
	var def := ItemRulesLoaderScript.item_definition(item_def_id)
	return bool(def.get("leveled_consumable", false))


static func icon_label(item: Dictionary, fallback_label: String) -> String:
	var def_id := str(item.get("item_def_id", ""))
	if not is_leveled_potion(def_id):
		return fallback_label
	var level := _potion_level(item)
	if level < 1:
		level = 1

	return str(level)


static func _potion_level(item: Dictionary) -> int:
	var top_level := int(item.get("item_level", 0))
	if top_level >= 1:
		return top_level

	return BlacksmithUpgradePreviewScript.item_level(item)
