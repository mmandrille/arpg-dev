class_name BlacksmithShardInventory
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const BlacksmithRecipesScript := preload("res://scripts/blacksmith_recipes.gd")


static func required_resource_level(recipe_id: String, item: Dictionary) -> int:
	return BlacksmithRecipesScript.required_resource_level(recipe_id, item)


static func required_shard_level(item: Dictionary) -> int:
	return maxi(1, BlacksmithUpgradePreviewScript.item_level(item))


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


static func leveled_consumable_bag_items(items: Array, item_def_id: String = "") -> Array:
	var out: Array = []
	for value in items:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var row := value as Dictionary
		var def_id := str(row.get("item_def_id", ""))
		if def_id != "upgrade_shard" and def_id != "renew_stone":
			continue
		if item_def_id != "" and def_id != item_def_id:
			continue
		if str(row.get("item_instance_id", "")) == "":
			continue
		out.append(row)
	return out


static func shard_stash_items(items: Array) -> Array:
	return leveled_consumable_bag_items(items)


static func resource_display_name(resource_item_def_id: String) -> String:
	if resource_item_def_id == "":
		return "resource"
	var def := ItemRulesLoader.item_definition(resource_item_def_id)
	return str(def.get("name", resource_item_def_id.replace("_", " ").capitalize()))
