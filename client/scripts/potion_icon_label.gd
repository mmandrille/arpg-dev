class_name PotionIconLabel
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")


static func is_leveled_potion(item_def_id: String) -> bool:
	match item_def_id:
		"red_potion", "blue_potion", "rejuv_potion":
			return true
		_:
			return false


static func icon_label(item: Dictionary, fallback_label: String) -> String:
	var def_id := str(item.get("item_def_id", ""))
	if not is_leveled_potion(def_id):
		return fallback_label
	var level := BlacksmithUpgradePreviewScript.item_level(item)
	if level < 1:
		level = 1

	return str(level)
