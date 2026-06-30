class_name BlacksmithShardInventory
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")


static func required_shard_level(item: Dictionary) -> int:
	return BlacksmithUpgradePreviewScript.item_level(item) + 1


static func resource_inventory_count(items: Array, resource_item_def_id: String, min_level: int = -1) -> int:
	if resource_item_def_id == "":
		return 0
	var count := 0
	for value in items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var row := value as Dictionary
		if str(row.get("item_def_id", "")) != resource_item_def_id:
			continue
		var shard_level := BlacksmithUpgradePreviewScript.shard_level(row)
		if min_level >= 0 and shard_level < min_level:
			continue
		count += 1
	return count


static func shard_stash_items(items: Array) -> Array:
	var out: Array = []
	for value in items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var row := value as Dictionary
		if str(row.get("item_def_id", "")) != "upgrade_shard":
			continue
		if str(row.get("stash_item_id", "")) == "":
			continue
		out.append(row)
	return out


static func resource_display_name(resource_item_def_id: String) -> String:
	if resource_item_def_id == "":
		return "resource"
	var def := ItemRulesLoader.item_definition(resource_item_def_id)
	return str(def.get("name", resource_item_def_id.replace("_", " ").capitalize()))
